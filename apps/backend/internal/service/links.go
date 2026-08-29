package service

import (
	"context"
	"math/rand"
	"time"

	"flux/apps/backend/internal/model"
	"flux/apps/backend/internal/model/link"
	"flux/apps/backend/internal/repository"

	"github.com/google/uuid"
)

type LinkService struct {
	repo *repository.LinkRepository
}

func NewLinkService(repo *repository.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) CreateLink(ctx context.Context, tenantID *uuid.UUID, payload *link.CreateLinkPayload) (*link.Link, error) {
	var shortCode string
	if payload.CustomCode != nil && *payload.CustomCode != "" {
		shortCode = *payload.CustomCode
	} else {
		shortCode = generateShortCode()
	}
	return s.repo.CreateLink(ctx, tenantID, payload, shortCode)
}

func (s *LinkService) GetLinks(ctx context.Context, tenantID *uuid.UUID, query *link.GetLinksQuery) (*model.PaginatedResponse[link.Link], error) {
	return s.repo.GetLinks(ctx, tenantID, query)
}

func (s *LinkService) GetLinkByID(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) (*link.Link, error) {
	return s.repo.GetLinkByID(ctx, tenantID, id)
}

func (s *LinkService) UpdateLink(ctx context.Context, tenantID *uuid.UUID, payload *link.UpdateLinkPayload) (*link.Link, error) {
	return s.repo.UpdateLink(ctx, tenantID, payload)
}

func (s *LinkService) DeleteLink(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) error {
	return s.repo.DeleteLink(ctx, tenantID, id)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode() string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
