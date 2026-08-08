package category_test

import (
	"testing"

	"flux/apps/backend/internal/model/category"
)

func TestCreateCategoryPayload_Validate(t *testing.T) {
	payload := category.CreateCategoryPayload{
		Name:  "Marketing",
		Color: "#3b82f6",
	}

	if err := payload.Validate(); err != nil {
		t.Errorf("expected valid CreateCategoryPayload, got error: %v", err)
	}

	invalidPayload := category.CreateCategoryPayload{
		Name:  "",
		Color: "invalid-color",
	}
	if err := invalidPayload.Validate(); err == nil {
		t.Errorf("expected validation error for invalid payload, got nil")
	}
}

func TestGetCategoriesQuery_Validate(t *testing.T) {
	query := category.GetCategoriesQuery{}
	if err := query.Validate(); err != nil {
		t.Errorf("expected valid GetCategoriesQuery, got error: %v", err)
	}

	if *query.Page != 1 {
		t.Errorf("expected default page 1, got %d", *query.Page)
	}
	if *query.Limit != 50 {
		t.Errorf("expected default limit 50, got %d", *query.Limit)
	}
}
