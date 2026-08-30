package campaign

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateCampaignPayload struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	UTMCampaign *string `json:"utmCampaign,omitempty" validate:"omitempty,max=255"`
}

func (p *CreateCampaignPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type UpdateCampaignPayload struct {
	ID          uuid.UUID `param:"id" validate:"required,uuid"`
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Status      *string   `json:"status,omitempty" validate:"omitempty,oneof=active paused completed"`
	UTMCampaign *string   `json:"utmCampaign,omitempty" validate:"omitempty,max=255"`
}

func (p *UpdateCampaignPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteCampaignPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteCampaignPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
