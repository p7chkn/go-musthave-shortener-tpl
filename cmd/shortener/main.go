package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/services"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/middlewares"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/database"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/filebase"
)

func setupRouter(repo handlers.RepositoryInterface, cfg *configuration.Config) *gin.Engine {
	router := gin.Default()

	handler := handlers.New(repo, cfg)

	router.Use(middlewares.GzipEncodeMiddleware())
	router.Use(middlewares.GzipDecodeMiddleware())
	router.Use(middlewares.CookiMiddleware(cfg))

	router.GET("/:id", handler.RetriveShortURL)
	router.POST("/", handler.CreateShortURL)
	router.POST("/api/shorten", handler.ShortenURL)
	router.GET("/user/urls", handler.GetUserURL)
	router.GET("/ping", handler.PingDB)
	router.POST("/api/shorten/batch", handler.CreateBatch)

	router.HandleMethodNotAllowed = true

	return router
}

func main() {

	cfg := configuration.New()

	var handler *gin.Engine

	if cfg.DataBase.DataBaseURI != "" {
		db, err := sql.Open("postgres", cfg.DataBase.DataBaseURI)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		services.SetUpDataBase(db)

		handler = setupRouter(database.NewDatabaseRepository(cfg.BaseURL, db), cfg)
	} else {
		handler = setupRouter(filebase.NewFileRepository(cfg), cfg)
	}

	server := &http.Server{
		Addr:    cfg.ServerAdress,
		Handler: handler,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		log.Println(server.ListenAndServe())
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
