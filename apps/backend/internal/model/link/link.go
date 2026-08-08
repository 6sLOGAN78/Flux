package link

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ShortCode      string     `json:"short_code" db:"short_code"`
	DestinationURL string     `json:"destination_url" db:"destination_url"`
	TenantID       *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	Title          *string    `json:"title,omitempty" db:"title"`
	Description    *string    `json:"description,omitempty" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}
