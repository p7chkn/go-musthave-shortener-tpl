package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
