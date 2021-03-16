package paperboy

import "time"

// Parameter represents a single function parameter.
type Parameter interface{}

// Task represents pointer to the scheduled function. It must return an error.
type Task interface{}

// TaskConfig holds the configuration for a specified task, such as the name, period,
// and other configurations.
type TaskConfig struct {
	Name   string
	Period time.Duration

	Recover       bool
	RecoverPeriod time.Duration
	RecoverStep   time.Duration // Not yet implemented.
}

// A TaskerFactory is responsible for creating tasks.
type TaskerFactory interface {
	CreateTasker(conf TaskConfig, task Task, params ...Parameter) (Tasker, error)
}

// A Tasker corresponds to a task, and is responsible for the execution.
// Execution begins in a seperate goroutine, and is triggered periodically
// as configured.
type Tasker interface {
	Start()
}
