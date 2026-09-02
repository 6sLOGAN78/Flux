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

// GenerateWebhookSecret securely generates a random signing secret for webhooks.
func GenerateWebhookSecret() (string, error) {
	// Generate 32 bytes of secure random entropy (64 hex characters)
	hexStr, err := GenerateRandomHex(32)
	if err != nil {
		return "", err
	}
	return "whsec_" + hexStr, nil
}
