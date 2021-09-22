package handlers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/middlewares"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(repo RepositoryInterface, baseURL string) (*gin.Engine, *configuration.Config) {
	router := gin.Default()
	key, _ := configuration.GenerateRandom(16)
	cfg := &configuration.Config{
		Key:     key,
		BaseURL: configuration.BaseURL,
	}
	handler := New(repo, cfg)
	router.Use(middlewares.CookiMiddleware(cfg))
	router.GET("/:id", handler.RetriveShortURL)
	router.POST("/", handler.CreateShortURL)
	router.POST("/api/shorten", handler.ShortenURL)
	router.GET("/user/urls", handler.GetUserURL)
	router.POST("/api/shorten/batch", handler.CreateBatch)
	router.HandleMethodNotAllowed = true
	return router, cfg
}
func TestRetriveShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		query  string
		err    error
		result string
		want   want
	}{
		{
			name:   "GET without id",
			query:  "",
			result: "",
			err:    errors.New("not found"),
			want: want{
				code:        405,
				response:    `405 method not allowed`,
				contentType: `text/plain`,
			},
		},
		{
			name:   "GET with correct id",
			query:  "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			result: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			err:    nil,
			want: want{
				code:        307,
				response:    ``,
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   "GET with incorrect id",
			query:  "12398fv58Wr3hGGIzm2-aH2zA628Ng=",
			result: "",
			err:    errors.New("not found"),
			want: want{
				code:        404,
				response:    `{"detail":"not found"}`,
				contentType: `application/json; charset=utf-8`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := new(MockRepositoryInterface)
			repoMock.On("GetURL", tt.query).Return(tt.result, tt.err)
			router, _ := setupRouter(repoMock, configuration.BaseURL)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/"+tt.query, nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, w.Header()["Content-Type"][0], tt.want.contentType)
			assert.Equal(t, tt.want.code, w.Code)
			resBody, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Fatal(err)
			}
			if w.Header()["Content-Type"][0] == `application/json; charset=utf-8` {
				assert.JSONEq(t, tt.want.response, string(resBody))
			} else {
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}
func TestCreateShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		query  string
		body   string
		result string
		want   want
	}{
		{
			name:   "correct POST",
			query:  "",
			body:   "http://iloverestaurant.ru/",
			result: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			want: want{
				code:        201,
				response:    `http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=`,
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   "incorrect POST",
			query:  "123",
			body:   "http://iloverestaurant.ru/",
			result: "",
			want: want{
				code:        405,
				response:    `405 method not allowed`,
				contentType: `text/plain`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := new(MockRepositoryInterface)
			repoMock.On("AddURL", tt.body, tt.result, mock.Anything).Return(nil)
			router, _ := setupRouter(repoMock, configuration.BaseURL)
			body := strings.NewReader(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/"+tt.query, body)
			router.ServeHTTP(w, req)
			assert.Equal(t, w.Header()["Content-Type"][0], tt.want.contentType)
			assert.Equal(t, tt.want.code, w.Code)
			resBody, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Fatal(err)
			}
			if w.Header()["Content-Type"][0] == `application/json; charset=utf-8` {
				assert.JSONEq(t, tt.want.response, string(resBody))
			} else {
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}
func TestShortenURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		query   string
		body    string
		rawData string
		result  string
		want    want
	}{
		{
			name:    "correct POST",
			query:   "api/shorten",
			body:    `{"url": "http://iloverestaurant.ru/"}`,
			rawData: "http://iloverestaurant.ru/",
			result:  "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			want: want{
				code:        201,
				response:    `{"result": "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng="}`,
				contentType: `application/json; charset=utf-8`,
			},
		},
		{
			name:    "incorrect POST",
			query:   "api/shorten",
			body:    `{"url2": "http://iloverestaurant.ru/"}`,
			rawData: "http://iloverestaurant.ru/",
			result:  "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			want: want{
				code:        400,
				response:    `{"detail": "Bad request"}`,
				contentType: `application/json; charset=utf-8`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := new(MockRepositoryInterface)
			repoMock.On("AddURL", tt.rawData, tt.result, mock.Anything).Return(nil)
			router, _ := setupRouter(repoMock, configuration.BaseURL)
			body := strings.NewReader(tt.body)
			w := httptest.NewRecorder()
			fmt.Println(tt.query)
			req, _ := http.NewRequest(http.MethodPost, "/"+tt.query, body)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.want.contentType, w.Header()["Content-Type"][0])
			assert.Equal(t, tt.want.code, w.Code)
			resBody, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Fatal(err)
			}
			if w.Header()["Content-Type"][0] == `application/json; charset=utf-8` {
				assert.JSONEq(t, tt.want.response, string(resBody))
			} else {
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}
func TestGetUserURL(t *testing.T) {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		name     string
		query    string
		response []ResponseGetURL
		want     want
	}{
		{
			name:  "correct GET",
			query: "user/urls",
			response: []ResponseGetURL{
				{
					ShortURL:    "http://localhost:8080/1yhVmSPGQlZn3EnrI2kd7Oxu5UM=",
					OriginalURL: "http://hbqouwjbx5jl.ru/lkm0skvkix1ejv",
				}},
			want: want{
				code:        200,
				contentType: `application/json; charset=utf-8`,
				response: `[{
					"short_url": "http://localhost:8080/1yhVmSPGQlZn3EnrI2kd7Oxu5UM=",
					"original_url": "http://hbqouwjbx5jl.ru/lkm0skvkix1ejv"
				}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, _ := uuid.NewV4()
			repoMock := new(MockRepositoryInterface)
			repoMock.On("GetUserURL", userID.String()).Return(tt.response, nil)
			router, cfg := setupRouter(repoMock, configuration.BaseURL)

			encoder, _ := utils.New(cfg.Key)

			cookie := http.Cookie{
				Name:  "userId",
				Value: encoder.EncodeUUIDtoString(userID.Bytes()),
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/"+tt.query, nil)
			req.AddCookie(&cookie)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.want.contentType, w.Header()["Content-Type"][0])
			assert.Equal(t, tt.want.code, w.Code)
			resBody, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.JSONEq(t, tt.want.response, string(resBody))

		})
	}
}

func TestCreateBatch(t *testing.T) {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		name     string
		query    string
		response string
		want     want
	}{
		{
			name:  "correct POST",
			query: "api/shorten/batch",
			want: want{
				code:        200,
				contentType: `application/json; charset=utf-8`,
				response:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
