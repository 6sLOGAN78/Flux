package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user mapped from Clerk.
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ClerkUserID string    `json:"clerk_user_id" db:"clerk_user_id"`
	Email       string    `json:"email" db:"email"`
	Name        string    `json:"name" db:"name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Workspace represents a tenant workspace mapped from Clerk Organizations.
type Workspace struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ClerkOrgID       *string   `json:"clerk_org_id" db:"clerk_org_id"` // Nullable for personal workspaces
	Name             string    `json:"name" db:"name"`
	TrackingClientID uuid.UUID `json:"tracking_client_id" db:"tracking_client_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
