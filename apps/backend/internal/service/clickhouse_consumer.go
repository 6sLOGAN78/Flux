package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"flux/apps/backend/internal/model/analytics"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RedisAnalyticsConsumer reads events from Redis Stream via Consumer Group
// and batches them into ClickHouse securely.
type RedisAnalyticsConsumer struct {
	redisClient   *redis.Client
	chConn        driver.Conn
	streamName    string
	groupName     string
	consumerName  string
	batchSize     int
	batchWait     time.Duration
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewRedisAnalyticsConsumer initializes the ClickHouse ingestion pipeline worker.
func NewRedisAnalyticsConsumer(redisClient *redis.Client, chConn driver.Conn, streamName string) *RedisAnalyticsConsumer {
	if streamName == "" {
		streamName = "analytics:events"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisAnalyticsConsumer{
		redisClient:   redisClient,
		chConn:        chConn,
		streamName:    streamName,
		groupName:     "analytics-clickhouse",
		consumerName:  fmt.Sprintf("backend-consumer-%s", uuid.New().String()),
		batchSize:     1000,
		batchWait:     1 * time.Second,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start ensures the consumer group exists and launches the ingestion loops.
func (c *RedisAnalyticsConsumer) Start() {
	// 1. Ensure Consumer Group exists
	err := c.redisClient.XGroupCreateMkStream(context.Background(), c.streamName, c.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Error().Err(err).Msg("failed to create redis consumer group for clickhouse ingestion")
	}

	c.wg.Add(2)
	go c.readLoop()
	go c.recoveryLoop()
}

// Stop initiates graceful shutdown.
func (c *RedisAnalyticsConsumer) Stop(timeout time.Duration) {
	c.cancel()
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("RedisAnalyticsConsumer shut down gracefully")
	case <-time.After(timeout):
		log.Warn().Msg("RedisAnalyticsConsumer shutdown timed out")
	}
}

// readLoop reads new messages blocking until batch size or timeout is reached.
func (c *RedisAnalyticsConsumer) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Read a batch using XREADGROUP. Block for batchWait.
		args := &redis.XReadGroupArgs{
			Group:    c.groupName,
			Consumer: c.consumerName,
			Streams:  []string{c.streamName, ">"},
			Count:    int64(c.batchSize),
			Block:    c.batchWait,
		}

		streams, err := c.redisClient.XReadGroup(c.ctx, args).Result()
		if err != nil {
			if err == redis.Nil || err == context.Canceled {
				continue
			}
			log.Warn().Err(err).Msg("redis consumer XREADGROUP error, retrying in 2s")
			time.Sleep(2 * time.Second)
			continue
		}

		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			c.processMessages(streams[0].Messages)
		}
	}
}

// recoveryLoop periodically reclaims and processes stuck pending messages from crashed consumers.
func (c *RedisAnalyticsConsumer) recoveryLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
		}

		// AutoClaim messages pending for more than 60 seconds
		args := &redis.XAutoClaimArgs{
			Stream:   c.streamName,
			Group:    c.groupName,
			Consumer: c.consumerName,
			MinIdle:  60 * time.Second,
			Start:    "0-0",
			Count:    100,
		}

		messages, _, err := c.redisClient.XAutoClaim(c.ctx, args).Result()
		if err != nil && err != redis.Nil {
			log.Warn().Err(err).Msg("XAUTOCLAIM error")
			continue
		}

		if len(messages) > 0 {
			log.Info().Msgf("Recovered %d pending analytics events", len(messages))
			c.processMessages(messages)
		}
	}
}

func (c *RedisAnalyticsConsumer) processMessages(messages []redis.XMessage) {
	if len(messages) == 0 {
		return
	}

	var events []analytics.AnalyticsEvent
	var validMessageIDs []string

	for _, msg := range messages {
		payloadStr, ok := msg.Values["payload"].(string)
		if !ok {
			log.Warn().Msgf("malformed redis message %s: payload not string", msg.ID)
			// ACK immediately so it doesn't block forever
			c.redisClient.XAck(c.ctx, c.streamName, c.groupName, msg.ID)
			continue
		}

		var event analytics.AnalyticsEvent
		if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
			log.Warn().Err(err).Msgf("malformed redis message JSON %s", msg.ID)
			c.redisClient.XAck(c.ctx, c.streamName, c.groupName, msg.ID)
			continue
		}

		events = append(events, event)
		validMessageIDs = append(validMessageIDs, msg.ID)
	}

	if len(events) == 0 {
		return
	}

	// Insert Batch into ClickHouse
	if err := c.insertToClickHouse(events); err != nil {
		log.Error().Err(err).Msg("failed to insert batch into ClickHouse. Will NOT ACK, allowing retry.")
		return // DO NOT ACK. They will become pending and be recovered by XAutoClaim later.
	}

	// Insert Succeeded. ACK all valid messages.
	if len(validMessageIDs) > 0 {
		err := c.redisClient.XAck(c.ctx, c.streamName, c.groupName, validMessageIDs...).Err()
		if err != nil {
			log.Error().Err(err).Msg("failed to ACK messages in redis after successful clickhouse insert")
			// Even if ACK fails, ClickHouse insert succeeded. 
			// Duplicate processing on retry is mitigated by ClickHouse replacing/query-time deduplication strategy.
		}
	}
}

func (c *RedisAnalyticsConsumer) insertToClickHouse(events []analytics.AnalyticsEvent) error {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	batch, err := c.chConn.PrepareBatch(ctx, "INSERT INTO analytics_events")
	if err != nil {
		return err
	}

	for _, e := range events {
		err := batch.Append(
			e.EventID,
			string(e.EventType),
			e.Timestamp,
			e.LinkID,
			e.WorkspaceID,
			e.ShortCode,
			e.Referrer,
			e.UserAgent,
			e.IPHash,
			e.CampaignID,
			e.UTMSource,
			e.UTMMedium,
			e.UTMCampaign,
			e.UTMTerm,
			e.UTMContent,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}
