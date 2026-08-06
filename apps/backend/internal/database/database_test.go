package database_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
)

func TestInitDBPool_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Invalid connection string should fail ping/connect
	_, err := database.InitDBPool(ctx, "postgres://invalid_user:invalid_pass@localhost:54321/non_existent_db?sslmode=disable")
	if err == nil {
		t.Errorf("expected error connecting to invalid database URL, got nil")
	}
}
