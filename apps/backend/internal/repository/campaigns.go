package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flux/apps/backend/internal/model/campaign"
	"flux/apps/backend/pkg/sqlerr"
)

type CampaignRepository struct {
	db *pgxpool.Pool
}

func NewCampaignRepository(db *pgxpool.Pool) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) CreateCampaign(ctx context.Context, camp *campaign.Campaign) error {
	query := `
		INSERT INTO campaigns (id, workspace_id, name, status, utm_campaign)
		VALUES (@id, @workspace_id, @name, @status, @utm_campaign)
		RETURNING created_at, updated_at
	`
	args := pgx.NamedArgs{
		"id":           camp.ID,
		"workspace_id": camp.WorkspaceID,
		"name":         camp.Name,
		"status":       camp.Status,
		"utm_campaign": camp.UTMCampaign,
	}

	err := r.db.QueryRow(ctx, query, args).Scan(&camp.CreatedAt, &camp.UpdatedAt)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	return nil
}

func (r *CampaignRepository) GetCampaign(ctx context.Context, workspaceID, id uuid.UUID) (*campaign.Campaign, error) {
	query := `
		SELECT id, workspace_id, name, status, utm_campaign, created_at, updated_at
		FROM campaigns
		WHERE id = $1 AND workspace_id = $2
	`
	var camp campaign.Campaign
	err := r.db.QueryRow(ctx, query, id, workspaceID).Scan(
		&camp.ID,
		&camp.WorkspaceID,
		&camp.Name,
		&camp.Status,
		&camp.UTMCampaign,
		&camp.CreatedAt,
		&camp.UpdatedAt,
	)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return &camp, nil
}

func (r *CampaignRepository) ListCampaigns(ctx context.Context, workspaceID uuid.UUID) ([]campaign.Campaign, error) {
	query := `
		SELECT id, workspace_id, name, status, utm_campaign, created_at, updated_at
		FROM campaigns
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	defer rows.Close()

	var campaigns []campaign.Campaign
	for rows.Next() {
		var camp campaign.Campaign
		if err := rows.Scan(
			&camp.ID,
			&camp.WorkspaceID,
			&camp.Name,
			&camp.Status,
			&camp.UTMCampaign,
			&camp.CreatedAt,
			&camp.UpdatedAt,
		); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		campaigns = append(campaigns, camp)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return campaigns, nil
}

func (r *CampaignRepository) UpdateCampaign(ctx context.Context, camp *campaign.Campaign) error {
	query := `
		UPDATE campaigns
		SET name = @name, status = @status, utm_campaign = @utm_campaign
		WHERE id = @id AND workspace_id = @workspace_id
		RETURNING updated_at
	`
	args := pgx.NamedArgs{
		"id":           camp.ID,
		"workspace_id": camp.WorkspaceID,
		"name":         camp.Name,
		"status":       camp.Status,
		"utm_campaign": camp.UTMCampaign,
	}

	err := r.db.QueryRow(ctx, query, args).Scan(&camp.UpdatedAt)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	return nil
}

func (r *CampaignRepository) DeleteCampaign(ctx context.Context, workspaceID, id uuid.UUID) error {
	query := `
		DELETE FROM campaigns
		WHERE id = $1 AND workspace_id = $2
	`
	cmd, err := r.db.Exec(ctx, query, id, workspaceID)
	if err != nil {
		return sqlerr.HandleError(err)
	}
	if cmd.RowsAffected() == 0 {
		return sqlerr.HandleError(pgx.ErrNoRows)
	}
	return nil
}
