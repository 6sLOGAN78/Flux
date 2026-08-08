package global_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/global"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnycast_CheckPoPHealth(t *testing.T) {
	pops := []global.PoPNode{
		{ID: "pop-us-east-1", Region: "us-east", AnycastIP: "198.51.100.1"},
		{ID: "pop-eu-west-1", Region: "eu-west", AnycastIP: "198.51.100.2"},
		{ID: "pop-ap-east-1", Region: "ap-east", AnycastIP: "198.51.100.3"},
	}

	monitor := global.NewAnycastMonitor(pops)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	results, err := monitor.CheckPoPHealth(ctx)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	for _, pop := range results {
		assert.True(t, pop.IsHealthy)
		assert.Equal(t, "advertised", pop.BGPState)
	}
}

func TestAnycast_WithdrawAndAdvertisePoP(t *testing.T) {
	pops := []global.PoPNode{
		{ID: "pop-us-east-1", Region: "us-east", AnycastIP: "198.51.100.1"},
	}

	monitor := global.NewAnycastMonitor(pops)
	ctx := context.Background()

	err := monitor.WithdrawUnhealthyPoP(ctx, "pop-us-east-1")
	require.NoError(t, err)

	status, err := monitor.GetPoPStatus(ctx, "pop-us-east-1")
	require.NoError(t, err)
	assert.Equal(t, "withdrawn", status.BGPState)

	err = monitor.AdvertiseHealthyPoP(ctx, "pop-us-east-1")
	require.NoError(t, err)

	status, err = monitor.GetPoPStatus(ctx, "pop-us-east-1")
	require.NoError(t, err)
	assert.Equal(t, "advertised", status.BGPState)
}

func TestTLSMeshManager_DeployAndRenewCertificate(t *testing.T) {
	manager := global.NewTLSMeshManager()
	ctx := context.Background()

	domain := "*.flux.dev"
	cert, err := manager.DeployWildcardCertificate(ctx, domain, "MOCK_CERT_PEM", "MOCK_KEY_PEM")
	require.NoError(t, err)
	assert.Equal(t, domain, cert.Domain)
	assert.Equal(t, "active", cert.Status)
	assert.NotEmpty(t, cert.Fingerprint)

	renewedCert, err := manager.RenewCertificate(ctx, domain)
	require.NoError(t, err)
	assert.Equal(t, domain, renewedCert.Domain)
	assert.Equal(t, "active", renewedCert.Status)
}
