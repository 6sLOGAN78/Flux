package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/webhook"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/lib/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type WebhookJob struct {
	EventID   string
	Webhook   webhook.Webhook
	Payload   []byte
	Signature string
}

type WebhookWorker struct {
	redisClient   *redis.Client
	repo          *repository.WebhookRepository
	httpClient    *http.Client
	cfg           *config.WebhookConfig
	streamName    string
	groupName     string
	consumerName  string
	jobChan       chan WebhookJob
	wg            sync.WaitGroup
	producerWg    sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}
func NewWebhookWorker(redisClient *redis.Client, repo *repository.WebhookRepository, streamName string, cfg *config.WebhookConfig) *WebhookWorker {
	if streamName == "" {
		streamName = "analytics:events"
	}
	timeout, _ := time.ParseDuration(cfg.DeliveryTimeout)
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &WebhookWorker{
		redisClient:  redisClient,
		repo:         repo,
		httpClient:   utils.SafeHTTPClient(timeout),
		cfg:          cfg,
		streamName:   streamName,
		groupName:    "analytics-webhooks",
		consumerName: fmt.Sprintf("webhook-consumer-%s", uuid.New().String()),
		jobChan:      make(chan WebhookJob, cfg.WorkerConcurrency*2),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// SetHTTPClient allows overriding the default SSRF-safe client for tests.
func (w *WebhookWorker) SetHTTPClient(client *http.Client) {
	w.httpClient = client
}

func (w *WebhookWorker) Start() {
	err := w.redisClient.XGroupCreateMkStream(context.Background(), w.streamName, w.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Error().Err(err).Msg("failed to create redis consumer group for webhook delivery")
	}

	for i := 0; i < w.cfg.WorkerConcurrency; i++ {
		w.wg.Add(1)
		go w.deliveryLoop()
	}

	w.producerWg.Add(2)
	go func() {
		w.producerWg.Wait()
		close(w.jobChan)
	}()
	go w.readLoop()
	go w.recoveryLoop()
}

func (w *WebhookWorker) Stop(timeout time.Duration) {
	w.cancel()

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("WebhookWorker shut down gracefully")
	case <-time.After(timeout):
		log.Warn().Msg("WebhookWorker shutdown timed out")
	}
}

func (w *WebhookWorker) readLoop() {
	defer w.producerWg.Done()
	

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		args := &redis.XReadGroupArgs{
			Group:    w.groupName,
			Consumer: w.consumerName,
			Streams:  []string{w.streamName, ">"},
			Count:    100,
			Block:    1 * time.Second,
		}

		streams, err := w.redisClient.XReadGroup(w.ctx, args).Result()
		if err != nil {
			if err == redis.Nil || err == context.Canceled {
				continue
			}
			log.Warn().Err(err).Msg("webhook redis consumer XREADGROUP error")
			time.Sleep(2 * time.Second)
			continue
		}

		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			w.processMessages(streams[0].Messages)
		}
	}
}

func (w *WebhookWorker) recoveryLoop() {
	defer w.producerWg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
		}

		args := &redis.XAutoClaimArgs{
			Stream:   w.streamName,
			Group:    w.groupName,
			Consumer: w.consumerName,
			MinIdle:  60 * time.Second,
			Start:    "0-0",
			Count:    50,
		}

		messages, _, err := w.redisClient.XAutoClaim(w.ctx, args).Result()
		if err != nil && err != redis.Nil {
			log.Warn().Err(err).Msg("webhook consumer XAUTOCLAIM error")
			continue
		}

		if len(messages) > 0 {
			w.processMessages(messages)
		}
	}
}

func (w *WebhookWorker) processMessages(messages []redis.XMessage) {
	var ackIDs []string

	for _, msg := range messages {
		payloadStr, ok := msg.Values["payload"].(string)
		if !ok {
			ackIDs = append(ackIDs, msg.ID)
			continue
		}

		var event analytics.AnalyticsEvent
		if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
			ackIDs = append(ackIDs, msg.ID)
			continue
		}

		w.dispatchToWebhooks(event)
		ackIDs = append(ackIDs, msg.ID)
	}

	if len(ackIDs) > 0 {
		w.redisClient.XAck(w.ctx, w.streamName, w.groupName, ackIDs...)
	}
}

func (w *WebhookWorker) dispatchToWebhooks(event analytics.AnalyticsEvent) {
	if event.WorkspaceID == "" {
		return
	}
	wsID, err := uuid.Parse(event.WorkspaceID)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	activeWebhooks, err := w.repo.GetActiveWebhooksForWorkspace(ctx, wsID)
	if err != nil {
		log.Error().Err(err).Str("workspace_id", event.WorkspaceID).Msg("failed to get webhooks for dispatch")
		return
	}

	for _, wh := range activeWebhooks {
		// Event matching
		if !contains(wh.Events, string(event.EventType)) {
			continue
		}

		// Construct stable payload
		payload := webhook.WebhookEventPayload{
			ID:        event.EventID,
			Type:      string(event.EventType),
			CreatedAt: event.Timestamp,
			Data:      event,
		}

		rawBody, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal webhook payload")
			continue
		}

		// Sign payload
		signature := utils.GenerateHMACSHA256(wh.Secret, rawBody)

		select {
		case w.jobChan <- WebhookJob{
			EventID:   event.EventID,
			Webhook:   wh,
			Payload:   rawBody,
			Signature: signature,
		}:
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *WebhookWorker) deliveryLoop() {
	defer w.wg.Done()
	

	for job := range w.jobChan {
		w.deliver(job)
	}
}

func (w *WebhookWorker) deliver(job WebhookJob) {
	req, err := http.NewRequestWithContext(w.ctx, "POST", job.Webhook.EndpointURL, bytes.NewReader(job.Payload))
	if err != nil {
		w.handleOutcome(job, nil, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Flux-Signature", "v1="+job.Signature)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.handleOutcome(job, nil, err)
		return
	}
	defer resp.Body.Close()
	// Drain body
	io.Copy(io.Discard, resp.Body)

	code := resp.StatusCode
	w.handleOutcome(job, &code, nil)
}

func (w *WebhookWorker) handleOutcome(job WebhookJob, statusCode *int, err error) {
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
		w.recordDelivery(job, "success", statusCode, errStr, nil)
		return
	}

	// Initial attempt failed.
	if !isRetryable || w.cfg.MaxRetries <= 1 {
		w.recordDelivery(job, "dead_letter", statusCode, errStr, nil)
		return
	}

	initialDelay, _ := time.ParseDuration(w.cfg.RetryInitialDelay)
	maxDelay, _ := time.ParseDuration(w.cfg.RetryMaxDelay)
	if initialDelay == 0 { initialDelay = 5 * time.Second }
	if maxDelay == 0 { maxDelay = 1 * time.Hour }
	
	delay := CalculateRetryDelay(1, initialDelay, maxDelay)
	nextAttemptAt := time.Now().Add(delay)

	w.recordDelivery(job, "retrying", statusCode, errStr, &nextAttemptAt)
}

func (w *WebhookWorker) recordDelivery(job WebhookJob, status string, statusCode *int, errStr *string, nextAttemptAt *time.Time) {
	delivery := &repository.WebhookDelivery{
		WebhookID:      job.Webhook.ID,
		EventID:        job.EventID,
		Status:         status,
		ResponseStatus: statusCode,
		AttemptCount:   1, // 15A-03 will handle > 1
		LastError:      errStr,
		Payload:        job.Payload,
		NextAttemptAt:  nextAttemptAt,
	}

	// Use background context for recording to avoid canceling record on shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := w.repo.RecordDelivery(ctx, delivery); err != nil {
		log.Error().Err(err).Str("webhook_id", job.Webhook.ID.String()).Msg("failed to record webhook delivery")
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
