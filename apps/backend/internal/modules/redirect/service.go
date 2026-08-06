package redirect

import (
	"context"
	"fmt"
	"time"
)

// RedirectService manages high-performance link resolution and cache management.
type RedirectService struct {
	repo  RedirectRepository
	cache RedirectCache
}

// NewRedirectService initializes a RedirectService instance.
func NewRedirectService(repo RedirectRepository, cache RedirectCache) *RedirectService {
	return &RedirectService{
		repo:  repo,
		cache: cache,
	}
}

// ResolveRedirect resolves a short link slug via Cache-Aside strategy.
func (s *RedirectService) ResolveRedirect(ctx context.Context, slug string) (*LinkRedirectTarget, error) {
	if ReservedSlugs[slug] {
		return nil, ErrNotFound
	}

	// 1. Redis / Cache Lookup
	if s.cache != nil {
		target, err := s.cache.Get(ctx, slug)
		if err == nil && target != nil {
			return target, nil
		}
	}

	// 2. Database Lookup (Cache Miss)
	target, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch link from repository: %w", err)
	}

	// 3. Update Cache on Miss (24 Hours TTL)
	if s.cache != nil && target != nil {
		_ = s.cache.Set(ctx, slug, target, 24*time.Hour)
	}

	return target, nil
}
