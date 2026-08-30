package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"flux/apps/backend/internal/lib/utils"
	"flux/apps/backend/internal/model/domain"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/errs"
	"flux/apps/backend/pkg/sqlerr"
)

type DomainService struct {
	repo           *repository.DomainRepository
	cache          repository.RedirectCache
	platformDomain string
}

func NewDomainService(repo *repository.DomainRepository, cache repository.RedirectCache, platformDomain string) *DomainService {
	return &DomainService{
		repo:           repo,
		cache:          cache,
		platformDomain: platformDomain,
	}
}

// CreateDomain normalizes, validates, generates a verification token, and creates a domain.
func (s *DomainService) CreateDomain(ctx context.Context, tenantID, hostname string) (*domain.CustomDomain, error) {
	hostname = utils.NormalizeHostname(hostname)

	// Basic validation
	if err := validateHostname(hostname, s.platformDomain); err != nil {
		return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, err)
	}

	// Generate verification token (e.g. flux-verify=hex)
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	token := "flux-verify=" + hex.EncodeToString(b)

	d, err := s.repo.CreateDomain(ctx, tenantID, hostname, token)
	if err != nil {
		var sqlErr *sqlerr.Error
		if errors.As(err, &sqlErr) && sqlErr.Code == sqlerr.UniqueViolation {
			code := "CONFLICT"
			return nil, &errs.HTTPError{StatusCode: 409, Message: "hostname is already registered", Code: &code}
		}
		if errors.As(err, &sqlErr) && sqlErr.Code == sqlerr.CheckViolation {
			return nil, errs.NewBadRequestError("invalid hostname configuration", false, nil, nil, err)
		}
		return nil, err
	}

	return d, nil
}

func (s *DomainService) GetDomains(ctx context.Context, tenantID string) ([]domain.CustomDomain, error) {
	return s.repo.GetDomainsByTenant(ctx, tenantID)
}

func (s *DomainService) GetDomainByID(ctx context.Context, tenantID, id string) (*domain.CustomDomain, error) {
	return s.repo.GetDomainByID(ctx, tenantID, id)
}

func (s *DomainService) DeleteDomain(ctx context.Context, tenantID, id string) error {
	// First fetch domain to get hostname for cache invalidation (if needed)
	d, err := s.repo.GetDomainByID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Don't leak existence across tenants, safe 404 is handled up stack
			return err
		}
		return err
	}

	// Delete from DB (links will be SET NULL on custom_domain_id)
	err = s.repo.DeleteDomain(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Invalidate any routing cache for this domain?
	// Currently links caching does not eagerly cache entire domains, it caches specific slugs.
	// We'd have to scan or let the TTL expire if the platform accesses it.
	// Actually, wait, cache invalidation is explicitly requested: 
	// "For delete domain... determine which existing cache invalidation methods need to be called. The invariant is: Deleted/inactive domain -> No stale Redis redirect."
	// Redis keys are redirect:{hostname}:{slug}. We cannot delete by wildcard easily.
	// In the real world we might publish an invalidation event. But wait, if they delete the domain, the links drop the custom_domain_id. The cache keeps redirect:{hostname}:{slug} which might STILL route if it's cached.
	// But `cacheRepo.Delete` takes `hostname, slug`. 
	// Is there a way to clear all for a hostname? No `DeleteByHost` in `RedirectCache`.
	// We can leave this as a TODO or implement a flush for the host. 
	// The prompt: "If 12E already handles the required invalidation, reuse it." 12E handles link updates. 
	// "After deletion, future routing must naturally stop resolving through that custom domain. Ensure the appropriate Redis invalidation path remains compatible with 12E."
	// Let's add an invalidation method.

	if s.cache != nil {
		_ = s.cache.DeleteHost(ctx, d.Hostname)
	}

	return nil
}

func validateHostname(host, platformDomain string) error {
	if host == "" {
		return errors.New("hostname cannot be empty")
	}
	if host == "localhost" || host == "127.0.0.1" {
		return errors.New("invalid hostname: reserved or local")
	}
	if strings.Contains(host, "/") || strings.Contains(host, " ") {
		return errors.New("invalid hostname characters")
	}
	if host == platformDomain || strings.HasSuffix(host, "."+platformDomain) {
		return errors.New("cannot register platform domain or its subdomains")
	}
	return nil
}
