package job

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/invitationcrypto"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const invitationSendDeadline = 10 * time.Second

// InvitationSender is the context-bounded provider acknowledgement boundary.
type InvitationSender interface {
	SendInvitation(context.Context, invitationcrypto.Message) error
}

func invitationIdentity(c repository.DeliveryClaim) string {
	return "flux-invitation/" + c.WorkspaceID.String() + "/" + c.DeliveryID.String()
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (j *JobService) prepareInvitation(ctx context.Context, claim repository.DeliveryClaim) error {
	return j.deliveries.Prepare(ctx, claim, func(intent repository.DeliveryIntent, workspace, role string) ([]byte, error) {
		envelope, err1 := invitationcrypto.Open(j.invitationConfig, claim.WorkspaceID, claim.InvitationID, claim.DeliveryID, intent.KeyID, intent.Ciphertext)
		if err1 != nil {
			return nil, err1
		}
		if err1 = j.validateInvitation(envelope); err1 != nil {
			return nil, err1
		}
		if envelope.Message != nil {
			return append([]byte(nil), intent.Ciphertext...), nil
		}
		body, err1 := (&email.Client{}).Render(email.TemplateInvitation, map[string]string{
			"WorkspaceName": workspace, "Role": role, "InvitationURL": envelope.PublicOrigin + "/invitations#token=" + envelope.Token})
		if err1 != nil {
			return nil, errors.New("invitation template unavailable")
		}
		envelope.Message = &invitationcrypto.Message{To: envelope.Email, From: envelope.Sender, Subject: "You are invited to a Flux workspace", HTML: body, IdempotencyKey: invitationIdentity(claim)}
		return invitationcrypto.Seal(j.invitationConfig, claim.WorkspaceID, claim.InvitationID, claim.DeliveryID, intent.KeyID, envelope)
	})
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (j *JobService) validateInvitation(envelope invitationcrypto.Envelope) error {
	raw, err2 := base64.RawURLEncoding.Strict().DecodeString(envelope.Token)
	address, addressErr := mail.ParseAddress(envelope.Email)
	policy := j.invitationConfig
	policy.Sender = envelope.Sender
	policy.PublicOrigin = envelope.PublicOrigin
	if err2 != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != envelope.Token || addressErr != nil || address.Address != envelope.Email || len(envelope.Email) > 254 || policy.Validate() != nil {
		return errors.New("invitation envelope unavailable")
	}
	return nil
}

// ConfigureInvitations binds the independent worker's durable delivery resources.
//
//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (j *JobService) ConfigureInvitations(pool *pgxpool.Pool, cfg config.InvitationConfig, sender InvitationSender) error {
	if pool == nil || sender == nil {
		return errors.New("invitation dependencies unavailable")
	}
	if err3 := cfg.Validate(); err3 != nil {
		return err3
	}
	j.deliveries = repository.NewDeliveryRepository(pool)
	j.invitationConfig = cfg
	j.invitationSender = sender
	return nil
}

//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func (j *JobService) handleInvitationTask(ctx context.Context, task *asynq.Task) error {
	claim, err4 := invitationClaim(task.Payload())
	if err4 != nil {
		return fmt.Errorf("invalid invitation reference: %w", asynq.SkipRetry)
	}
	if j.deliveries == nil {
		return errors.New("invitation delivery unavailable")
	}
	err4 = j.deliveries.Deliver(ctx, claim, func(ctx context.Context, intent repository.DeliveryIntent, _, _ string) error {
		envelope, err5 := invitationcrypto.Open(j.invitationConfig, claim.WorkspaceID, claim.InvitationID, claim.DeliveryID, intent.KeyID, intent.Ciphertext)
		if err5 != nil {
			return errors.New("invitation envelope unavailable")
		}
		if j.validateInvitation(envelope) != nil || envelope.Message == nil || envelope.Message.To != envelope.Email || envelope.Message.From != envelope.Sender || envelope.Message.HTML == "" || envelope.Message.Subject == "" || envelope.Message.IdempotencyKey != invitationIdentity(claim) {
			return errors.New("invitation envelope unavailable")
		}
		sendCtx, cancel := context.WithTimeout(ctx, invitationSendDeadline)
		defer cancel()
		// The intent UUID is immutable across lease generations and retries. A crash
		// after acknowledgement but before SQL commit reuses this provider key.
		return j.invitationSender.SendInvitation(sendCtx, *envelope.Message)
	})
	if err4 != nil {
		return errors.New("invitation delivery unavailable")
	}
	return nil
}
