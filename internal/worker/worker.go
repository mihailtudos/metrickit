package worker

import (
	"sync"
)

type Task interface {
	Process()
}

type WorkerPool struct {
	Tasks       []Task
	concurrency int
	taskChan    chan Task
	wg          sync.WaitGroup
}

func (wp *WorkerPool) worker() {
	for task := range wp.taskChan {
		task.Process()
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Run() {
	// initialize the task chan
	wp.taskChan = make(chan Task, len(wp.Tasks))

	for i := 0; i < wp.concurrency; i++ {
		go wp.worker()
	}

	wp.wg.Add(len(wp.Tasks))

	for _, task := range wp.Tasks {
		wp.taskChan <- task
	}

	close(wp.taskChan)

	wp.wg.Wait()
}
