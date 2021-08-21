package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name   string
		query  string
		method string
		body   string
		want   want
	}{
		{
			name:   "GET without id",
			query:  "",
			method: "GET",
			body:   "",
			want: want{
				code:        400,
				response:    `{"detail":"Bad request"}`,
				contentType: "application/json",
			},
		},
		{
			name:   "correct POST",
			query:  "",
			method: "POST",
			body:   "http://iloverestaurant.ru/",
			want: want{
				code:        201,
				response:    "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=",
				contentType: "application/json",
			},
		},
		{
			name:   "incorrect POST",
			query:  "/122",
			method: "POST",
			body:   "http://iloverestaurant.ru/",
			want: want{
				code:        400,
				response:    `{"detail":"Bad request"}`,
				contentType: "application/json",
			},
		},
		{
			name:   "GET with correct id",
			query:  "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			method: "GET",
			body:   "",
			want: want{
				code:        307,
				response:    "",
				contentType: "application/json",
			},
		},
		{
			name:   "GET with incorrect id",
			query:  "98fv58Wr3hGGIzm2-aH2zA6",
			method: "GET",
			body:   "",
			want: want{
				code:        404,
				response:    `{"detail":"Not found"}`,
				contentType: "application/json",
			},
		},
		{
			name:   "incorrect method",
			query:  "",
			method: "PUT",
			body:   "",
			want: want{
				code:        405,
				response:    `{"detail":"Method not allowed"}`,
				contentType: "application/json",
			},
		},
		{
			name:   "incorrect url",
			query:  "/users/1",
			method: "GET",
			body:   "",
			want: want{
				code:        404,
				response:    `{"detail":"Page not found"}`,
				contentType: "application/json",
			},
		},
	}

	data := url.Values{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)

			request := httptest.NewRequest(tt.method, "/"+tt.query, body)

			w := httptest.NewRecorder()

			h := http.HandlerFunc(URLHandler(data))

			h.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, res.StatusCode, tt.want.code)

			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)

			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, string(resBody), tt.want.response)
		})
	}
}
