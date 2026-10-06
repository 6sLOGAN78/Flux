package repository

import "github.com/6sLOGAN78/flux/internal/server"

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}