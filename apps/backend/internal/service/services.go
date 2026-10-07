package service

import (
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/server"
)

// Services collects explicitly injected application services.
type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

// NewServices constructs application service dependencies.
func NewServices(s *server.Server, _ *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}
