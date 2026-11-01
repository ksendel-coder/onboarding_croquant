package domains

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type eventRepositoryStub struct {
	recorded TourEvent
	created  bool
}

func (s *eventRepositoryStub) Record(
	_ context.Context,
	event TourEvent,
) (bool, error) {
	s.recorded = event
	return s.created, nil
}

func (s *eventRepositoryStub) Analytics(
	_ context.Context,
	tourId uuid.UUID,
	versionId uuid.UUID,
) (*TourAnalytics, error) {
	return &TourAnalytics{TourId: tourId, TourVersionId: versionId}, nil
}

func validTourEvent() TourEvent {
	hintId := uuid.New()
	return TourEvent{
		AppId:         uuid.New(),
		EventId:       uuid.New(),
		AttemptId:     uuid.New(),
		SubjectId:     "visitor-123",
		TourId:        uuid.New(),
		TourVersionId: uuid.New(),
		HintId:        &hintId,
		Type:          EventHintShown,
	}
}

func TestEventDomainRecord(t *testing.T) {
	repo := &eventRepositoryStub{created: true}
	domain := NewEventDomain(repo)
	event := validTourEvent()

	created, err := domain.Record(context.Background(), event)

	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, event, repo.recorded)
}

func TestEventDomainRejectsInvalidEvents(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*TourEvent)
	}{
		{name: "missing app", mutate: func(e *TourEvent) { e.AppId = uuid.Nil }},
		{name: "missing event id", mutate: func(e *TourEvent) { e.EventId = uuid.Nil }},
		{name: "missing attempt", mutate: func(e *TourEvent) { e.AttemptId = uuid.Nil }},
		{name: "blank subject", mutate: func(e *TourEvent) { e.SubjectId = "  " }},
		{name: "unknown type", mutate: func(e *TourEvent) { e.Type = "unknown" }},
		{name: "hint event without hint", mutate: func(e *TourEvent) { e.HintId = nil }},
		{name: "tour event with hint", mutate: func(e *TourEvent) { e.Type = EventTourStarted }},
		{name: "nil hint id", mutate: func(e *TourEvent) {
			nilId := uuid.Nil
			e.HintId = &nilId
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain := NewEventDomain(&eventRepositoryStub{})
			event := validTourEvent()
			tc.mutate(&event)

			created, err := domain.Record(context.Background(), event)

			assert.False(t, created)
			var logicError *LogicError
			require.ErrorAs(t, err, &logicError)
			assert.Equal(t, http.StatusBadRequest, logicError.StatusCode())
			assert.ErrorIs(t, err, ErrInvalidEvent)
		})
	}
}

func TestEventDomainAnalyticsValidation(t *testing.T) {
	domain := NewEventDomain(&eventRepositoryStub{})

	_, err := domain.Analytics(context.Background(), uuid.Nil, uuid.New())

	var logicError *LogicError
	require.ErrorAs(t, err, &logicError)
	assert.Equal(t, http.StatusBadRequest, logicError.StatusCode())
}
