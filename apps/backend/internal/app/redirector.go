package app

import (
	"context"

	"github.com/6sLOGAN78/flux/internal/config"
)

type Redirector struct{ *RoleRuntime }

// NewRedirector constructs only the Phase 1 system HTTP and logging graph.
// Link configuration lookup and data-plane resources belong to later phases.
func NewRedirector(ctx context.Context, cfg *config.Config) (*Redirector, error) {
	runtime, err := newRole(ctx, config.RoleRedirector, cfg, defaultRoleFactories())
	if err != nil {
		return nil, err
	}
	return &Redirector{runtime}, nil
}
