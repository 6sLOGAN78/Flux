package app

import (
	"context"
	"errors"
	"strings"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/server"
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
	r.readiness = workerReadinessChecks(r.Server, adapter)
	return nil
}

func workerReadinessChecks(srv *server.Server, adapter *email.Client) []handler.ReadinessCheck {
	return []handler.ReadinessCheck{queueReadinessCheck(srv), {Name: "email", Check: func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if adapter == nil || strings.TrimSpace(srv.Config.Integration.ResendAPIKey) == "" {
			return errors.New("email adapter unconfigured")
		}
		// Exercise required embedded assets locally; never send email or ping
		// the provider during queue-consumer readiness.
		_, err := adapter.Render(email.TemplateWelcome, map[string]string{"UserFirstName": "Health"})
		return err
	}}}
}
