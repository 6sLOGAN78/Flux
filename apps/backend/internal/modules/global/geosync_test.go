package global_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/global"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKVClient struct {
	updates map[string]string
	deleted map[string]bool
}

func newMockKVClient() *mockKVClient {
	return &mockKVClient{
		updates: make(map[string]string),
		deleted: make(map[string]bool),
	}
}

func (m *mockKVClient) Put(ctx context.Context, slug, targetURL string) error {
	m.updates[slug] = targetURL
	delete(m.deleted, slug)
	return nil
}

func (m *mockKVClient) Delete(ctx context.Context, slug string) error {
	m.deleted[slug] = true
	delete(m.updates, slug)
	return nil
}

func TestGeoSync_BroadcastLinkUpdate(t *testing.T) {
	mockClient := newMockKVClient()
	regions := []string{"us-east", "eu-west", "ap-east"}

	broadcaster := global.NewEdgeKVBroadcaster(regions, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := broadcaster.BroadcastLinkUpdate(ctx, "openai", "https://openai.com")
	require.NoError(t, err)

	assert.Equal(t, "https://openai.com", mockClient.updates["openai"])
	assert.False(t, mockClient.deleted["openai"])
}

func TestGeoSync_BroadcastLinkDelete(t *testing.T) {
	mockClient := newMockKVClient()
	regions := []string{"us-east", "eu-west", "ap-east"}

	mockClient.updates["openai"] = "https://openai.com"

	broadcaster := global.NewEdgeKVBroadcaster(regions, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := broadcaster.BroadcastLinkDelete(ctx, "openai")
	require.NoError(t, err)

	assert.True(t, mockClient.deleted["openai"])
	_, exists := mockClient.updates["openai"]
	assert.False(t, exists)
}

func TestGeoSync_CheckRegionalHealth(t *testing.T) {
	regions := []string{"us-east", "eu-west", "ap-east"}
	cluster := global.NewGeoDBCluster(regions)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	statuses, err := cluster.CheckHealth(ctx)
	require.NoError(t, err)
	assert.Len(t, statuses, 3)

	for _, s := range statuses {
		assert.True(t, s.IsHealthy)
		assert.Less(t, s.LatencyMs, int64(500))
	}
}
