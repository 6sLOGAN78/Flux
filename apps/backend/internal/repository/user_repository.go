package repository

import (
	"context"
	"fmt"
	"strings"

	"flux/apps/backend/internal/model/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// SyncUser finds an existing user by Clerk ID or creates a new one.
func (r *UserRepository) SyncUser(ctx context.Context, clerkID, email, name string) (*user.User, error) {
	if strings.TrimSpace(name) == "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			name = parts[0]
		}
	}

	query := `
		INSERT INTO users (clerk_user_id, email, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (clerk_user_id) DO UPDATE 
		SET email = EXCLUDED.email, name = EXCLUDED.name, updated_at = CURRENT_TIMESTAMP
		RETURNING id, clerk_user_id, email, name, created_at, updated_at
	`
	
	u := &user.User{}
	err := r.db.QueryRow(ctx, query, clerkID, email, name).Scan(
		&u.ID, &u.ClerkUserID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("syncing user: %w", err)
	}
	return u, nil
}

// SyncWorkspace finds an existing workspace by Clerk Org ID or creates a new one.
// If clerkOrgID is empty, it returns the user's personal workspace (null clerk_org_id).
func (r *UserRepository) SyncWorkspace(ctx context.Context, clerkOrgID, name string, userID uuid.UUID) (*user.Workspace, error) {
	var query string
	var args []interface{}

	if clerkOrgID == "" {
		// Personal workspace fallback for the user
		
		findPersonalQuery := `
			SELECT w.id, w.clerk_org_id, w.name, w.created_at, w.updated_at
			FROM workspaces w
			JOIN workspace_members wm ON wm.workspace_id = w.id
			WHERE w.clerk_org_id IS NULL AND wm.user_id = $1
			LIMIT 1
		`
		w := &user.Workspace{}
		err := r.db.QueryRow(ctx, findPersonalQuery, userID).Scan(&w.ID, &w.ClerkOrgID, &w.Name, &w.CreatedAt, &w.UpdatedAt)
		if err == nil {
			return w, nil
		}
		
		query = `
			WITH w AS (
				INSERT INTO workspaces (clerk_org_id, name)
				VALUES (NULL, $1)
				RETURNING id, clerk_org_id, name, created_at, updated_at
			),
			wm AS (
				INSERT INTO workspace_members (workspace_id, user_id, role)
				SELECT id, $2, 'owner' FROM w
			)
			SELECT id, clerk_org_id, name, created_at, updated_at FROM w
		`
		args = []interface{}{name, userID}
	} else {
		// Org workspace
		query = `
			INSERT INTO workspaces (clerk_org_id, name)
			VALUES ($1, $2)
			ON CONFLICT (clerk_org_id) DO UPDATE 
			SET name = EXCLUDED.name, updated_at = CURRENT_TIMESTAMP
			RETURNING id, clerk_org_id, name, created_at, updated_at
		`
		args = []interface{}{clerkOrgID, name}
	}

	w := &user.Workspace{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&w.ID, &w.ClerkOrgID, &w.Name, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("syncing workspace: %w", err)
	}

	// Ensure membership exists
	if clerkOrgID != "" {
		_, err = r.db.Exec(ctx, `
			INSERT INTO workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, 'admin')
			ON CONFLICT (workspace_id, user_id) DO NOTHING
		`, w.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("adding workspace member: %w", err)
		}
	}

	return w, nil
}
