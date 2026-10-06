package middleware

import (
	"github.com/6sLOGAN78/flux/internal/server"
)

type RateLimitMiddleware struct {
	server *server.Server
}

func NewRateLimitMiddleware(s *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		server: s,
	}
}

func (r *RateLimitMiddleware) RecordRateLimitHit(endpoint string) {
	r.server.Logger.Warn().Int("http.response.status_code", 429).Msg("http.request")
}
