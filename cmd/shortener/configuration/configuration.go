package configuration

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAdress string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
}

func New() *Config {
	var cfg Config

	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	if cfg.ServerAdress == "" {
		cfg.ServerAdress = "localhost:8080"
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://%s/", cfg.ServerAdress)
	}
	if string(cfg.BaseURL[len(cfg.BaseURL)-1]) != "/" {
		cfg.BaseURL += "/"
	}

	return &cfg
}
