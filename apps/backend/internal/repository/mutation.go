package repository

import (
	"bytes"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// bootstrapRequest reads an actor-scoped ledger while its durable user is locked.
// Cleanup never touches another actor and only removes records past retention.
func bootstrapRequest(ctx context.Context, tx pgx.Tx, actor uuid.UUID, key string) (uuid.UUID, []byte, error) {
	if _, err := tx.Exec(ctx, "DELETE FROM workspace_bootstrap_requests "+
		"WHERE actor_id=$1 AND retain_until < now()", actor); err != nil {
		return uuid.Nil, nil, err
	}
	var workspace uuid.UUID
	var hash []byte
	err := tx.QueryRow(ctx, "SELECT workspace_id,request_hash FROM workspace_bootstrap_requests "+
		"WHERE actor_id=$1 AND request_key=$2", actor, key).Scan(&workspace, &hash)
	return workspace, hash, err
}

// recordBootstrap shares the transaction with workspace and owner creation.
func recordBootstrap(ctx context.Context, tx pgx.Tx, actor, workspace uuid.UUID, key string, hash []byte) error {
	_, err := tx.Exec(ctx, "INSERT INTO workspace_bootstrap_requests "+
		"(actor_id,request_key,request_hash,workspace_id) VALUES($1,$2,$3,$4)", actor, key, hash, workspace)
	return err
}

// matchRequest is used only AFTER fresh authorization of the committed result.
func matchRequest(request, stored []byte) error {
	if !bytes.Equal(request, stored) {
		return ErrWorkspaceConflict
	}
	return nil
}
