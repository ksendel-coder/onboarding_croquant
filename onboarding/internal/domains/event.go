package domains

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidEvent = errors.New("invalid onboarding event")

type EventDomain struct {
	repo EventRepositoryInterface
}

func NewEventDomain(repo EventRepositoryInterface) *EventDomain {
	return &EventDomain{repo: repo}
}

func (d *EventDomain) Record(ctx context.Context, event TourEvent) (bool, error) {
	validType := event.Type == EventTourStarted || event.Type == EventHintShown ||
		event.Type == EventHintCompleted || event.Type == EventTourCompleted ||
		event.Type == EventTourDismissed
	isHint := event.Type == EventHintShown || event.Type == EventHintCompleted

	if !validType || event.AppId == uuid.Nil || event.EventId == uuid.Nil ||
		event.AttemptId == uuid.Nil || event.TourId == uuid.Nil ||
		event.TourVersionId == uuid.Nil || strings.TrimSpace(event.SubjectId) == "" ||
		(event.HintId != nil) != isHint ||
		(event.HintId != nil && *event.HintId == uuid.Nil) {
		return false, logicErr(ErrInvalidEvent, "event validation", http.StatusBadRequest)
	}

	return d.repo.Record(ctx, event)
}

func (d *EventDomain) Analytics(
	ctx context.Context,
	tourId uuid.UUID,
	versionId uuid.UUID,
) (*TourAnalytics, error) {
	if tourId == uuid.Nil || versionId == uuid.Nil {
		return nil, logicErr(ErrInvalidEvent, "analytics validation", http.StatusBadRequest)
	}

	return d.repo.Analytics(ctx, tourId, versionId)
}
