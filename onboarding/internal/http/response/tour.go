package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type Tour struct {
	Id          uuid.UUID `json:"id"`
	AppId       uuid.UUID `json:"app_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	VersionId  uuid.UUID `json:"version_id,omitempty"`
	Version    int       `json:"version,omitempty"`
	Status     string    `json:"status,omitempty"`
	TargetPath string    `json:"target_path,omitempty"`
	Hints      []Hint    `json:"hints,omitempty"`
	StepsCount int       `json:"steps_count"`
}

func NewTour(t *domains.Tour) *Tour {
	if t == nil {
		return nil
	}

	return &Tour{
		Id:          t.Id,
		AppId:       t.AppId,
		Title:       t.Title,
		Description: t.Description,
		Enabled:     t.Enabled,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		VersionId:   t.VersionId,
		Version:     t.Version,
		Status:      string(t.Status),
		TargetPath:  t.TargetPath,
		Hints:       NewHints(t.Hints),
		StepsCount:  t.StepsCount,
	}
}

func NewTours(
	tours []*domains.Tour,
) []*Tour {
	if tours == nil {
		return nil
	}

	out := make(
		[]*Tour,
		0,
		len(tours),
	)

	for _, t := range tours {
		out = append(
			out,
			NewTour(t),
		)
	}

	return out
}
