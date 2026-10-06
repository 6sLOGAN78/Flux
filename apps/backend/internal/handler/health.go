package handler

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/transport"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ReadinessCheck declares a required, context-aware dependency of a role.
// Names are fixed by composition, never derived from request input or errors.
type ReadinessCheck struct {
	Name  string
	Check func(context.Context) error
}

type HealthHandler struct {
	logger  *zerolog.Logger
	timeout time.Duration
	checks  []ReadinessCheck
	ready   func() bool
}

// SetReadinessGate connects supervisor shutdown without dependency probes.
// Composition calls this before the listener starts serving requests.
func (h *HealthHandler) SetReadinessGate(ready func() bool) { h.ready = ready }

// NewHealthHandler preserves registry construction without inferring dependencies
// from the server container. Role composition injects its own required checks.
func NewHealthHandler(s *server.Server) *HealthHandler {
	return NewReadinessHandler(s.Logger, s.Config.ForRole(s.Role).ReadinessTimeout, nil)
}

// NewReadinessHandler snapshots required checks; optional integrations are absent.
func NewReadinessHandler(log *zerolog.Logger, timeout time.Duration, checks []ReadinessCheck) *HealthHandler {
	if timeout <= 0 {
		timeout = time.Second
	}
	return &HealthHandler{logger: log, timeout: timeout, checks: append([]ReadinessCheck(nil), checks...)}
}

// Live reflects the running HTTP process and never contacts dependencies.
func (h *HealthHandler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, transport.HealthLiveResponse{Status: transport.Alive})
}

type healthComponent = struct {
	Name  string                                            `json:"name"`
	State transport.TransportHealthReadyResponseChecksState `json:"state"`
}

type readinessResult struct {
	index    int
	err      error
	duration time.Duration
}

// Ready checks independent dependencies concurrently within one request deadline.
// Only this goroutine writes the response or logs results. Late completions send
// to the buffered channel without retaining the Echo request context.
func (h *HealthHandler) Ready(c echo.Context) error {
	if h.ready != nil && !h.ready() {
		return c.JSON(http.StatusServiceUnavailable, transport.HealthReadyResponse{
			Status: transport.TransportHealthReadyResponseStatusNotReady, Checks: []healthComponent{},
		})
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), h.timeout)
	defer cancel()
	response := transport.HealthReadyResponse{Status: transport.TransportHealthReadyResponseStatusReady, Checks: make([]healthComponent, len(h.checks))}
	results := make(chan readinessResult, len(h.checks))
	for i, check := range h.checks {
		response.Checks[i] = healthComponent{Name: check.Name, State: transport.TransportHealthReadyResponseChecksStateNotReady}
		go func(index int, check ReadinessCheck) {
			start := time.Now()
			var err error
			if check.Check == nil {
				err = errors.New("required check is unconfigured")
			} else if ctx.Err() != nil {
				err = ctx.Err()
			} else {
				err = check.Check(ctx)
			}
			results <- readinessResult{index: index, err: err, duration: time.Since(start)}
		}(i, check)
	}
	pending := make([]bool, len(h.checks))
	for i := range pending {
		pending[i] = true
	}
	remaining := len(h.checks)
	for remaining > 0 {
		select {
		case result := <-results:
			pending[result.index] = false
			remaining--
			if result.err == nil {
				response.Checks[result.index].State = transport.TransportHealthReadyResponseChecksStateReady
			} else {
				h.logFailure(c, h.checks[result.index].Name, result.err, result.duration)
			}
		case <-ctx.Done():
			for i, waiting := range pending {
				if waiting {
					h.logFailure(c, h.checks[i].Name, ctx.Err(), h.timeout)
				}
			}
			remaining = 0
		}
	}
	status := http.StatusOK
	for _, check := range response.Checks {
		if check.State != transport.TransportHealthReadyResponseChecksStateReady {
			response.Status = transport.TransportHealthReadyResponseStatusNotReady
			status = http.StatusServiceUnavailable
			break
		}
	}
	if ctx.Err() != nil || (h.ready != nil && !h.ready()) {
		response.Status = transport.TransportHealthReadyResponseStatusNotReady
		status = http.StatusServiceUnavailable
	}
	return c.JSON(status, response)
}

func (h *HealthHandler) logFailure(c echo.Context, name string, err error, duration time.Duration) {
	log := h.logger
	if contextual, ok := c.Get(middleware.LoggerKey).(*zerolog.Logger); ok {
		log = contextual
	}
	if log == nil {
		return
	}
	kind := "dependency_error"
	var networkErr net.Error
	switch {
	case errors.Is(err, context.Canceled):
		kind = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		kind = "deadline_exceeded"
	case errors.As(err, &networkErr):
		if networkErr.Timeout() {
			kind = "network_timeout"
		} else {
			kind = "network_error"
		}
	}
	// Driver/provider text may contain credentials or payloads. Emit safe
	// classification, component and timing once without serializing Err.
	dependency := name
	if dependency == "database" {
		dependency = "postgres"
	}
	category := "unavailable"
	if errors.Is(err, context.Canceled) {
		category = "canceled"
	} else if errors.Is(err, context.DeadlineExceeded) || kind == "network_timeout" {
		category = "timeout"
	}
	attrs := observability.SanitizeAttributes([]attribute.KeyValue{
		attribute.String("operation", "dependency.check"),
		attribute.String("dependency", dependency),
		attribute.String("error.category", category),
	}, false)
	trace.SpanFromContext(c.Request().Context()).AddEvent("dependency.check", trace.WithAttributes(attrs...))
	log.Error().Str("operation", "dependency.check").Str("dependency", dependency).Str("error.category", category).
		Str("component", name).Str("failure_kind", kind).Dur("duration", duration).Msg("readiness check failed")
}
