package domains

import (
	"context"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTourStarted   EventType = "tour_started"
	EventHintShown     EventType = "hint_shown"
	EventHintCompleted EventType = "hint_completed"
	EventTourCompleted EventType = "tour_completed"
	EventTourDismissed EventType = "tour_dismissed"
)

type TourEvent struct {
	AppId         uuid.UUID
	EventId       uuid.UUID
	AttemptId     uuid.UUID
	SubjectId     string
	TourId        uuid.UUID
	TourVersionId uuid.UUID
	HintId        *uuid.UUID
	Type          EventType
}

type StepAnalytics struct {
	HintId    uuid.UUID `json:"hint_id"`
	Step      int       `json:"step"`
	Shown     int       `json:"shown"`
	Completed int       `json:"completed"`
	DropOff   int       `json:"drop_off"`
}

type TourAnalytics struct {
	TourId         uuid.UUID       `json:"tour_id"`
	TourVersionId  uuid.UUID       `json:"tour_version_id"`
	Started        int             `json:"started"`
	Completed      int             `json:"completed"`
	Dismissed      int             `json:"dismissed"`
	CompletionRate float64         `json:"completion_rate"`
	Steps          []StepAnalytics `json:"steps"`
}

type EventRepositoryInterface interface {
	Record(ctx context.Context, event TourEvent) (bool, error)
	Analytics(ctx context.Context, tourId, versionId uuid.UUID) (*TourAnalytics, error)
}
