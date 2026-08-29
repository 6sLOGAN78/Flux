package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories holds all data access repositories for the application.
type Repositories struct {
	Link     *LinkRepository
	Category *CategoryRepository
	User     *UserRepository
}

// NewRepositories initializes and returns a Repositories instance.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Link:     NewLinkRepository(pool),
		Category: NewCategoryRepository(pool),
		User:     NewUserRepository(pool),
	}
}
