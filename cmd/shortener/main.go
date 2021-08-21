package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/handlers"
)

func main() {
	data := url.Values{}

	http.HandleFunc("/", handlers.URLHandler(data))

	server := &http.Server{
		Addr: "localhost:8080",
	}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	server.Shutdown(context.Background())
}
