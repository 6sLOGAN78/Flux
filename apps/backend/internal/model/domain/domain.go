package domain

import "time"

type CustomDomain struct {
	ID                string    `json:"id" db:"id"`
	TenantID          string    `json:"tenant_id" db:"tenant_id"`
	Hostname          string    `json:"hostname" db:"hostname"`
	Status            string    `json:"status" db:"status"`
	VerificationToken string    `json:"verification_token" db:"verification_token"`
	SSLStatus         string    `json:"ssl_status" db:"ssl_status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
