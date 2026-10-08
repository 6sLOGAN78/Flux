package database

import (
	"context"
	"fmt"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	loggerConfig "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Database owns a PostgreSQL pool and its contextual logger.
type Database struct {
	Pool *pgxpool.Pool
	log  *zerolog.Logger
}

type queryTracer struct {
	tracer trace.Tracer
	log    *zerolog.Logger
}
type querySpanKey struct{}

// Query data (including SQL and arguments) is deliberately never inspected.
//
//nolint:spancheck // pgx transfers this span to TraceQueryEnd, which always ends it after the query.
func (t queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	ctx = baggage.ContextWithBaggage(ctx, baggage.Baggage{})
	ctx = trace.ContextWithSpanContext(ctx, observability.CleanSpanContext(trace.SpanContextFromContext(ctx)))
	ids := observability.CorrelationFromContext(ctx)
	ctx,
		span := t.tracer.Start(ctx,
		"database.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("dependency",
			"postgres"),
			attribute.String("operation",
				"database.query"),
			attribute.String("db.system.name",
				"postgresql"),
			attribute.String("request_id",
				ids.RequestID),
			attribute.String("correlation_id",
				ids.CorrelationID)))
	return context.WithValue(ctx, querySpanKey{}, span)
}
func (t queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span, ok := ctx.Value(querySpanKey{}).(trace.Span)
	if !ok {
		return
	}
	defer span.End()
	outcome := "success"
	if data.Err != nil {
		outcome = "error"
		span.SetStatus(codes.Error, "")
		span.SetAttributes(attribute.String("error.category", "unknown"))
	}
	span.SetAttributes(attribute.String("outcome", outcome))
	log := loggerConfig.WithContext(*t.log, ctx)
	log.Debug().
		Str("dependency",
			"postgres").
		Str("operation",
			"database.query").
		Str("outcome",
			outcome).
		Msg("database.query")
}

// DatabasePingTimeout bounds the initial PostgreSQL probe in seconds.
const DatabasePingTimeout = 10

// New constructs a PostgreSQL pool and closes it if its initial probe fails.
func New(cfg *config.Config,
	logger *zerolog.Logger,
	_ *loggerConfig.LoggerService,
	telemetry ...*observability.Telemetry) (*Database,
	error) {
	pgxPoolConfig, err := pgxpool.ParseConfig(databaseURL(cfg.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
	}

	tracer := noop.NewTracerProvider().Tracer("flux.database")
	if len(telemetry) > 0 && telemetry[0] != nil && telemetry[0].Tracer != nil {
		tracer = telemetry[0].Tracer
	}
	pgxPoolConfig.ConnConfig.Tracer = queryTracer{tracer: tracer, log: logger}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	database := &Database{
		Pool: pool,
		log:  logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DatabasePingTimeout*time.Second)
	defer cancel()
	if err = pingPool(ctx, pool); err != nil {
		return nil, err
	}

	logger.Info().Msg("connected to the database")

	return database, nil
}

// pingPool transfers ownership only after a successful initial ping.
func pingPool(ctx context.Context, pool interface {
	Ping(context.Context) error
	Close()
}) error {
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// Close closes the owned PostgreSQL pool.
func (db *Database) Close() error {
	db.log.Info().Msg("closing database connection pool")
	db.Pool.Close()
	return nil
}
