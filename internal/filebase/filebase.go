package filebase

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
)

func NewFileRepository(ctx context.Context, filePath string, baseURL string) handlers.RepositoryInterface {
	return handlers.RepositoryInterface(NewRepositoryMap(ctx, filePath, baseURL))
}

type RepositoryMap struct {
	values   map[string]string
	filePath string
	baseURL  string
	usersURL map[string][]string
}

func NewRepositoryMap(ctx context.Context, filePath string, baseURL string) *RepositoryMap {
	repo := RepositoryMap{
		values:   map[string]string{},
		filePath: filePath,
		baseURL:  baseURL,
		usersURL: map[string][]string{},
	}
	file, err := os.OpenFile(repo.filePath, os.O_RDONLY|os.O_CREATE, configuration.FilePerm)
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

func (repo *RepositoryMap) AddURL(ctx context.Context, longURL string, shortURL string, user string) error {
	repo.values[shortURL] = longURL
	repo.writeRow(longURL, shortURL, repo.filePath, user)
	repo.usersURL[user] = append(repo.usersURL[user], shortURL)
	return nil
}

func (repo *RepositoryMap) GetURL(ctx context.Context, shortURL string) (string, error) {
	resultURL, okey := repo.values[shortURL]
	if !okey {
		return "", errors.New("not found")
	}
	return resultURL, nil
}

func (repo *RepositoryMap) GetUserURL(ctx context.Context, user string) ([]handlers.ResponseGetURL, error) {
	var result []handlers.ResponseGetURL
	for _, url := range repo.usersURL[user] {
		temp := handlers.ResponseGetURL{
			ShortURL:    repo.baseURL + url,
			OriginalURL: repo.values[url],
		}
		result = append(result, temp)
	}

	return result, nil
}

func (repo *RepositoryMap) Ping(ctx context.Context) error {
	return nil
}

func (repo *RepositoryMap) AddManyURL(ctx context.Context, urls []handlers.ManyPostURL, user string) ([]handlers.ManyPostResponse, error) {
	return nil, nil
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
	file, err := os.OpenFile(repo.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, configuration.FilePerm)

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

func (repo *RepositoryMap) DeleteManyURL(ctx context.Context, urls []string, user string) error {
	return nil
}
