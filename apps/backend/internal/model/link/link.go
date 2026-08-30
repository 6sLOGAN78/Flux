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
	CampaignID     *uuid.UUID `json:"campaign_id,omitempty" db:"campaign_id"`
	UTMSource      *string    `json:"utm_source,omitempty" db:"utm_source"`
	UTMMedium      *string    `json:"utm_medium,omitempty" db:"utm_medium"`
	UTMCampaign    *string    `json:"utm_campaign,omitempty" db:"utm_campaign"`
	UTMTerm        *string    `json:"utm_term,omitempty" db:"utm_term"`
	UTMContent     *string    `json:"utm_content,omitempty" db:"utm_content"`
	CustomDomainID *uuid.UUID `json:"custom_domain_id,omitempty" db:"custom_domain_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}
