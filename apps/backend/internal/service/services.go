// Package service provides domain business logic services.
package service

import (
	"context"
	"time"

	"flux/apps/backend/internal/model/redirect"
	"flux/apps/backend/internal/repository"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// RedirectService manages high-performance link resolution and cache management.
type RedirectService struct {
	repo  repository.RedirectRepository
	cache repository.RedirectCache
	sg    singleflight.Group
}

// NewRedirectService initializes a RedirectService instance.
func NewRedirectService(repo repository.RedirectRepository, cache repository.RedirectCache) *RedirectService {
	return &RedirectService{
		repo:  repo,
		cache: cache,
	}
}

// ResolveRedirect resolves a short link slug via Cache-Aside strategy.
func (s *RedirectService) ResolveRedirect(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
	if redirect.ReservedSlugs[slug] {
		return nil, repository.ErrNotFound
	}

	// 1. Redis / Cache Lookup
	if s.cache != nil {
		target, err := s.cache.Get(ctx, hostname, slug)
		if err == nil && target != nil {
			log.Debug().Str("hostname", hostname).Str("slug", slug).Msg("cache_hit")
			return target, nil
		}
		if err == repository.ErrNotFound {
			log.Debug().Str("hostname", hostname).Str("slug", slug).Msg("cache_miss")
		} else {
			log.Warn().Err(err).Str("hostname", hostname).Str("slug", slug).Msg("cache_error")
		}
	}

	// 2. Database Lookup (Cache Miss) with Singleflight
	// Singleflight key uses hostname+slug to prevent simultaneous lookups for same link
	sfKey := hostname + ":" + slug
	v, err, _ := s.sg.Do(sfKey, func() (interface{}, error) {
		if s.repo != nil {
			target, repoErr := s.repo.GetByHostAndSlug(ctx, hostname, slug)
			if repoErr == nil && target != nil {
				if s.cache != nil {
					// 24 hours TTL for active links.
					// The cache will be explicitly invalidated by LinkService on update/delete.
					_ = s.cache.Set(ctx, hostname, slug, target, 24*time.Hour)
				}
				return target, nil
			}
			return nil, repoErr
		}
		return nil, repository.ErrNotFound
	})

	if err != nil {
		return nil, err
	}
	return v.(*redirect.LinkRedirectTarget), nil
}
