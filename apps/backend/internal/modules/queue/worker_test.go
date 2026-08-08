package queue_test

import (
	"errors"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/queue"
)

func TestCalculateBackoff(t *testing.T) {
	baseSeconds := 5

	// Retry 0: 5 * 2^0 = 5s
	if delay := queue.CalculateBackoff(0, baseSeconds); delay != 5*time.Second {
		t.Errorf("expected 5s for retry 0, got %v", delay)
	}

	// Retry 1: 5 * 2^1 = 10s
	if delay := queue.CalculateBackoff(1, baseSeconds); delay != 10*time.Second {
		t.Errorf("expected 10s for retry 1, got %v", delay)
	}

	// Retry 2: 5 * 2^2 = 20s
	if delay := queue.CalculateBackoff(2, baseSeconds); delay != 20*time.Second {
		t.Errorf("expected 20s for retry 2, got %v", delay)
	}

	// Retry 3: 5 * 2^3 = 40s
	if delay := queue.CalculateBackoff(3, baseSeconds); delay != 40*time.Second {
		t.Errorf("expected 40s for retry 3, got %v", delay)
	}
}

func TestJobWorker_SuccessfulProcessing(t *testing.T) {
	worker := queue.NewJobWorker()
	executed := false

	worker.RegisterHandler("analytics_rollup", func(j queue.Job) error {
		executed = true
		return nil
	})

	job := queue.Job{
		JobID:      "job_101",
		Queue:      "analytics",
		Task:       "analytics_rollup",
		MaxRetries: 3,
	}

	err := worker.ProcessJob(job)
	if err != nil {
		t.Fatalf("unexpected error processing job: %v", err)
	}

	if !executed {
		t.Error("expected task handler to be executed")
	}

	if len(worker.GetDLQJobs()) != 0 {
		t.Errorf("expected DLQ to be empty, got %d jobs", len(worker.GetDLQJobs()))
	}
}

func TestJobWorker_RetryAndDLQRouting(t *testing.T) {
	worker := queue.NewJobWorker()

	// Handler that always fails
	worker.RegisterHandler("failing_task", func(j queue.Job) error {
		return errors.New("database connection timeout")
	})

	job := queue.Job{
		JobID:             "job_fail_999",
		Queue:             "emails",
		Task:              "failing_task",
		MaxRetries:        2,
		RetryCount:        2, // Already at max retries
		RetryDelaySeconds: 5,
	}

	err := worker.ProcessJob(job)
	if err == nil {
		t.Error("expected error for failed task")
	}

	dlqJobs := worker.GetDLQJobs()
	if len(dlqJobs) != 1 {
		t.Fatalf("expected 1 job in DLQ, got %d", len(dlqJobs))
	}

	if dlqJobs[0].JobID != "job_fail_999" {
		t.Errorf("expected DLQ jobID 'job_fail_999', got %q", dlqJobs[0].JobID)
	}

	if dlqJobs[0].LastError != "database connection timeout" {
		t.Errorf("expected LastError to be saved in DLQ job, got %q", dlqJobs[0].LastError)
	}
}
