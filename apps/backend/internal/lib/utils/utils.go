// Package utils provides common utility functions.
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
)

// GenerateRandomHex generates a crypto-secure random hex string of given byte length.
func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsValidURL checks whether a raw string is a valid HTTP/HTTPS URL.
func IsValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
