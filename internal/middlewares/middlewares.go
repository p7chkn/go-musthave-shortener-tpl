package middlewares

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func GzipEncodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, _ := zap.NewDevelopment()

		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			gz.Reset(c.Writer)
			if err == nil {
				c.Header("Content-Encoding", "gzip")
				c.Header("Vary", "Accept-Encoding")
				c.Writer = &gzipWriter{c.Writer, gz}
			} else {
				logger.Error("Problem with creating gzipWriter")
			}
		}
	}
}

func GzipDecodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") {
			return
		}
		logger, _ := zap.NewDevelopment()

		defer c.Request.Body.Close()

		body, err := io.ReadAll(c.Request.Body)

		if err != nil {
			logger.Error(err.Error())
			return
		}

		reader, err := gzip.NewReader(bytes.NewReader(body))

		if err != nil {
			logger.Error(err.Error())
			return
		}
		fmt.Println("--------------------")
		// fmt.Println(reader)

		c.Request.Body = io.NopCloser(reader)

	}
}
