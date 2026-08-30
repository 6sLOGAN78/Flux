package service_test

import (
	"context"
	"net"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResolver struct {
	txtRecords map[string][]string
	err        error
	delay      time.Duration
}

func (m *mockResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	records, ok := m.txtRecords[name]
	if !ok {
		return nil, &net.DNSError{IsNotFound: true}
	}
	return records, nil
}

func TestDomainVerificationWorker(t *testing.T) {
	ctx := context.Background()
	
	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	pool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)
	defer pool.Close()

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	domainRepo := repository.NewDomainRepository(pool)
	
	wsID := uuid.New().String()
	_, err = pool.Exec(ctx, "INSERT INTO workspaces (id, name) VALUES ($1, $2)", wsID, "Test Workspace")
	require.NoError(t, err)

	setupDomain := func(hostname, token, status string, overrideCreatedAt ...time.Time) *repository.DomainRepository {
		d, err := domainRepo.CreateDomain(ctx, wsID, hostname, token)
		require.NoError(t, err)

		err = domainRepo.UpdateDomainStatus(ctx, d.ID, status)
		require.NoError(t, err)

		if len(overrideCreatedAt) > 0 {
			_, err = pool.Exec(ctx, "UPDATE custom_domains SET created_at = $1 WHERE id = $2", overrideCreatedAt[0], d.ID)
			require.NoError(t, err)
		}

		return domainRepo
	}

	t.Run("success verifying valid token", func(t *testing.T) {
		hostname := "valid.flux.ly"
		token := "flux-verify=123"
		setupDomain(hostname, token, "pending")

		res := &mockResolver{
			txtRecords: map[string][]string{
				hostname: {token},
			},
		}

		worker := service.NewDomainVerificationWorker(domainRepo, nil, res, 100*time.Millisecond)
		worker.Start()
		time.Sleep(200 * time.Millisecond)
		worker.Stop(ctx)

		domains, _ := domainRepo.GetDomainsByTenant(ctx, wsID)
		for _, d := range domains {
			if d.Hostname == hostname {
				assert.Equal(t, "active", d.Status)
			}
		}
	})

	t.Run("fails on invalid token, but remains verifying until expiry", func(t *testing.T) {
		hostname := "invalid.flux.ly"
		token := "flux-verify=abc"
		setupDomain(hostname, token, "pending")

		res := &mockResolver{
			txtRecords: map[string][]string{
				hostname: {"flux-verify=wrong"},
			},
		}

		worker := service.NewDomainVerificationWorker(domainRepo, nil, res, 100*time.Millisecond)
		worker.Start()
		time.Sleep(200 * time.Millisecond)
		worker.Stop(ctx)

		domains, _ := domainRepo.GetDomainsByTenant(ctx, wsID)
		for _, d := range domains {
			if d.Hostname == hostname {
				assert.Equal(t, "verifying", d.Status)
			}
		}
	})

	t.Run("temporary dns error does not transition out of verifying", func(t *testing.T) {
		hostname := "timeout.flux.ly"
		setupDomain(hostname, "token", "verifying")

		res := &mockResolver{
			err: &net.DNSError{IsTimeout: true}, // Temporary error
		}

		worker := service.NewDomainVerificationWorker(domainRepo, nil, res, 100*time.Millisecond)
		worker.Start()
		time.Sleep(200 * time.Millisecond)
		worker.Stop(ctx)

		domains, _ := domainRepo.GetDomainsByTenant(ctx, wsID)
		for _, d := range domains {
			if d.Hostname == hostname {
				assert.Equal(t, "verifying", d.Status)
			}
		}
	})

	t.Run("pending expiration marks as failed", func(t *testing.T) {
		hostname := "expired.flux.ly"
		oldTime := time.Now().Add(-8 * 24 * time.Hour)
		setupDomain(hostname, "token", "pending", oldTime)

		res := &mockResolver{
			txtRecords: map[string][]string{
				hostname: {"token"}, // even if dns is valid, it's expired
			},
		}

		worker := service.NewDomainVerificationWorker(domainRepo, nil, res, 100*time.Millisecond)
		worker.Start()
		time.Sleep(200 * time.Millisecond)
		worker.Stop(ctx)

		domains, _ := domainRepo.GetDomainsByTenant(ctx, wsID)
		for _, d := range domains {
			if d.Hostname == hostname {
				assert.Equal(t, "failed", d.Status)
			}
		}
	})

	t.Run("active domain fails if token removed", func(t *testing.T) {
		hostname := "removed.flux.ly"
		setupDomain(hostname, "token", "active")

		res := &mockResolver{
			err: &net.DNSError{IsNotFound: true}, // Record deleted
		}

		worker := service.NewDomainVerificationWorker(domainRepo, nil, res, 100*time.Millisecond)
		worker.Start()
		time.Sleep(200 * time.Millisecond)
		worker.Stop(ctx)

		domains, _ := domainRepo.GetDomainsByTenant(ctx, wsID)
		for _, d := range domains {
			if d.Hostname == hostname {
				assert.Equal(t, "failed", d.Status)
			}
		}
	})
}
