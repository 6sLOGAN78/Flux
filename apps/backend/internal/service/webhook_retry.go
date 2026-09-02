package service

import (
	"crypto/rand"
	"math/big"
	"time"
)

// CalculateRetryDelay computes exponential backoff with jitter.
// attemptCount starts at 1 for the first failure.
func CalculateRetryDelay(attemptCount int, initialDelay, maxDelay time.Duration) time.Duration {
	if attemptCount < 1 {
		attemptCount = 1
	}

	// Exponential backoff: initialDelay * 2^(attempt-1)
	delay := float64(initialDelay)
	for i := 1; i < attemptCount; i++ {
		delay *= 2
	}

	if time.Duration(delay) > maxDelay {
		delay = float64(maxDelay)
	}

	// Apply 20% jitter
	jitterFactor := 0.2
	jitterRange := delay * jitterFactor * 2 // +/- 20%
	
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	jitterAmount := (float64(n.Int64()) / 10000.0) * jitterRange
	
	finalDelay := delay - (delay * jitterFactor) + jitterAmount
	
	return time.Duration(finalDelay)
}

// IsRetryableError determines if an HTTP delivery error or status code warrants a retry.
func IsRetryableError(statusCode int, err error) bool {
	if err != nil {
		// Network errors / timeouts are retryable. 
		// SSRF is permanent.
		if err.Error() == "ssrf protection: request blocked to private or restricted network" {
			return false
		}
		// If it's a generic connection/timeout err, it is retryable
		return true
	}

	// HTTP Status Classification
	switch statusCode {
	case 429:
		return true // Rate limited
	case 408:
		return true // Request Timeout
	case 400, 401, 403, 404, 410, 422:
		return false // Client permanent errors
	}

	if statusCode >= 500 && statusCode <= 599 {
		// 501 Not Implemented might be permanent, but generally 5xx is retryable.
		if statusCode == 501 {
			return false
		}
		return true
	}

	return false
}
