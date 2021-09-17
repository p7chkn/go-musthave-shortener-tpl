package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/middlewares"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/models"
)

func setupRouter(repo models.RepositoryInterface, cfg *configuration.Config) *gin.Engine {
	router := gin.Default()

	handler := handlers.New(repo, cfg.BaseURL)

	router.Use(middlewares.GzipEncodeMiddleware())
	router.Use(middlewares.GzipDecodeMiddleware())
	router.Use(middlewares.CookiMiddleware(cfg))

	router.GET("/:id", handler.RetriveShortURL)
	router.POST("/", handler.CreateShortURL)
	router.POST("/api/shorten", handler.ShortenURL)
	router.GET("/user/urls", handler.GetUserURL)

	router.HandleMethodNotAllowed = true

	return router
}

func main() {

	cfg := configuration.New()

	handler := setupRouter(models.NewFileRepository(cfg.FilePath), cfg)

	server := &http.Server{
		Addr:    cfg.ServerAdress,
		Handler: handler,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		log.Fatal(server.ListenAndServe())
		cancel()
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		cancel()
	case <-ctx.Done():
	}
	server.Shutdown(context.Background())
}
