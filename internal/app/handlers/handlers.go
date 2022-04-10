// Package handlers - выполняет обработку HTTP запросов.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/responses"
	custom_errors "github.com/p7chkn/go-musthave-shortener-tpl/internal/errors"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:generate mockery --name=UserUseCaseInterface --case camel --inpackage

// RepositoryInterface - интерфейс для взаимодействия с репозиторием.
type UserUseCaseInterface interface {
	GetURL(ctx context.Context, userID string) (string, error)
	CreateURL(ctx context.Context, longURL string, user string) (string, error)
	GetUserURL(ctx context.Context, userId string) ([]responses.GetURL, error)
	PingDB(ctx context.Context) error
	CreateBatch(ctx context.Context, urls []responses.ManyPostURL, userId string) ([]responses.ManyPostResponse, error)
	DeleteBatch(urls []string, userId string)
	GetStats(ctx context.Context, ip net.IP) (bool, responses.StatResponse, error)
}

// Handler - структура обработчика запросов.
type Handler struct {
	useCase UserUseCaseInterface
}

// New - функция создания нового обработчика.
func New(service UserUseCaseInterface) *Handler {
	return &Handler{
		useCase: service,
	}
}

// RetrieveShortURL - получение оригинальной ссылки по укороченному URL.
// Обязательный параметр URL - id.
// Если ссылка верная - код ответа 307 и заголовок "location" с искомой ссылкой.
// Если ссылка была удалена - код ответа 410.
// Если ссылка не найдена - код ответа 404.
func (h *Handler) RetrieveShortURL(c *gin.Context) {
	result := map[string]string{}
	long, err := h.useCase.GetURL(c.Request.Context(), c.Param("id"))

	if err != nil {
		var ue *custom_errors.ErrorWithDB
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
	responseURL, err := h.useCase.CreateURL(c.Request.Context(), string(body), c.GetString("userId"))
	if err != nil {
		var ue *custom_errors.ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "UniqConstraint" {
			c.String(http.StatusConflict, responseURL)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, responseURL)
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
	var url responses.PostURL

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
	responseURL, err := h.useCase.CreateURL(c.Request.Context(), url.URL, c.GetString("userId"))
	if err != nil {
		var ue *custom_errors.ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "UniqConstraint" {
			result["result"] = responseURL
			c.IndentedJSON(http.StatusConflict, result)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	result["result"] = responseURL
	c.IndentedJSON(http.StatusCreated, result)
}

// GetUserURL - получение списка URL пользователя.
// При успешном запросе - код ответа 200 и списко URL пользователя в
// формате GetURL.
// В случае ошибки получение ссылок из базы данных - код ответа 500.
// В случае отсутствия ссылок у пользователя - код ответа 204.
func (h *Handler) GetUserURL(c *gin.Context) {
	result, err := h.useCase.GetUserURL(c.Request.Context(), c.GetString("userId"))
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
	err := h.useCase.PingDB(c.Request.Context())
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
	var data []responses.ManyPostURL
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
	response, err := h.useCase.CreateBatch(c.Request.Context(), data, c.GetString("userId"))
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
	h.useCase.DeleteBatch(data, c.GetString("userId"))

	c.Status(http.StatusAccepted)
}

func (h *Handler) GetStats(c *gin.Context) {
	hasPermission, response, err := h.useCase.GetStats(c.Request.Context(), net.ParseIP(c.GetHeader("X-Real-IP")))
	if !hasPermission {
		c.Status(http.StatusForbidden)
		return
	}
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
