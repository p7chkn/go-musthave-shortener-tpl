package configuration

import (
	"crypto/rand"
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
	EnableHttps  = false
	// DataBaseURI  = "postgresql://postgres:1234@localhost:5432?sslmode=disable"
	DataBaseURI   = ""
	NumOfWorkers  = 10
	WorkersBuffer = 100
)

// Config - структура для кофигурации сервиса.
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	NumOfWorkers  int    `env:"NUMBER_OF_WORKERS"`
	EnableHttps   bool   `env:"ENABLE_HTTPS"`
	DataBase      ConfigDatabase
	Key           []byte
	WorkersBuffer int `env:"WORKERS_BUFFER"`
}

type ConfigDatabase struct {
	DataBaseURI string `env:"DATABASE_DSN"`
}

// New - создание новой конфигурации для сервиса, парсинг env и флагов в
// струтуру Config.
func New() *Config {
	dbCfg := ConfigDatabase{
		DataBaseURI: DataBaseURI,
	}

	flagServerAddress := flag.String("a", ServerAdress, "server adress")
	flagBaseURL := flag.String("b", BaseURL, "base url")
	flagFilePath := flag.String("f", FileName, "file path")
	flagDataBaseURI := flag.String("d", DataBaseURI, "URI for database")
	flagNumOfWorkers := flag.Int("w", NumOfWorkers, "Number of workers")
	flagBufferOfWorkers := flag.Int("wb", WorkersBuffer, "Workers channel buffer")
	flagEnableHttps := flag.Bool("s", EnableHttps, "Enable https")
	flag.Parse()

	if *flagDataBaseURI != DataBaseURI {
		dbCfg.DataBaseURI = *flagDataBaseURI
	}

	cfg := Config{
		ServerAddress: ServerAdress,
		FilePath:      FileName,
		BaseURL:       BaseURL,
		DataBase:      dbCfg,
		Key:           make([]byte, 16),
		NumOfWorkers:  NumOfWorkers,
		WorkersBuffer: WorkersBuffer,
		EnableHttps:   EnableHttps,
	}
	cfg.BaseURL = fmt.Sprintf("http://%s/", cfg.ServerAddress)

	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	if *flagServerAddress != ServerAdress {
		cfg.ServerAddress = *flagServerAddress
	}
	if *flagBaseURL != BaseURL {
		cfg.BaseURL = *flagBaseURL
	}
	if *flagFilePath != FileName {
		cfg.FilePath = *flagFilePath
	}
	if *flagNumOfWorkers != NumOfWorkers {
		cfg.NumOfWorkers = *flagNumOfWorkers
	}

	if *flagBufferOfWorkers != WorkersBuffer {
		cfg.WorkersBuffer = *flagBufferOfWorkers
	}

	if *flagEnableHttps != EnableHttps {
		cfg.EnableHttps = *flagEnableHttps
	}

	if cfg.FilePath != FileName {
		if _, err = os.Stat(filepath.Dir(cfg.FilePath)); os.IsNotExist(err) {
			log.Println("Creating folder")
			err = os.Mkdir(filepath.Dir(cfg.FilePath), FilePerm)
			if err != nil {
				log.Printf("Error: %v \n", err)
			}
		}
	}

	if string(cfg.BaseURL[len(cfg.BaseURL)-1]) != "/" {
		cfg.BaseURL += "/"
	}

	file, err := os.Open("key")

	if err != nil {
		cfg.Key, _ = GenerateRandom(16)
		file, _ = os.Create("key")
		file.Write(cfg.Key)
	} else {
		file.Read(cfg.Key)
	}

	return &cfg
}

// GenerateRandom - генерация случайной последоватльности байтов.
func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
