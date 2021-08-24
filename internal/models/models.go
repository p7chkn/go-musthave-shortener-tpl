package models

import (
	"errors"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
	"github.com/stretchr/testify/mock"
)

func SetupRepository() RepositoryInterface {
	var respoInterface RepositoryInterface

	dataMap := new(RepositoryMap)
	dataMap.SetupValues()
	respoInterface = dataMap
	return respoInterface
}

type RepositoryMap struct {
	values map[string]string
}

func (repo *RepositoryMap) SetupValues() {
	repo.values = make(map[string]string)
}

func (repo *RepositoryMap) AddUrl(longURL string) string {
	shortURL := shortener.ShorterURL(longURL)
	repo.values[shortURL] = longURL
	return shortURL
}

func (repo *RepositoryMap) GetUrl(shortURL string) (string, error) {
	resultUrl, okey := repo.values[shortURL]
	if !okey {
		return "", errors.New("not found")
	}
	return resultUrl, nil
}

type RepositoryInterface interface {
	AddUrl(longURL string) string
	GetUrl(shortURL string) (string, error)
}

type RepositoryMock struct {
	mock.Mock
}

func (m *RepositoryMock) AddUrl(longURL string) string {
	args := m.Called(longURL)
	return args.String(0)
}

func (m *RepositoryMock) GetUrl(shortURL string) (string, error) {
	args := m.Called(shortURL)
	return args.String(0), args.Error(1)
}
