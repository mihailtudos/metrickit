package worker

import (
	"context"
	"sync"
)

type Task interface {
	Process()
}

type WorkerPool struct {
	taskChan    chan Task
	wg          sync.WaitGroup
	concurrency int
}

func NewWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{
		concurrency: concurrency,
		taskChan:    make(chan Task),
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	for {
		select {
		case task, ok := <-wp.taskChan:
			if !ok {
				return
			}
			task.Process()
			wp.wg.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) Run(ctx context.Context) {
	for range wp.concurrency {
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) AddTask(task Task) {
	wp.wg.Add(1)
	wp.taskChan <- task
}

func (wp *WorkerPool) Wait() {
	close(wp.taskChan)
	wp.wg.Wait()
}
