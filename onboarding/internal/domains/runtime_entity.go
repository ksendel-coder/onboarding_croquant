package domains

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	Id             uuid.UUID
	Name           string
	PublicKey      string
	AllowedOrigins []string
	CreatedAt      time.Time
	ArchivedAt     *time.Time
}

type ProgressStatus string

const (
	ProgressInProgress ProgressStatus = "in_progress"
	ProgressCompleted  ProgressStatus = "completed"
	ProgressDismissed  ProgressStatus = "dismissed"
)

type Progress struct {
	AppId         uuid.UUID
	SubjectId     string
	TourId        uuid.UUID
	TourVersionId uuid.UUID
	CurrentHintId *uuid.UUID
	Status        ProgressStatus
	ShowsCount    int
	StartedAt     time.Time
	UpdatedAt     time.Time
	FinishedAt    *time.Time
}

type ResolveRequest struct {
	AppId     uuid.UUID
	Url       string
	SubjectId string
	SessionId string
	Props     map[string]any
}

type ResolveResult struct {
	Tour          *Tour
	TourVersionId uuid.UUID
	Version       int
	CurrentHintId *uuid.UUID
}
