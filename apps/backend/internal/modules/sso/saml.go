package sso

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SSOConfig struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	IDPType        string    `json:"idp_type"` // 'saml', 'oidc'
	EntityID       string    `json:"entity_id"`
	SSOURL         string    `json:"sso_url"`
	Certificate    string    `json:"certificate"`
	EnforceSSO     bool      `json:"enforce_sso"`
}

type SAMLAssertion struct {
	EntityID     string            `json:"entity_id"`
	NameID       string            `json:"name_id"`
	SessionIndex string            `json:"session_index"`
	IssueInstant time.Time         `json:"issue_instant"`
	Attributes   map[string]string `json:"attributes"`
}

type SAMLValidationResult struct {
	IsValid      bool              `json:"is_valid"`
	UserEmail    string            `json:"user_email"`
	SessionIndex string            `json:"session_index"`
	Attributes   map[string]string `json:"attributes"`
}

type SAMLValidator struct{}

func NewSAMLValidator() *SAMLValidator {
	return &SAMLValidator{}
}

// ValidateAssertion verifies SAML 2.0 Identity Provider assertions against workspace SSO configs.
func (v *SAMLValidator) ValidateAssertion(config SSOConfig, assertion SAMLAssertion) (*SAMLValidationResult, error) {
	if config.EntityID != "" && assertion.EntityID != config.EntityID {
		return nil, fmt.Errorf("SAML entity ID mismatch: expected %s, got %s", config.EntityID, assertion.EntityID)
	}

	if assertion.NameID == "" {
		return nil, fmt.Errorf("SAML assertion missing NameID subject")
	}

	if !assertion.IssueInstant.IsZero() && time.Since(assertion.IssueInstant) > 24*time.Hour {
		return nil, fmt.Errorf("SAML assertion expired")
	}

	return &SAMLValidationResult{
		IsValid:      true,
		UserEmail:    assertion.NameID,
		SessionIndex: assertion.SessionIndex,
		Attributes:   assertion.Attributes,
	}, nil
}

// SCIM 2.0 Protocol Models
type SCIMName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

type SCIMUser struct {
	Schemas    []string    `json:"schemas"`
	ID         string      `json:"id"`
	ExternalID string      `json:"externalId,omitempty"`
	UserName   string      `json:"userName"`
	Name       SCIMName    `json:"name"`
	Emails     []SCIMEmail `json:"emails"`
	Active     bool        `json:"active"`
}

type SCIMMember struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
}

type SCIMGroup struct {
	Schemas     []string     `json:"schemas"`
	ID          string       `json:"id"`
	DisplayName string       `json:"displayName"`
	Members     []SCIMMember `json:"members"`
}

type SCIMManager struct {
	mu    sync.RWMutex
	users map[string]SCIMUser
}

func NewSCIMManager() *SCIMManager {
	return &SCIMManager{
		users: make(map[string]SCIMUser),
	}
}

func (m *SCIMManager) CreateUser(user SCIMUser) (*SCIMUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	m.users[user.ID] = user
	return &user, nil
}

func (m *SCIMManager) UpdateUser(id string, user SCIMUser) (*SCIMUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		return nil, fmt.Errorf("SCIM user %s not found", id)
	}

	user.ID = id
	m.users[id] = user
	return &user, nil
}

func (m *SCIMManager) DeleteUser(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		return fmt.Errorf("SCIM user %s not found", id)
	}

	delete(m.users, id)
	return nil
}

type IPAllowlistChecker struct{}

func NewIPAllowlistChecker() *IPAllowlistChecker {
	return &IPAllowlistChecker{}
}

// IsIPAllowed evaluates client IP addresses against an array of CIDR blocks.
func (c *IPAllowlistChecker) IsIPAllowed(clientIP string, cidrBlocks []string) bool {
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return false
	}

	for _, cidr := range cidrBlocks {
		_, ipNet, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}
