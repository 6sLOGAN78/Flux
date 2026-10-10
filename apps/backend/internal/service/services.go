package service

import (
	"github.com/6sLOGAN78/flux/internal/lib/job"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Services collects explicitly injected application services.
type Services struct {
	Auth        *AuthService
	Identity    IdentityResolver
	Workspace   *WorkspaceService
	Links       *LinkService
	Team        *TeamService
	Invitations *InvitationService
	Job         *job.JobService
}

// NewServices constructs application service dependencies.
func NewServices(s *server.Server, stores *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	if stores == nil {
		var pool *pgxpool.Pool
		if s.DB != nil {
			pool = s.DB.Pool
		}
		stores = repository.NewRepositories(pool)
	}

	workspace := NewWorkspaceService(stores.Workspaces)
	var pool *pgxpool.Pool
	if s.DB != nil {
		pool = s.DB.Pool
	}
	team := NewTeamService(stores.Team, workspace)
	return &Services{
		Job:         s.Job,
		Auth:        authService,
		Identity:    NewIdentityService(stores.Users, authService),
		Workspace:   workspace,
		Team:        team,
		Invitations: NewInvitationService(repository.NewInvitationRepository(pool), team, s.Config.Invitations),
		Links:       NewLinkService(repository.NewLinkRepository(pool), workspace, s.Config.Links),
	}, nil
}
