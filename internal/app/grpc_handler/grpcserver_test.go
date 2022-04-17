package grpchandler

import (
	"context"
	"errors"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/responses"
	custom_errors "github.com/p7chkn/go-musthave-shortener-tpl/internal/errors"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/pb"
	"github.com/stretchr/testify/mock"
	"net"
	"net/http"
	"reflect"
	"testing"
)

func TestURLServer_Create(t *testing.T) {
	type result struct {
		res string
		err error
	}

	tests := []struct {
		name    string
		query   string
		request *pb.CreateRequest
		result  result
		want    *pb.CreateResponse
		wantErr bool
	}{
		{
			name:  "correct POST",
			query: "http://iloverestaurant.ru/",
			request: &pb.CreateRequest{
				OriginalUrl: "http://iloverestaurant.ru/",
				UserId:      "1",
			},
			result: result{
				res: "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=",
				err: nil,
			},
			want: &pb.CreateResponse{
				Status:      "ok",
				ResponseUrl: "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=",
			},
		},
		{
			name:  "conflict POST",
			query: "http://iloverestaurant.ru/",
			request: &pb.CreateRequest{
				OriginalUrl: "http://iloverestaurant.ru/",
				UserId:      "1",
			},
			result: result{
				res: "http://localhost:8080/98fv58Wr3hGGIzm2-aH2zA628Ng=",
				err: custom_errors.NewCustomError(errors.New("conflict"), http.StatusConflict),
			},
			want: &pb.CreateResponse{
				Status: "conflict",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceMock := new(handlers.MockUserUseCaseInterface)

			serviceMock.On("CreateURL", mock.Anything, tt.query, mock.Anything).
				Return(tt.result.res, tt.result.err)

			us := NewGRPCHandler(serviceMock)
			got, err := us.Create(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create () error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create () got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLServer_CreateBatch(t *testing.T) {
	type result struct {
		res []responses.ManyPostResponse
		err error
	}

	tests := []struct {
		name    string
		query   []responses.ManyPostURL
		request *pb.CreateBatchRequest
		result  result
		want    *pb.CreateBatchResponse
		wantErr bool
	}{
		{
			name: "correct POST",
			query: []responses.ManyPostURL{
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
			request: &pb.CreateBatchRequest{
				UserId: "1",
				Urls: []*pb.CreateBatchRequest_URL{
					{
						CorrelationId: 1,
						OriginalUrl:   "https://www.postgresqltutorial.com/postgresql-create-table/",
					},
					{
						CorrelationId: 2,
						OriginalUrl:   "https://twitter.com/home",
					},
					{
						CorrelationId: 3,
						OriginalUrl:   "https://www.gismeteo.ru/weather-sankt-peterburg-4079/10-days/",
					},
				},
			},
			result: result{
				res: []responses.ManyPostResponse{
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
				err: nil,
			},
			want: &pb.CreateBatchResponse{
				Status: "ok",
				Urls: []*pb.CreateBatchResponse_URL{
					{
						CorrelationId: 1,
						ShortUrl:      "http://localhost:8080/Kkm_RJeyfdOxwVZoQA1k9E8Sg8Q=",
					},
					{
						CorrelationId: 2,
						ShortUrl:      "http://localhost:8080/RrbgmrELxXSzwnYKBcJInKtp-_I=",
					},
					{
						CorrelationId: 3,
						ShortUrl:      "http://localhost:8080/LuHrl3OJA_f61piIambybX8cNvA=",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceMock := new(handlers.MockUserUseCaseInterface)

			serviceMock.On("CreateBatch", mock.Anything, tt.query, mock.Anything).
				Return(tt.result.res, tt.result.err)

			us := NewGRPCHandler(serviceMock)
			got, err := us.CreateBatch(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateBatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLServer_DeleteBatch(t *testing.T) {

	tests := []struct {
		name    string
		request *pb.DeleteBatchRequest
		want    *pb.DeleteBatchResponse
		wantErr bool
	}{
		{
			name: "success batch delete",
			request: &pb.DeleteBatchRequest{
				UserId: "1",
				Urls:   []string{"1", "2", "3", "4"},
			},
			want: &pb.DeleteBatchResponse{
				Status: "accepted",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceMock := new(handlers.MockUserUseCaseInterface)

			serviceMock.On("DeleteBatch", mock.Anything, mock.Anything).Return(nil)

			us := NewGRPCHandler(serviceMock)
			got, err := us.DeleteBatch(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteBatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLServer_GetStats(t *testing.T) {
	type result struct {
		hasPermission bool
		res           responses.StatResponse
		err           error
	}
	tests := []struct {
		name    string
		query   net.IP
		request *pb.GetStatsRequest
		result  result
		want    *pb.GetStatsResponse
		wantErr bool
	}{
		{
			name:  "success get stats",
			query: net.IP("127.0.0.1"),
			request: &pb.GetStatsRequest{
				IpAddress: "127.0.0.1",
			},
			result: result{
				hasPermission: true,
				res: responses.StatResponse{
					CountURL:  2,
					CountUser: 3,
				},
				err: nil,
			},
			want: &pb.GetStatsResponse{
				Status: "ok",
				Users:  3,
				Urls:   2,
			},
		},
		{
			name:  "forbidden get stats",
			query: net.IP("127.0.0.1"),
			request: &pb.GetStatsRequest{
				IpAddress: "127.0.0.1",
			},
			result: result{
				hasPermission: false,
				res:           responses.StatResponse{},
				err:           nil,
			},
			want: &pb.GetStatsResponse{
				Status: "forbidden",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceMock := new(handlers.MockUserUseCaseInterface)

			serviceMock.On("GetStats", mock.Anything, tt.query).
				Return(tt.result.hasPermission, tt.result.res, tt.result.err)

			us := NewGRPCHandler(serviceMock)
			got, err := us.GetStats(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStats() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLServer_GetUserURLs(t *testing.T) {
	type result struct {
		res []responses.GetURL
		err error
	}
	tests := []struct {
		name    string
		result  result
		request *pb.GetUserURLsRequest
		want    *pb.GetUserURLsResponse
		wantErr bool
	}{
		{
			name: "success get users urls",
			result: result{
				res: []responses.GetURL{
					{
						ShortURL:    "http://localhost:8080/1yhVmSPGQlZn3EnrI2kd7Oxu5UM=",
						OriginalURL: "http://hbqouwjbx5jl.ru/lkm0skvkix1ejv",
					},
				},
				err: nil,
			},
			request: &pb.GetUserURLsRequest{
				UserId: "1",
			},
			want: &pb.GetUserURLsResponse{
				Status: "ok",
				Urls: []*pb.GetUserURLsResponse_URL{
					{
						ShortUrl:    "http://localhost:8080/1yhVmSPGQlZn3EnrI2kd7Oxu5UM=",
						OriginalUrl: "http://hbqouwjbx5jl.ru/lkm0skvkix1ejv",
					},
				},
			},
		},
		{
			name: "no content get users urls",
			result: result{
				res: []responses.GetURL{},
				err: custom_errors.NewCustomError(errors.New("no content"), http.StatusNoContent),
			},
			request: &pb.GetUserURLsRequest{
				UserId: "1",
			},
			want: &pb.GetUserURLsResponse{
				Status: "no content",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceMock := new(handlers.MockUserUseCaseInterface)

			serviceMock.On("GetUserURL", mock.Anything, mock.Anything).
				Return(tt.result.res, tt.result.err)

			us := NewGRPCHandler(serviceMock)
			got, err := us.GetUserURLs(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserURLs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLServer_Retrieve(t *testing.T) {
	type result struct {
		res string
		err error
	}

	tests := []struct {
		name    string
		query   string
		request *pb.RetrieveRequest
		result  result
		want    *pb.RetrieveResponse
		wantErr bool
	}{
		{
			name:  "GET without incorrect id",
			query: "1234",
			request: &pb.RetrieveRequest{
				ShortUrlId: "1234",
			},
			result: result{
				res: "",
				err: custom_errors.NewCustomError(errors.New("not found"), http.StatusNotFound),
			},
			want: &pb.RetrieveResponse{
				Status: "not found",
			},
		},
		{
			name:  "GET with correct id",
			query: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			request: &pb.RetrieveRequest{
				ShortUrlId: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			},
			result: result{
				res: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
				err: nil,
			},
			want: &pb.RetrieveResponse{
				Status:      "ok",
				RedirectUrl: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			},
		},
		{
			name:  "GET with error request",
			query: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			request: &pb.RetrieveRequest{
				ShortUrlId: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			},
			result: result{
				res: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
				err: errors.New("500 error"),
			},
			want: &pb.RetrieveResponse{
				Status: "internal server error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceMock := new(handlers.MockUserUseCaseInterface)

			serviceMock.On("GetURL", mock.Anything, tt.query).Return(tt.result.res, tt.result.err)

			us := NewGRPCHandler(serviceMock)
			got, err := us.Retrieve(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Retrieve() got = %v, want %v", got, tt.want)
			}
		})
	}
}
