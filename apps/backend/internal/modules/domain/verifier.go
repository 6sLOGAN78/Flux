// Package domain handles custom domain DNS verification, CNAME checking, and ACME SSL certificate provisioning.
package domain

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CustomDomain represents a custom branded domain configuration.
type CustomDomain struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	Domain             string     `json:"domain"`
	VerificationToken  string     `json:"verification_token"`
	IsVerified         bool       `json:"is_verified"`
	SSLStatus          string     `json:"ssl_status"` // 'pending', 'provisioning', 'active', 'failed', 'expired'
	SSLExpiresAt       *time.Time `json:"ssl_expires_at,omitempty"`
	CustomRootRedirect *string    `json:"custom_root_redirect,omitempty"`
	Custom404Redirect  *string    `json:"custom_404_redirect,omitempty"`
	IsWildcard         bool       `json:"is_wildcard"`
	LastHealthCheck    *time.Time `json:"last_health_check,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// DNSResolver defines the interface for DNS resolution queries.
type DNSResolver interface {
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

// SystemDNSResolver uses standard Go net package for live DNS resolution.
type SystemDNSResolver struct{}

func NewSystemDNSResolver() *SystemDNSResolver {
	return &SystemDNSResolver{}
}

func (r *SystemDNSResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	return net.DefaultResolver.LookupCNAME(ctx, host)
}

func (r *SystemDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return net.DefaultResolver.LookupTXT(ctx, name)
}

// DomainVerifier handles CNAME record validation and ACME SSL status tracking.
type DomainVerifier struct {
	resolver DNSResolver
}

func NewDomainVerifier(resolver DNSResolver) *DomainVerifier {
	if resolver == nil {
		resolver = NewSystemDNSResolver()
	}
	return &DomainVerifier{resolver: resolver}
}

// VerifyCNAME checks if the domain's CNAME record points to the expected target host.
func (v *DomainVerifier) VerifyCNAME(ctx context.Context, domain, expectedCNAME string) (bool, error) {
	cname, err := v.resolver.LookupCNAME(ctx, domain)
	if err != nil {
		return false, nil // Return false on lookup failure rather than erroring out worker
	}

	normalizedCNAME := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(cname)), ".")
	normalizedExpected := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(expectedCNAME)), ".")

	return normalizedCNAME == normalizedExpected, nil
}

// VerifyTXT checks if the domain's TXT record contains the expected verification token.
func (v *DomainVerifier) VerifyTXT(ctx context.Context, domain, expectedToken string) (bool, error) {
	challengeHost := fmt.Sprintf("_flux-challenge.%s", domain)
	txts, err := v.resolver.LookupTXT(ctx, challengeHost)
	if err != nil {
		return false, nil
	}

	for _, txt := range txts {
		if strings.Contains(txt, expectedToken) {
			return true, nil
		}
	}
	return false, nil
}

// CheckSSLRenewal determines if an SSL certificate requires renewal based on remaining validity days threshold.
func (v *DomainVerifier) CheckSSLRenewal(expiresAt time.Time, thresholdDays int) bool {
	if expiresAt.IsZero() {
		return true
	}
	remaining := time.Until(expiresAt)
	threshold := time.Duration(thresholdDays) * 24 * time.Hour
	return remaining <= threshold
}

// ProcessDomainVerification runs verification checks on a domain and updates its state.
func (v *DomainVerifier) ProcessDomainVerification(ctx context.Context, d *CustomDomain, expectedCNAME string) (bool, string, error) {
	verified, err := v.VerifyCNAME(ctx, d.Domain, expectedCNAME)
	if err != nil {
		return false, d.SSLStatus, err
	}

	now := time.Now()
	d.LastHealthCheck = &now

	if verified {
		d.IsVerified = true
		d.SSLStatus = "provisioning"
		return true, "provisioning", nil
	}

	return false, d.SSLStatus, nil
}
