// Package worker provides a worker pool implementation for concurrent task processing.
package worker

import (
	"context"
	"sync"
)

// Task defines the interface that must be implemented by any task that can be processed by the worker pool.
type Task interface {
	// Process executes the task.
	Process()
}

// WorkerPool manages a pool of workers to process tasks concurrently.
// It provides methods to add tasks and wait for their completion.
type WorkerPool struct {
	taskChan    chan Task       // Channel to receive tasks for processing.
	wg          *sync.WaitGroup // WaitGroup to wait for all tasks to complete.
	concurrency int             // Number of concurrent workers.
}

// NewWorkerPool creates a new WorkerPool with the specified concurrency level.
// It initializes the task channel and wait group.
func NewWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{
		concurrency: concurrency,
		taskChan:    make(chan Task, concurrency),
		wg:          &sync.WaitGroup{},
	}
}

// worker is a method that processes tasks from the task channel.
// It runs as a goroutine and listens for tasks until the channel is closed or the context is done.
func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done() // Signal that the worker is done when the function exits.
	for {
		select {
		case task, ok := <-wp.taskChan: // Receive task from the task channel.
			if !ok {
				return // Exit if the channel is closed.
			}
			task.Process() // Process the received task.
		case <-ctx.Done(): // Exit if the context is done.
			return
		}
	}
}

// Run starts the worker pool by creating the specified number of worker goroutines.
// It adds the number of workers to the wait group.
func (wp *WorkerPool) Run(ctx context.Context) {
	wp.wg.Add(wp.concurrency)             // Add workers to the wait group.
	for i := 0; i < wp.concurrency; i++ { //nolint:intrange //requires go v1.22
		go wp.worker(ctx) // Start a worker goroutine.
	}
}

// AddTask adds a new task to the worker pool for processing.
// It sends the task to the task channel.
func (wp *WorkerPool) AddTask(task Task) {
	wp.taskChan <- task // Send task to the channel.
}

// Wait closes the task channel and waits for all tasks to be processed.
// It ensures that all worker goroutines complete their work before returning.
func (wp *WorkerPool) Wait() {
	close(wp.taskChan) // Close the task channel to signal no more tasks.
	wp.wg.Wait()       // Wait for all workers to complete.
}
