package middleware

import (
	"net/http"

	"github.com/6sLOGAN78/flux/internal/server"
)

// RateLimitMiddleware bounds request rates before handler execution.
type RateLimitMiddleware struct {
	server *server.Server
}

// NewRateLimitMiddleware constructs request throttling middleware.
func NewRateLimitMiddleware(s *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		server: s,
	}
}

// RecordRateLimitHit records a sanitized throttling event.
func (r *RateLimitMiddleware) RecordRateLimitHit(_ string) {
	r.server.Logger.Warn().Int("http.response.status_code", http.StatusTooManyRequests).Msg("http.request")
}
