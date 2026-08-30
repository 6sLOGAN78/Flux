package service

import (
	"context"

	"github.com/google/uuid"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model/campaign"
	"flux/apps/backend/internal/repository"
)

type CampaignService struct {
	repo       *repository.CampaignRepository
	linkRepo   *repository.LinkRepository
	cache      repository.RedirectCache
}

func NewCampaignService(repo *repository.CampaignRepository, linkRepo *repository.LinkRepository, cache repository.RedirectCache) *CampaignService {
	return &CampaignService{repo: repo, linkRepo: linkRepo, cache: cache}
}

func (s *CampaignService) CreateCampaign(ctx context.Context, workspaceID uuid.UUID, payload *campaign.CreateCampaignPayload) (*campaign.Campaign, error) {
	if err := payload.Validate(); err != nil {
		return nil, errs.NewBadRequestError("validation failed", true, nil, nil, err)
	}

	camp := &campaign.Campaign{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        payload.Name,
		Status:      "active",
		UTMCampaign: payload.UTMCampaign,
	}

	if err := s.repo.CreateCampaign(ctx, camp); err != nil {
		return nil, err
	}

	return camp, nil
}

func (s *CampaignService) GetCampaign(ctx context.Context, workspaceID, id uuid.UUID) (*campaign.Campaign, error) {
	return s.repo.GetCampaign(ctx, workspaceID, id)
}

func (s *CampaignService) ListCampaigns(ctx context.Context, workspaceID uuid.UUID) ([]campaign.Campaign, error) {
	return s.repo.ListCampaigns(ctx, workspaceID)
}

func (s *CampaignService) UpdateCampaign(ctx context.Context, workspaceID uuid.UUID, payload *campaign.UpdateCampaignPayload) (*campaign.Campaign, error) {
	if err := payload.Validate(); err != nil {
		return nil, errs.NewBadRequestError("validation failed", true, nil, nil, err)
	}

	camp, err := s.repo.GetCampaign(ctx, workspaceID, payload.ID)
	if err != nil {
		return nil, err
	}

	if payload.Name != nil {
		camp.Name = *payload.Name
	}
	if payload.Status != nil {
		camp.Status = *payload.Status
	}
	
	utmChanged := false
	if payload.UTMCampaign != nil {
		if camp.UTMCampaign == nil || *camp.UTMCampaign != *payload.UTMCampaign {
			utmChanged = true
		}
		camp.UTMCampaign = payload.UTMCampaign
	} else {
		// If payload provides explicit null (if supported), but actually payload.UTMCampaign is pointer to string.
		// Wait, if it's not provided, we don't change it. So UTM doesn't change unless it's in payload.
	}

	if err := s.repo.UpdateCampaign(ctx, camp); err != nil {
		return nil, err
	}

	if utmChanged && s.cache != nil && s.linkRepo != nil {
		s.invalidateCampaignLinksCache(ctx, camp.ID)
	}

	return camp, nil
}

func (s *CampaignService) DeleteCampaign(ctx context.Context, workspaceID, id uuid.UUID) error {
	// Need to invalidate cache BEFORE deleting, or after?
	// If after, the links still exist, just campaign_id is NULL. So we can get short codes BEFORE deleting!
	var cacheKeys []repository.LinkCacheKey
	if s.linkRepo != nil && s.cache != nil {
		cacheKeys, _ = s.linkRepo.GetCacheKeysByCampaign(ctx, id)
	}

	err := s.repo.DeleteCampaign(ctx, workspaceID, id)
	
	if err == nil && len(cacheKeys) > 0 {
		for _, key := range cacheKeys {
			_ = s.cache.Delete(ctx, key.Hostname, key.Slug)
		}
	}
	
	return err
}

func (s *CampaignService) invalidateCampaignLinksCache(ctx context.Context, campaignID uuid.UUID) {
	if s.linkRepo == nil || s.cache == nil {
		return
	}
	
	cacheKeys, err := s.linkRepo.GetCacheKeysByCampaign(ctx, campaignID)
	if err == nil {
		for _, key := range cacheKeys {
			_ = s.cache.Delete(ctx, key.Hostname, key.Slug)
		}
	}
}
