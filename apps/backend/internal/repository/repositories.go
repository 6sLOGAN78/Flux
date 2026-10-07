// Package repository collects explicitly injected data-access dependencies.
package repository

import "github.com/6sLOGAN78/flux/internal/server"

// Repositories collects explicitly injected persistence dependencies.
type Repositories struct{}

// NewRepositories constructs the persistence registry.
func NewRepositories(_ *server.Server) *Repositories {
	return &Repositories{}
}
