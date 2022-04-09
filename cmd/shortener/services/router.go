package services

import (
	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/middlewares"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/workers"
)

// SetupRouter - подготоваливает роутер для обработки запросов.
func SetupRouter(repo handlers.RepositoryInterface, cfg *configuration.Config, wp *workers.WorkerPool) *gin.Engine {
	router := gin.Default()

	handler := handlers.New(repo, cfg.BaseURL, wp, cfg.TrustedSubnet)

	router.Use(middlewares.GzipEncodeMiddleware())
	router.Use(middlewares.GzipDecodeMiddleware())
	router.Use(middlewares.CookiMiddleware(cfg))

	router.GET("/:id", handler.RetrieveShortURL)
	router.POST("/", handler.CreateShortURL)
	router.POST("/api/shorten", handler.ShortenURL)
	router.GET("/api/user/urls", handler.GetUserURL)
	router.GET("/ping", handler.PingDB)
	router.POST("/api/shorten/batch", handler.CreateBatch)
	router.DELETE("/api/user/urls", handler.DeleteBatch)
	router.GET("/api/internal/stats", handler.GetStats)

	router.HandleMethodNotAllowed = true

	return router
}
