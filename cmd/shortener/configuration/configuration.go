package configuration

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
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

	return &cfg
}
