package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func NewFileRepository(filePath string) RepositoryInterface {
	return RepositoryInterface(NewRepositoryMap(filePath))
}

type RepositoryMap struct {
	values   map[string]string
	filePath string
}

func NewRepositoryMap(filePath string) *RepositoryMap {
	values := make(map[string]string)
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, configuration.FilePerm)
	if err != nil {
		log.Printf("Error with reading file: %v\n", err)
	}
	defer file.Close()
	reader := bufio.NewScanner(file)

	for {
		data, err := readRow(reader)

		if err != nil {
			log.Printf("Error while parsing file: %v\n", err)
		}

		if data == nil {
			break
		}
		values[data.ShortURL] = data.LongURL
	}

	return &RepositoryMap{
		values:   values,
		filePath: filePath,
	}
}

func (repo *RepositoryMap) AddURL(longURL string) string {
	shortURL := shortener.ShorterURL(longURL)
	repo.values[shortURL] = longURL
	writeRow(longURL, shortURL, repo.filePath)
	return shortURL
}

func (repo *RepositoryMap) GetURL(shortURL string) (string, error) {
	resultURL, okey := repo.values[shortURL]
	if !okey {
		return "", errors.New("not found")
	}
	return resultURL, nil
}

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func readRow(reader *bufio.Scanner) (*row, error) {

	if !reader.Scan() {
		return nil, reader.Err()
	}
	data := reader.Bytes()

	row := &row{}

	err := json.Unmarshal(data, row)

	if err != nil {
		return nil, err
	}

	return row, nil
}

func writeRow(longURL string, shortURL string, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, configuration.FilePerm)

	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)

	data, err := json.Marshal(&row{
		LongURL:  longURL,
		ShortURL: shortURL,
	})
	if err != nil {
		return err
	}

	if _, err := writer.Write(data); err != nil {
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		return err
	}

	return writer.Flush()
}

//go:generate mockery -name=RepositoryInterface
type RepositoryInterface interface {
	AddURL(longURL string) string
	GetURL(shortURL string) (string, error)
}
