package configuration

import (
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
		cfg.BaseURL = "http://localhost:8080/"
	}

	return &cfg
}
