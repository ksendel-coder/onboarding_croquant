package request

import (
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type Tour struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Priority    int    `json:"priority"`

	TargetPath string `json:"target_path"`
}

type TourMetaPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
	Priority    *int    `json:"priority"`
}

type TourSettingsPatch struct {
	TargetPath       *string                `json:"target_path"`
	TriggerType      *domains.TriggerType   `json:"trigger_type"`
	TriggerConfig    *domains.TriggerConfig `json:"trigger_config"`
	AudienceShowOnce *bool                  `json:"audience_show_once"`
	AudienceMaxShows *int                   `json:"audience_max_shows"`
	AudienceOnlyNew  *bool                  `json:"audience_only_new"`
}

func (t *Tour) ToDomain() *domains.Tour {
	if t == nil {
		return nil
	}

	return &domains.Tour{
		Title:       t.Title,
		Description: t.Description,
		Enabled:     t.Enabled,
		Priority:    t.Priority,
		TargetPath:  t.TargetPath,
	}
}

func (p *TourMetaPatch) ToDomain() domains.TourMetaPatch {
	return domains.TourMetaPatch{
		Title:       p.Title,
		Description: p.Description,
		Enabled:     p.Enabled,
		Priority:    p.Priority,
	}
}

func (p *TourSettingsPatch) ToDomain() domains.TourSettingsPatch {
	return domains.TourSettingsPatch{
		TargetPath:       p.TargetPath,
		TriggerType:      p.TriggerType,
		TriggerConfig:    p.TriggerConfig,
		AudienceShowOnce: p.AudienceShowOnce,
		AudienceMaxShows: p.AudienceMaxShows,
		AudienceOnlyNew:  p.AudienceOnlyNew,
	}
}
