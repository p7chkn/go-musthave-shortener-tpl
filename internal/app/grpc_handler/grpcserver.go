package grpchandler

import (
	"context"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/responses"
	custom_errors "github.com/p7chkn/go-musthave-shortener-tpl/internal/errors"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/pb"
	"net"
	"net/http"
	"strconv"
)

func NewGRPCHandler(service handlers.URLServiceInterface) *URLServer {
	return &URLServer{
		service: service,
	}
}

type URLServer struct {
	pb.UnimplementedURLServer
	service handlers.URLServiceInterface
}

func (us *URLServer) Retrieve(ctx context.Context, in *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {

	long, err := us.service.GetURL(ctx, in.ShortUrlId)
	if err != nil {
		statusCode := custom_errors.ParseError(err)
		switch statusCode {
		case http.StatusGone:
			return &pb.RetrieveResponse{
				Status: "gone",
			}, nil
		case http.StatusNotFound:
			return &pb.RetrieveResponse{
				Status: "not found",
			}, nil
		default:
			return &pb.RetrieveResponse{
				Status: "internal server error",
			}, nil
		}
	}
	return &pb.RetrieveResponse{
		RedirectUrl: long,
		Status:      "ok",
	}, nil
}

func (us *URLServer) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	responseURL, err := us.service.CreateURL(ctx, in.OriginalUrl, in.UserId)
	if err != nil {
		statusCode := custom_errors.ParseError(err)
		switch statusCode {
		case http.StatusConflict:
			return &pb.CreateResponse{
				Status: "conflict",
			}, nil
		default:
			return &pb.CreateResponse{
				Status: "internal server error",
			}, nil
		}
	}
	return &pb.CreateResponse{
		Status:      "ok",
		ResponseUrl: responseURL,
	}, nil
}

func (us *URLServer) GetUserURLs(ctx context.Context, in *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	urls, err := us.service.GetUserURL(ctx, in.UserId)
	if err != nil {
		statusCode := custom_errors.ParseError(err)
		switch statusCode {
		case http.StatusNoContent:
			return &pb.GetUserURLsResponse{
				Status: "no content",
			}, nil
		default:
			return &pb.GetUserURLsResponse{
				Status: "internal server error",
			}, nil
		}
	}
	var result []*pb.GetUserURLsResponse_URL
	for i := 0; i < len(urls); i++ {
		result = append(result, &pb.GetUserURLsResponse_URL{
			OriginalUrl: urls[0].OriginalURL,
			ShortUrl:    urls[0].ShortURL,
		})
	}
	return &pb.GetUserURLsResponse{
		Status: "ok",
		Urls:   result,
	}, nil
}

func (us *URLServer) CreateBatch(ctx context.Context, in *pb.CreateBatchRequest) (*pb.CreateBatchResponse, error) {
	var data []responses.ManyPostURL
	for i := 0; i < len(in.Urls); i++ {
		data = append(data, responses.ManyPostURL{
			CorrelationID: strconv.Itoa(int(in.Urls[i].CorrelationId)),
			OriginalURL:   in.Urls[i].OriginalUrl,
		})
	}
	urls, err := us.service.CreateBatch(ctx, data, in.UserId)
	if err != nil {
		return &pb.CreateBatchResponse{
			Status: "internal server error",
		}, nil
	}
	var response []*pb.CreateBatchResponse_URL
	for i := 0; i < len(urls); i++ {
		id, _ := strconv.ParseInt(urls[i].CorrelationID, 10, 32)
		response = append(response, &pb.CreateBatchResponse_URL{
			CorrelationId: int32(id),
			ShortUrl:      urls[i].ShortURL,
		})
	}
	return &pb.CreateBatchResponse{
		Status: "ok",
		Urls:   response,
	}, nil
}

func (us *URLServer) DeleteBatch(ctx context.Context, in *pb.DeleteBatchRequest) (*pb.DeleteBatchResponse, error) {
	us.service.DeleteBatch(in.Urls, in.UserId)
	return &pb.DeleteBatchResponse{
		Status: "accepted",
	}, nil
}

func (us *URLServer) GetStats(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	hasPermission, response, err := us.service.GetStats(ctx, net.IP(in.IpAddress))
	if !hasPermission {
		return &pb.GetStatsResponse{
			Status: "forbidden",
		}, nil
	}
	if err != nil {
		return &pb.GetStatsResponse{
			Status: "bad request",
		}, nil
	}
	return &pb.GetStatsResponse{
		Status: "ok",
		Users:  int32(response.CountUser),
		Urls:   int32(response.CountURL),
	}, nil
}
