package service

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"flux/apps/backend/internal/repository"
	"github.com/rs/zerolog/log"
)

type DNSResolver interface {
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

type DefaultDNSResolver struct {
	resolver *net.Resolver
}

func NewDefaultDNSResolver() *DefaultDNSResolver {
	return &DefaultDNSResolver{
		resolver: &net.Resolver{
			PreferGo: true,
		},
	}
}

func (r *DefaultDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return r.resolver.LookupTXT(ctx, name)
}

type DomainVerificationWorker struct {
	repo          *repository.DomainRepository
	cache         repository.RedirectCache
	resolver      DNSResolver
	interval      time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	pendingExpiry time.Duration
}

func NewDomainVerificationWorker(repo *repository.DomainRepository, cache repository.RedirectCache, resolver DNSResolver, interval time.Duration) *DomainVerificationWorker {
	if resolver == nil {
		resolver = NewDefaultDNSResolver()
	}
	if interval == 0 {
		interval = 5 * time.Minute
	}
	return &DomainVerificationWorker{
		repo:          repo,
		cache:         cache,
		resolver:      resolver,
		interval:      interval,
		stopChan:      make(chan struct{}),
		pendingExpiry: 7 * 24 * time.Hour,
	}
}

func (w *DomainVerificationWorker) Start() {
	w.wg.Add(1)
	go w.run()
}

func (w *DomainVerificationWorker) Stop(ctx context.Context) {
	close(w.stopChan)
	
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Warn().Msg("domain verification worker shutdown timed out")
	}
}

func (w *DomainVerificationWorker) run() {
	defer w.wg.Done()
	
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Initial run
	w.processBatch()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.processBatch()
		}
	}
}

func (w *DomainVerificationWorker) processBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	domains, err := w.repo.GetDomainsToVerify(ctx, 100)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch domains for verification")
		return
	}

	for _, d := range domains {
		select {
		case <-w.stopChan:
			return
		default:
		}

		w.verifyDomain(ctx, d.ID, d.Hostname, d.Status, d.VerificationToken, d.CreatedAt)
	}
}

type VerificationResult int

const (
	ResultSuccess VerificationResult = iota
	ResultFailed
	ResultTemporaryError
)

func (w *DomainVerificationWorker) verifyDomain(ctx context.Context, id, hostname, currentStatus, token string, createdAt time.Time) {
	// If domain is pending/verifying for > 7 days, mark as failed
	if time.Since(createdAt) > w.pendingExpiry && (currentStatus == "pending" || currentStatus == "verifying") {
		log.Info().Str("domain_id", id).Str("hostname", hostname).Msg("domain pending expiration reached, marking as failed")
		if err := w.repo.UpdateDomainStatus(ctx, id, "failed"); err != nil {
			log.Error().Err(err).Str("domain_id", id).Msg("failed to update domain status to failed")
		}
		return
	}

	// Move to verifying if it was pending
	if currentStatus == "pending" {
		if err := w.repo.UpdateDomainStatus(ctx, id, "verifying"); err != nil {
			log.Error().Err(err).Str("domain_id", id).Msg("failed to update domain status to verifying")
			return
		}
		currentStatus = "verifying" // Update local state for logic below
	}

	// Verify DNS
	result := w.checkTXT(ctx, hostname, token)

	if result == ResultTemporaryError {
		// Do not downgrade or change status on temporary error
		return
	}

	newStatus := "failed"
	if result == ResultSuccess {
		newStatus = "active"
	}

	// Only active -> failed or verifying -> active/failed should update
	if currentStatus != newStatus {
		// Actually, if it's verifying -> failed, it means they set wrong TXT. We shouldn't permanently fail it until 7 days.
		// Wait, the prompt says "For example: PENDING_DNS -> VERIFYING -> ACTIVE, and VERIFYING -> FAILED. Do not allow arbitrary state manipulation".
		// But "Do not immediately mark a domain permanently failed because a DNS record is temporarily unavailable... Distinguish between record not found... incorrect verification token."
		// If they have the wrong token, it's not verified. We can leave it in 'verifying' and it will keep checking until 7 days, then 'failed'.
		// If it was 'active' and they remove the token, we can change to 'failed' (or 'disabled'). Let's change to 'failed'.
		if currentStatus == "verifying" && newStatus == "failed" {
			// Leave it as verifying so we can retry, until 7 days expire.
			return
		}

		if err := w.repo.UpdateDomainStatus(ctx, id, newStatus); err != nil {
			log.Error().Err(err).Str("domain_id", id).Msg("failed to update domain status")
			return
		}
		log.Info().Str("domain_id", id).Str("hostname", hostname).Str("new_status", newStatus).Msg("domain status updated")

		// If it became inactive, invalidate cache
		if newStatus != "active" {
			if w.cache != nil {
				if err := w.cache.DeleteHost(ctx, hostname); err != nil {
					log.Error().Err(err).Str("hostname", hostname).Msg("failed to invalidate cache for domain")
				}
			}
		}
	}
}

func (w *DomainVerificationWorker) checkTXT(ctx context.Context, hostname, expectedToken string) VerificationResult {
	lookupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	txtRecords, err := w.resolver.LookupTXT(lookupCtx, hostname)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok {
			if dnsErr.IsNotFound {
				// Record explicitly does not exist
				return ResultFailed
			}
			if dnsErr.Timeout() || dnsErr.Temporary() {
				log.Debug().Err(err).Str("hostname", hostname).Msg("dns lookup timeout or temporary error")
				return ResultTemporaryError
			}
		}
		log.Debug().Err(err).Str("hostname", hostname).Msg("dns txt lookup failed")
		return ResultFailed
	}

	for _, txt := range txtRecords {
		if strings.TrimSpace(txt) == expectedToken {
			return ResultSuccess
		}
	}

	return ResultFailed
}
