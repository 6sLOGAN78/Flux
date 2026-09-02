package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// GenerateHMACSHA256 creates a hex-encoded HMAC-SHA256 signature for the given body and secret.
func GenerateHMACSHA256(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSHA256 securely verifies a hex-encoded HMAC-SHA256 signature in constant time.
func VerifyHMACSHA256(secret string, body []byte, signature string) bool {
	expectedMAC := GenerateHMACSHA256(secret, body)
	return subtle.ConstantTimeCompare([]byte(expectedMAC), []byte(signature)) == 1
}
