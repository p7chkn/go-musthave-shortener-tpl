package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

//go:generate mockery --name=RepositoryInterface --structname=MockRepositoryInterface --inpackage
type RepositoryInterface interface {
	AddURL(ctx context.Context, longURL string, shortURL string, user string) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetUserURL(ctx context.Context, user string) ([]ResponseGetURL, error)
	AddManyURL(ctx context.Context, urls []ManyPostURL, user string) ([]ManyPostResponse, error)
	Ping(ctx context.Context) error
}

type PostURL struct {
	URL string
}

type ManyPostURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ManyPostResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ResponseGetURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UniqueConstraintError struct {
	Err error
}

func (ue *UniqueConstraintError) Error() string {
	return fmt.Sprintf("%v", ue.Err)
}

func (ue *UniqueConstraintError) Unwrap() error {
	return ue.Err
}

func NewUniqueConstraintError(err error) error {
	return &UniqueConstraintError{
		Err: err,
	}
}

type Handler struct {
	repo    RepositoryInterface
	baseURL string
}

func New(repo RepositoryInterface, basURL string) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: basURL,
	}
}

func (h *Handler) RetriveShortURL(c *gin.Context) {
	result := map[string]string{}
	long, err := h.repo.GetURL(c.Request.Context(), c.Param("id"))

	if err != nil {
		result["detail"] = err.Error()
		c.IndentedJSON(http.StatusNotFound, result)
		return
	}

	c.Header("Location", long)
	c.String(http.StatusTemporaryRedirect, "")
}

func (h *Handler) CreateShortURL(c *gin.Context) {
	result := map[string]string{}
	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	longURL := string(body)
	shortURL := shortener.ShorterURL(longURL)
	err = h.repo.AddURL(c.Request.Context(), longURL, shortURL, c.GetString("userId"))
	if err != nil {
		var ue *UniqueConstraintError
		if errors.As(err, &ue) {
			c.String(http.StatusConflict, h.baseURL+shortURL)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, h.baseURL+shortURL)
}

func (h *Handler) ShortenURL(c *gin.Context) {
	result := map[string]string{}
	var url PostURL

	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	json.Unmarshal(body, &url)
	if url.URL == "" {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	shortURL := shortener.ShorterURL(url.URL)
	err = h.repo.AddURL(c.Request.Context(), url.URL, shortURL, c.GetString("userId"))
	if err != nil {
		var ue *UniqueConstraintError
		if errors.As(err, &ue) {
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

func (h *Handler) PingDB(c *gin.Context) {
	err := h.repo.Ping(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		return
	}
	c.String(http.StatusOK, "")
}

func (h *Handler) CreateBatch(c *gin.Context) {
	var data []ManyPostURL

	c.BindJSON(&data)
	response, err := h.repo.AddManyURL(c.Request.Context(), data, c.GetString("userId"))
	if err != nil {
		message := make(map[string]string)
		message["detail"] = err.Error()
		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}
	if response == nil {
		message := make(map[string]string)
		message["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}
	c.IndentedJSON(http.StatusCreated, response)
}
