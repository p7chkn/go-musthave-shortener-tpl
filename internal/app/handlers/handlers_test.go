package handlers

import (
	"context"
	"errors"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/responses"
	custom_errors "github.com/p7chkn/go-musthave-shortener-tpl/internal/errors"
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
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/workers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(useCase URLServiceInterface) (*gin.Engine, *configuration.Config) {
	router := gin.Default()
	key, _ := configuration.GenerateRandom(16)
	cfg := &configuration.Config{
		Key:     key,
		BaseURL: configuration.BaseURL,
	}
	handler := New(useCase)
	router.Use(middlewares.CookiMiddleware(cfg))
	router.GET("/:id", handler.RetrieveShortURL)
	router.POST("/", handler.CreateShortURL)
	router.POST("/api/shorten", handler.ShortenURL)
	router.GET("/user/urls", handler.GetUserURL)
	router.POST("/api/shorten/batch", handler.CreateBatch)
	router.DELETE("/api/user/urls", handler.DeleteBatch)
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
			err:    custom_errors.NewCustomError(errors.New("not found"), http.StatusNotFound),
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
			err:    custom_errors.NewCustomError(errors.New("not found"), http.StatusNotFound),
			want: want{
				code:        404,
				response:    `{"detail":"not found"}`,
				contentType: `application/json; charset=utf-8`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("GetURL", mock.Anything, tt.query).Return(tt.result, tt.err)
			router, _ := setupRouter(useCaseMock)
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
			result: "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=",
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
			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("CreateURL", mock.Anything, tt.body, mock.Anything).Return(tt.result, nil)
			router, _ := setupRouter(useCaseMock)
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
			result:  "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=",
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
				response:    `{"detail": "bad request"}`,
				contentType: `application/json; charset=utf-8`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("CreateURL", mock.Anything, tt.rawData, mock.Anything).Return(tt.result, nil)
			router, _ := setupRouter(useCaseMock)
			body := strings.NewReader(tt.body)
			w := httptest.NewRecorder()
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
		response []responses.GetURL
		want     want
	}{
		{
			name:  "correct GET",
			query: "user/urls",
			response: []responses.GetURL{
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

			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			userID, _ := uuid.NewV4()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("GetUserURL", mock.Anything, userID.String()).Return(tt.response, nil)
			router, cfg := setupRouter(useCaseMock)

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
		name         string
		query        string
		body         string
		mockData     []responses.ManyPostURL
		mockResponce []responses.ManyPostResponse
		want         want
	}{
		{
			name:  "correct POST",
			query: "api/shorten/batch",
			body: `[
				{
					"correlation_id": "1",
					"original_url": "https://www.postgresqltutorial.com/postgresql-create-table/"
				},
				  {
					"correlation_id": "2",
					"original_url": "https://twitter.com/home"
				},
				  {
					"correlation_id": "3",
					"original_url": "https://www.gismeteo.ru/weather-sankt-peterburg-4079/10-days/"
				}
			] `,
			mockData: []responses.ManyPostURL{
				{
					CorrelationID: "1",
					OriginalURL:   "https://www.postgresqltutorial.com/postgresql-create-table/",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://twitter.com/home",
				},
				{
					CorrelationID: "3",
					OriginalURL:   "https://www.gismeteo.ru/weather-sankt-peterburg-4079/10-days/",
				},
			},
			mockResponce: []responses.ManyPostResponse{
				{
					CorrelationID: "1",
					ShortURL:      "http://localhost:8080/Kkm_RJeyfdOxwVZoQA1k9E8Sg8Q=",
				},
				{
					CorrelationID: "2",
					ShortURL:      "http://localhost:8080/RrbgmrELxXSzwnYKBcJInKtp-_I=",
				},
				{
					CorrelationID: "3",
					ShortURL:      "http://localhost:8080/LuHrl3OJA_f61piIambybX8cNvA=",
				},
			},
			want: want{
				code:        201,
				contentType: `application/json; charset=utf-8`,
				response: `[
					{
						"correlation_id": "1",
						"short_url": "http://localhost:8080/Kkm_RJeyfdOxwVZoQA1k9E8Sg8Q="
					},
					{
						"correlation_id": "2",
						"short_url": "http://localhost:8080/RrbgmrELxXSzwnYKBcJInKtp-_I="
					},
					{
						"correlation_id": "3",
						"short_url": "http://localhost:8080/LuHrl3OJA_f61piIambybX8cNvA="
					}
				]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			userID, _ := uuid.NewV4()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("CreateBatch", mock.Anything, tt.mockData, userID.String()).Return(tt.mockResponce, nil)

			router, cfg := setupRouter(useCaseMock)
			encoder, _ := utils.New(cfg.Key)

			cookie := http.Cookie{
				Name:  "userId",
				Value: encoder.EncodeUUIDtoString(userID.Bytes()),
			}

			body := strings.NewReader(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/"+tt.query, body)
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

func TestDeleteBatch(t *testing.T) {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		name  string
		query string
		body  string
		want  want
	}{
		{
			name:  "correct DELETE",
			query: "api/user/urls",
			body:  `["1", "2", "3", "4"]`,
			want: want{
				code:        202,
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

			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			userID, _ := uuid.NewV4()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("DeleteBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			router, cfg := setupRouter(useCaseMock)

			encoder, _ := utils.New(cfg.Key)

			cookie := http.Cookie{
				Name:  "userId",
				Value: encoder.EncodeUUIDtoString(userID.Bytes()),
			}
			body := strings.NewReader(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/"+tt.query, body)
			req.AddCookie(&cookie)

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.want.code, w.Code)

		})
	}
}

func BenchmarkHandler_GetUserURL(b *testing.B) {
	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			wp := workers.New(ctx, configuration.NumOfWorkers, configuration.WorkersBuffer)

			go func() {
				wp.Run(ctx)
			}()
			userID, _ := uuid.NewV4()
			useCaseMock := new(MockUserUseCaseInterface)
			useCaseMock.On("DeleteBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			router, cfg := setupRouter(useCaseMock)

			encoder, _ := utils.New(cfg.Key)

			cookie := http.Cookie{
				Name:  "userId",
				Value: encoder.EncodeUUIDtoString(userID.Bytes()),
			}
			body := strings.NewReader(`["1", "2", "3", "4"]`)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/user/urls", body)
			req.AddCookie(&cookie)

			router.ServeHTTP(w, req)
		}
	})
}

func ExampleHandler_CreateBatch() {
	router := gin.Default()
	userCase := new(MockUserUseCaseInterface)
	handler := New(userCase)
	router.POST("/api/shorten/batch", handler.CreateBatch)
}

func ExampleHandler_CreateShortURL() {
	router := gin.Default()
	userCase := new(MockUserUseCaseInterface)
	handler := New(userCase)
	router.POST("/api/shorten", handler.CreateShortURL)
}
