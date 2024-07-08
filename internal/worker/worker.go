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
	wg          *sync.WaitGroup
	concurrency int
}

func NewWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{
		concurrency: concurrency,
		taskChan:    make(chan Task, concurrency),
		wg:          &sync.WaitGroup{},
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()
	for {
		select {
		case task, ok := <-wp.taskChan:
			if !ok {
				return
			}
			task.Process()
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) Run(ctx context.Context) {
	wp.wg.Add(wp.concurrency)
	for i := 0; i < wp.concurrency; i++ { //nolint:intrange //requires go v1.22
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) AddTask(task Task) {
	wp.taskChan <- task
}

func (wp *WorkerPool) Wait() {
	close(wp.taskChan)
	wp.wg.Wait()
}
