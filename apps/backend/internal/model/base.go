// Package model contains core domain entities and data transfer objects.
package model

import "time"

// BaseModel provides common metadata fields for database entities.
type BaseModel struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
