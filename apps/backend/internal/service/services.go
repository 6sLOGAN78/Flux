package service

import (
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/server"
)

// Services collects explicitly injected application services.
type Services struct {
	Auth      *AuthService
	Identity  IdentityResolver
	Workspace *WorkspaceService
	Job       *job.JobService
}

// NewServices constructs application service dependencies.
func NewServices(s *server.Server, stores *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	if stores == nil {
		stores = repository.NewRepositories(s)
	}

	return &Services{
		Job:       s.Job,
		Auth:      authService,
		Identity:  NewIdentityService(stores.Users, authService),
		Workspace: NewWorkspaceService(stores.Workspaces),
	}, nil
}
