// Package handlers - выполняет обработку HTTP запросов.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/workers"
)

//go:generate mockery --name=RepositoryInterface -case camel -inpkg

// RepositoryInterface - интерфейс для взаимодействия с репозиторием.
type RepositoryInterface interface {
	AddURL(ctx context.Context, longURL string, shortURL string, user string) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetUserURL(ctx context.Context, user string) ([]ResponseGetURL, error)
	AddManyURL(ctx context.Context, urls []ManyPostURL, user string) ([]ManyPostResponse, error)
	DeleteManyURL(ctx context.Context, urls []string, user string) error
	GetStats(ctx context.Context) (StatResponse, error)
	Ping(ctx context.Context) error
}

// PostURL - структура запроса на создание URL.
type PostURL struct {
	URL string
}

// ManyPostURL - структура запроса на создание нескольких URL.
type ManyPostURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ManyPostResponse - структура ответа создания множества URL.
type ManyPostResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ResponseGetURL - структура ответа на запрос о записанных URL.
type ResponseGetURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type StatResponse struct {
	CountURL  int `json:"urls"`
	CountUser int `json:"users"`
}

// ErrorWithDB - тип ошибки от базы данных.
type ErrorWithDB struct {
	Err   error
	Title string
}

func (err *ErrorWithDB) Error() string {
	return fmt.Sprintf("%v", err.Err)
}

func (err *ErrorWithDB) Unwrap() error {
	return err.Err
}

func NewErrorWithDB(err error, title string) error {
	return &ErrorWithDB{
		Err:   err,
		Title: title,
	}
}

// Handler - структура обработчика запросов.
type Handler struct {
	repo          RepositoryInterface
	baseURL       string
	wp            workers.WorkerPool
	trustedSubnet string
}

// New - функция создания нового обработчика.
func New(repo RepositoryInterface, basURL string, wp *workers.WorkerPool,
	trustedSubnet string) *Handler {
	return &Handler{
		repo:          repo,
		baseURL:       basURL,
		wp:            *wp,
		trustedSubnet: trustedSubnet,
	}
}

// RetrieveShortURL - получение оригинальной ссылки по укороченному URL.
// Обязательный параметр URL - id.
// Если ссылка верная - код ответа 307 и заголовок "location" с искомой ссылкой.
// Если ссылка была удалена - код ответа 410.
// Если ссылка не найдена - код ответа 404.
func (h *Handler) RetrieveShortURL(c *gin.Context) {
	result := map[string]string{}
	long, err := h.repo.GetURL(c.Request.Context(), c.Param("id"))

	if err != nil {
		var ue *ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "deleted" {
			c.Status(http.StatusGone)
			return
		}
		result["detail"] = err.Error()
		c.IndentedJSON(http.StatusNotFound, result)
		return
	}

	c.Header("Location", long)
	c.String(http.StatusTemporaryRedirect, "")
}

// CreateShortURL - создание укороченной ссылки.
// Формат запроса - строка с URL (plain text).
// При успешном создании код ответа 201, а так же в ответе будет укороченная ссылка.
// В случае ошибки в формате запроса - код ответа 400.
// В случае ошибки при записи в базу данных - код ответа 500.
func (h *Handler) CreateShortURL(c *gin.Context) {
	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		h.handleError(c, err)
		return
	}
	longURL := string(body)
	shortURL := shortener.ShorterURL(longURL)
	err = h.repo.AddURL(c.Request.Context(), longURL, shortURL, c.GetString("userId"))
	if err != nil {
		var ue *ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "UniqConstraint" {
			c.String(http.StatusConflict, h.baseURL+shortURL)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, h.baseURL+shortURL)
}

// ShortenURL - создание укороченной ссылки.
// Формат запроса PostURL.
// При успешном создании код ответа 201, а так же в ответе будет укороченная ссылка
// в result.
// В случае ошибки в формате запроса - код ответа 400.
// В случае, если такая ссылка уже имеется - код ответа 409.
// В случае ошибки при записи в базу данных - код ответа 500.
func (h *Handler) ShortenURL(c *gin.Context) {

	result := map[string]string{}
	var url PostURL

	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		h.handleError(c, err)
		return
	}
	err = json.Unmarshal(body, &url)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if url.URL == "" {
		h.handleError(c, errors.New("bad request"))
		return
	}
	shortURL := shortener.ShorterURL(url.URL)
	err = h.repo.AddURL(c.Request.Context(), url.URL, shortURL, c.GetString("userId"))
	if err != nil {
		var ue *ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "UniqConstraint" {
			result["result"] = h.baseURL + shortURL
			c.IndentedJSON(http.StatusConflict, result)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	result["result"] = h.baseURL + shortURL
	c.IndentedJSON(http.StatusCreated, result)
}

// GetUserURL - получение списка URL пользователя.
// При успешном запросе - код ответа 200 и списко URL пользователя в
// формате ResponseGetURL.
// В случае ошибки получение ссылок из базы данных - код ответа 500.
// В случае отсутствия ссылок у пользователя - код ответа 204.
func (h *Handler) GetUserURL(c *gin.Context) {
	result, err := h.repo.GetUserURL(c.Request.Context(), c.GetString("userId"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	if len(result) == 0 {
		c.IndentedJSON(http.StatusNoContent, result)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

// PingDB - проверка соединения с базой данных.
// В случае нормального соединения - код ответа 200.
// В случае ошибки с базой данных - код ответа 500.
func (h *Handler) PingDB(c *gin.Context) {
	err := h.repo.Ping(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		return
	}
	c.String(http.StatusOK, "")
}

// CreateBatch - создание нескольких коротких URL сразу.
// Формат запроса json в виде списка объектов формата ManyPostURL.
// В случае успешного создания - код ответа 201, а так же списко созданных
// URL в формате ManyPostResponse.
// В случае ошибки в формате запроса - код ответа 400.
// В случае ошибки записи в базу данных - код ответа 400.
func (h *Handler) CreateBatch(c *gin.Context) {
	var data []ManyPostURL
	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		h.handleError(c, err)
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response, err := h.repo.AddManyURL(c.Request.Context(), data, c.GetString("userId"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	if response == nil {
		h.handleError(c, errors.New("bad request"))
		return
	}
	c.IndentedJSON(http.StatusCreated, response)
}

// DeleteBatch - удаление нескольких сокращенных URL.
// В запросе ожидается список коротких URL.
// В случае успешной обработки запроса - код ответа 202.
// В случае ошибки в запросе - код ответа 400.
// Ссылки удаляются не сразу, а выставляются в очередь на удаление в
// WorkerPool(wp).
func (h *Handler) DeleteBatch(c *gin.Context) {
	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		h.handleError(c, err)
		return
	}
	var data []string
	err = json.Unmarshal(body, &data)
	if err != nil {
		h.handleError(c, err)
		return
	}
	var sliceData [][]string
	for i := 10; i <= len(data); i += 10 {
		sliceData = append(sliceData, data[i-10:i])
	}
	rem := len(data) % 10
	if rem > 0 {
		sliceData = append(sliceData, data[len(data)-rem:])
	}
	for _, item := range sliceData {
		func(taskData []string) {
			h.wp.Push(func(ctx context.Context) error {
				err := h.repo.DeleteManyURL(ctx, taskData, c.GetString("userId"))
				return err
			})
		}(item)
	}

	c.Status(http.StatusAccepted)
}

func (h *Handler) GetStats(c *gin.Context) {
	if h.trustedSubnet == "" {
		c.Status(http.StatusForbidden)
		return
	}
	realAddress := net.ParseIP(c.GetHeader("X-Real-IP"))
	_, subnet, err := net.ParseCIDR(h.trustedSubnet)
	if err != nil {
		h.handleError(c, errors.New("bad request"))
		return
	}
	if !subnet.Contains(realAddress) {
		c.Status(http.StatusForbidden)
		return
	}
	response, err := h.repo.GetStats(c.Request.Context())
	if err != nil {
		h.handleError(c, errors.New("bad request"))
		return
	}
	c.IndentedJSON(http.StatusOK, response)
}

// handleError обработка типовых ошибок.
func (h *Handler) handleError(c *gin.Context, err error) {
	message := make(map[string]string)
	message["detail"] = err.Error()
	c.IndentedJSON(http.StatusBadRequest, message)
}
