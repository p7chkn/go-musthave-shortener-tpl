package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func NewFileRepository(cfg *configuration.Config) RepositoryInterface {
	return RepositoryInterface(NewRepositoryMap(cfg))
}

type RepositoryMap struct {
	values   map[string]string
	Cfg      *configuration.Config
	usersURL map[string][]string
	raw      []string
}

func NewRepositoryMap(cfg *configuration.Config) *RepositoryMap {
	repo := RepositoryMap{
		values:   map[string]string{},
		Cfg:      cfg,
		usersURL: map[string][]string{},
	}
	file, err := os.OpenFile(cfg.FilePath, os.O_RDONLY|os.O_CREATE, configuration.FilePerm)
	if err != nil {
		log.Printf("Error with reading file: %v\n", err)
	}
	defer file.Close()
	reader := bufio.NewScanner(file)

	for {
		ok, err := repo.readRow(reader)

		if err != nil {
			log.Printf("Error while parsing file: %v\n", err)
		}

		if !ok {
			break
		}
	}

	return &repo
}

func (repo *RepositoryMap) AddURL(longURL string, user string) string {
	shortURL := shortener.ShorterURL(longURL)
	repo.values[shortURL] = longURL
	repo.writeRow(longURL, shortURL, repo.Cfg.FilePath, user)
	repo.usersURL[user] = append(repo.usersURL[user], shortURL)
	return shortURL
}

func (repo *RepositoryMap) GetURL(shortURL string) (string, error) {
	resultURL, okey := repo.values[shortURL]
	if !okey {
		return "", errors.New("not found")
	}
	return resultURL, nil
}

func (repo *RepositoryMap) GetUserURL(user string) []ResponseGetURL {
	result := []ResponseGetURL{}
	// for _, url := range repo.usersURL[user] {
	// 	temp := ResponseGetURL{
	// 		ShortURL:    repo.Cfg.BaseURL + url,
	// 		OriginalURL: repo.values[url],
	// 	}
	// 	result = append(result, temp)
	// }
	s := ""
	for _, raw := range repo.raw {
		s += raw
	}
	result = append(result, ResponseGetURL{
		ShortURL:    fmt.Sprint(repo.usersURL),
		OriginalURL: fmt.Sprint(repo.values),
	})
	return result
}

type row struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	User     string `json:"user"`
}

func (repo *RepositoryMap) readRow(reader *bufio.Scanner) (bool, error) {

	if !reader.Scan() {
		return false, reader.Err()
	}
	data := reader.Bytes()

	repo.raw = append(repo.raw, string(data))

	row := &row{}

	err := json.Unmarshal(data, row)

	if err != nil {
		return false, err
	}
	repo.values[row.ShortURL] = row.LongURL
	repo.usersURL[row.User] = append(repo.usersURL[row.User], row.ShortURL)

	return true, nil
}

func (repo *RepositoryMap) writeRow(longURL string, shortURL string, filePath string, user string) error {
	file, err := os.OpenFile(repo.Cfg.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, configuration.FilePerm)

	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)

	data, err := json.Marshal(&row{
		LongURL:  longURL,
		ShortURL: shortURL,
		User:     user,
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

type ResponseGetURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//go:generate mockery -name=RepositoryInterface
type RepositoryInterface interface {
	AddURL(longURL string, user string) string
	GetURL(shortURL string) (string, error)
	GetUserURL(user string) []ResponseGetURL
}
