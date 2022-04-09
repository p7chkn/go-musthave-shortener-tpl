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
	EnableHTTPS  = false
	//DataBaseURI  = "postgresql://pavelchuykin:1234@localhost:5432?sslmode=disable"
	DataBaseURI   = ""
	NumOfWorkers  = 10
	WorkersBuffer = 100
	TrustedSubnet = "127.0.0.1/24"
)

// Config - структура для кофигурации сервиса.
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	NumOfWorkers  int    `env:"NUMBER_OF_WORKERS"`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`
	DataBase      ConfigDatabase
	Key           []byte
	WorkersBuffer int    `env:"WORKERS_BUFFER"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
}

type ConfigDatabase struct {
	DataBaseURI string `env:"DATABASE_DSN"`
}

// New - создание новой конфигурации для сервиса, парсинг env и флагов в
// струтуру Config.
func New() *Config {

	flagServerAddress := flag.String("a", ServerAdress, "server adress")
	flagBaseURL := flag.String("b", BaseURL, "base url")
	flagFilePath := flag.String("f", FileName, "file path")
	flagDataBaseURI := flag.String("d", DataBaseURI, "URI for database")
	flagNumOfWorkers := flag.Int("w", NumOfWorkers, "Number of workers")
	flagBufferOfWorkers := flag.Int("wb", WorkersBuffer, "Workers channel buffer")
	flagEnableHTTPS := flag.Bool("s", EnableHTTPS, "Enable https")
	flagConfigFile := flag.String("c", "", "configuration file")
	flagTrustedSubnet := flag.String("t", TrustedSubnet, "trusted subnet")
	flag.Parse()

	cfg := Config{}

	if *flagConfigFile != "" {
		cfg = getConfigFromFIle(*flagConfigFile)
	} else {
		cfg.ServerAddress = ServerAdress
		cfg.FilePath = FileName
		cfg.BaseURL = BaseURL
		cfg.DataBase.DataBaseURI = DataBaseURI
		cfg.Key = make([]byte, 16)
		cfg.NumOfWorkers = NumOfWorkers
		cfg.WorkersBuffer = WorkersBuffer
		cfg.EnableHTTPS = EnableHTTPS
		cfg.TrustedSubnet = TrustedSubnet
	}

	cfg.BaseURL = fmt.Sprintf("http://%s/", cfg.ServerAddress)

	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	if *flagDataBaseURI != DataBaseURI {
		cfg.DataBase.DataBaseURI = *flagDataBaseURI
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

	if *flagEnableHTTPS {
		cfg.EnableHTTPS = *flagEnableHTTPS
	}

	if *flagTrustedSubnet != TrustedSubnet {
		cfg.TrustedSubnet = *flagTrustedSubnet
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
