package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DeliveryIntent is private encrypted storage, never serialized to clients.
type DeliveryIntent struct {
	KeyID        string
	Ciphertext   []byte
	ID           uuid.UUID
	InvitationID uuid.UUID
}

func recordInvitationDelivery(ctx context.Context, tx pgx.Tx, scope Scope, intent DeliveryIntent) error {
	_, err := tx.Exec(ctx, "INSERT INTO invitation_delivery_intents"+
		"(workspace_id,id,invitation_id,key_id,ciphertext) VALUES($1,$2,$3,$4,$5)",
		scope.WorkspaceID, intent.ID, intent.InvitationID, intent.KeyID, intent.Ciphertext)
	return err
}
