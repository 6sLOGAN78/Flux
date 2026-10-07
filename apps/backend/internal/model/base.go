// Package model defines shared entity metadata and pagination responses.
package model

import (
	"time"

	"github.com/google/uuid"
)

// BaseWithID contains the canonical entity identifier.
type BaseWithID struct {
	ID uuid.UUID `json:"id" db:"id"`
}

// BaseWithCreatedAt contains an entity creation timestamp.
type BaseWithCreatedAt struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// BaseWithUpdatedAt contains an entity update timestamp.
type BaseWithUpdatedAt struct {
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Base contains common entity identifier and timestamp fields.
type Base struct {
	BaseWithCreatedAt
	BaseWithUpdatedAt
	BaseWithID
}

// PaginatedResponse contains page items and pagination metadata.
type PaginatedResponse[T interface{}] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
