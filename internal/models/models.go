package models

import (
	"errors"
	"log"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/files"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func NewRepository() RepositoryInterface {
	return RepositoryInterface(NewRepositoryMap())
}

type RepositoryMap struct {
	values map[string]string
}

func NewRepositoryMap() *RepositoryMap {
	values := make(map[string]string)

	reader := files.NewFileReader()

	defer reader.Close()
	for {
		data, err := reader.ReadRow()

		if err != nil {
			log.Printf("Error while parsing file: %v\n", err)
		}

		if data == nil {
			break
		}
		values[data.ShortURL] = data.LongURL
	}

	return &RepositoryMap{
		values: values,
	}
}

func (repo *RepositoryMap) AddURL(longURL string) string {
	shortURL := shortener.ShorterURL(longURL)
	data := files.Row{
		LongURL:  longURL,
		ShortURL: shortURL,
	}
	writer := files.NewFileWriter()
	defer writer.Close()

	writer.WriteRow(&data)

	repo.values[shortURL] = longURL
	return shortURL
}

func (repo *RepositoryMap) GetURL(shortURL string) (string, error) {
	resultURL, okey := repo.values[shortURL]
	if !okey {
		return "", errors.New("not found")
	}
	return resultURL, nil
}

//go:generate mockery -name=RepositoryInterface
type RepositoryInterface interface {
	AddURL(longURL string) string
	GetURL(shortURL string) (string, error)
}
