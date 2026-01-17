package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// MockClient simulates an RPC client with configurable latency.
type MockClient struct {
	name       string
	latency    time.Duration
	taskCount  atomic.Int32
	shouldFail bool
}

func (m *MockClient) Name() string { return m.name }

func (m *MockClient) Process(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.latency):
	}
	m.taskCount.Add(1)
	if m.shouldFail {
		return context.DeadlineExceeded
	}
	return nil
}

func TestPullBasedDistribution(t *testing.T) {
	// Create a shared task queue
	tasks := make(chan int, 100)
	var wg sync.WaitGroup

	// Create workers with different "speeds"
	fastWorker := &MockClient{name: "fast", latency: 10 * time.Millisecond}
	slowWorker := &MockClient{name: "slow", latency: 100 * time.Millisecond}

	// Worker function - pulls tasks from queue
	worker := func(client *MockClient) {
		defer wg.Done()
		for range tasks {
			client.Process(context.Background())
		}
	}

	// Start workers
	wg.Add(2)
	go worker(fastWorker)
	go worker(slowWorker)

	// Send 20 tasks
	for i := 0; i < 20; i++ {
		tasks <- i
	}
	close(tasks)

	wg.Wait()

	fastCount := fastWorker.taskCount.Load()
	slowCount := slowWorker.taskCount.Load()

	t.Logf("Fast worker completed: %d tasks", fastCount)
	t.Logf("Slow worker completed: %d tasks", slowCount)

	// Fast worker should complete more tasks
	if fastCount <= slowCount {
		t.Errorf("Expected fast worker (%d) to complete more tasks than slow worker (%d)",
			fastCount, slowCount)
	}

	// Total should equal 20
	if fastCount+slowCount != 20 {
		t.Errorf("Expected 20 total tasks, got %d", fastCount+slowCount)
	}
}

func TestBackoffCalculation(t *testing.T) {
	w := &Worker{}

	tests := []struct {
		failures int
		expected time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second}, // capped at max
		{10, 30 * time.Second},
	}

	for _, tt := range tests {
		got := w.calculateBackoff(tt.failures)
		if got != tt.expected {
			t.Errorf("calculateBackoff(%d) = %v, want %v", tt.failures, got, tt.expected)
		}
	}
}

func TestTaskGeneration(t *testing.T) {
	s := &Scheduler{batchSize: 1000}

	tests := []struct {
		start, end uint64
		expected   int
	}{
		{0, 999, 1},
		{0, 1000, 2},
		{0, 2999, 3},
		{1000, 2999, 2},
		{0, 0, 1},
	}

	for _, tt := range tests {
		got := s.countTasks(tt.start, tt.end)
		if got != tt.expected {
			t.Errorf("countTasks(%d, %d) = %d, want %d", tt.start, tt.end, got, tt.expected)
		}
	}
}
