//nolint:testpackage // Verify worker-owned resource factories and actual dependencies.
package app

import (
	"context"
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	backendTesting "github.com/6sLOGAN78/flux/internal/testing"
)

func TestInvitationWorkerActualRedisPostgres(t *testing.T) {
	pg, closePG := backendTesting.SetupTestDB(t)
	defer closePG()
	queue, closeQueue := backendTesting.SetupTestRedis(t)
	defer closeQueue()
	cfg := roleTestConfig()
	cfg.Database, cfg.Server, cfg.Redis = pg.Config.Database, pg.Config.Server, queue.Config
	cfg.Integration = config.IntegrationConfig{ResendAPIKey: "test-only"}
	worker, err := NewWorker(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := worker.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	if worker.Server.DB == nil {
		t.Fatal("durable invitation worker does not own PostgreSQL")
	}
}
