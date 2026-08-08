package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/domain"

	"github.com/google/uuid"
)

type MockDNSResolver struct {
	cnames map[string]string
	txts   map[string][]string
}

func (m *MockDNSResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	if target, exists := m.cnames[host]; exists {
		return target, nil
	}
	return "", errors.New("dns lookup failed: no CNAME record")
}

func (m *MockDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if txts, exists := m.txts[name]; exists {
		return txts, nil
	}
	return nil, errors.New("dns lookup failed: no TXT record")
}

func TestVerifyCNAME_Success(t *testing.T) {
	resolver := &MockDNSResolver{
		cnames: map[string]string{
			"link.acme.com": "cname.flux.dev.",
		},
	}
	verifier := domain.NewDomainVerifier(resolver)

	verified, err := verifier.VerifyCNAME(context.Background(), "link.acme.com", "cname.flux.dev")
	if err != nil {
		t.Fatalf("expected verification to succeed, got error: %v", err)
	}

	if !verified {
		t.Errorf("expected CNAME verification to be true, got false")
	}
}

func TestVerifyCNAME_Mismatch(t *testing.T) {
	resolver := &MockDNSResolver{
		cnames: map[string]string{
			"link.acme.com": "other.domain.com.",
		},
	}
	verifier := domain.NewDomainVerifier(resolver)

	verified, err := verifier.VerifyCNAME(context.Background(), "link.acme.com", "cname.flux.dev")
	if err != nil {
		t.Fatalf("expected no execution error on mismatch, got: %v", err)
	}

	if verified {
		t.Errorf("expected CNAME verification to fail for mismatched target")
	}
}

func TestCheckSSLRenewal(t *testing.T) {
	verifier := domain.NewDomainVerifier(nil)

	// Expires in 10 days -> needs renewal (threshold = 14 days)
	expiresSoon := time.Now().Add(10 * 24 * time.Hour)
	if !verifier.CheckSSLRenewal(expiresSoon, 14) {
		t.Errorf("expected SSL expiring in 10 days to require renewal")
	}

	// Expires in 60 days -> does NOT need renewal
	expiresFar := time.Now().Add(60 * 24 * time.Hour)
	if verifier.CheckSSLRenewal(expiresFar, 14) {
		t.Errorf("expected SSL expiring in 60 days NOT to require renewal")
	}
}

func TestProcessDomainVerification(t *testing.T) {
	resolver := &MockDNSResolver{
		cnames: map[string]string{
			"brand.link.com": "cname.flux.dev.",
		},
	}
	verifier := domain.NewDomainVerifier(resolver)

	d := &domain.CustomDomain{
		ID:                uuid.New(),
		Domain:            "brand.link.com",
		VerificationToken: "flux-verify=xyz123",
		IsVerified:        false,
		SSLStatus:         "pending",
	}

	success, sslStatus, err := verifier.ProcessDomainVerification(context.Background(), d, "cname.flux.dev")
	if err != nil {
		t.Fatalf("expected ProcessDomainVerification to succeed, got error: %v", err)
	}

	if !success {
		t.Errorf("expected domain verification to succeed")
	}
	if sslStatus != "provisioning" {
		t.Errorf("expected SSL status to transition to 'provisioning', got '%s'", sslStatus)
	}
	if !d.IsVerified {
		t.Errorf("expected d.IsVerified to be true")
	}
}
