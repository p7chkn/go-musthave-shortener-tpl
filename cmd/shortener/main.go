package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	_ "github.com/lib/pq"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/services"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/middlewares"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/database"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/filebase"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/workers"
)

var (
	httpServer *http.Server
)

func setupRouter(repo handlers.RepositoryInterface, cfg *configuration.Config, wp *workers.WorkerPool) *gin.Engine {
	router := gin.Default()

	handler := handlers.New(repo, cfg.BaseURL, wp)

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

	router.HandleMethodNotAllowed = true

	return router
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	cfg := configuration.New()

	var handler *gin.Engine

	wp := workers.New(ctx, cfg.NumOfWorkers, cfg.WorekersBuffer)

	go func() {
		wp.Run(ctx)
	}()

	if cfg.DataBase.DataBaseURI != "" {
		db, err := sql.Open("postgres", cfg.DataBase.DataBaseURI)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		err = services.SetUpDataBase(db, ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		handler = setupRouter(database.NewDatabaseRepository(cfg.BaseURL, db), cfg, wp)
	} else {
		handler = setupRouter(filebase.NewFileRepository(ctx, cfg.FilePath, cfg.BaseURL), cfg, wp)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		httpServer = &http.Server{
			Addr:    cfg.ServerAdress,
			Handler: handler,
		}
		log.Printf("httpServer starting at: %v", cfg.ServerAdress)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	log.Println("Receive shutdown signal")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer shutdownCancel()

	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}

	err := g.Wait()
	if err != nil {
		log.Printf("server returning an error: %v", err)
		os.Exit(2)
	}

}
