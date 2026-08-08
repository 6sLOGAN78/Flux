package sso

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSAMLValidator_ValidateAssertion(t *testing.T) {
	validator := NewSAMLValidator()

	config := SSOConfig{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		IDPType:        "saml",
		EntityID:       "https://idp.okta.com/app/exk123",
		SSOURL:         "https://idp.okta.com/app/exk123/sso/saml",
		Certificate:    "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
		EnforceSSO:     true,
	}

	assertion := SAMLAssertion{
		EntityID:     "https://idp.okta.com/app/exk123",
		NameID:       "user@enterprise.com",
		SessionIndex: "sess_123456",
		IssueInstant: time.Now(),
		Attributes: map[string]string{
			"firstName": "John",
			"lastName":  "Doe",
			"role":      "admin",
		},
	}

	res, err := validator.ValidateAssertion(config, assertion)
	if err != nil {
		t.Fatalf("unexpected SAML assertion validation error: %v", err)
	}

	if !res.IsValid {
		t.Errorf("expected valid SAML assertion")
	}

	if res.UserEmail != "user@enterprise.com" {
		t.Errorf("expected user email 'user@enterprise.com', got '%s'", res.UserEmail)
	}
}

func TestSAMLValidator_EntityMismatch(t *testing.T) {
	validator := NewSAMLValidator()

	config := SSOConfig{
		EntityID: "https://idp.okta.com/app/correct",
	}

	assertion := SAMLAssertion{
		EntityID: "https://idp.okta.com/app/wrong",
		NameID:   "user@enterprise.com",
	}

	_, err := validator.ValidateAssertion(config, assertion)
	if err == nil {
		t.Errorf("expected error for entity ID mismatch")
	}
}

func TestSCIMManager_UserProvisioning(t *testing.T) {
	scim := NewSCIMManager()

	newUser := SCIMUser{
		UserName: "alice@acme.com",
		Name:     SCIMName{GivenName: "Alice", FamilyName: "Smith"},
		Emails:   []SCIMEmail{{Value: "alice@acme.com", Primary: true}},
		Active:   true,
	}

	created, err := scim.CreateUser(newUser)
	if err != nil {
		t.Fatalf("unexpected error creating SCIM user: %v", err)
	}

	if created.ID == "" {
		t.Errorf("expected non-empty SCIM user ID")
	}

	// Update user
	created.Active = false
	updated, err := scim.UpdateUser(created.ID, *created)
	if err != nil {
		t.Fatalf("unexpected error updating SCIM user: %v", err)
	}

	if updated.Active != false {
		t.Errorf("expected user active status to be false")
	}
}

func TestIPAllowlistChecker_CIDRValidation(t *testing.T) {
	checker := NewIPAllowlistChecker()

	allowedCIDRs := []string{"192.168.1.0/24", "10.0.0.0/16"}

	if !checker.IsIPAllowed("192.168.1.45", allowedCIDRs) {
		t.Errorf("expected 192.168.1.45 to be allowed")
	}

	if checker.IsIPAllowed("172.16.0.1", allowedCIDRs) {
		t.Errorf("expected 172.16.0.1 to be rejected")
	}
}
