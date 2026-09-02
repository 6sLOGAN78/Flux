package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/lib/utils"

	"github.com/rs/zerolog/log"
)

// WebhookRetryWorker manages the polling and execution of dead-letter/retry queues for webhook deliveries.
type WebhookRetryWorker struct {
	repo         *repository.WebhookRepository
	httpClient   *http.Client
	config       *config.WebhookConfig
	jobChan      chan repository.WebhookDelivery
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewWebhookRetryWorker(repo *repository.WebhookRepository, cfg *config.WebhookConfig) *WebhookRetryWorker {
	ctx, cancel := context.WithCancel(context.Background())
	deliveryTimeout, _ := time.ParseDuration(cfg.DeliveryTimeout)
	if deliveryTimeout == 0 {
		deliveryTimeout = 10 * time.Second
	}
	
	concurrency := cfg.RetryConcurrency
	if concurrency <= 0 {
		concurrency = 5
	}
	
	return &WebhookRetryWorker{
		repo:       repo,
		httpClient: utils.SafeHTTPClient(deliveryTimeout),
		config:     cfg,
		jobChan:    make(chan repository.WebhookDelivery, concurrency*2),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start launches the background scheduler and execution pool.
func (w *WebhookRetryWorker) Start() {
	for i := 0; i < w.config.RetryConcurrency; i++ {
		w.wg.Add(1)
		go w.deliveryLoop()
	}

	w.wg.Add(1)
	go w.schedulerLoop()
}

// Stop gracefully stops the retry scheduler and allows executing retries to drain.
func (w *WebhookRetryWorker) Stop(timeout time.Duration) {
	w.cancel()
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("WebhookRetryWorker shut down gracefully")
	case <-time.After(timeout):
		log.Warn().Msg("WebhookRetryWorker shutdown timed out")
	}
}

func (w *WebhookRetryWorker) schedulerLoop() {
	defer w.wg.Done()
	defer close(w.jobChan)

	pollInterval, _ := time.ParseDuration(w.config.RetryPollInterval)
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
		}

		// Pull batch equal to concurrency to avoid pulling more than we can process.
		deliveries, err := w.repo.ClaimDueRetries(w.ctx, w.config.RetryConcurrency)
		if err != nil {
			log.Error().Err(err).Msg("failed to claim due webhook retries")
			continue
		}

		for _, d := range deliveries {
			select {
			case w.jobChan <- d:
			case <-w.ctx.Done():
				return
			}
		}
	}
}

func (w *WebhookRetryWorker) deliveryLoop() {
	defer w.wg.Done()

	for delivery := range w.jobChan {
		w.processRetry(delivery)
	}
}

func (w *WebhookRetryWorker) processRetry(d repository.WebhookDelivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // isolated background context
	defer cancel()

	// 1. Verify webhook still active
	wh, err := w.repo.GetWebhookByID(ctx, d.WebhookID)
	if err != nil || wh == nil {
		if err == repository.ErrNotFound {
			// Webhook deleted. Fail the delivery permanently.
			msg := "webhook deleted"
			w.repo.UpdateDeliveryState(ctx, d.ID, "dead_letter", nil, d.AttemptCount, &msg, nil)
			return
		}
		// DB failure, revert to retrying
		w.repo.UpdateDeliveryState(ctx, d.ID, "retrying", nil, d.AttemptCount, nil, d.NextAttemptAt)
		return
	}

	if !wh.Active {
		msg := "webhook deactivated"
		w.repo.UpdateDeliveryState(ctx, d.ID, "dead_letter", nil, d.AttemptCount, &msg, nil)
		return
	}

	// 2. Perform HTTP attempt
	req, err := http.NewRequestWithContext(ctx, "POST", wh.EndpointURL, bytes.NewReader(d.Payload))
	if err != nil {
		w.handleOutcome(ctx, d, nil, err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Regenerate signature natively in case the secret was rotated
	signature := utils.GenerateHMACSHA256(wh.Secret, d.Payload)
	req.Header.Set("X-Flux-Signature", "v1="+signature)

	resp, err := w.httpClient.Do(req)
	var statusCode *int
	if resp != nil {
		code := resp.StatusCode
		statusCode = &code
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	w.handleOutcome(ctx, d, statusCode, err)
}

func (w *WebhookRetryWorker) handleOutcome(ctx context.Context, d repository.WebhookDelivery, statusCode *int, err error) {
	var errStr *string
	if err != nil {
		msg := err.Error()
		errStr = &msg
	}

	isRetryable := true
	if err != nil || (statusCode != nil && (*statusCode < 200 || *statusCode >= 300)) {
		if statusCode != nil {
			isRetryable = IsRetryableError(*statusCode, err)
		} else {
			isRetryable = IsRetryableError(0, err)
		}
	} else {
		// Success
		if err := w.repo.UpdateDeliveryState(ctx, d.ID, "success", statusCode, d.AttemptCount, nil, nil); err != nil {
			log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("failed to update webhook delivery to success")
		}
		return
	}

	// Failure case
	newAttemptCount := d.AttemptCount + 1
	if !isRetryable || newAttemptCount >= w.config.MaxRetries {
		// Dead Letter
		if err := w.repo.UpdateDeliveryState(ctx, d.ID, "dead_letter", statusCode, newAttemptCount, errStr, nil); err != nil {
			log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("failed to move webhook delivery to dead_letter")
		}
		return
	}

	// Calculate backoff
	initialDelay, _ := time.ParseDuration(w.config.RetryInitialDelay)
	maxDelay, _ := time.ParseDuration(w.config.RetryMaxDelay)
	if initialDelay == 0 {
		initialDelay = 5 * time.Second
	}
	if maxDelay == 0 {
		maxDelay = 1 * time.Hour
	}

	delay := CalculateRetryDelay(newAttemptCount, initialDelay, maxDelay)
	nextAttemptAt := time.Now().Add(delay)

	if err := w.repo.UpdateDeliveryState(ctx, d.ID, "retrying", statusCode, newAttemptCount, errStr, &nextAttemptAt); err != nil {
		log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("failed to schedule webhook retry")
	}
}

// SetHTTPClient allows overriding the default SSRF-safe client for tests.
func (w *WebhookRetryWorker) SetHTTPClient(client *http.Client) {
	w.httpClient = client
}
