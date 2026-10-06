package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/observability"
)

func TestRoleMigratorProviderOwnership(t *testing.T) {
	for _, stage := range []string{"success", "telemetry", "migration", "close"} {
		t.Run(stage, func(t *testing.T) {
			cfg := &config.Config{Observability: config.DefaultObservabilityConfig()}
			cfg.Observability.HealthChecks.Timeout = 100 * time.Millisecond
			var events []string
			var exit context.Context
			cause := errors.New("SECRET-MARKER")
			f := migratorFactories{
				telemetry: func(ctx context.Context, settings observability.Settings, role string) (*observability.Telemetry, func(context.Context) error, error) {
					events = append(events, "telemetry")
					if role != "migrator" || settings != cfg.Observability.TelemetrySettings() {
						t.Error("incorrect provider settings")
					}
					if stage == "telemetry" {
						return nil, nil, cause
					}
					owner, err := observability.New(ctx, settings, role)
					return owner, func(ctx context.Context) error {
						events = append(events, "flush")
						if ctx != exit {
							t.Error("flush created a fresh exit budget")
						}
						if stage == "close" {
							return errors.Join(cause, owner.Shutdown(ctx))
						}
						return owner.Shutdown(ctx)
					}, err
				},
				migrate: func(ctx context.Context, cfg *config.Config, closeContext func() context.Context) (database.MigrationResult, error) {
					events = append(events, "migration", "postgres_close")
					exit = closeContext()
					deadline, ok := exit.Deadline()
					if !ok || time.Until(deadline) > cfg.Observability.HealthChecks.Timeout {
						t.Error("unbounded exit")
					}
					if stage == "migration" {
						return database.MigrationResult{}, cause
					}
					return database.MigrationResult{}, nil
				},
			}
			var out bytes.Buffer
			code := runConfigured(context.Background(), cfg, &out, f)
			wantCode := 1
			if stage == "success" {
				wantCode = 0
			}
			if code != wantCode || strings.Contains(out.String(), "SECRET-MARKER") {
				t.Fatalf("code %d logs %s", code, out.String())
			}
			want := []string{"telemetry", "migration", "postgres_close", "flush"}
			if stage == "telemetry" {
				want = []string{"telemetry"}
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events %v want %v", events, want)
			}
		})
	}
}
