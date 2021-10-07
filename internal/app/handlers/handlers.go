package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

//go:generate mockery --name=RepositoryInterface -case camel -inpkg

type RepositoryInterface interface {
	AddURL(longURL string, shortURL string, user string) error
	GetURL(shortURL string) (string, error)
	GetUserURL(user string) ([]ResponseGetURL, error)
	AddManyURL(urls []ManyPostURL, user string) ([]ManyPostResponse, error)
	DeleteManyURL(urls []string, user string) error
	IsOwner(url string, user string) bool
	Ping() error
}

type PostURL struct {
	URL string
}

type ManyPostURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ManyPostResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ResponseGetURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ErrorWithDB struct {
	Err   error
	Title string
}

func (err *ErrorWithDB) Error() string {
	return fmt.Sprintf("%v", err.Err)
}

func (err *ErrorWithDB) Unwrap() error {
	return err.Err
}

func NewErrorWithDB(err error, title string) error {
	return &ErrorWithDB{
		Err:   err,
		Title: title,
	}
}

type Handler struct {
	repo    RepositoryInterface
	baseURL string
}

func New(repo RepositoryInterface, basURL string) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: basURL,
	}
}

func (h *Handler) RetriveShortURL(c *gin.Context) {
	result := map[string]string{}
	long, err := h.repo.GetURL(c.Param("id"))

	if err != nil {
		var ue *ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "Deleted" {
			c.Status(http.StatusGone)
			return
		}
		result["detail"] = err.Error()
		c.IndentedJSON(http.StatusNotFound, result)
		return
	}

	c.Header("Location", long)
	c.String(http.StatusTemporaryRedirect, "")
}

func (h *Handler) CreateShortURL(c *gin.Context) {
	result := map[string]string{}
	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	longURL := string(body)
	shortURL := shortener.ShorterURL(longURL)
	err = h.repo.AddURL(longURL, shortURL, c.GetString("userId"))
	if err != nil {
		var ue *ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "UniqConstraint" {
			c.String(http.StatusConflict, h.baseURL+shortURL)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, h.baseURL+shortURL)
}

func (h *Handler) ShortenURL(c *gin.Context) {
	result := map[string]string{}
	var url PostURL

	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	json.Unmarshal(body, &url)
	if url.URL == "" {
		result["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, result)
		return
	}
	shortURL := shortener.ShorterURL(url.URL)
	err = h.repo.AddURL(url.URL, shortURL, c.GetString("userId"))
	if err != nil {
		var ue *ErrorWithDB
		if errors.As(err, &ue) && ue.Title == "UniqConstraint" {
			result["result"] = h.baseURL + shortURL
			c.IndentedJSON(http.StatusConflict, result)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	result["result"] = h.baseURL + shortURL
	c.IndentedJSON(http.StatusCreated, result)
}

func (h *Handler) GetUserURL(c *gin.Context) {
	result, err := h.repo.GetUserURL(c.GetString("userId"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	if len(result) == 0 {
		c.IndentedJSON(http.StatusNoContent, result)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

func (h *Handler) PingDB(c *gin.Context) {
	err := h.repo.Ping()
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		return
	}
	c.String(http.StatusOK, "")
}

func (h *Handler) CreateBatch(c *gin.Context) {
	var data []ManyPostURL

	c.BindJSON(&data)
	response, err := h.repo.AddManyURL(data, c.GetString("userId"))
	if err != nil {
		message := make(map[string]string)
		message["detail"] = err.Error()
		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}
	if response == nil {
		message := make(map[string]string)
		message["detail"] = "Bad request"
		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}
	c.IndentedJSON(http.StatusCreated, response)
}

func (h *Handler) DeleteBatch(c *gin.Context) {
	defer c.Request.Body.Close()

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		message := make(map[string]string)
		message["detail"] = err.Error()
		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}
	var data []string
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		message := make(map[string]string)
		message["detail"] = err.Error()
		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}
	go func() {
		wp := NewWorkerPool(10, h.repo)
		wp.DeleteURL(data, c.GetString("userId"))
		h.repo.DeleteManyURL(data, c.GetString("userId"))
	}()

	c.Status(http.StatusAccepted)
}

type WorkerPool struct {
	NumOfWorkers int
	inputCh      chan []string
	workers      []chan string
	pool         []chan []string
	repo         RepositoryInterface
}

func NewWorkerPool(numOfWorkers int, repo RepositoryInterface) *WorkerPool {
	return &WorkerPool{
		NumOfWorkers: numOfWorkers,
		workers:      make([]chan string, 0, numOfWorkers),
		inputCh:      make(chan []string),
		repo:         repo,
	}
}

func (wp *WorkerPool) fanOut() {
	cs := make([]chan []string, 0, wp.NumOfWorkers)
	for i := 0; i < wp.NumOfWorkers; i++ {
		cs = append(cs, make(chan []string))
	}
	go func() {
		defer func(cs []chan []string) {
			for _, c := range cs {
				close(c)
			}
		}(cs)

		for i := 0; i < len(cs); i++ {
			if i == len(cs)-1 {
				i = 0
			}

			url, ok := <-wp.inputCh
			if !ok {
				return
			}

			cs[i] <- url
		}
	}()
	wp.pool = cs
}

func (wp *WorkerPool) fanIn() chan string {
	out := make(chan string)

	go func() {
		wg := &sync.WaitGroup{}

		for _, ch := range wp.workers {
			wg.Add(1)

			go func(items chan string) {
				defer wg.Done()
				for item := range items {

					out <- item

				}
			}(ch)
		}
		wg.Wait()
		close(out)
	}()
	return out
}

func (wp *WorkerPool) NewWorker(input <-chan []string) chan string {
	out := make(chan string)

	go func() {
		for item := range input {
			isOwner := wp.repo.IsOwner(item[0], item[1])
			if isOwner {
				out <- item[0]
			}
		}

		close(out)
	}()
	return out
}

func (wp *WorkerPool) DeleteURL(urls []string, user string) {
	go func() {
		for _, url := range urls {
			wp.inputCh <- []string{url, user}
		}
		close(wp.inputCh)
	}()
	urlsToDelete := []string{"1"}
	// urlsToDelete = append(urlsToDelete, "s")
	wp.fanOut()

	for _, ch := range wp.pool {
		wp.workers = append(wp.workers, wp.NewWorker(ch))
	}
	for url := range wp.fanIn() {
		urlsToDelete = append(urlsToDelete, url)
	}

	// wp.repo.DeleteManyURL(urlsToDelete, user)
}
