package request

import (
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type RecordEventRequest struct {
	EventId       uuid.UUID         `json:"event_id" binding:"required"`
	AttemptId     uuid.UUID         `json:"attempt_id" binding:"required"`
	SubjectId     string            `json:"subject_id" binding:"required"`
	TourId        uuid.UUID         `json:"tour_id" binding:"required"`
	TourVersionId uuid.UUID         `json:"tour_version_id" binding:"required"`
	HintId        *uuid.UUID        `json:"hint_id"`
	Type          domains.EventType `json:"type" binding:"required"`
}

func (r *RecordEventRequest) ToDomain(appId uuid.UUID) domains.TourEvent {
	return domains.TourEvent{
		AppId:         appId,
		EventId:       r.EventId,
		AttemptId:     r.AttemptId,
		SubjectId:     r.SubjectId,
		TourId:        r.TourId,
		TourVersionId: r.TourVersionId,
		HintId:        r.HintId,
		Type:          r.Type,
	}
}
