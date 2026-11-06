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

	VersionId        uuid.UUID              `json:"version_id,omitempty"`
	Version          int                    `json:"version,omitempty"`
	Status           string                 `json:"status,omitempty"`
	TargetPath       string                 `json:"target_path,omitempty"`
	TriggerType      string                 `json:"trigger_type,omitempty"`
	TriggerConfig    *domains.TriggerConfig `json:"trigger_config,omitempty"`
	AudienceShowOnce bool                   `json:"audience_show_once"`
	AudienceMaxShows int                    `json:"audience_max_shows"`
	AudienceOnlyNew  bool                   `json:"audience_only_new"`
	Hints            []Hint                 `json:"hints,omitempty"`
	StepsCount       int                    `json:"steps_count"`
}

func NewTour(t *domains.Tour) *Tour {
	if t == nil {
		return nil
	}

	out := &Tour{
		Id:            t.Id,
		AppId:         t.AppId,
		Title:         t.Title,
		Description:   t.Description,
		Enabled:       t.Enabled,
		Priority:      t.Priority,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		VersionId:     t.VersionId,
		Version:       t.Version,
		Status:        string(t.Status),
		TargetPath:    t.TargetPath,
		TriggerType:   string(t.TriggerType),
		TriggerConfig: t.TriggerConfig,
		Hints:         NewHints(t.Hints),
		StepsCount:    t.StepsCount,
	}

	if t.Audience != nil {
		out.AudienceShowOnce = t.Audience.ShowOnce
		out.AudienceMaxShows = t.Audience.MaxShows
		out.AudienceOnlyNew = t.Audience.OnlyNew
	}

	return out
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
