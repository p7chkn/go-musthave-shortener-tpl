package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type ConfigFile struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

func getConfigFromFIle(fileName string) Config {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal()
	}
	cfg := ConfigFile{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		ServerAddress: cfg.ServerAddress,
		BaseURL:       cfg.BaseURL,
		FilePath:      cfg.FileStoragePath,
		EnableHTTPS:   cfg.EnableHTTPS,
		TrustedSubnet: cfg.TrustedSubnet,
		DataBase: ConfigDatabase{
			DataBaseURI: cfg.DatabaseDSN,
		},
	}
}
