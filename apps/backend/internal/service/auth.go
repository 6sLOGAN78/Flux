// Package service collects application services and authentication configuration.
package service

import (
	"github.com/6sLOGAN78/flux/internal/server"

	"github.com/clerk/clerk-sdk-go/v2"
)

// AuthService holds authentication configuration.
type AuthService struct {
	server *server.Server
}

// NewAuthService constructs authentication settings without global initialization.
func NewAuthService(s *server.Server) *AuthService {
	clerk.SetKey(s.Config.Auth.SecretKey)
	return &AuthService{
		server: s,
	}
}
