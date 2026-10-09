// Package repository collects explicitly injected data-access dependencies.
package repository

import (
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories collects explicitly injected persistence dependencies.
type Repositories struct {
	Users      *UserRepository
	Workspaces *WorkspaceRepository
}

// NewRepositories constructs the persistence registry.
func NewRepositories(s *server.Server) *Repositories {
	var pool *pgxpool.Pool
	if s.DB != nil {
		pool = s.DB.Pool
	}
	return &Repositories{Users: NewUserRepository(pool), Workspaces: NewWorkspaceRepository(pool)}
}
