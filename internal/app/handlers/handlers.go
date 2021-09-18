package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/models"
)

type PostURL struct {
	URL string
}

type Handler struct {
	repo    models.RepositoryInterface
	baseURL string
}

func New(repo models.RepositoryInterface, baseURL string) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: baseURL,
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
	short := h.repo.AddURL(string(body), c.GetString("userId"))
	c.String(http.StatusCreated, h.baseURL+short)
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

	short := h.repo.AddURL(url.URL, c.GetString("userId"))
	result["result"] = h.baseURL + short
	c.IndentedJSON(http.StatusCreated, result)

	// err := c.BindJSON(&url)
	// if err != nil || url.URL == "" {
	// 	result["detail"] = "Bad request"
	// 	c.IndentedJSON(http.StatusBadRequest, result)
	// 	return
	// }
	// short := h.repo.AddURL(url.URL)
	// result["result"] = "http://localhost:8080/" + short
	// c.IndentedJSON(http.StatusCreated, result)
}

func (h *Handler) GetUserURL(c *gin.Context) {
	result := h.repo.GetUserURL(c.GetString("userId"))
	if len(result) == 0 {
		c.IndentedJSON(http.StatusNoContent, result)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}
