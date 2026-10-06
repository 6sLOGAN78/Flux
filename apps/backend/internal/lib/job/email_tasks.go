package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/propagation"
)

const (
	TaskWelcome = "email:welcome"
)

type WelcomeEmailPayload struct {
	To        string    `json:"to"`
	FirstName string    `json:"first_name"`
	Metadata  *Metadata `json:"metadata,omitempty"`
}

// Metadata is the entire approved queue propagation envelope. No private state.
type Metadata struct {
	Version       int    `json:"version"`
	Traceparent   string `json:"traceparent,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

func cleanMetadata(m *Metadata) *Metadata {
	if m == nil || (m.Version != 0 && m.Version != 1) {
		return nil
	}
	clean := &Metadata{Version: 1}
	if len(m.Traceparent) == 55 {
		ctx := observability.TraceparentPropagator{}.Extract(context.Background(), propagation.MapCarrier{"traceparent": m.Traceparent})
		carrier := propagation.MapCarrier{}
		observability.TraceparentPropagator{}.Inject(ctx, carrier)
		clean.Traceparent = carrier["traceparent"]
	}
	for _, pair := range []struct {
		input  string
		output *string
	}{{m.RequestID, &clean.RequestID}, {m.CorrelationID, &clean.CorrelationID}} {
		if len(pair.input) != 36 {
			continue
		}
		id, err := uuid.Parse(pair.input)
		if err == nil && id != uuid.Nil && id.String() == pair.input {
			*pair.output = pair.input
		}
	}
	if clean.Traceparent == "" && clean.RequestID == "" && clean.CorrelationID == "" {
		return nil
	}
	return clean
}

func NewWelcomeEmailTask(to, firstName string) (*asynq.Task, error) {
	return NewWelcomeEmailTaskContext(context.Background(), to, firstName)
}

// NewWelcomeEmailTaskContext stores only safe IDs; old tasks omit metadata.
func NewWelcomeEmailTaskContext(ctx context.Context, to, firstName string) (*asynq.Task, error) {
	carrier := propagation.MapCarrier{}
	observability.TraceparentPropagator{}.Inject(ctx, carrier)
	ids := observability.CorrelationFromContext(ctx)
	payload, err := json.Marshal(WelcomeEmailPayload{
		To:        to,
		FirstName: firstName,
		Metadata:  cleanMetadata(&Metadata{Version: 1, Traceparent: carrier["traceparent"], RequestID: ids.RequestID, CorrelationID: ids.CorrelationID}),
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskWelcome, payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
		asynq.Timeout(30*time.Second)), nil
}
