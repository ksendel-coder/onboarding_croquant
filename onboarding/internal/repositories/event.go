package repositories

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func invalidEvent(message string, code int) error {
	return &domains.LogicError{
		Err:   errors.New(message),
		Stage: "event",
		Code:  code,
	}
}

func (r *EventRepository) Record(
	ctx context.Context,
	event domains.TourEvent,
) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var version uuid.UUID
	var status string
	var currentHintId *uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT p.tour_version_id, p.status, p.current_hint_id
		FROM tour_progress p
		WHERE p.app_id=$1 AND p.subject_id=$2 AND p.tour_id=$3
		FOR UPDATE`,
		event.AppId,
		event.SubjectId,
		event.TourId,
	).Scan(&version, &status, &currentHintId)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, invalidEvent("progress not found; call /resolve first", http.StatusConflict)
	}
	if err != nil {
		return false, err
	}
	if version != event.TourVersionId {
		return false, invalidEvent("tour version does not match progress", http.StatusConflict)
	}

	var published bool
	err = tx.QueryRow(ctx, `
		SELECT (v.status='published')
		FROM tours t
		JOIN tour_versions v ON v.tour_id=t.id
		WHERE t.id=$1 AND t.app_id=$2 AND v.id=$3`,
		event.TourId,
		event.AppId,
		event.TourVersionId,
	).Scan(&published)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, invalidEvent("tour or version not found", http.StatusNotFound)
	}
	if err != nil {
		return false, err
	}
	if !published {
		return false, invalidEvent("only published versions can receive events", http.StatusConflict)
	}

	if event.HintId != nil {
		var exists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM hints WHERE id=$1 AND tour_version_id=$2
			)`,
			*event.HintId,
			event.TourVersionId,
		).Scan(&exists)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, invalidEvent("hint not in tour version", http.StatusBadRequest)
		}
	}

	var previous domains.TourEvent
	err = tx.QueryRow(ctx, `
		SELECT attempt_id, subject_id, tour_id, tour_version_id, hint_id, event_type
		FROM tour_events
		WHERE app_id=$1 AND event_id=$2`,
		event.AppId,
		event.EventId,
	).Scan(
		&previous.AttemptId,
		&previous.SubjectId,
		&previous.TourId,
		&previous.TourVersionId,
		&previous.HintId,
		&previous.Type,
	)
	if err == nil {
		sameHint := (previous.HintId == nil && event.HintId == nil) ||
			(previous.HintId != nil && event.HintId != nil && *previous.HintId == *event.HintId)
		if previous.AttemptId != event.AttemptId || previous.SubjectId != event.SubjectId ||
			previous.TourId != event.TourId || previous.TourVersionId != event.TourVersionId ||
			!sameHint || previous.Type != event.Type {
			return false, invalidEvent("event_id reused with different data", http.StatusConflict)
		}

		return false, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	if status != string(domains.ProgressInProgress) {
		return false, invalidEvent("scenario already finished", http.StatusConflict)
	}

	var sameTypeExists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tour_events
			WHERE app_id=$1 AND attempt_id=$2 AND tour_id=$3
				AND event_type=$4 AND hint_id IS NOT DISTINCT FROM $5::uuid
		)`,
		event.AppId,
		event.AttemptId,
		event.TourId,
		string(event.Type),
		event.HintId,
	).Scan(&sameTypeExists)
	if err != nil {
		return false, err
	}
	if sameTypeExists {
		return false, tx.Commit(ctx)
	}

	var anotherAttempt bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tour_events
			WHERE app_id=$1
				AND subject_id=$2
				AND tour_id=$3
				AND tour_version_id=$4
				AND attempt_id<>$5
		)`,
		event.AppId,
		event.SubjectId,
		event.TourId,
		event.TourVersionId,
		event.AttemptId,
	).Scan(&anotherAttempt)
	if err != nil {
		return false, err
	}
	if anotherAttempt {
		return false, invalidEvent("attempt_id does not match", http.StatusConflict)
	}

	if event.Type == domains.EventHintCompleted {
		if currentHintId == nil || event.HintId == nil || *currentHintId != *event.HintId {
			return false, invalidEvent("hint is not current", http.StatusConflict)
		}
	}

	if event.Type == domains.EventTourCompleted {
		var totalHints int
		err = tx.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM hints
			WHERE tour_version_id=$1`,
			event.TourVersionId,
		).Scan(&totalHints)
		if err != nil {
			return false, err
		}
		if totalHints == 0 || currentHintId != nil {
			return false, invalidEvent("scenario steps are not completed", http.StatusConflict)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO tour_events (
			app_id, event_id, attempt_id, subject_id, tour_id,
			tour_version_id, hint_id, event_type
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		event.AppId,
		event.EventId,
		event.AttemptId,
		event.SubjectId,
		event.TourId,
		event.TourVersionId,
		event.HintId,
		event.Type,
	)
	if err != nil {
		return false, err
	}

	switch event.Type {
	case domains.EventHintCompleted:
		_, err = tx.Exec(ctx, `
			UPDATE tour_progress p
			SET current_hint_id=(
				SELECT h2.id
				FROM hints h1
				JOIN hints h2
					ON h2.tour_version_id=h1.tour_version_id AND h2.step>h1.step
				WHERE h1.id=$4 AND h1.tour_version_id=$5
				ORDER BY h2.step
				LIMIT 1
			)
			WHERE p.app_id=$1 AND p.subject_id=$2 AND p.tour_id=$3`,
			event.AppId,
			event.SubjectId,
			event.TourId,
			*event.HintId,
			event.TourVersionId,
		)
	case domains.EventTourCompleted, domains.EventTourDismissed:
		newStatus := string(domains.ProgressCompleted)
		if event.Type == domains.EventTourDismissed {
			newStatus = string(domains.ProgressDismissed)
		}

		_, err = tx.Exec(ctx, `
			UPDATE tour_progress
			SET status=$4, finished_at=now(), current_hint_id=NULL
			WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`,
			event.AppId,
			event.SubjectId,
			event.TourId,
			newStatus,
		)
	}
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}

func (r *EventRepository) Analytics(
	ctx context.Context,
	tourId uuid.UUID,
	versionId uuid.UUID,
) (*domains.TourAnalytics, error) {
	result := &domains.TourAnalytics{
		TourId:        tourId,
		TourVersionId: versionId,
		Steps:         make([]domains.StepAnalytics, 0),
	}

	var found uuid.UUID
	if err := r.pool.QueryRow(ctx, `
		SELECT id
		FROM tour_versions
		WHERE id=$1 AND tour_id=$2`,
		versionId,
		tourId,
	).Scan(&found); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, invalidEvent("version not found for tour", http.StatusNotFound)
		}
		return nil, err
	}

	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT attempt_id) FILTER (WHERE event_type='tour_started'),
			COUNT(DISTINCT attempt_id) FILTER (WHERE event_type='tour_completed'),
			COUNT(DISTINCT attempt_id) FILTER (WHERE event_type='tour_dismissed')
		FROM tour_events
		WHERE tour_id=$1 AND tour_version_id=$2`,
		tourId,
		versionId,
	).Scan(&result.Started, &result.Completed, &result.Dismissed)
	if err != nil {
		return nil, err
	}
	if result.Started > 0 {
		result.CompletionRate = float64(result.Completed) * 100 / float64(result.Started)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			h.id,
			h.step,
			COUNT(DISTINCT e.attempt_id) FILTER (WHERE e.event_type='hint_shown') AS shown,
			COUNT(DISTINCT e.attempt_id) FILTER (WHERE e.event_type='hint_completed') AS completed
		FROM hints h
		LEFT JOIN tour_events e
			ON e.hint_id=h.id AND e.tour_version_id=h.tour_version_id
		WHERE h.tour_version_id=$1
		GROUP BY h.id,h.step
		ORDER BY h.step`,
		versionId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var step domains.StepAnalytics
		if err := rows.Scan(&step.HintId, &step.Step, &step.Shown, &step.Completed); err != nil {
			return nil, err
		}
		step.DropOff = step.Shown - step.Completed
		if step.DropOff < 0 {
			step.DropOff = 0
		}
		result.Steps = append(result.Steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

var _ domains.EventRepositoryInterface = (*EventRepository)(nil)
