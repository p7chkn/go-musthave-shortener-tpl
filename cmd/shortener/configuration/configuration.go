package configuration

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/files"
)

type Config struct {
	ServerAdress string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
	FilePath     string `env:"FILE_STORAGE_PATH"`
}

func New() *Config {
	cfg := Config{
		ServerAdress: "localhost:8080",
	}
	cfg.BaseURL = fmt.Sprintf("http://%s/", cfg.ServerAdress)

	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	if string(cfg.BaseURL[len(cfg.BaseURL)-1]) != "/" {
		cfg.BaseURL += "/"
	}

	if cfg.FilePath == "" {
		cfg.FilePath = files.FileName
	}

	return &cfg
}
