package app

import (
	"context"
	"errors"
	"strings"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/server"
)

// Worker owns the worker role runtime.
type Worker struct{ *RoleRuntime }

// NewWorker constructs PostgreSQL, Redis, the email adapter and consumer processing, plus
// a management-only listener with no public API or documentation routes.
func NewWorker(ctx context.Context, cfg *config.Config) (*Worker, error) {
	runtime, err1 := newRole(ctx, config.RoleWorker, cfg, defaultRoleFactories())
	if err1 != nil {
		return nil, err1
	}
	return &Worker{runtime}, nil
}

// NewWorkerWithInvitationSender injects an explicit provider adapter while retaining
// the complete worker resource graph. No environment setting selects this seam.
func NewWorkerWithInvitationSender(ctx context.Context,
	cfg *config.Config, sender job.InvitationSender,
) (*Worker, error) {
	f := defaultRoleFactories()
	f.consumer = func(srv *server.Server, adapter *email.Client) (*job.JobService, func(context.Context) error, error) {
		consumer := job.NewConsumer(srv.Logger, srv.Config, srv.Redis, adapter, srv.Telemetry)
		err2 := consumer.ConfigureInvitations(srv.DB.Pool, srv.Config.Invitations, sender)
		return consumer, func(context.Context) error { return consumer.Stop() }, err2
	}
	runtime, err3 := newRole(ctx, config.RoleWorker, cfg, f)
	if err3 != nil {
		return nil, err3
	}
	return &Worker{runtime}, nil
}

func (r *RoleRuntime) constructWorker(f roleFactories) error {
	adapter, err4 := f.email(r.Server)
	var closeEmail func(context.Context) error
	if adapter != nil {
		closeEmail = func(context.Context) error { return adapter.Close() }
	}
	if err5 := r.own("email", closeEmail, err4); err5 != nil {
		return err5
	}
	consumer, closeConsumer, err4 := f.consumer(r.Server, adapter)
	if err32 := r.own("consumer", closeConsumer, err4); err32 != nil {
		return err32
	}
	r.Server.Job = consumer
	if consumer != nil {
		r.lifecycle.stop = consumer.StopIntake
	}
	r.readiness = workerReadinessChecks(r.Server, adapter)
	return nil
}

func workerReadinessChecks(srv *server.Server, adapter *email.Client) []handler.ReadinessCheck {
	return append(apiReadinessChecks(srv, true),
		handler.ReadinessCheck{Name: "email",
			Check: func(ctx context.Context) error {
				if err6 := ctx.Err(); err6 != nil {
					return err6
				}
				if adapter == nil || strings.TrimSpace(srv.Config.Integration.ResendAPIKey) == "" {
					return errors.New("email adapter unconfigured")
				}
				if err7 := srv.Config.Invitations.Validate(); err7 != nil {
					return err7
				}
				return adapter.CheckLocal(ctx)
			}})
}
