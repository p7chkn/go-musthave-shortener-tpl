package handlers

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func RetriveShortURL(data url.Values) func(c *gin.Context) {
	return func(c *gin.Context) {
		result := map[string]string{}
		long, err := shortener.GetURL(c.Param("id"), data)

		if err != nil {
			result["detail"] = err.Error()
			c.IndentedJSON(http.StatusNotFound, result)
			return
		}

		c.Header("Location", long)
		c.String(http.StatusTemporaryRedirect, "")
	}
}

func CreateShortURL(data url.Values) func(c *gin.Context) {
	return func(c *gin.Context) {
		result := map[string]string{}
		defer c.Request.Body.Close()

		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			result["detail"] = "Bad request"
			c.IndentedJSON(http.StatusBadRequest, result)
			return
		}
		short := shortener.AddURL(string(body), data)
		c.String(http.StatusCreated, "http://localhost:8080/"+short)
	}
}
