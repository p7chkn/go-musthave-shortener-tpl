package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/models"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

type PostURL struct {
	URL string
}

type Handler struct {
	repo    models.RepositoryInterface
	baseURL string
}

func New(repo models.RepositoryInterface, cfg *configuration.Config) *Handler {
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
	h.repo.AddURL(longURL, shortURL, c.GetString("userId"))
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
	h.repo.AddURL(url.URL, shortURL, c.GetString("userId"))
	result["result"] = h.baseURL + shortURL
	c.IndentedJSON(http.StatusCreated, result)
}

func (h *Handler) GetUserURL(c *gin.Context) {
	result := h.repo.GetUserURL(c.GetString("userId"))
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
