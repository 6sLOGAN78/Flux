package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
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
func (t queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	ctx = baggage.ContextWithBaggage(ctx, baggage.Baggage{})
	ctx = trace.ContextWithSpanContext(ctx, observability.CleanSpanContext(trace.SpanContextFromContext(ctx)))
	ids := observability.CorrelationFromContext(ctx)
	ctx, span := t.tracer.Start(ctx, "database.query", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("dependency", "postgres"), attribute.String("operation", "database.query"), attribute.String("db.system.name", "postgresql"), attribute.String("request_id", ids.RequestID), attribute.String("correlation_id", ids.CorrelationID)))
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
	log.Debug().Str("dependency", "postgres").Str("operation", "database.query").Str("outcome", outcome).Msg("database.query")
}

const DatabasePingTimeout = 10

func New(cfg *config.Config, logger *zerolog.Logger, _ *loggerConfig.LoggerService, telemetry ...*observability.Telemetry) (*Database, error) {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// URL-encode the password
	encodedPassword := url.QueryEscape(cfg.Database.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	pgxPoolConfig, err := pgxpool.ParseConfig(dsn)
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

func (db *Database) Close() error {
	db.log.Info().Msg("closing database connection pool")
	db.Pool.Close()
	return nil
}
