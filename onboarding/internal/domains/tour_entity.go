package domains

import (
	"time"

	"github.com/google/uuid"
)

type TourStatus string

const (
	TourDraft     TourStatus = "draft"
	TourPublished TourStatus = "published"
	TourArchived  TourStatus = "archived"
)

type TriggerType string

const (
	TriggerOnLoad         TriggerType = "on_load"
	TriggerDelay          TriggerType = "delay"
	TriggerExitIntent     TriggerType = "exit_intent"
	TriggerManual         TriggerType = "manual"
	TriggerScrollDepth    TriggerType = "scroll_depth"
	TriggerInactivity     TriggerType = "inactivity"
	TriggerElementVisible TriggerType = "element_visible"
)

type TriggerConfig struct {
	ScrollDepth     *int    `json:"scroll_depth,omitempty"`
	InactivitySecs  *int    `json:"inactivity_secs,omitempty"`
	ElementSelector *string `json:"element_selector,omitempty"`
	DelayMs         *int    `json:"delay_ms,omitempty"`
}

type AudienceRule struct {
	Type      string
	Key       string
	Operator  string
	Value     any
	Timeframe string
}

type Audience struct {
	ShowOnce bool
	MaxShows int
	OnlyNew  bool
	Rules    []AudienceRule
}

type Tour struct {
	Id          uuid.UUID
	AppId       uuid.UUID
	Title       string
	Description string
	Enabled     bool
	Priority    int
	CreatedAt   time.Time
	UpdatedAt   time.Time

	VersionId uuid.UUID
	Version   int
	Status    TourStatus

	TriggerType   TriggerType
	TriggerConfig *TriggerConfig

	TargetPath string
	Audience   *Audience
	Hints      []Hint
	StepsCount int `json:"steps_count"`
}

type TourVersion struct {
	Id      uuid.UUID
	TourId  uuid.UUID
	Version int
	Status  TourStatus
}

type PublishResult struct {
	TourId             uuid.UUID `json:"tour_id"`
	PublishedVersionId uuid.UUID `json:"published_version_id"`
	Version            int       `json:"version"`
	NewDraftVersionId  uuid.UUID `json:"new_draft_version_id"`
}

type TourVersionHistoryItem struct {
	Id          uuid.UUID  `json:"id"`
	Version     int        `json:"version"`
	Status      TourStatus `json:"status"`
	StepsCount  int        `json:"steps_count"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at"`
}

type RestoreVersionResult struct {
	TourId                uuid.UUID `json:"tour_id"`
	RestoredFromVersionId uuid.UUID `json:"restored_from_version_id"`
	DraftVersionId        uuid.UUID `json:"draft_version_id"`
	DraftVersion          int       `json:"draft_version"`
	StepsCount            int       `json:"steps_count"`
}
