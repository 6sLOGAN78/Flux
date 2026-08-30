package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flux/apps/backend/internal/model/redirect"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
)

type MockRedirectRepository struct {
	GetByHostAndSlugFunc func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error)
}

func (m *MockRedirectRepository) GetByHostAndSlug(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
	return m.GetByHostAndSlugFunc(ctx, hostname, slug)
}

type MockRedirectCache struct {
	GetFunc    func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error)
	SetFunc    func(ctx context.Context, hostname, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error
	DeleteFunc func(ctx context.Context, hostname, slug string) error
	DeleteHostFunc func(ctx context.Context, hostname string) error
}

func (m *MockRedirectCache) Get(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
	return m.GetFunc(ctx, hostname, slug)
}
func (m *MockRedirectCache) Set(ctx context.Context, hostname, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error {
	return m.SetFunc(ctx, hostname, slug, target, ttl)
}
func (m *MockRedirectCache) Delete(ctx context.Context, hostname, slug string) error {
	return m.DeleteFunc(ctx, hostname, slug)
}
func (m *MockRedirectCache) DeleteHost(ctx context.Context, hostname string) error {
	return m.DeleteHostFunc(ctx, hostname)
}

func TestResolveRedirect_CacheHit(t *testing.T) {
	target := &redirect.LinkRedirectTarget{Slug: "hit"}
	repoCalled := false

	repo := &MockRedirectRepository{
		GetByHostAndSlugFunc: func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
			repoCalled = true
			return nil, repository.ErrNotFound
		},
	}
	cache := &MockRedirectCache{
		GetFunc: func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
			return target, nil
		},
	}

	svc := service.NewRedirectService(repo, cache)
	res, err := svc.ResolveRedirect(context.Background(), "flux.ly", "hit")
	if err != nil || res != target || repoCalled {
		t.Fatalf("Cache hit failed")
	}
}

func TestResolveRedirect_CacheMiss(t *testing.T) {
	target := &redirect.LinkRedirectTarget{Slug: "miss"}
	setCalled := false

	repo := &MockRedirectRepository{
		GetByHostAndSlugFunc: func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
			return target, nil
		},
	}
	cache := &MockRedirectCache{
		GetFunc: func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
			return nil, repository.ErrNotFound
		},
		SetFunc: func(ctx context.Context, hostname, slug string, targ *redirect.LinkRedirectTarget, ttl time.Duration) error {
			setCalled = true
			return nil
		},
	}

	svc := service.NewRedirectService(repo, cache)
	res, err := svc.ResolveRedirect(context.Background(), "flux.ly", "miss")
	if err != nil || res != target || !setCalled {
		t.Fatalf("Cache miss failed")
	}
}

func TestResolveRedirect_CacheErrorBypass(t *testing.T) {
	target := &redirect.LinkRedirectTarget{Slug: "error"}

	repo := &MockRedirectRepository{
		GetByHostAndSlugFunc: func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
			return target, nil
		},
	}
	cache := &MockRedirectCache{
		GetFunc: func(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
			return nil, errors.New("redis down")
		},
		SetFunc: func(ctx context.Context, hostname, slug string, targ *redirect.LinkRedirectTarget, ttl time.Duration) error {
			return errors.New("redis down")
		},
	}

	svc := service.NewRedirectService(repo, cache)
	res, err := svc.ResolveRedirect(context.Background(), "flux.ly", "error")
	if err != nil || res != target {
		t.Fatalf("Cache error bypass failed")
	}
}
