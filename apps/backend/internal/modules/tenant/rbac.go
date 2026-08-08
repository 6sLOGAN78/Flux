// Package tenant provides multi-tenant Organization/Workspace scoping and Role-Based Access Control (RBAC).
package tenant

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Permission constants
const (
	PermissionRead            = "read"
	PermissionWrite           = "write"
	PermissionAdmin           = "admin"
	PermissionManageBilling   = "manage_billing"
	PermissionDeleteWorkspace = "delete_workspace"
)

// Organization represents an enterprise or SaaS organization account.
type Organization struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	BillingEmail string    `json:"billing_email" db:"billing_email"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Workspace represents an isolated tenant workspace within an organization.
type Workspace struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Slug           string    `json:"slug" db:"slug"`
	IsDefault      bool      `json:"is_default" db:"is_default"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// WorkspaceMember represents a user membership within a workspace with RBAC role.
type WorkspaceMember struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	WorkspaceID       uuid.UUID       `json:"workspace_id" db:"workspace_id"`
	UserID            uuid.UUID       `json:"user_id" db:"user_id"`
	Role              string          `json:"role" db:"role"`
	CustomPermissions map[string]bool `json:"custom_permissions,omitempty" db:"custom_permissions"`
	InvitedBy         *uuid.UUID      `json:"invited_by,omitempty" db:"invited_by"`
	JoinedAt          time.Time       `json:"joined_at" db:"joined_at"`
}

// IsValidRole validates if a role string is a supported RBAC role.
func IsValidRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleOwner, RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

// HasPermission checks if the given role possesses a requested permission.
func HasPermission(role, permission string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	permission = strings.ToLower(strings.TrimSpace(permission))

	switch role {
	case RoleOwner:
		return true
	case RoleAdmin:
		return permission == PermissionRead || permission == PermissionWrite || permission == PermissionAdmin
	case RoleEditor:
		return permission == PermissionRead || permission == PermissionWrite
	case RoleViewer:
		return permission == PermissionRead
	default:
		return false
	}
}

// HasPermissionWithCustom evaluates permission including custom permission overrides.
func HasPermissionWithCustom(role, permission string, custom map[string]bool) bool {
	if custom != nil {
		if allowed, exists := custom[permission]; exists {
			return allowed
		}
	}
	return HasPermission(role, permission)
}
