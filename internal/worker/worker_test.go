package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockTask struct {
	processed *atomic.Int32
}

func newMockTask(counter *atomic.Int32) *mockTask {
	return &mockTask{processed: counter}
}

func (t *mockTask) Process() {
	t.processed.Add(1)
}

func TestWorkerPool(t *testing.T) {
	t.Run("processes all tasks with multiple workers", func(t *testing.T) {
		wp := NewWorkerPool(3)
		ctx := context.Background()
		processed := &atomic.Int32{}

		wp.Run(ctx)

		const numTasks = 10
		for range [numTasks]int{} {
			wp.AddTask(newMockTask(processed))
		}

		wp.Wait()
		assert.Equal(t, int32(numTasks), processed.Load())
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		wp := NewWorkerPool(2)
		ctx, cancel := context.WithCancel(context.Background())
		processed := &atomic.Int32{}

		wp.Run(ctx)

		// Add some tasks
		for range [5]int{} {
			wp.AddTask(newMockTask(processed))
		}

		// Cancel context before all tasks complete
		cancel()
		wp.Wait()

		// Verify some tasks were processed
		assert.True(t, processed.Load() > 0)
	})

	t.Run("worker pool with zero tasks", func(t *testing.T) {
		wp := NewWorkerPool(1)
		ctx := context.Background()

		wp.Run(ctx)
		wp.Wait()
		// Test passes if no panic occurs
	})

	t.Run("concurrent task addition", func(t *testing.T) {
		wp := NewWorkerPool(4)
		ctx := context.Background()
		processed := &atomic.Int32{}

		wp.Run(ctx)

		// Concurrently add tasks
		go func() {
			for range [5]int{} {
				wp.AddTask(newMockTask(processed))
			}
		}()

		go func() {
			for range [5]int{} {
				wp.AddTask(newMockTask(processed))
			}
		}()

		// Give time for tasks to be added
		time.Sleep(100 * time.Millisecond)
		wp.Wait()

		assert.Equal(t, int32(10), processed.Load())
	})
}
