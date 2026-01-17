package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/sepolia-sh/ch2/pkg/rpc"
)

// Scheduler orchestrates work distribution across multiple RPC endpoints.
// It uses a pull-based model where workers independently pull tasks from a shared queue.
type Scheduler struct {
	clients  []*rpc.Client
	contract common.Address
	topic    common.Hash

	batchSize  uint64
	bufferSize int
}

// Config holds scheduler configuration.
type Config struct {
	Contract   common.Address
	Topic      common.Hash
	BatchSize  uint64 // blocks per task
	BufferSize int    // task queue buffer size
}

// New creates a new scheduler with the given RPC clients.
func New(clients []*rpc.Client, cfg Config) *Scheduler {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 1000
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = len(clients) * 2
	}

	return &Scheduler{
		clients:    clients,
		contract:   cfg.Contract,
		topic:      cfg.Topic,
		batchSize:  cfg.BatchSize,
		bufferSize: cfg.BufferSize,
	}
}

// Run executes the scheduler from startBlock to endBlock.
// Returns the total number of logs found and any error.
func (s *Scheduler) Run(ctx context.Context, startBlock, endBlock uint64) (int, error) {
	// Create channels
	tasks := make(chan Task, s.bufferSize)
	results := make(chan Result, s.bufferSize)

	// Start workers
	var wg sync.WaitGroup
	for _, client := range s.clients {
		worker := NewWorker(client, s.contract, s.topic, tasks, results)
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker.Run(ctx)
		}()
	}

	// Start task generator
	go s.generateTasks(ctx, startBlock, endBlock, tasks)

	// Start result collector
	totalTasks := s.countTasks(startBlock, endBlock)
	totalLogs, err := s.collectResults(ctx, results, totalTasks)

	// Wait for workers to finish
	wg.Wait()

	// Print stats
	s.printStats()

	return totalLogs, err
}

// generateTasks creates tasks and sends them to the task channel.
func (s *Scheduler) generateTasks(ctx context.Context, start, end uint64, tasks chan<- Task) {
	defer close(tasks)

	taskID := 0
	for from := start; from <= end; from += s.batchSize {
		to := from + s.batchSize - 1
		if to > end {
			to = end
		}

		task := Task{
			ID:        taskID,
			FromBlock: from,
			ToBlock:   to,
		}

		select {
		case <-ctx.Done():
			return
		case tasks <- task:
			taskID++
		}
	}
}

// countTasks calculates the total number of tasks.
func (s *Scheduler) countTasks(start, end uint64) int {
	if end < start {
		return 0
	}
	return int((end-start)/s.batchSize) + 1
}

// collectResults gathers results from workers.
func (s *Scheduler) collectResults(ctx context.Context, results <-chan Result, totalTasks int) (int, error) {
	totalLogs := 0
	completed := 0

	for completed < totalTasks {
		select {
		case <-ctx.Done():
			return totalLogs, ctx.Err()
		case result := <-results:
			completed++
			if result.Err != nil {
				log.Printf("Task %d failed: %v", result.Task.ID, result.Err)
				// In production, you might want to retry failed tasks
				continue
			}
			totalLogs += result.LogCount
		}
	}

	return totalLogs, nil
}

// printStats logs the final statistics for each RPC.
func (s *Scheduler) printStats() {
	log.Println("=== RPC Statistics ===")
	for _, client := range s.clients {
		requests, failures, avgLatency := client.Stats().GetStats()
		log.Printf("[%s] requests=%d failures=%d avg_latency=%v",
			client.Name(), requests, failures, avgLatency)
	}
}
