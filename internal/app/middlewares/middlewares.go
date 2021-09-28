package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/utils"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func GzipEncodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			defer gz.Close()
			c.Header("Vary", "Accept-Encoding")
			c.Header("Content-Encoding", "gzip")
			c.Writer = &gzipWriter{c.Writer, gz}
		}
		c.Next()
	}
}

func GzipDecodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") {
			return
		}

		r, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Request.Body = r

		c.Next()

	}
}

func CookiMiddleware(cfg *configuration.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		cookie, _ := c.Request.Cookie("userId")
		encryptor, err := utils.New(cfg.Key)
		if err != nil {
			return
		}
		if cookie != nil {
			value, err := encryptor.DecodeUUIDfromString(cookie.Value)
			if err == nil {
				c.Set("userId", value)
				return
			}
		}
		id, err := uuid.NewV4()
		if err != nil {
			return
		}
		value := encryptor.EncodeUUIDtoString(id.Bytes())
		c.SetCookie("userId", value, 864000, "/", cfg.BaseURL, false, false)
		c.Set("userId", id.String())
	}
}
