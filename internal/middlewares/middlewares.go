package middlewares

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
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
		cookie, _ := c.Request.Cookie("userId")
		verifyOrCreateCookie(cookie, c, cfg)
		c.Next()
	}
}

func verifyOrCreateCookie(cookie *http.Cookie, c *gin.Context, cfg *configuration.Config) {
	defer c.Next()
	h := hmac.New(sha256.New, cfg.Key)
	u, _ := uuid.NewV4()
	h.Write(u.Bytes())
	value := h.Sum(nil)

	if cookie == nil {
		fmt.Println("Set new cookie")
		c.SetCookie("userId", string(value), 864000, "/", cfg.BaseURL, false, false)
		c.Set("userId", string(value))
	} else {
		c.Set("userId", cookie.Value)
	}
}
