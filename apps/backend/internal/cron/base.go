// Package cron manages scheduled background job execution.
package cron

import (
	"context"
)

// Job defines the interface for executable background cron jobs.
type Job interface {
	Name() string
	Execute(ctx context.Context) error
}

// BaseJob provides standard job metadata.
type BaseJob struct {
	JobName string
}

func (b *BaseJob) Name() string {
	return b.JobName
}
