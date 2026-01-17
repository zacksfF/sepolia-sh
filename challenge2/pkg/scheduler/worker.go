package scheduler

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/sepolia-sh/ch2/pkg/rpc"
)

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
	backoffFactor  = 2.0
)

// Worker processes tasks from a shared queue using its RPC client.
// It implements pull-based scheduling - taking tasks when ready.
type Worker struct {
	id       string
	client   *rpc.Client
	contract common.Address
	topic    common.Hash
	tasks    <-chan Task
	results  chan<- Result
}

// NewWorker creates a new worker with the given RPC client.
func NewWorker(
	client *rpc.Client,
	contract common.Address,
	topic common.Hash,
	tasks <-chan Task,
	results chan<- Result,
) *Worker {
	return &Worker{
		id:       client.Name(),
		client:   client,
		contract: contract,
		topic:    topic,
		tasks:    tasks,
		results:  results,
	}
}

// Run starts the worker loop. It pulls tasks from the queue and processes them.
// The worker stops when the context is cancelled or the task channel is closed.
func (w *Worker) Run(ctx context.Context) {
	consecutiveFailures := 0

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-w.tasks:
			if !ok {
				return // channel closed, no more tasks
			}

			// Apply backoff if we've had recent failures
			if consecutiveFailures > 0 {
				backoff := w.calculateBackoff(consecutiveFailures)
				log.Printf("[%s] backing off for %v after %d failures", w.id, backoff, consecutiveFailures)
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}
			}

			// Process the task
			result := w.processTask(ctx, task)

			// Track consecutive failures for backoff
			if result.Err != nil {
				consecutiveFailures++
				log.Printf("[%s] task %d failed: %v", w.id, task.ID, result.Err)
			} else {
				consecutiveFailures = 0
				log.Printf("[%s] completed task %d (blocks %d-%d): %d logs",
					w.id, task.ID, task.FromBlock, task.ToBlock, result.LogCount)
			}

			// Send result
			select {
			case <-ctx.Done():
				return
			case w.results <- result:
			}
		}
	}
}

// processTask fetches logs for the given task's block range.
func (w *Worker) processTask(ctx context.Context, task Task) Result {
	query := rpc.FilterQuery(w.contract, w.topic, task.FromBlock, task.ToBlock)

	logs, err := w.client.FilterLogs(ctx, query)

	return Result{
		Task:     task,
		WorkerID: w.id,
		LogCount: len(logs),
		Err:      err,
	}
}

// calculateBackoff returns the backoff duration based on failure count.
// Uses exponential backoff with a maximum cap.
func (w *Worker) calculateBackoff(failures int) time.Duration {
	backoff := float64(initialBackoff) * math.Pow(backoffFactor, float64(failures-1))
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}
	return time.Duration(backoff)
}
