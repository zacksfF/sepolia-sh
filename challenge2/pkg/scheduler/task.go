package scheduler

// Task represents a unit of work to be processed by a worker.
// Each task is a block range to fetch logs from.
type Task struct {
	ID        int
	FromBlock uint64
	ToBlock   uint64
}

// Result contains the outcome of processing a task.
type Result struct {
	Task     Task
	WorkerID string
	LogCount int
	Err      error
}
