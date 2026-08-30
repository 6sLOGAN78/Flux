package campaign

import (
	"time"

	"github.com/google/uuid"
)

type Campaign struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Status      string    `json:"status" db:"status"` // active, paused, archived
	UTMCampaign *string   `json:"utm_campaign,omitempty" db:"utm_campaign"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
