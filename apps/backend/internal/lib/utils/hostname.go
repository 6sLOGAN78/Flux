package utils

import "strings"

// NormalizeHostname strips ports, lowercases, and removes trailing dots.
func NormalizeHostname(host string) string {
	// Remove port if present
	if idx := strings.IndexByte(host, ':'); idx != -1 {
		host = host[:idx]
	}

	host = strings.ToLower(host)

	// Remove trailing dot
	if strings.HasSuffix(host, ".") {
		host = host[:len(host)-1]
	}

	return host
}
