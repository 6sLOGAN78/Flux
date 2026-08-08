package link_test

import (
	"testing"

	"flux/apps/backend/internal/model/link"
)

func TestCreateLinkPayload_Validate(t *testing.T) {
	payload := link.CreateLinkPayload{
		DestinationURL: "https://example.com/target",
	}

	if err := payload.Validate(); err != nil {
		t.Errorf("expected valid CreateLinkPayload, got error: %v", err)
	}

	invalidPayload := link.CreateLinkPayload{
		DestinationURL: "not-a-valid-url",
	}
	if err := invalidPayload.Validate(); err == nil {
		t.Errorf("expected validation error for invalid URL, got nil")
	}
}

func TestGetLinksQuery_Validate(t *testing.T) {
	query := link.GetLinksQuery{}
	if err := query.Validate(); err != nil {
		t.Errorf("expected valid GetLinksQuery, got error: %v", err)
	}

	if *query.Page != 1 {
		t.Errorf("expected default page 1, got %d", *query.Page)
	}
	if *query.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", *query.Limit)
	}
}
