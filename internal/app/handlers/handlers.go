package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

//go:generate mockery --name=RepositoryInterface --structname=MockRepositoryInterface --inpackage
type RepositoryInterface interface {
	AddURL(longURL string, shortURL string, user string) error
	GetURL(shortURL string) (string, error)
	GetUserURL(user string) ([]ResponseGetURL, error)
	AddManyURL(urls []ManyPostURL, user string) ([]ManyPostResponse, error)
	Ping() error
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

type Handler struct {
	repo    RepositoryInterface
	baseURL string
}

func New(repo RepositoryInterface, cfg *configuration.Config) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: cfg.BaseURL,
	}
}

func (h *Handler) RetriveShortURL(c *gin.Context) {
	result := map[string]string{}
	long, err := h.repo.GetURL(c.Param("id"))

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
	err = h.repo.AddURL(longURL, shortURL, c.GetString("userId"))
	if err != nil {
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
	err = h.repo.AddURL(url.URL, shortURL, c.GetString("userId"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	result["result"] = h.baseURL + shortURL
	c.IndentedJSON(http.StatusCreated, result)
}

func (h *Handler) GetUserURL(c *gin.Context) {
	result, err := h.repo.GetUserURL(c.GetString("userId"))
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
	err := h.repo.Ping()
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		return
	}
	c.String(http.StatusOK, "")
}

func (h *Handler) CreateBatch(c *gin.Context) {
	var data []ManyPostURL

	c.BindJSON(&data)
	response, err := h.repo.AddManyURL(data, c.GetString("userId"))
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
