package request

import (
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type CreateAppRequest struct {
	Name           string   `json:"name" binding:"required"`
	AllowedOrigins []string `json:"allowed_origins" binding:"required,min=1,dive,required"`
}

type UpdateAppRequest struct {
	Name           *string  `json:"name,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
}

func (a *CreateAppRequest) ToDomain() *domains.App {
	return &domains.App{
		Name:           a.Name,
		AllowedOrigins: a.AllowedOrigins,
	}
}
