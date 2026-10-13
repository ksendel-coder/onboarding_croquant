package request

import (
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type ResolveRequest struct {
	Url       string         `json:"url"`
	SubjectId string         `json:"subject_id" binding:"required"`
	SessionId string         `json:"session_id"`
	Props     map[string]any `json:"props"`
}

type CreateAppRequest struct {
	Name           string   `json:"name" binding:"required"`
	AllowedOrigins []string `json:"allowed_origins" binding:"required,min=1,dive,required"`
}

type UpdateAppRequest struct {
	Name           *string  `json:"name,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
}

func (r *ResolveRequest) ToDomain(
	appId uuid.UUID,
) domains.ResolveRequest {
	return domains.ResolveRequest{
		AppId:     appId,
		Url:       r.Url,
		SubjectId: r.SubjectId,
		SessionId: r.SessionId,
		Props:     r.Props,
	}
}

func (a *CreateAppRequest) ToDomain() *domains.App {
	return &domains.App{
		Name:           a.Name,
		AllowedOrigins: a.AllowedOrigins,
	}
}
