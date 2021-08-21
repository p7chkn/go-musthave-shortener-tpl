package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRuter(data url.Values) *gin.Engine {
	r := gin.Default()
	r.GET("/:id", RetriveShortURL(data))
	r.POST("/", CreateShortURL(data))
	return r
}

func TestRetriveShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name     string
		query    string
		longURL  string
		shortURL string
		want     want
	}{
		{
			name:     "GET without id",
			query:    "",
			longURL:  "",
			shortURL: "",
			want: want{
				code:        404,
				response:    `404 page not found`,
				contentType: `text/plain`,
			},
		},
		{
			name:     "GET with correct id",
			query:    "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			longURL:  "http://iloverestaurant.ru/",
			shortURL: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			want: want{
				code:        307,
				response:    ``,
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:     "GET with incorrect id",
			query:    "12398fv58Wr3hGGIzm2-aH2zA628Ng=",
			longURL:  "http://iloverestaurant.ru/",
			shortURL: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			want: want{
				code:        404,
				response:    `{"detail":"not found"}`,
				contentType: `application/json; charset=utf-8`,
			},
		},
	}
	data := url.Values{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data.Set(tt.shortURL, tt.longURL)

			router := setupRuter(data)

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
		name  string
		query string
		body  string
		want  want
	}{
		{
			name:  "correct POST",
			query: "",
			body:  "http://iloverestaurant.ru/",
			want: want{
				code:        201,
				response:    `http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=`,
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:  "incorrect POST",
			query: "123",
			want: want{
				code:        404,
				response:    `404 page not found`,
				contentType: `text/plain`,
			},
		},
	}
	data := url.Values{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			router := setupRuter(data)
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
