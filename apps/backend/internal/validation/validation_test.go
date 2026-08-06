package validation_test

import (
	"testing"

	"flux/apps/backend/internal/validation"
)

func TestValidateSlug(t *testing.T) {
	validSlugs := []string{"openai", "link-123", "short_code", "ABC09"}
	for _, slug := range validSlugs {
		if err := validation.ValidateSlug(slug); err != nil {
			t.Errorf("expected slug '%s' to be valid, got: %v", slug, err)
		}
	}

	invalidSlugs := []string{"", "invalid slug!", "openai@dev", "bad/slug"}
	for _, slug := range invalidSlugs {
		if err := validation.ValidateSlug(slug); err == nil {
			t.Errorf("expected slug '%s' to be invalid, got nil", slug)
		}
	}
}
