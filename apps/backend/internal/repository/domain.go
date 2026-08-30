package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flux/apps/backend/internal/model/domain"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DomainRepository struct {
	pool *pgxpool.Pool
}

func NewDomainRepository(pool *pgxpool.Pool) *DomainRepository {
	return &DomainRepository{pool: pool}
}

func (r *DomainRepository) CreateDomain(ctx context.Context, tenantID, hostname, verificationToken string) (*domain.CustomDomain, error) {
	// Normalize hostname
	hostname = strings.ToLower(hostname)

	stmt := `
		INSERT INTO custom_domains (tenant_id, hostname, verification_token)
		VALUES (@tenant_id, @hostname, @verification_token)
		RETURNING id, tenant_id, hostname, status, verification_token, ssl_status, created_at, updated_at
	`
	args := pgx.NamedArgs{
		"tenant_id":          tenantID,
		"hostname":           hostname,
		"verification_token": verificationToken,
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to create domain: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.CustomDomain])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect domain row: %w", err))
	}

	return &item, nil
}

func (r *DomainRepository) GetDomainByID(ctx context.Context, tenantID, id string) (*domain.CustomDomain, error) {
	stmt := `
		SELECT id, tenant_id, hostname, status, verification_token, ssl_status, created_at, updated_at
		FROM custom_domains
		WHERE id = @id AND tenant_id = @tenant_id
	`
	args := pgx.NamedArgs{
		"id":        id,
		"tenant_id": tenantID,
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get domain: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect domain row: %w", err))
	}

	return &item, nil
}

func (r *DomainRepository) GetDomainByHostname(ctx context.Context, hostname string) (*domain.CustomDomain, error) {
	stmt := `
		SELECT id, tenant_id, hostname, status, verification_token, ssl_status, created_at, updated_at
		FROM custom_domains
		WHERE hostname = @hostname
	`
	args := pgx.NamedArgs{
		"hostname": strings.ToLower(hostname),
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get domain by hostname: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect domain row: %w", err))
	}

	return &item, nil
}

func (r *DomainRepository) GetDomainsByTenant(ctx context.Context, tenantID string) ([]domain.CustomDomain, error) {
	stmt := `
		SELECT id, tenant_id, hostname, status, verification_token, ssl_status, created_at, updated_at
		FROM custom_domains
		WHERE tenant_id = @tenant_id
		ORDER BY created_at DESC
	`
	args := pgx.NamedArgs{
		"tenant_id": tenantID,
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to query domains: %w", err))
	}
	defer rows.Close()

	domains, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.CustomDomain])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect domain rows: %w", err))
	}

	return domains, nil
}

func (r *DomainRepository) DeleteDomain(ctx context.Context, tenantID, id string) error {
	stmt := `
		DELETE FROM custom_domains
		WHERE id = @id AND tenant_id = @tenant_id
	`
	args := pgx.NamedArgs{
		"id":        id,
		"tenant_id": tenantID,
	}

	ct, err := r.pool.Exec(ctx, stmt, args)
	if err != nil {
		return sqlerr.HandleError(fmt.Errorf("failed to delete domain: %w", err))
	}

	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *DomainRepository) GetDomainsToVerify(ctx context.Context, batchSize int) ([]domain.CustomDomain, error) {
	stmt := `
		SELECT id, tenant_id, hostname, status, verification_token, ssl_status, created_at, updated_at
		FROM custom_domains
		WHERE status IN ('pending', 'verifying', 'active')
		ORDER BY updated_at ASC
		LIMIT @limit
	`
	args := pgx.NamedArgs{
		"limit": batchSize,
	}

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to query domains for verification: %w", err))
	}
	defer rows.Close()

	domains, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.CustomDomain])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect domain rows: %w", err))
	}

	return domains, nil
}

func (r *DomainRepository) UpdateDomainStatus(ctx context.Context, id, status string) error {
	stmt := `
		UPDATE custom_domains
		SET status = @status
		WHERE id = @id
	`
	args := pgx.NamedArgs{
		"id":     id,
		"status": status,
	}

	ct, err := r.pool.Exec(ctx, stmt, args)
	if err != nil {
		return sqlerr.HandleError(fmt.Errorf("failed to update domain status: %w", err))
	}

	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
