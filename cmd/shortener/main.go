package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/p7chkn/go-musthave-shortener-tpl/iternal/handlers"
)

func main() {
	http.HandleFunc("/url", handlers.HelloWorld)

	server := &http.Server{
		Addr: "localhost:8000",
	}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	server.Shutdown(context.Background())
}
