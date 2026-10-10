package job

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// TaskInvitation contains only durable SQL references, never recipient or bearer material.
const invitationTaskDeadline = 15 * time.Second

// TaskInvitation identifies opaque durable invitation references.
const TaskInvitation = "email:invitation"

// NewInvitationTask serializes a current SQL lease reference.
func NewInvitationTask(claim repository.DeliveryClaim) (*asynq.Task, error) {
	if !validClaim(claim) {
		return nil, errors.New("invalid invitation reference")
	}
	payload, err := json.Marshal(claim)
	if err != nil {
		return nil, errors.New("invalid invitation reference")
	}
	// PostgreSQL schedules the eight bounded delivery attempts. Redis retries do
	// not spend another provider attempt; stale references become harmless no-ops.
	return asynq.NewTask(TaskInvitation, payload,
		asynq.MaxRetry(0), asynq.Queue("default"), asynq.Timeout(invitationTaskDeadline)), nil
}

func validClaim(c repository.DeliveryClaim) bool {
	return c.WorkspaceID != uuid.Nil && c.InvitationID != uuid.Nil && c.DeliveryID != uuid.Nil && c.Generation > 0
}

func invitationClaim(payload []byte) (repository.DeliveryClaim, error) {
	var claim repository.DeliveryClaim
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if len(payload) > 512 || decoder.Decode(&claim) != nil || !validClaim(claim) {
		return claim, errors.New("invalid invitation reference")
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return repository.DeliveryClaim{}, errors.New("invalid invitation reference")
	}
	return claim, nil
}
