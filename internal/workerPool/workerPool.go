package workerPool

import (
	"fmt"
	"sync"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
)

type WorkerPool struct {
	NumOfWorkers int
	inputCh      chan []string
	outCh        chan string
	workers      []chan string
	pool         []chan []string
	repo         handlers.RepositoryInterface
}

func New(numOfWorkers int) *WorkerPool {
	return &WorkerPool{
		NumOfWorkers: numOfWorkers,
		workers:      make([]chan string, 0, numOfWorkers),
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
		close(wp.outCh)
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
	fmt.Println("In Delete")
	go func() {
		for _, url := range urls {
			fmt.Println("InitLoop")
			wp.inputCh <- []string{url, user}
		}
		close(wp.inputCh)
	}()
	urlsToDelete := []string{}

	wp.fanOut()

	for _, ch := range wp.pool {
		fmt.Println("poll")
		wp.workers = append(wp.workers, wp.NewWorker(ch))
	}

	for url := range wp.fanIn() {
		fmt.Println("fanIN")
		urlsToDelete = append(urlsToDelete, url)
	}

	wp.repo.DeleteManyURL(urlsToDelete, user)
}
