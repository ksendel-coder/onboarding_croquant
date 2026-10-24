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
	ScrollDepth     *int
	InactivitySecs  *int
	ElementSelector *string
	DelayMs         *int
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
}

type TourVersion struct {
	Id      uuid.UUID
	TourId  uuid.UUID
	Version int
	Status  TourStatus
}
