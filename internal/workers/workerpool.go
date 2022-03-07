// Package workers - пакет для работы с асинхронными задачами.
package workers

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// WorkerPool - структура для создания и управление пулом воркеров.
type WorkerPool struct {
	numOfWorkers int
	inputCh      chan func(ctx context.Context) error
}

// New - создание структуры WorkerPool.
func New(ctx context.Context, numOfWorkers int, buffer int) *WorkerPool {
	wp := &WorkerPool{
		numOfWorkers: numOfWorkers,
		inputCh:      make(chan func(ctx context.Context) error, buffer),
	}
	return wp
}

// Run - запуск работы WorkerPool.
// Запускается numOfWorkers горутин, которые выполняют полезную рабту.
// Функция ждет завершения всех горутин.
func (wp *WorkerPool) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}
	for i := 0; i < wp.numOfWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Printf("Worker #%v start \n", i)
		outer:
			for {
				select {
				case f := <-wp.inputCh:
					err := f(ctx)
					if err != nil {
						fmt.Printf("Error on worker #%v: %v\n", i, err.Error())
					}
				case <-ctx.Done():
					break outer
				}

			}
			log.Printf("Worker #%v close\n", i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(wp.inputCh)
}

// Push - загрузка задачи в канал выполнения.
func (wp *WorkerPool) Push(task func(ctx context.Context) error) {
	wp.inputCh <- task
}
