package handlers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/models"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/models/mocks"
	"github.com/stretchr/testify/assert"
)

func setupRouter(repo models.RepositoryInterface) *gin.Engine {
	router := gin.Default()

	handler := New(repo)

	router.GET("/:id", handler.RetriveShortURL)
	router.POST("/", handler.CreateShortURL)

	router.HandleMethodNotAllowed = true

	return router
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

			repoMock := new(mocks.RepositoryInterface)
			repoMock.On("GetURL", tt.query).Return(tt.result, tt.err)

			router := setupRouter(repoMock)

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

			repoMock := new(mocks.RepositoryInterface)
			repoMock.On("AddURL", tt.body).Return(tt.result, nil)

			router := setupRouter(repoMock)

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
