package services

import (
	"context"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/responses"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/workers"
	"net"
)

type UserRepositoryInterface interface {
	AddURL(ctx context.Context, longURL string, shortURL string, user string) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetUserURL(ctx context.Context, user string) ([]responses.GetURL, error)
	AddManyURL(ctx context.Context, urls []responses.ManyPostURL, user string) ([]responses.ManyPostResponse, error)
	DeleteManyURL(ctx context.Context, urls []string, user string) error
	GetStats(ctx context.Context) (responses.StatResponse, error)
	Ping(ctx context.Context) error
}

func NewURLService(repo UserRepositoryInterface, baseURL string, wp *workers.WorkerPool, subnet *net.IPNet) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
		wp:      wp,
		subnet:  subnet,
	}
}

type URLService struct {
	repo    UserRepositoryInterface
	baseURL string
	wp      *workers.WorkerPool
	subnet  *net.IPNet
}

func (us *URLService) GetURL(ctx context.Context, userID string) (string, error) {
	return us.repo.GetURL(ctx, userID)
}

func (us *URLService) CreateURL(ctx context.Context, longURL string, user string) (string, error) {
	shortURL := shortener.ShorterURL(longURL)
	err := us.repo.AddURL(ctx, longURL, shortURL, user)
	return us.baseURL + shortURL, err
}

func (us *URLService) GetUserURL(ctx context.Context, userID string) ([]responses.GetURL, error) {
	return us.repo.GetUserURL(ctx, userID)
}

func (us *URLService) PingDB(ctx context.Context) error {
	return us.repo.Ping(ctx)
}

func (us *URLService) CreateBatch(ctx context.Context, urls []responses.ManyPostURL, userID string) ([]responses.ManyPostResponse, error) {
	return us.repo.AddManyURL(ctx, urls, userID)
}

func (us *URLService) DeleteBatch(urls []string, userID string) {
	var sliceData [][]string
	for i := 10; i <= len(urls); i += 10 {
		sliceData = append(sliceData, urls[i-10:i])
	}
	rem := len(urls) % 10
	if rem > 0 {
		sliceData = append(sliceData, urls[len(urls)-rem:])
	}
	for _, item := range sliceData {
		func(taskData []string) {
			us.wp.Push(func(ctx context.Context) error {
				err := us.repo.DeleteManyURL(ctx, taskData, userID)
				return err
			})
		}(item)
	}
}

func (us *URLService) GetStats(ctx context.Context, ip net.IP) (bool, responses.StatResponse, error) {
	if us.subnet == nil || !us.subnet.Contains(ip) {
		return false, responses.StatResponse{}, nil
	}
	response, err := us.repo.GetStats(ctx)
	return true, response, err
}
