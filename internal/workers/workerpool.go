package workers

import (
	"context"
	"errors"
	"fmt"
	"log"

	"golang.org/x/sync/errgroup"
)

type WorkerPool struct {
	numOfWorkers int
	inputCh      chan func(ctx context.Context) error
	errorCh      chan error
}

func New(ctx context.Context, numOfWorkers int, cancel context.CancelFunc) *WorkerPool {
	wp := &WorkerPool{
		numOfWorkers: numOfWorkers,
		inputCh:      make(chan func(ctx context.Context) error),
		errorCh:      make(chan error),
	}
	wp.run(ctx)
	return wp
}

func (wp *WorkerPool) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	g, _ := errgroup.WithContext(ctx)
	for i := 0; i < wp.numOfWorkers; i++ {
		g.Go(func() error {
			fmt.Println("Worker start")
		outer:
			for {
				select {
				case f := <-wp.inputCh:
					err := f(ctx)
					if err != nil {
						fmt.Println("WORKER CLOSE with error")
						return errors.New(err.Error())
					}
				case <-ctx.Done():
					break outer
				}

			}
			fmt.Println("WORKER CLOSE")
			return nil
		})
	}
	go func() {
		if err := g.Wait(); err != nil {
			log.Println(err)
			close(wp.inputCh)
			close(wp.errorCh)
			cancel()
			return
		}

	}()
}

func (wp *WorkerPool) Push(task func(ctx context.Context) error) {
	wp.inputCh <- task
}
