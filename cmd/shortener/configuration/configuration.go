package configuration

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/caarlos0/env"
)

const (
	FileName     = "urls.log"
	FilePerm     = 0755
	ServerAdress = "localhost:8080"
	BaseURL      = "http://localhost:8080/"
)

var Configuration config

func init() {
	Configuration = *newConfig()
}

type config struct {
	ServerAdress string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
	FilePath     string `env:"FILE_STORAGE_PATH"`
}

func newConfig() *config {
	cfg := config{
		ServerAdress: ServerAdress,
		FilePath:     FileName,
		BaseURL:      BaseURL,
	}
	cfg.BaseURL = fmt.Sprintf("http://%s/", cfg.ServerAdress)

	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}
	flagServerAdress := flag.String("a", ServerAdress, "server adress")
	flagBaseURL := flag.String("b", BaseURL, "base url")
	flagFilePath := flag.String("c", FileName, "file path")
	flag.Parse()

	if *flagServerAdress != ServerAdress {
		cfg.ServerAdress = *flagServerAdress
	}
	if *flagBaseURL != BaseURL {
		cfg.BaseURL = *flagBaseURL
	}
	if *flagFilePath != FileName {
		cfg.FilePath = *flagFilePath
	}

	if cfg.FilePath != FileName {
		if _, err := os.Stat(filepath.Dir(cfg.FilePath)); os.IsNotExist(err) {
			log.Println("Creating folder")
			err := os.Mkdir(filepath.Dir(cfg.FilePath), FilePerm)
			if err != nil {
				log.Printf("Error: %v \n", err)
			}
		}
	}

	if string(cfg.BaseURL[len(cfg.BaseURL)-1]) != "/" {
		cfg.BaseURL += "/"
	}

	return &cfg
}
