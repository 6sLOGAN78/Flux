package audit

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty"`
	UserID         uuid.UUID              `json:"user_id"`
	Action         string                 `json:"action"` // e.g. "link.created", "member.invited", "domain.verified"
	ResourceType   string                 `json:"resource_type"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	IPAddress      string                 `json:"ip_address,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type AuditLogger struct {
	mu   sync.RWMutex
	logs []*AuditLog
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]*AuditLog, 0),
	}
}

// Log records an immutable audit log entry.
func (l *AuditLogger) Log(entry *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	l.logs = append(l.logs, entry)
	return nil
}

// GetLogsByOrg retrieves audit logs associated with a specific organization.
func (l *AuditLogger) GetLogsByOrg(orgID uuid.UUID) ([]*AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]*AuditLog, 0)
	for _, entry := range l.logs {
		if entry.OrganizationID != nil && *entry.OrganizationID == orgID {
			result = append(result, entry)
		}
	}
	return result, nil
}

// GetAllLogs returns all audit trail entries for super-admin inspection.
func (l *AuditLogger) GetAllLogs() ([]*AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]*AuditLog, len(l.logs))
	copy(result, l.logs)
	return result, nil
}
