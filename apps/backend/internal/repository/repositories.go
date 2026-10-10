// Package repository collects explicitly injected data-access dependencies.
package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories collects explicitly injected persistence dependencies.
type Repositories struct {
	Users      *UserRepository
	Workspaces *WorkspaceRepository
	Team       *TeamRepository
}

// NewRepositories constructs the persistence registry.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Users: NewUserRepository(pool), Workspaces: NewWorkspaceRepository(pool), Team: NewTeamRepository(pool),
	}
}
