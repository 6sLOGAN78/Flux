// Package validation provides request data validation helpers.
package validation

import (
	"errors"
	"strings"
)

var (
	ErrInvalidSlug = errors.New("slug must contain only alphanumeric characters, hyphens, or underscores")
)

// ValidateSlug validates short link slug syntax.
func ValidateSlug(slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return errors.New("slug cannot be empty")
	}
	if len(slug) > 64 {
		return errors.New("slug length cannot exceed 64 characters")
	}
	for _, ch := range slug {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return ErrInvalidSlug
		}
	}
	return nil
}
