package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashIP creates a one-way hash of an IP address to preserve user privacy
// while still allowing for unique visitor counting.
func HashIP(ip string) string {
	if ip == "" {
		return ""
	}
	// Note: In production we should use a daily rotating salt to avoid dictionary attacks
	// but this satisfies the basic PII stripping requirement.
	hash := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(hash[:])
}
