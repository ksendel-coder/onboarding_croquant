package response

import (
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type PreviewTriggerConfig struct {
	ScrollDepth     *int    `json:"scroll_depth,omitempty"`
	InactivitySecs  *int    `json:"inactivity_secs,omitempty"`
	ElementSelector *string `json:"element_selector,omitempty"`
	DelayMs         *int    `json:"delay_ms,omitempty"`
}

type PreviewAudience struct {
	ShowOnce bool                   `json:"show_once"`
	MaxShows int                    `json:"max_shows"`
	OnlyNew  bool                   `json:"only_new"`
	Rules    []domains.AudienceRule `json:"rules"`
}

type Preview struct {
	Tour          *Tour                 `json:"tour"`
	TriggerType   string                `json:"trigger_type"`
	TriggerConfig *PreviewTriggerConfig `json:"trigger_config"`
	Audience      *PreviewAudience      `json:"audience"`
	Preview       bool                  `json:"preview"`
}

func NewPreview(t *domains.Tour) *Preview {
	out := &Preview{
		Tour:        NewTour(t),
		TriggerType: string(t.TriggerType),
		Preview:     true,
	}
	if c := t.TriggerConfig; c != nil {
		out.TriggerConfig = &PreviewTriggerConfig{
			ScrollDepth:     c.ScrollDepth,
			InactivitySecs:  c.InactivitySecs,
			ElementSelector: c.ElementSelector,
			DelayMs:         c.DelayMs,
		}
	}
	if a := t.Audience; a != nil {
		out.Audience = &PreviewAudience{
			ShowOnce: a.ShowOnce,
			MaxShows: a.MaxShows,
			OnlyNew:  a.OnlyNew,
			Rules:    a.Rules,
		}
	}
	return out
}
