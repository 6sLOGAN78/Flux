// Package service provides domain business logic services.
package service

import (
	"context"
	"time"

	"flux/apps/backend/internal/model/redirect"
	"flux/apps/backend/internal/repository"
)

// RedirectService manages high-performance link resolution and cache management.
type RedirectService struct {
	repo  repository.RedirectRepository
	cache repository.RedirectCache
}

// NewRedirectService initializes a RedirectService instance.
func NewRedirectService(repo repository.RedirectRepository, cache repository.RedirectCache) *RedirectService {
	return &RedirectService{
		repo:  repo,
		cache: cache,
	}
}

// ResolveRedirect resolves a short link slug via Cache-Aside strategy.
func (s *RedirectService) ResolveRedirect(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error) {
	if redirect.ReservedSlugs[slug] {
		return nil, repository.ErrNotFound
	}

	// 1. Redis / Cache Lookup
	if s.cache != nil {
		target, err := s.cache.Get(ctx, slug)
		if err == nil && target != nil {
			return target, nil
		}
	}

	// 2. Database Lookup (Cache Miss)
	if s.repo != nil {
		target, err := s.repo.GetBySlug(ctx, slug)
		if err == nil && target != nil {
			if s.cache != nil {
				_ = s.cache.Set(ctx, slug, target, 24*time.Hour)
			}
			return target, nil
		}
	}

	return nil, repository.ErrNotFound
}
