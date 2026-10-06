package job

import (
	"context"
	"sync"

	"github.com/6sLOGAN78/flux/internal/app"
	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/hibiken/asynq"
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
		var cleanup app.Cleanup
		// Producer is allocated first and outlives its dependent consumer.
		_ = cleanup.Push("job producer", func(context.Context) error { return j.closeClient() })
		_ = cleanup.Push("job consumer", func(context.Context) error { return j.shutdown() })
		j.stopErr = cleanup.Close(context.Background())
	})
	return j.stopErr
}
