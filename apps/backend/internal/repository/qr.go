package repository

import (
	"context"
	"fmt"
)

// MockQRCache provides an in-memory QRCache implementation for testing.
type MockQRCache struct {
	store map[string][]byte
}

// NewMockQRCache initializes an in-memory QRCache.
func NewMockQRCache() *MockQRCache {
	return &MockQRCache{store: make(map[string][]byte)}
}

func (m *MockQRCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, exists := m.store[key]
	if !exists {
		return nil, fmt.Errorf("cache miss")
	}
	return val, nil
}

func (m *MockQRCache) Set(ctx context.Context, key string, data []byte) error {
	m.store[key] = data
	return nil
}
