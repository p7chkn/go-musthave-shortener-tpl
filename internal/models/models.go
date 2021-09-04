package models

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/files"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func NewRepository(filePath string) RepositoryInterface {
	return RepositoryInterface(NewRepositoryMap(filePath))
}

type RepositoryMap struct {
	values   map[string]string
	filePath string
}

func NewRepositoryMap(filePath string) *RepositoryMap {
	values := make(map[string]string)

	if filePath != files.FileName {
		if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
			fmt.Println("Creating folder")
			err := os.Mkdir(filepath.Dir(filePath), files.FilePerm)
			if err != nil {
				fmt.Printf("Error: %v \n", err)
			}
		}
	}
	reader := files.NewFileReader(filePath)

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
		values:   values,
		filePath: filePath,
	}
}

func (repo *RepositoryMap) AddURL(longURL string) string {
	shortURL := shortener.ShorterURL(longURL)
	data := files.Row{
		LongURL:  longURL,
		ShortURL: shortURL,
	}
	writer := files.NewFileWriter(repo.filePath)
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
