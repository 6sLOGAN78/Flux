//nolint:testpackage // These tests verify package-private lifecycle and failure-injection seams.
package job

import (
	"errors"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/rs/zerolog"
)

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestJobStopClosesAll(t *testing.T) {
	consumerCause := errors.New("SECRET-MARKER consumer")
	producerCause := errors.New("SECRET-MARKER producer")
	for _, test := range []struct {
		consumer error
		producer error
		name     string
	}{{name: "consumer",
		consumer: consumerCause,
		producer: nil},
		{name: "producer",
			consumer: nil,
			producer: producerCause},
		{name: "both",
			consumer: consumerCause,
			producer: producerCause},
		{name: "success",
			consumer: nil,
			producer: nil}} {
		t.Run(test.name, func(t *testing.T) {
			var order []string
			log := zerolog.Nop()
			j := &JobService{logger: &log,
				shutdown: func() error {
					order = append(order,
						"consumer")
					return test.consumer
				},
				closeClient: func() error {
					order = append(order,
						"producer")
					return test.producer
				}}
			err := j.Stop()
			if strings.Join(order, ",") != "consumer,producer" {
				t.Fatalf("did not stop all dependents in order: %v", order)
			}
			for _, cause := range []error{test.consumer, test.producer} {
				if cause != nil && !errors.Is(err, cause) {
					t.Fatalf("lost cause: %v", err)
				}
			}
			if test.consumer == nil && test.producer == nil && err != nil {
				t.Fatal(err)
			}
			if err != nil && strings.Contains(err.Error(), "SECRET-MARKER") {
				t.Fatal("shutdown diagnostic leaked secret")
			}
			//nolint:errorlint // Idempotent stop must return the identical error object, including nil on success.
			if again := j.Stop(); again != err || len(order) != 2 {
				t.Fatal("Stop was not idempotent")
			}
		})
	}
}

func TestJobOwnsEmailClient(t *testing.T) {
	log := zerolog.Nop()
	first := NewJobService(&log, &config.Config{Redis: config.RedisConfig{Address: "localhost:6379"}})
	second := NewJobService(&log, &config.Config{Redis: config.RedisConfig{Address: "localhost:6379"}})
	defer first.Stop()
	defer second.Stop()
	if first.emailClient == nil || second.emailClient == nil || first.emailClient == second.emailClient {
		t.Fatal("instances do not own separate email clients")
	}
	owned := first.emailClient
	second.InitHandlers(&config.Config{}, &log)
	if first.emailClient != owned {
		t.Fatal("initializing another instance replaced owned client")
	}
}
