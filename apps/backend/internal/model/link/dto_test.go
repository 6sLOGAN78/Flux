package link_test

import (
	"testing"

	"flux/apps/backend/internal/model/link"

	"github.com/google/uuid"
)

func TestCreateLinkPayload_WithCategoryID(t *testing.T) {
	catID := uuid.New()
	payload := link.CreateLinkPayload{
		DestinationURL: "https://example.com/summer-sale",
		CategoryID:     &catID,
	}

	if err := payload.Validate(); err != nil {
		t.Fatalf("expected payload with category_id to be valid, got: %v", err)
	}
}

func TestUpdateLinkPayload_WithCategoryID(t *testing.T) {
	linkID := uuid.New()
	catID := uuid.New()
	payload := link.UpdateLinkPayload{
		ID:         linkID,
		CategoryID: &catID,
	}

	if err := payload.Validate(); err != nil {
		t.Fatalf("expected update payload with category_id to be valid, got: %v", err)
	}
}

func TestBulkCategorizePayload_Validation(t *testing.T) {
	catID := uuid.New()
	payload := link.BulkCategorizePayload{
		LinkIDs:    []uuid.UUID{uuid.New(), uuid.New()},
		CategoryID: &catID,
	}

	if err := payload.Validate(); err != nil {
		t.Fatalf("expected bulk categorize payload to be valid, got: %v", err)
	}

	invalidPayload := link.BulkCategorizePayload{
		LinkIDs: []uuid.UUID{},
	}
	if err := invalidPayload.Validate(); err == nil {
		t.Fatal("expected empty link_ids to fail validation")
	}
}

func TestGetLinksQuery_WithCategoryID(t *testing.T) {
	catID := uuid.New()
	query := link.GetLinksQuery{
		CategoryID: &catID,
	}

	if err := query.Validate(); err != nil {
		t.Fatalf("expected get links query with category_id to be valid, got: %v", err)
	}

	if *query.Page != 1 || *query.Limit != 20 {
		t.Fatalf("expected query defaults page=1 limit=20, got page=%d limit=%d", *query.Page, *query.Limit)
	}
}
