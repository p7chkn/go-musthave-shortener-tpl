package workers

import (
	"context"
	"sync"
)

type workerRepoInterface interface {
	DeleteManyURL(urls []string, user string) error
	IsOwner(url string, user string) bool
}

type WorkerURLDelete struct {
	ctx          context.Context
	NumOfWorkers int
	inputCh      chan []string
	workers      []chan string
	pool         []chan []string
	repo         workerRepoInterface
}

func NewWorkerURLDelete(ctx context.Context, numOfWorkers int, repo workerRepoInterface) *WorkerURLDelete {
	return &WorkerURLDelete{
		ctx:          ctx,
		NumOfWorkers: numOfWorkers,
		workers:      make([]chan string, 0, numOfWorkers),
		inputCh:      make(chan []string),
		repo:         repo,
	}
}

func (wp *WorkerURLDelete) fanOut() {
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

func (wp *WorkerURLDelete) fanIn() chan string {
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

func (wp *WorkerURLDelete) newWorker(input <-chan []string) chan string {
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

func (wp *WorkerURLDelete) DeleteURL(urls []string, user string) {
	go func() {
	outer:
		for _, url := range urls {
			select {
			case <-wp.ctx.Done():
				close(wp.inputCh)
				break outer
			case wp.inputCh <- []string{url, user}:
			}

		}
		close(wp.inputCh)
	}()
	urlsToDelete := []string{}
	wp.fanOut()

	for _, ch := range wp.pool {
		wp.workers = append(wp.workers, wp.newWorker(ch))
	}
	for url := range wp.fanIn() {
		urlsToDelete = append(urlsToDelete, url)
	}

	wp.repo.DeleteManyURL(urlsToDelete, user)
}
