package worker

import (
	"context"
	"sync"
)

type Task interface {
	Process()
}

type WorkerPool struct {
	concurrency int
	taskChan    chan Task
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewWorkerPool(concurrency int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		concurrency: concurrency,
		ctx:         ctx,
		cancel:      cancel,
		taskChan:    make(chan Task),
	}
}

func (wp *WorkerPool) worker() {
	for task := range wp.taskChan {
		task.Process()
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.concurrency; i++ {
		go wp.worker()
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

func (wp *WorkerPool) Stop() {
	wp.cancel()
}
