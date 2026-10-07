package job

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lifecycle"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	workerConcurrency   = 10
	criticalQueueWeight = 6
	defaultQueueWeight  = 3
	drainPollDivisor    = 10
)

// JobService owns asynchronous producers and worker resources.
//
//nolint:revive // Preserve the established service type used by constructors and role adapters.
type JobService struct {
	emailClient  WelcomeEmailSender
	stopErr      error
	drainErr     error
	tracer       trace.Tracer
	replacements *asynq.Client
	server       *asynq.Server
	logger       *zerolog.Logger
	shutdown     func() error
	closeClient  func() error
	Client       *asynq.Client
	inspector    *asynq.Inspector
	drainDone    chan struct{}
	stopOnce     sync.Once
	drainOnce    sync.Once
}

// WelcomeEmailSender is the worker's narrow, injectable delivery boundary.
type WelcomeEmailSender interface {
	SendWelcomeEmail(to, firstName string) error
}

// StopIntake prevents new claims without closing the shared Redis connection.
func (j *JobService) StopIntake() {
	if j.server != nil {
		j.server.Stop()
	}
}

// Drain joins one consumer shutdown within the caller's overall deadline.
func (j *JobService) Drain(ctx context.Context) error {
	j.drainOnce.Do(func() {
		j.drainDone = make(chan struct{})
		go func() {
			if j.shutdown != nil {
				j.drainErr = j.shutdown()
			}
			close(j.drainDone)
		}()
	})
	select {
	case <-j.drainDone:
		return j.drainErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NewProducer allocates only an enqueue client; it never constructs a consumer
// or email transport. Its shared Redis connection is owned by the role.
func NewProducer(logger *zerolog.Logger, client *redis.Client, telemetry ...*observability.Telemetry) *JobService {
	return &JobService{Client: asynq.NewClientFromRedisClient(client), logger: logger, tracer: jobTracer(telemetry)}
}

func jobTracer(telemetry []*observability.Telemetry) trace.Tracer {
	if len(telemetry) > 0 && telemetry[0] != nil && telemetry[0].Tracer != nil {
		return telemetry[0].Tracer
	}
	return noop.NewTracerProvider().Tracer("flux.jobs")
}

// NewConsumer owns only worker processing over the role's shared Redis client.
// Sharing the connection lets partial startup close it even before Start runs.
func NewConsumer(logger *zerolog.Logger,
	cfg *config.Config,
	client *redis.Client,
	emailClient WelcomeEmailSender,
	telemetry ...*observability.Telemetry) *JobService {
	var tel *observability.Telemetry
	if len(telemetry) > 0 {
		tel = telemetry[0]
	}
	return newConsumer(logger, client, emailClient, tel, asynq.Config{
		Concurrency: workerConcurrency,
		Queues: map[string]int{"critical": criticalQueueWeight,
			"default": defaultQueueWeight,
			"low":     1},
		ShutdownTimeout: cfg.Worker.DrainTimeout,
		// Asynq Stop waits for its idle poll sleep (up to 1.5 intervals).
		// Keep that intake delay inside short configured shutdown budgets.
		TaskCheckInterval: min(time.Second, cfg.Worker.DrainTimeout/drainPollDivisor),
	})
}

func newConsumer(logger *zerolog.Logger,
	client *redis.Client,
	emailClient WelcomeEmailSender,
	tel *observability.Telemetry,
	cfg asynq.Config) *JobService {
	cfg.Logger = safeJobLogger{log: logger}
	server := asynq.NewServerFromRedisClient(client, cfg)
	return &JobService{server: server,
		logger:       logger,
		emailClient:  emailClient,
		tracer:       jobTracer([]*observability.Telemetry{tel}),
		replacements: asynq.NewClientFromRedisClient(client),
		inspector:    asynq.NewInspectorFromRedisClient(client),
		shutdown:     func() error { server.Shutdown(); return nil }}
}

// NewJobService constructs a compatibility job producer.
func NewJobService(logger *zerolog.Logger, cfg *config.Config) *JobService {
	redisAddr := cfg.Redis.Address

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Logger:      safeJobLogger{log: logger},
			Concurrency: workerConcurrency,
			Queues: map[string]int{
				"critical": criticalQueueWeight, // Higher priority queue for important emails
				"default":  defaultQueueWeight,  // Default priority for most emails
				"low":      1,                   // Lower priority for non-urgent emails
			},
		},
	)
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})

	return &JobService{
		Client:      client,
		server:      server,
		logger:      logger,
		emailClient: email.NewClient(cfg, logger),
		shutdown: func() error {
			// Asynq's Shutdown has no error return; the seam also allows
			// deterministic failure tests without starting a worker process.
			server.Shutdown()
			return nil
		},
		closeClient:  func() error { return errors.Join(client.Close(), inspector.Close()) },
		tracer:       jobTracer(nil),
		replacements: client,
		inspector:    inspector,
	}
}

// Asynq's operational messages can embed task/provider values. Do not format them.
type safeJobLogger struct{ log *zerolog.Logger }

func (l safeJobLogger) Debug(...interface{}) { l.log.Debug().Msg("job.process") }
func (l safeJobLogger) Info(...interface{})  { l.log.Info().Msg("job.process") }
func (l safeJobLogger) Warn(...interface{})  { l.log.Warn().Msg("job.process") }
func (l safeJobLogger) Error(...interface{}) { l.log.Error().Msg("job.process") }
func (l safeJobLogger) Fatal(...interface{}) { l.log.Error().Msg("job.process") }

// Start starts the configured worker consumer.
func (j *JobService) Start() error {
	if j.server == nil {
		return errors.New("job consumer not configured")
	}
	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)

	j.logger.Info().Msg("Starting background job server")
	if err := j.server.Start(mux); err != nil {
		return err
	}

	return nil
}

// Stop drains and closes all owned job resources.
func (j *JobService) Stop() error {
	j.stopOnce.Do(func() {
		j.logger.Info().Msg("Stopping background job server")
		var cleanup lifecycle.Cleanup
		// Producer is allocated first and outlives its dependent consumer.
		if j.closeClient != nil {
			_ = cleanup.Push("job producer", func(context.Context) error { return j.closeClient() })
		}
		if j.shutdown != nil {
			_ = cleanup.Push("job consumer", j.Drain)
		}
		j.stopErr = cleanup.Close(context.Background())
	})
	return j.stopErr
}
