// Package queue provides background job processing, exponential backoff retries, and Dead-Letter Queue (DLQ) management.
package queue

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Job represents a background task unit dispatched to the worker queue.
type Job struct {
	JobID             string         `json:"job_id"`
	Queue             string         `json:"queue"`
	Task              string         `json:"task"`
	Payload           map[string]any `json:"payload,omitempty"`
	MaxRetries        int            `json:"max_retries"`
	RetryCount        int            `json:"retry_count"`
	RetryDelaySeconds int            `json:"retry_delay_seconds"`
	LastError         string         `json:"last_error,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
}

// JobHandler is a function type for processing background tasks.
type JobHandler func(job Job) error

// JobWorker manages background task execution, handlers, and dead-letter queues.
type JobWorker struct {
	mu       sync.RWMutex
	handlers map[string]JobHandler
	dlq      []Job
}

// NewJobWorker initializes a new JobWorker instance.
func NewJobWorker() *JobWorker {
	return &JobWorker{
		handlers: make(map[string]JobHandler),
		dlq:      make([]Job, 0),
	}
}

// CalculateBackoff computes exponential backoff duration: delay = base_seconds * (2 ^ retryCount).
func CalculateBackoff(retryCount int, baseSeconds int) time.Duration {
	if baseSeconds <= 0 {
		baseSeconds = 5
	}
	multiplier := math.Pow(2, float64(retryCount))
	seconds := float64(baseSeconds) * multiplier
	return time.Duration(seconds) * time.Second
}

// RegisterHandler registers a task handler function for a specific task name.
func (w *JobWorker) RegisterHandler(taskName string, handler JobHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[taskName] = handler
}

// ProcessJob executes the registered handler for the given job.
func (w *JobWorker) ProcessJob(job Job) error {
	w.mu.RLock()
	handler, exists := w.handlers[job.Task]
	w.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no handler registered for task %q", job.Task)
		job.LastError = err.Error()
		w.moveToDLQ(job)
		return err
	}

	err := handler(job)
	if err != nil {
		job.LastError = err.Error()
		if job.RetryCount >= job.MaxRetries {
			// Max retries exceeded -> Route to Dead-Letter Queue (DLQ)
			w.moveToDLQ(job)
			return fmt.Errorf("job %s failed after %d retries and moved to DLQ: %w", job.JobID, job.MaxRetries, err)
		}

		// Schedule retry with exponential backoff
		job.RetryCount++
		return fmt.Errorf("job %s failed (retry %d/%d): %w", job.JobID, job.RetryCount, job.MaxRetries, err)
	}

	return nil
}

func (w *JobWorker) moveToDLQ(job Job) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dlq = append(w.dlq, job)
}

// GetDLQJobs returns all failed jobs currently in the Dead-Letter Queue.
func (w *JobWorker) GetDLQJobs() []Job {
	w.mu.RLock()
	defer w.mu.RUnlock()
	copied := make([]Job, len(w.dlq))
	copy(copied, w.dlq)
	return copied
}

// PurgeDLQ clears all jobs from the Dead-Letter Queue.
func (w *JobWorker) PurgeDLQ() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dlq = make([]Job, 0)
}
