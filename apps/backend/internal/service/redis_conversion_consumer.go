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

type RedisConversionConsumer struct {
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

func NewRedisConversionConsumer(redisClient *redis.Client, chConn driver.Conn, streamName string) *RedisConversionConsumer {
	if streamName == "" {
		streamName = "analytics:conversions"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisConversionConsumer{
		redisClient:   redisClient,
		chConn:        chConn,
		streamName:    streamName,
		groupName:     "conversions-clickhouse",
		consumerName:  fmt.Sprintf("conversions-consumer-%s", uuid.New().String()),
		batchSize:     1000,
		batchWait:     1 * time.Second,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (c *RedisConversionConsumer) Start() {
	err := c.redisClient.XGroupCreateMkStream(context.Background(), c.streamName, c.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Error().Err(err).Msg("failed to create redis consumer group for conversions")
	}

	c.wg.Add(2)
	go c.readLoop()
	go c.recoveryLoop()
}

func (c *RedisConversionConsumer) Stop(timeout time.Duration) {
	c.cancel()
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("RedisConversionConsumer shut down gracefully")
	case <-time.After(timeout):
		log.Warn().Msg("RedisConversionConsumer shutdown timed out")
	}
}

func (c *RedisConversionConsumer) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

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
			log.Warn().Err(err).Msg("conversions consumer XREADGROUP error, retrying in 2s")
			time.Sleep(2 * time.Second)
			continue
		}

		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			c.processMessages(streams[0].Messages)
		}
	}
}

func (c *RedisConversionConsumer) recoveryLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
		}

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
			log.Warn().Err(err).Msg("conversions XAUTOCLAIM error")
			continue
		}

		if len(messages) > 0 {
			log.Info().Msgf("Recovered %d pending conversions", len(messages))
			c.processMessages(messages)
		}
	}
}

func (c *RedisConversionConsumer) processMessages(messages []redis.XMessage) {
	if len(messages) == 0 {
		return
	}

	var events []analytics.ConversionEvent
	var validMessageIDs []string

	for _, msg := range messages {
		payloadStr, ok := msg.Values["payload"].(string)
		if !ok {
			log.Warn().Msgf("malformed redis conversions message %s", msg.ID)
			c.redisClient.XAck(c.ctx, c.streamName, c.groupName, msg.ID)
			continue
		}

		var event analytics.ConversionEvent
		if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
			log.Warn().Err(err).Msgf("malformed conversions JSON %s", msg.ID)
			c.redisClient.XAck(c.ctx, c.streamName, c.groupName, msg.ID)
			continue
		}

		events = append(events, event)
		validMessageIDs = append(validMessageIDs, msg.ID)
	}

	if len(events) == 0 {
		return
	}

	if err := c.insertToClickHouse(events); err != nil {
		log.Error().Err(err).Msg("failed to insert conversions batch. Will retry.")
		return
	}

	if len(validMessageIDs) > 0 {
		c.redisClient.XAck(c.ctx, c.streamName, c.groupName, validMessageIDs...)
	}
}

func (c *RedisConversionConsumer) insertToClickHouse(events []analytics.ConversionEvent) error {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	batch, err := c.chConn.PrepareBatch(ctx, "INSERT INTO conversions")
	if err != nil {
		return err
	}

	for _, e := range events {
		err := batch.Append(
			e.ConversionID,
			e.WorkspaceID,
			e.Timestamp,
			e.ConversionName,
			e.Revenue,
			e.Currency,
			e.ClickIDs,
			e.VisitorID,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}
