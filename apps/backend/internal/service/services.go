package service

import (
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Services collects explicitly injected application services.
type Services struct {
	Auth      *AuthService
	Identity  IdentityResolver
	Workspace *WorkspaceService
	Job       *job.JobService
}

// NewServices constructs application service dependencies.
func NewServices(s *server.Server, _ *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	var pool *pgxpool.Pool
	if s.DB != nil {
		pool = s.DB.Pool
	}

	return &Services{
		Job:       s.Job,
		Auth:      authService,
		Identity:  NewIdentityService(repository.NewUserRepository(pool), authService),
		Workspace: NewWorkspaceService(repository.NewWorkspaceRepository(pool)),
	}, nil
}
