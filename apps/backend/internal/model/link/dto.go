package link

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ------------------------------------------------------------

type CreateLinkPayload struct {
	DestinationURL string     `json:"destinationUrl" validate:"required,url"`
	CustomCode     *string    `json:"customCode,omitempty" validate:"omitempty,min=3,max=20,alphanum"`
	CategoryID     *uuid.UUID `json:"categoryId,omitempty" validate:"omitempty,uuid"`
	Title          *string    `json:"title,omitempty" validate:"omitempty,max=100"`
	Description    *string    `json:"description,omitempty" validate:"omitempty,max=255"`
}

func (p *CreateLinkPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type UpdateLinkPayload struct {
	ID             uuid.UUID  `param:"id" validate:"required,uuid"`
	DestinationURL *string    `json:"destinationUrl,omitempty" validate:"omitempty,url"`
	CategoryID     *uuid.UUID `json:"categoryId,omitempty" validate:"omitempty,uuid"`
	Title          *string    `json:"title,omitempty" validate:"omitempty,max=100"`
	Description    *string    `json:"description,omitempty" validate:"omitempty,max=255"`
}

func (p *UpdateLinkPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type BulkCategorizePayload struct {
	LinkIDs    []uuid.UUID `json:"linkIds" validate:"required,min=1"`
	CategoryID *uuid.UUID  `json:"categoryId" validate:"omitempty,uuid"`
}

func (p *BulkCategorizePayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetLinksQuery struct {
	Page       *int       `query:"page" validate:"omitempty,min=1"`
	Limit      *int       `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort       *string    `query:"sort" validate:"omitempty,oneof=created_at updated_at short_code"`
	Order      *string    `query:"order" validate:"omitempty,oneof=asc desc"`
	Search     *string    `query:"search" validate:"omitempty,min=1"`
	CategoryID *uuid.UUID `query:"category_id" validate:"omitempty,uuid"`
}

func (q *GetLinksQuery) Validate() error {
	validate := validator.New()

	if err := validate.Struct(q); err != nil {
		return err
	}

	// Set defaults
	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}
	if q.Limit == nil {
		defaultLimit := 20
		q.Limit = &defaultLimit
	}
	if q.Sort == nil {
		defaultSort := "created_at"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "desc"
		q.Order = &defaultOrder
	}

	return nil
}

// ------------------------------------------------------------

type DeleteLinkPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteLinkPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
