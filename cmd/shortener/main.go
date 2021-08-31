package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/models"
)

func setupRouter(repo models.RepositoryInterface) *gin.Engine {
	router := gin.Default()

	handler := handlers.New(repo)

	router.GET("/:id", handler.RetriveShortURL)
	router.POST("/", handler.CreateShortURL)
	router.POST("/api/shorten", handler.ShortenURL)

	router.HandleMethodNotAllowed = true

	return router
}

func main() {
	handler := setupRouter(models.SetupRepository())

	server := &http.Server{
		Addr:    "localhost:8080",
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
