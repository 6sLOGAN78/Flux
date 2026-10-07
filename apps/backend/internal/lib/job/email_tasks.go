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
	traceparentLength   = 55
	canonicalUUIDLength = 36
	welcomeMaxRetries   = 3
	welcomeTaskTimeout  = 30 * time.Second
)

// TaskWelcome identifies asynchronous welcome email jobs.
const (
	TaskWelcome = "email:welcome"
)

// WelcomeEmailPayload contains welcome email values and safe correlation metadata.
type WelcomeEmailPayload struct {
	Metadata  *Metadata `json:"metadata,omitempty"`
	To        string    `json:"to"`
	FirstName string    `json:"first_name"`
}

// Metadata is the entire approved queue propagation envelope. No private state.
type Metadata struct {
	Traceparent   string `json:"traceparent,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Version       int    `json:"version"`
}

func cleanMetadata(m *Metadata) *Metadata {
	if m == nil || (m.Version != 0 && m.Version != 1) {
		return nil
	}
	clean := &Metadata{Version: 1}
	if len(m.Traceparent) == traceparentLength {
		ctx := observability.TraceparentPropagator{}.Extract(context.Background(),
			propagation.MapCarrier{"traceparent": m.Traceparent})
		carrier := propagation.MapCarrier{}
		observability.TraceparentPropagator{}.Inject(ctx, carrier)
		clean.Traceparent = carrier["traceparent"]
	}
	for _, pair := range []struct {
		output *string
		input  string
	}{{input: m.RequestID, output: &clean.RequestID}, {input: m.CorrelationID, output: &clean.CorrelationID}} {
		if len(pair.input) != canonicalUUIDLength {
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

// NewWelcomeEmailTask validates and serializes a welcome email job.
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
		Metadata: cleanMetadata(&Metadata{Version: 1,
			Traceparent:   carrier["traceparent"],
			RequestID:     ids.RequestID,
			CorrelationID: ids.CorrelationID}),
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskWelcome, payload,
		asynq.MaxRetry(welcomeMaxRetries),
		asynq.Queue("default"),
		asynq.Timeout(welcomeTaskTimeout)), nil
}
