package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model"
	"flux/apps/backend/internal/model/link"
	"flux/apps/backend/internal/modules/billing"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/google/uuid"
)

type LinkService struct {
	repo        *repository.LinkRepository
	cache       repository.RedirectCache
	campRepo    *repository.CampaignRepository
	billingRepo *repository.BillingRepository
}

func NewLinkService(repo *repository.LinkRepository, cache repository.RedirectCache, campRepo *repository.CampaignRepository, billingRepo *repository.BillingRepository) *LinkService {
	return &LinkService{repo: repo, cache: cache, campRepo: campRepo, billingRepo: billingRepo}
}

func (s *LinkService) CreateLink(ctx context.Context, tenantID *uuid.UUID, payload *link.CreateLinkPayload) (*link.Link, error) {
	if payload.CampaignID != nil && tenantID != nil {
		_, err := s.campRepo.GetCampaign(ctx, *tenantID, *payload.CampaignID)
		if err != nil {
			return nil, errs.NewBadRequestError("invalid campaign id or cross-workspace association denied", false, nil, nil, err)
		}
	}

	var maxLinks *int64
	if tenantID != nil && s.billingRepo != nil {
		sub, err := s.billingRepo.GetSubscriptionByWorkspace(ctx, *tenantID)
		var planTier string
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) || err.Error() == "repository: resource not found" {
				planTier = "free"
			} else {
				return nil, err
			}
		} else {
			if sub.Status == "active" || sub.Status == "trialing" {
				planTier = sub.PlanTier
			} else {
				planTier = "free"
			}
		}
		limits := billing.GetTierLimits(planTier)
		maxLinks = &limits.MaxLinks
	}

	if payload.CustomCode != nil && *payload.CustomCode != "" {
		return s.repo.CreateLink(ctx, tenantID, payload, *payload.CustomCode, maxLinks)
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		shortCode, err := generateShortCode()
		if err != nil {
			return nil, err
		}

		result, err := s.repo.CreateLink(ctx, tenantID, payload, shortCode, maxLinks)
		if err == nil {
			return result, nil
		}

		var pgErr *sqlerr.Error
		if errors.As(err, &pgErr) && pgErr.Code == sqlerr.UniqueViolation && strings.Contains(pgErr.ConstraintName, "short_code") {
			continue // Retry
		}

		return nil, err
	}

	return nil, errors.New("failed to generate unique short code after maximum retries")
}

func (s *LinkService) GetLinks(ctx context.Context, tenantID *uuid.UUID, query *link.GetLinksQuery) (*model.PaginatedResponse[link.Link], error) {
	return s.repo.GetLinks(ctx, tenantID, query)
}

func (s *LinkService) GetLinkByID(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) (*link.Link, error) {
	return s.repo.GetLinkByID(ctx, tenantID, id)
}

func (s *LinkService) UpdateLink(ctx context.Context, tenantID *uuid.UUID, payload *link.UpdateLinkPayload) (*link.Link, error) {
	if payload.CampaignID != nil && tenantID != nil {
		_, err := s.campRepo.GetCampaign(ctx, *tenantID, *payload.CampaignID)
		if err != nil {
			return nil, errs.NewBadRequestError("invalid campaign id or cross-workspace association denied", false, nil, nil, err)
		}
	}

	// 1. Fetch current hostname before update if we need to invalidate
	var oldHostname string
	if s.cache != nil {
		if h, err := s.repo.GetHostnameForLink(ctx, payload.ID); err == nil && h != nil {
			oldHostname = *h
		}
	}

	result, err := s.repo.UpdateLink(ctx, tenantID, payload)
	
	// 2. Fetch new hostname after update
	var newHostname string
	if err == nil && s.cache != nil {
		if h, err := s.repo.GetHostnameForLink(ctx, payload.ID); err == nil && h != nil {
			newHostname = *h
		}
		
		// Invalidate old and new hosts
		_ = s.cache.Delete(ctx, oldHostname, result.ShortCode)
		if newHostname != oldHostname {
			_ = s.cache.Delete(ctx, newHostname, result.ShortCode)
		}
	}
	return result, err
}

func (s *LinkService) DeleteLink(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) error {
	var shortCode string
	var hostname string
	if s.cache != nil {
		if existing, err := s.repo.GetLinkByID(ctx, tenantID, id); err == nil && existing != nil {
			shortCode = existing.ShortCode
		}
		if h, err := s.repo.GetHostnameForLink(ctx, id); err == nil && h != nil {
			hostname = *h
		}
	}

	err := s.repo.DeleteLink(ctx, tenantID, id)
	if err == nil && s.cache != nil && shortCode != "" {
		_ = s.cache.Delete(ctx, hostname, shortCode)
	}
	return err
}

func generateShortCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 7)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}
