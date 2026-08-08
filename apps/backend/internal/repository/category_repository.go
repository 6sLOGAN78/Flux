package repository

import (
	"context"
	"fmt"
	"strings"

	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/model"
	"flux/apps/backend/internal/model/category"
	"flux/apps/backend/pkg/sqlerr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) CreateCategory(
	ctx context.Context,
	tenantID *uuid.UUID,
	payload *category.CreateCategoryPayload,
) (*category.Category, error) {
	stmt := `
		INSERT INTO link_categories (
			tenant_id,
			name,
			color,
			description
		)
		VALUES (
			@tenant_id,
			@name,
			@color,
			@description
		)
		RETURNING *
	`

	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{
		"tenant_id":   tenantID,
		"name":        payload.Name,
		"color":       payload.Color,
		"description": payload.Description,
	})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to execute create category query: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect row from link_categories: %w", err))
	}

	return &item, nil
}

func (r *CategoryRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*category.Category, error) {
	stmt := `
		SELECT *
		FROM link_categories
		WHERE id = @id
	`

	rows, err := r.pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get category by id: %w", err))
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect category row: %w", err))
	}

	return &item, nil
}

func (r *CategoryRepository) GetCategories(
	ctx context.Context,
	tenantID *uuid.UUID,
	query *category.GetCategoriesQuery,
) (*model.PaginatedResponse[category.Category], error) {
	args := pgx.NamedArgs{}
	conditions := []string{}

	if tenantID != nil {
		conditions = append(conditions, "tenant_id = @tenant_id")
		args["tenant_id"] = *tenantID
	}

	if query.Search != nil && *query.Search != "" {
		conditions = append(conditions, "(name ILIKE @search OR description ILIKE @search)")
		args["search"] = "%" + *query.Search + "%"
	}

	baseQuery := "FROM link_categories"
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	countStmt := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to count categories: %w", err))
	}

	stmt := "SELECT * " + baseQuery

	sortCol := "name"
	if query.Sort != nil {
		sortCol = *query.Sort
	}
	orderDir := "ASC"
	if query.Order != nil && strings.ToLower(*query.Order) == "desc" {
		orderDir = "DESC"
	}
	stmt += fmt.Sprintf(" ORDER BY %s %s", sortCol, orderDir)

	limit := 50
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
		return nil, sqlerr.HandleError(fmt.Errorf("failed to get categories list: %w", err))
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect category rows: %w", err))
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &model.PaginatedResponse[category.Category]{
		Data:       items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (r *CategoryRepository) UpdateCategory(
	ctx context.Context,
	tenantID *uuid.UUID,
	payload *category.UpdateCategoryPayload,
) (*category.Category, error) {
	args := pgx.NamedArgs{
		"id": payload.ID,
	}
	setClauses := []string{}

	if payload.Name != nil {
		setClauses = append(setClauses, "name = @name")
		args["name"] = *payload.Name
	}
	if payload.Color != nil {
		setClauses = append(setClauses, "color = @color")
		args["color"] = *payload.Color
	}
	if payload.Description != nil {
		setClauses = append(setClauses, "description = @description")
		args["description"] = *payload.Description
	}

	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError("no fields provided for update", false, nil, nil, nil)
	}

	stmt := "UPDATE link_categories SET " + strings.Join(setClauses, ", ") + " WHERE id = @id"
	if tenantID != nil {
		stmt += " AND tenant_id = @tenant_id"
		args["tenant_id"] = *tenantID
	}
	stmt += " RETURNING *"

	rows, err := r.pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to update category: %w", err))
	}

	updated, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, sqlerr.HandleError(fmt.Errorf("failed to collect updated category row: %w", err))
	}

	return &updated, nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, tenantID *uuid.UUID, id uuid.UUID) error {
	stmt := "DELETE FROM link_categories WHERE id = @id"
	args := pgx.NamedArgs{"id": id}

	if tenantID != nil {
		stmt += " AND tenant_id = @tenant_id"
		args["tenant_id"] = *tenantID
	}

	result, err := r.pool.Exec(ctx, stmt, args)
	if err != nil {
		return sqlerr.HandleError(fmt.Errorf("failed to delete category: %w", err))
	}

	if result.RowsAffected() == 0 {
		code := "CATEGORY_NOT_FOUND"
		return errs.NewNotFoundError("category not found", false, &code)
	}

	return nil
}
