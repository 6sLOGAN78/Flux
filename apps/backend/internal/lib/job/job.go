package job

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lifecycle"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type JobService struct {
	Client      *asynq.Client
	server      *asynq.Server
	logger      *zerolog.Logger
	emailClient *email.Client
	shutdown    func() error
	closeClient func() error
	stopOnce    sync.Once
	stopErr     error
	drainErr    error
	drainOnce   sync.Once
	drainDone   chan struct{}
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
func NewProducer(logger *zerolog.Logger, client *redis.Client) *JobService {
	return &JobService{Client: asynq.NewClientFromRedisClient(client), logger: logger}
}

// NewConsumer owns only worker processing over the role's shared Redis client.
// Sharing the connection lets partial startup close it even before Start runs.
func NewConsumer(logger *zerolog.Logger, cfg *config.Config, client *redis.Client, emailClient *email.Client) *JobService {
	server := asynq.NewServerFromRedisClient(client, asynq.Config{
		Concurrency:     10,
		Queues:          map[string]int{"critical": 6, "default": 3, "low": 1},
		ShutdownTimeout: cfg.Worker.DrainTimeout,
		// Asynq Stop waits for its idle poll sleep (up to 1.5 intervals).
		// Keep that intake delay inside short configured shutdown budgets.
		TaskCheckInterval: min(time.Second, cfg.Worker.DrainTimeout/10),
	})
	return &JobService{server: server, logger: logger, emailClient: emailClient, shutdown: func() error { server.Shutdown(); return nil }}
}

func NewJobService(logger *zerolog.Logger, cfg *config.Config) *JobService {
	redisAddr := cfg.Redis.Address

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // Higher priority queue for important emails
				"default":  3, // Default priority for most emails
				"low":      1, // Lower priority for non-urgent emails
			},
		},
	)

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
		closeClient: client.Close,
	}
}

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
