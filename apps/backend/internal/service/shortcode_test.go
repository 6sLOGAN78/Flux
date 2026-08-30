package service

import (
	"regexp"
	"sync"
	"testing"
)

func TestGenerateShortCode_FormatAndLength(t *testing.T) {
	code, err := generateShortCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(code) != 7 {
		t.Errorf("expected length 7, got %d for code %s", len(code), code)
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9]{7}$", code)
	if !matched {
		t.Errorf("expected base62 characters, got %s", code)
	}
}

func TestGenerateShortCode_ConcurrencyAndUniqueness(t *testing.T) {
	// Generate 1000 codes concurrently and ensure no exact duplicates
	// (While mathematically possible to collide, 1000 codes in 62^7 space has negligible chance of collision)
	const count = 1000
	var wg sync.WaitGroup
	codes := make(chan string, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, _ := generateShortCode()
			codes <- code
		}()
	}

	wg.Wait()
	close(codes)

	seen := make(map[string]bool)
	for code := range codes {
		if seen[code] {
			t.Fatalf("found duplicate shortcode during concurrent generation: %s", code)
		}
		seen[code] = true
	}
}
