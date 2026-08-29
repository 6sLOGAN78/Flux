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
// If clerkOrgID is empty, it returns the user's personal workspace (null clerk_org_id, bound to owner_id).
func (r *UserRepository) SyncWorkspace(ctx context.Context, clerkOrgID, name string, userID uuid.UUID) (*user.Workspace, error) {
	var query string
	var args []interface{}

	if clerkOrgID == "" {
		// Personal workspace fallback for the user
		findPersonalQuery := `
			SELECT id, clerk_org_id, name, created_at, updated_at
			FROM workspaces
			WHERE clerk_org_id IS NULL AND owner_id = $1
			LIMIT 1
		`
		w := &user.Workspace{}
		err := r.db.QueryRow(ctx, findPersonalQuery, userID).Scan(&w.ID, &w.ClerkOrgID, &w.Name, &w.CreatedAt, &w.UpdatedAt)
		if err == nil {
			return w, nil
		}
		
		query = `
			INSERT INTO workspaces (clerk_org_id, name, owner_id)
			VALUES (NULL, $1, $2)
			RETURNING id, clerk_org_id, name, created_at, updated_at
		`
		args = []interface{}{name, userID}
	} else {
		// Org workspace (Source of truth is Clerk; no owner_id in Postgres, no local membership table)
		query = `
			INSERT INTO workspaces (clerk_org_id, name, owner_id)
			VALUES ($1, $2, NULL)
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

	return w, nil
}
