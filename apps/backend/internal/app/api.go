// Package app composes independently owned process resources and lifecycle operations.
package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	loggerPkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/router"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// StartupError exposes the failed role stage without provider diagnostics.
type StartupError struct {
	cause error
	Role  config.Role
	Stage string
}

func (e *StartupError) Error() string { return string(e.Role) + " startup " + e.Stage + " failed" }
func (e *StartupError) Unwrap() error { return e.cause }

// RoleRuntime owns a bounded resource graph and its HTTP transport.
// The worker's HTTP transport is exclusively a management listener.
type RoleRuntime struct {
	listener  net.Listener
	Server    *server.Server
	HTTP      *echo.Echo
	Role      config.Role
	lifecycle Lifecycle
	readiness []handler.ReadinessCheck
	settings  config.RoleConfig
	cleanup   Cleanup
}

// API owns the API role runtime.
type API struct{ *RoleRuntime }

// NewAPI constructs PostgreSQL and explicitly enabled producers only.
// It never migrates a database or allocates/starts job consumers.
func NewAPI(ctx context.Context, cfg *config.Config) (*API, error) {
	runtime, err := newRole(ctx, config.RoleAPI, cfg, defaultRoleFactories())
	if err != nil {
		return nil, err
	}
	return &API{runtime}, nil
}

type roleFactories struct {
	telemetry func(context.Context,
		*config.Config,
		config.Role) (*observability.Telemetry,
		func(context.Context) error,
		error)
	logger func(*config.Config,
		*observability.Telemetry) (*zerolog.Logger,
		func(context.Context) error,
		error)
	database      func(context.Context, *server.Server) (*database.Database, func(context.Context) error, error)
	redis         func(context.Context, *server.Server) (*redis.Client, func(context.Context) error, error)
	producer      func(*server.Server) (*job.JobService, func(context.Context) error, error)
	email         func(*server.Server) (*email.Client, error)
	consumer      func(*server.Server, *email.Client) (*job.JobService, func(context.Context) error, error)
	startConsumer func(*job.JobService) error
	router        func(config.Role, *server.Server) (*echo.Echo, error)
	listen        func(string) (net.Listener, error)
}

func defaultRoleFactories() roleFactories {
	return roleFactories{
		logger: func(cfg *config.Config,
			owner *observability.Telemetry) (*zerolog.Logger,
			func(context.Context) error,
			error) {
			log := loggerPkg.NewLogger(cfg.Observability, os.Stdout, owner.Logger)
			return &log, nil, nil
		},
		telemetry: func(ctx context.Context,
			cfg *config.Config,
			role config.Role) (*observability.Telemetry,
			func(context.Context) error,
			error) {
			owner, err := observability.New(ctx, cfg.Observability.TelemetrySettings(), string(role))
			if err != nil {
				return nil, nil, err
			}
			return owner, owner.Shutdown, nil
		},
		database: func(_ context.Context,
			srv *server.Server) (*database.Database,
			func(context.Context) error,
			error) {
			db, err := database.New(srv.Config, srv.Logger, nil, srv.Telemetry)
			if err != nil {
				return nil, nil, err
			}
			return db, func(context.Context) error { return db.Close() }, nil
		},
		redis: func(ctx context.Context,
			srv *server.Server) (*redis.Client,
			func(context.Context) error,
			error) {
			client := redis.NewClient(&redis.Options{Addr: srv.Config.Redis.Address})
			probeCtx, cancel := context.WithTimeout(ctx, srv.Config.ForRole(srv.Role).ReadinessTimeout)
			defer cancel()
			if err := client.Ping(probeCtx).Err(); err != nil {
				return nil, nil, errors.Join(err, client.Close())
			}
			return client, func(context.Context) error { return client.Close() }, nil
		},
		producer: func(srv *server.Server) (*job.JobService, func(context.Context) error, error) {
			producer := job.NewProducer(srv.Logger, srv.Redis, srv.Telemetry)
			return producer, func(context.Context) error { return producer.Stop() }, nil
		},
		email: func(srv *server.Server) (*email.Client,
			error) {
			return email.NewClient(srv.Config,
					srv.Logger),
				nil
		},
		consumer: func(srv *server.Server,
			adapter *email.Client) (*job.JobService,
			func(context.Context) error,
			error) {
			consumer := job.NewConsumer(srv.Logger, srv.Config, srv.Redis, adapter, srv.Telemetry)
			return consumer, func(context.Context) error { return consumer.Stop() }, nil
		},
		startConsumer: func(consumer *job.JobService) error { return consumer.Start() },
		router:        defaultRoleRouter,
		listen: func(address string) (net.Listener, error) {
			var listenerConfig net.ListenConfig
			return listenerConfig.Listen(context.Background(), "tcp", address)
		},
	}
}

//nolint:nonamedreturns // Partial startup cleanup joins errors and clears the returned runtime in a defer.
func newRole(ctx context.Context,
	role config.Role,
	cfg *config.Config,
	factories roleFactories) (runtime *RoleRuntime,
	err error) {
	if cfg == nil || cfg.Observability == nil {
		return nil, &StartupError{Role: role, Stage: "config", cause: errors.New("configuration required")}
	}
	runtime = &RoleRuntime{Role: role, settings: cfg.ForRole(role)}
	runtime.lifecycle.cleanup = &runtime.cleanup
	if runtime.settings.DrainTimeout <= 0 ||
		runtime.settings.ReadinessTimeout <= 0 ||
		runtime.settings.ListenAddress == "" {
		return nil,
			&StartupError{Role: role,
				Stage: "config",
				cause: errors.New("normalized role configuration required")}
	}
	defer func() {
		if err != nil {
			closeCtx, cancel := context.WithTimeout(context.Background(), runtime.settings.DrainTimeout)
			defer cancel()
			err = errors.Join(err, runtime.Close(closeCtx))
			runtime = nil
		}
	}()
	if err = runtime.constructResources(ctx, cfg, factories); err != nil {
		return runtime, err
	}
	if err = runtime.constructHTTP(factories); err != nil {
		return runtime, err
	}
	if role == config.RoleWorker {
		if cause := factories.startConsumer(runtime.Server.Job); cause != nil {
			return runtime, &StartupError{Role: role, Stage: "start consumer", cause: cause}
		}
		if runtime.Server.Job != nil {
			runtime.lifecycle.drain = func(ctx context.Context) error {
				// Management stays reachable and unready while active jobs finish.
				if err148 := runtime.Server.Job.Drain(ctx); err148 != nil {
					return err148
				}
				return runtime.Server.Shutdown(ctx)
			}
		}
	}
	return runtime, nil
}

func (r *RoleRuntime) constructResources(ctx context.Context, cfg *config.Config, f roleFactories) error {
	owner, closeTelemetry, err := f.telemetry(ctx, cfg, r.Role)
	if err160 := r.own("telemetry", closeTelemetry, err); err160 != nil {
		return err160
	}
	if owner == nil {
		return &StartupError{Role: r.Role, Stage: "telemetry", cause: errors.New("provider required")}
	}
	log, closeLogger, cause := f.logger(cfg, owner)
	if err167 := r.own("logger", closeLogger, cause); err167 != nil {
		return err167
	}
	r.Server, err = server.New(cfg, log, nil)
	if err != nil {
		return &StartupError{Role: r.Role, Stage: "server", cause: err}
	}
	r.Server.Role = r.Role
	r.Server.Telemetry = owner
	if r.Role == config.RoleRedirector {
		r.readiness = redirectorReadinessChecks()
	}
	if r.Role == config.RoleAPI {
		db, closeDB, err180 := f.database(ctx, r.Server)
		if err181 := r.own("database", closeDB, err180); err181 != nil {
			return err181
		}
		r.Server.DB = db
		r.readiness = apiReadinessChecks(r.Server, r.settings.ProducerEnabled)
	}
	if r.Role == config.RoleWorker || (r.Role == config.RoleAPI && r.settings.ProducerEnabled) {
		client, closeRedis, err188 := f.redis(ctx, r.Server)
		if err189 := r.own("redis", closeRedis, err188); err189 != nil {
			return err189
		}
		r.Server.Redis = client
	}
	if r.Role == config.RoleAPI && r.settings.ProducerEnabled {
		producer, closeProducer, err195 := f.producer(r.Server)
		if err196 := r.own("producer", closeProducer, err195); err196 != nil {
			return err196
		}
		r.Server.Job = producer
	}
	if r.Role == config.RoleWorker {
		return r.constructWorker(f)
	}
	return nil
}

func (r *RoleRuntime) own(stage string, cleanup func(context.Context) error, cause error) error {
	if cleanup != nil {
		if err := r.cleanup.Push(stage, cleanup); err != nil {
			return err
		}
	}
	if cause != nil {
		return &StartupError{Role: r.Role, Stage: stage, cause: cause}
	}
	return nil
}

func (r *RoleRuntime) constructHTTP(f roleFactories) error {
	httpRouter, err := f.router(r.Role, r.Server)
	if err != nil {
		return &StartupError{Role: r.Role, Stage: "router", cause: err}
	}
	r.HTTP = httpRouter
	health := handler.NewReadinessHandler(r.Server.Logger, r.settings.ReadinessTimeout, r.readiness)
	health.SetReadinessGate(r.lifecycle.Ready)
	router.RegisterHealthRoutes(httpRouter, health)
	r.Server.SetupHTTPServerAt(r.settings.ListenAddress, httpRouter)
	listener, err := f.listen(r.settings.ListenAddress)
	if err != nil {
		return &StartupError{Role: r.Role, Stage: "listener", cause: err}
	}
	r.listener = listener
	if err234 := r.cleanup.Push("listener", func(context.Context) error {
		err235 := listener.Close()
		if errors.Is(err235, net.ErrClosed) {
			return nil
		}
		return err235
	}); err234 != nil {
		return err234
	}
	r.lifecycle.drain = r.Server.Shutdown
	// Also attempt HTTP shutdown after a worker drain failure or partial startup.
	return r.cleanup.Push("http", r.Server.Shutdown)
}

func defaultRoleRouter(role config.Role, srv *server.Server) (*echo.Echo, error) {
	if role != config.RoleAPI {
		e := echo.New()
		global := middleware.NewGlobalMiddlewares(srv)
		e.HTTPErrorHandler = global.GlobalErrorHandler
		e.Use(middleware.RequestID(), middleware.NewTracingMiddleware(srv, srv.Telemetry).EnhanceTracing(),
			middleware.NewContextEnhancer(srv).EnhanceContext(), global.RequestLogger(), global.Recover())
		return e, nil
	}
	services := &service.Services{Job: srv.Job, Auth: service.NewAuthService(srv)}
	return router.NewRouter(srv, &handler.Handlers{OpenAPI: handler.NewOpenAPIHandler(srv)}, services), nil
}

func apiReadinessChecks(srv *server.Server, producerEnabled bool) []handler.ReadinessCheck {
	checks := []handler.ReadinessCheck{{Name: "database", Check: func(ctx context.Context) error {
		if srv.DB == nil || srv.DB.Pool == nil {
			return errors.New("database unconfigured")
		}
		return srv.DB.Pool.Ping(ctx)
	}}}
	if producerEnabled {
		checks = append(checks, queueReadinessCheck(srv))
	}
	return checks
}

func queueReadinessCheck(srv *server.Server) handler.ReadinessCheck {
	return handler.ReadinessCheck{Name: "redis", Check: func(ctx context.Context) error {
		if srv.Redis == nil {
			return errors.New("queue unconfigured")
		}
		return srv.Redis.Ping(ctx).Err()
	}}
}

// Run serves the role transport until cancellation or a serving error.
func (r *RoleRuntime) Run(ctx context.Context) error {
	result := make(chan error, 1)
	go func() { result <- r.Server.Serve(r.listener) }()
	var serveErr error
	select {
	case serveErr = <-result:
	case <-ctx.Done():
	}
	closeCtx, cancel := context.WithTimeout(context.Background(), r.settings.DrainTimeout)
	defer cancel()
	closeErr := r.Close(closeCtx)
	if errors.Is(serveErr, http.ErrServerClosed) || errors.Is(serveErr, net.ErrClosed) {
		serveErr = nil
	}
	return errors.Join(serveErr, closeErr)
}

// Close drains HTTP, then closes dependents and shared resources exactly once.
func (r *RoleRuntime) Close(ctx context.Context) error { return r.lifecycle.Shutdown(ctx) }

// Address exposes the bound address, including automatically assigned test ports.
func (r *RoleRuntime) Address() string { return r.listener.Addr().String() }
