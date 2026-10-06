package app

import (
	"context"

	"github.com/6sLOGAN78/flux/internal/config"
)

type Worker struct{ *RoleRuntime }

// NewWorker constructs Redis, the email adapter and consumer processing, plus
// a management-only listener with no public API or documentation routes.
func NewWorker(ctx context.Context, cfg *config.Config) (*Worker, error) {
	runtime, err := newRole(ctx, config.RoleWorker, cfg, defaultRoleFactories())
	if err != nil {
		return nil, err
	}
	return &Worker{runtime}, nil
}

func (r *RoleRuntime) constructWorker(f roleFactories) error {
	adapter, err := f.email(r.Server)
	if err != nil {
		return &StartupError{Role: r.Role, Stage: "email", cause: err}
	}
	consumer, closeConsumer, err := f.consumer(r.Server, adapter)
	if err := r.own("consumer", closeConsumer, err); err != nil {
		return err
	}
	r.Server.Job = consumer
	return nil
}
