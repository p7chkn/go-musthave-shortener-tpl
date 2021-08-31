package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/models"
)

type PostURL struct {
	Url string `json: "url"`
}

type Handler struct {
	repo models.RepositoryInterface
}

func New(repo models.RepositoryInterface) *Handler {
	return &Handler{
		repo: repo,
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
	short := h.repo.AddURL(string(body))
	c.String(http.StatusCreated, "http://localhost:8080/"+short)
}

func (h *Handler) ShortenURL(c *gin.Context) {
	result := map[string]string{}
	var url PostURL
	err := c.BindJSON(&url)
	if err != nil || url.Url == "" {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	short := h.repo.AddURL(url.Url)
	result["result"] = "http://localhost:8080/" + short
	c.IndentedJSON(http.StatusCreated, result)
}
