package main

import (
	"context"
	"crypto/tls"
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
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/database"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/filebase"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/workers"
)

var (
	httpServer   *http.Server
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {

	log.Printf("Build version: %v\n", buildVersion)
	log.Printf("Build date: %v\n", buildDate)
	log.Printf("Build commit: %v\n", buildCommit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	cfg := configuration.New()

	var handler *gin.Engine

	wp := workers.New(ctx, cfg.NumOfWorkers, cfg.WorkersBuffer)

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

		handler = services.SetupRouter(database.NewDatabaseRepository(cfg.BaseURL, db), cfg, wp)
	} else {
		handler = services.SetupRouter(filebase.NewFileRepository(ctx, cfg.FilePath, cfg.BaseURL), cfg, wp)
	}

	g, ctx := errgroup.WithContext(ctx)

	tlsS := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	g.Go(func() error {
		httpServer = &http.Server{
			Addr:      cfg.ServerAddress,
			Handler:   handler,
			TLSConfig: tlsS,
		}
		log.Printf("httpServer starting at: %v", cfg.ServerAddress)
		if cfg.EnableHTTPS {
			if err := httpServer.ListenAndServeTLS(
				"localhost.crt",
				"localhost.key"); err != http.ErrServerClosed {
				return err
			}
		} else {
			if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
				return err
			}
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
	}

}
