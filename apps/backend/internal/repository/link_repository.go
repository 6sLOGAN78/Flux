package repository

import (
	"context"
	"fmt"
	"strings"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model"
	"flux/apps/backend/internal/model/link"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepository struct {
	pool *pgxpool.Pool
}

func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{pool: pool}
}

func (r *LinkRepository) CreateLink(
	ctx context.Context,
	tenantID *uuid.UUID,
	payload *link.CreateLinkPayload,
	shortCode string,
) (*link.Link, error) {
	stmt := `
		INSERT INTO links (
			short_code,
			destination_url,
			tenant_id,
			title,
			description
		)
		VALUES (
			@short_code,
			@destination_url,
			@tenant_id,
			@title,
			@description
		)
		RETURNING *
	`

	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{
		"short_code":      shortCode,
		"destination_url": payload.DestinationURL,
		"tenant_id":       tenantID,
		"title":          payload.Title,
		"description":    payload.Description,
	})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to execute create link query: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[link.Link])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect row from links: %w", err))
	}

	return &item, nil
}

func (r *LinkRepository) GetLinkByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	stmt := `
		SELECT *
		FROM links
		WHERE id = @id
	`

	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get link by id: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[link.Link])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect link row: %w", err))
	}

	return &item, nil
}

func (r *LinkRepository) GetLinkByShortCode(ctx context.Context, shortCode string) (*link.Link, error) {
	stmt := `
		SELECT *
		FROM links
		WHERE short_code = @short_code
	`

	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{
		"short_code": shortCode,
	})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get link by short code: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[link.Link])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect link row: %w", err))
	}

	return &item, nil
}

func (r *LinkRepository) GetLinks(
	ctx context.Context,
	tenantID *uuid.UUID,
	query *link.GetLinksQuery,
) (*model.PaginatedResponse[link.Link], error) {
	args := pgx.NamedArgs{}
	conditions := []string{}

	if tenantID != nil {
		conditions = append(conditions, "tenant_id = @tenant_id")
		args["tenant_id"] = *tenantID
	}

	if query.Search != nil && *query.Search != "" {
		conditions = append(conditions, "(short_code ILIKE @search OR destination_url ILIKE @search OR title ILIKE @search)")
		args["search"] = "%" + *query.Search + "%"
	}

	baseQuery := "FROM links"
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	countStmt := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to count links: %w", err))
	}

	stmt := "SELECT * " + baseQuery

	sortCol := "created_at"
	if query.Sort != nil {
		sortCol = *query.Sort
	}
	orderDir := "DESC"
	if query.Order != nil && strings.ToLower(*query.Order) == "asc" {
		orderDir = "ASC"
	}
	stmt += fmt.Sprintf(" ORDER BY %s %s", sortCol, orderDir)

	limit := 20
	if query.Limit != nil {
		limit = *query.Limit
	}
	page := 1
	if query.Page != nil {
		page = *query.Page
	}
	offset := (page - 1) * limit

	stmt += " LIMIT @limit OFFSET @offset"
	args["limit"] = limit
	args["offset"] = offset

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get links list: %w", err))
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[link.Link])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect links rows: %w", err))
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &model.PaginatedResponse[link.Link]{
		Data:       items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (r *LinkRepository) UpdateLink(
	ctx context.Context,
	tenantID *uuid.UUID,
	payload *link.UpdateLinkPayload,
) (*link.Link, error) {
	args := pgx.NamedArgs{
		"id": payload.ID,
	}
	setClauses := []string{}

	if payload.DestinationURL != nil {
		setClauses = append(setClauses, "destination_url = @destination_url")
		args["destination_url"] = *payload.DestinationURL
	}
	if payload.Title != nil {
		setClauses = append(setClauses, "title = @title")
		args["title"] = *payload.Title
	}
	if payload.Description != nil {
		setClauses = append(setClauses, "description = @description")
		args["description"] = *payload.Description
	}

	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError("no fields provided for update", false, nil, nil, nil)
	}

	stmt := "UPDATE links SET " + strings.Join(setClauses, ", ") + " WHERE id = @id"
	if tenantID != nil {
		stmt += " AND tenant_id = @tenant_id"
		args["tenant_id"] = *tenantID
	}
	stmt += " RETURNING *"

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to update link: %w", err))
	}

	updated, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[link.Link])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect updated link row: %w", err))
	}

	return &updated, nil
}

func (r *LinkRepository) DeleteLink(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) error {
	stmt := "DELETE FROM links WHERE id = @id"
	args := pgx.NamedArgs{"id": id}

	if tenantID != nil {
		stmt += " AND tenant_id = @tenant_id"
		args["tenant_id"] = *tenantID
	}

	result, err := r.pool.Exec(ctx, stmt, args)
	if err != nil {
		return sqlerr.HandleError(fmt.Errorf("failed to delete link: %w", err))
	}

	if result.RowsAffected() == 0 {
		code := "LINK_NOT_FOUND"
		return errs.NewNotFoundError("link not found", false, &code)
	}

	return nil
}
