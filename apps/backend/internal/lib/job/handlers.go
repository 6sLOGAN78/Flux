// Package job owns asynchronous email producers, consumers, and graceful drain.
package job

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	loggerpkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const stageEnqueue = "enqueue"

// InitHandlers registers the service-owned job handlers.
func (j *JobService) InitHandlers(config *config.Config, logger *zerolog.Logger) {
	j.emailClient = email.NewClient(config, logger)
}

func (j *JobService) handleWelcomeEmailTask(ctx context.Context, t *asynq.Task) (err error) {
	defer func() {
		if recover() != nil {
			err = jobError{stage: "process", cause: errors.New("job processing failed")}
		}
	}()
	var p WelcomeEmailPayload
	if err = json.Unmarshal(t.Payload(), &p); err != nil {
		return jobError{stage: "validate", cause: err}
	}
	p.Metadata = cleanMetadata(p.Metadata)
	canonical, marshalErr := json.Marshal(p)
	if marshalErr != nil {
		return jobError{stage: "validate", cause: marshalErr}
	}
	var input, clean any
	_ = json.Unmarshal(t.Payload(), &input)
	_ = json.Unmarshal(canonical, &clean)
	if !reflect.DeepEqual(input, clean) || len(t.Headers()) > 0 {
		return j.replaceLegacy(ctx, t, canonical)
	}
	carrier := propagation.MapCarrier{}
	if p.Metadata != nil {
		carrier["traceparent"] = p.Metadata.Traceparent
	}
	ctx = observability.TraceparentPropagator{}.Extract(ctx, carrier)
	if p.Metadata != nil {
		ctx = observability.WithCorrelation(ctx, p.Metadata.RequestID, p.Metadata.CorrelationID)
	} else {
		ctx = observability.WithCorrelation(ctx, "", "")
	}
	parent := observability.CleanSpanContext(trace.SpanContextFromContext(ctx))
	var links []trace.Link
	if parent.IsValid() {
		links = []trace.Link{{SpanContext: parent}}
	}
	ids := observability.CorrelationFromContext(ctx)
	retry, _ := asynq.GetRetryCount(ctx)
	tracer := j.tracer
	if tracer == nil {
		tracer = jobTracer(nil)
	}
	ctx,
		span := tracer.Start(ctx,
		"job.process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithLinks(links...),
		trace.WithAttributes(attribute.String("job.type",
			TaskWelcome),
			attribute.String("operation",
				"job.process"),
			attribute.String("dependency",
				"email"),
			attribute.String("request_id",
				ids.RequestID),
			attribute.String("correlation_id",
				ids.CorrelationID),
			attribute.Int("retry.count",
				retry)))
	defer span.End()
	log := loggerpkg.WithContext(*j.logger, ctx)
	log.Info().Str("job.type", TaskWelcome).Int("retry.count", retry).Msg("job.process")
	if err = j.emailClient.SendWelcomeEmail(p.To, p.FirstName); err != nil {
		span.SetStatus(codes.Error, "")
		span.SetAttributes(attribute.String("outcome", "error"), attribute.String("error.category", "unknown"))
		log.Error().
			Str("job.type",
				TaskWelcome).
			Str("outcome",
				"error").
			Str("error.category",
				"unknown").
			Msg("job.process")
		return jobError{stage: "send", cause: err}
	}
	span.SetAttributes(attribute.String("outcome", "success"))
	log.Info().Str("job.type", TaskWelcome).Str("outcome", "success").Msg("job.process")
	return nil
}

// jobError keeps the cause available without Asynq persisting private error text.
type jobError struct {
	cause error
	stage string
}

func (e jobError) Error() string { return "job " + e.stage + ": " + observability.SafeError(e.cause) }
func (e jobError) Unwrap() error { return e.cause }

// Asynq v0.26 shares Payload/Headers with its retry message; cancellation can
// persist that message concurrently. Never mutate those shared values. Migrate
// through public enqueue/revoke APIs, preserving the remaining retry budget.
func (j *JobService) replaceLegacy(ctx context.Context, t *asynq.Task, payload []byte) error {
	if j.replacements == nil || j.inspector == nil {
		return jobError{stage: stageEnqueue, cause: errors.New("job migration unavailable")}
	}
	id, ok := asynq.GetTaskID(ctx)
	if !ok {
		return jobError{stage: stageEnqueue, cause: errors.New("task identity unavailable")}
	}
	queue, _ := asynq.GetQueueName(ctx)
	info, err := j.inspector.GetTaskInfo(queue, id)
	if err != nil {
		return jobError{stage: stageEnqueue, cause: err}
	}
	opts := []asynq.Option{asynq.TaskID("safe:" + id),
		asynq.Queue(queue),
		asynq.MaxRetry(max(0,
			info.MaxRetry-info.Retried)),
		asynq.Retention(info.Retention)}
	if info.Timeout > 0 {
		opts = append(opts, asynq.Timeout(info.Timeout))
	}
	if !info.Deadline.IsZero() {
		opts = append(opts, asynq.Deadline(info.Deadline))
	}
	_, err = j.replacements.EnqueueContext(ctx, asynq.NewTask(t.Type(), payload, opts...))
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		existing, lookupErr := j.inspector.GetTaskInfo(queue, "safe:"+id)
		if lookupErr != nil {
			return jobError{stage: stageEnqueue, cause: lookupErr}
		}
		if existing.Type != t.Type() || !bytes.Equal(existing.Payload, payload) {
			return jobError{stage: stageEnqueue, cause: errors.New("job replacement conflict")}
		}
		return asynq.RevokeTask
	}
	if err != nil {
		return jobError{stage: stageEnqueue, cause: err}
	}
	return asynq.RevokeTask
}
