package webhook

import (
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	EndpointURL string    `json:"endpoint_url" db:"endpoint_url"`
	Secret      string    `json:"secret,omitempty" db:"secret"` // Omit from general JSON output by default
	Active      bool      `json:"active" db:"active"`
	Events      []string  `json:"events" db:"events"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateWebhookPayload struct {
	EndpointURL string   `json:"endpoint_url" validate:"required,url"`
	Events      []string `json:"events" validate:"required,min=1"`
}

type UpdateWebhookPayload struct {
	EndpointURL *string   `json:"endpoint_url,omitempty" validate:"omitempty,url"`
	Events      *[]string `json:"events,omitempty" validate:"omitempty,min=1"`
	Active      *bool     `json:"active,omitempty"`
}
