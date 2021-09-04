package models

import (
	"errors"
	"fmt"
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
	if filePath != files.FileName {
		path, _ := os.Getwd()
		fmt.Printf("--------------------------- %v\n", filepath.Dir(filePath))
		fmt.Printf("--------------------------- %v\n", path+filepath.Dir(filePath))
		if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
			fmt.Println("Creating folder")
			err := os.MkdirAll(filepath.Dir(filePath), files.FilePerm)
			if err != nil {
				fmt.Printf("Error: %v \n", err)
			}
		}
	}
	return &RepositoryMap{
		values:   make(map[string]string),
		filePath: filePath,
	}
}

func (repo *RepositoryMap) AddURL(longURL string) string {
	shortURL := shortener.ShorterURL(longURL)
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
