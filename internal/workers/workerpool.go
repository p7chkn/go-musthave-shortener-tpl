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
	stopCh       chan struct{}
}

func New(ctx context.Context, numOfWorkers int, cancel context.CancelFunc) *WorkerPool {
	wp := &WorkerPool{
		numOfWorkers: numOfWorkers,
		inputCh:      make(chan func(ctx context.Context) error),
		errorCh:      make(chan error),
		stopCh:       make(chan struct{}),
	}
	wp.run(ctx, cancel)
	return wp
}

func (wp *WorkerPool) run(ctx context.Context, cancel context.CancelFunc) {
	g, _ := errgroup.WithContext(ctx)
	for i := 0; i < wp.numOfWorkers; i++ {
		fmt.Println("--------------- start worker", i)
		g.Go(func() error {
		outer:
			for {
				select {
				case f := <-wp.inputCh:
					err := f(ctx)
					if err != nil {
						close(wp.stopCh)
						fmt.Println("SEND ERROR")
						return errors.New(err.Error())
					}
				case <-ctx.Done():
					fmt.Println("CONTEXT DONE")
					break outer
				case <-wp.stopCh:
					fmt.Println("RECIVIE ERROR")
					return errors.New("GONE")
				}

			}
			fmt.Println("Close worker")
			return nil
		})
	}

	go func() {
		fmt.Println("_____--------------------------START")
		err := g.Wait()
		fmt.Println("------------------", err)
		if err != nil {
			log.Println(err)
			cancel()
			return
		}
		fmt.Println(" ----------------------- CLOSE")
	}()

}

func (wp *WorkerPool) Push(task func(ctx context.Context) error) {
	wp.inputCh <- task
}
