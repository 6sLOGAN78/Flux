package handler

import (
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
)

// Handlers collects explicitly injected HTTP handlers.
type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
}

// NewHandlers constructs the role-injected health and documentation handlers.
func NewHandlers(s *server.Server, _ *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
	}
}
