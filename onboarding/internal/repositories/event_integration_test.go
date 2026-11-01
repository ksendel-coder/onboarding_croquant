package repositories

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventRepositoryLifecycleIntegration(t *testing.T) {
	dsn := os.Getenv("EVENT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("EVENT_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	appId := uuid.New()
	foreignAppId := uuid.New()
	tourId := uuid.New()
	foreignTourId := uuid.New()
	versionId := uuid.New()
	foreignVersionId := uuid.New()
	firstHintId := uuid.New()
	secondHintId := uuid.New()
	foreignHintId := uuid.New()
	subjectOne := "event-test-" + uuid.NewString()
	subjectTwo := "event-test-" + uuid.NewString()
	foreignSubject := "event-test-" + uuid.NewString()

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM tour_events WHERE app_id=ANY($1::uuid[])`, []uuid.UUID{appId, foreignAppId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM tour_progress WHERE app_id=ANY($1::uuid[])`, []uuid.UUID{appId, foreignAppId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM hints WHERE tour_version_id=ANY($1::uuid[])`, []uuid.UUID{versionId, foreignVersionId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM tour_versions WHERE id=ANY($1::uuid[])`, []uuid.UUID{versionId, foreignVersionId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM tours WHERE id=ANY($1::uuid[])`, []uuid.UUID{tourId, foreignTourId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=ANY($1::uuid[])`, []uuid.UUID{appId, foreignAppId})
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO apps(id,name,public_key) VALUES
			($1,'Event test app',$2),
			($3,'Foreign event test app',$4)`,
		appId,
		"pk_"+appId.String(),
		foreignAppId,
		"pk_"+foreignAppId.String(),
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tours(id,app_id,title,enabled) VALUES
			($1,$2,'Event test tour',true),
			($3,$4,'Foreign event test tour',true)`,
		tourId,
		appId,
		foreignTourId,
		foreignAppId,
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_versions(
			id,tour_id,version,status,target_path,published_at
		) VALUES
			($1,$2,1,'published','/dashboard',now()),
			($3,$4,1,'published','/foreign',now())`,
		versionId,
		tourId,
		foreignVersionId,
		foreignTourId,
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO hints(id,tour_version_id,step,title,content,placement) VALUES
			($1,$2,1,'First','First step','bottom'),
			($3,$2,2,'Second','Second step','bottom'),
			($4,$5,1,'Foreign','Foreign step','bottom')`,
		firstHintId,
		versionId,
		secondHintId,
		foreignHintId,
		foreignVersionId,
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_progress(
			app_id,subject_id,tour_id,tour_version_id,current_hint_id,status,shows_count
		) VALUES
			($1,$2,$3,$4,$5,'in_progress',3),
			($1,$6,$3,$4,$5,'in_progress',1),
			($1,$7,$8,$9,$10,'in_progress',1)`,
		appId,
		subjectOne,
		tourId,
		versionId,
		firstHintId,
		subjectTwo,
		foreignSubject,
		foreignTourId,
		foreignVersionId,
		foreignHintId,
	)
	require.NoError(t, err)

	repo := NewEventRepository(pool)
	attemptOne := uuid.New()
	attemptTwo := uuid.New()
	newEvent := func(subject string, attemptId uuid.UUID, eventType domains.EventType, hintId *uuid.UUID) domains.TourEvent {
		return domains.TourEvent{
			AppId:         appId,
			EventId:       uuid.New(),
			AttemptId:     attemptId,
			SubjectId:     subject,
			TourId:        tourId,
			TourVersionId: versionId,
			HintId:        hintId,
			Type:          eventType,
		}
	}
	recordCreated := func(event domains.TourEvent) {
		t.Helper()
		created, recordErr := repo.Record(ctx, event)
		require.NoError(t, recordErr)
		assert.True(t, created)
	}

	startedOne := newEvent(subjectOne, attemptOne, domains.EventTourStarted, nil)
	recordCreated(startedOne)
	created, err := repo.Record(ctx, startedOne)
	require.NoError(t, err)
	assert.False(t, created)
	var repeatedEventCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tour_events WHERE app_id=$1 AND event_id=$2`,
		appId,
		startedOne.EventId,
	).Scan(&repeatedEventCount))
	assert.Equal(t, 1, repeatedEventCount)

	duplicateStart := startedOne
	duplicateStart.EventId = uuid.New()
	created, err = repo.Record(ctx, duplicateStart)
	require.NoError(t, err)
	assert.False(t, created)

	var showsCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT shows_count FROM tour_progress
		WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`,
		appId,
		subjectOne,
		tourId,
	).Scan(&showsCount))
	assert.Equal(t, 3, showsCount)

	wrongAttemptEvent := newEvent(
		subjectOne,
		uuid.New(),
		domains.EventHintShown,
		&firstHintId,
	)
	_, err = repo.Record(ctx, wrongAttemptEvent)
	assertLogicErrorStatus(t, err, http.StatusConflict)

	wrongCurrentHintEvent := newEvent(
		subjectOne,
		attemptOne,
		domains.EventHintCompleted,
		&secondHintId,
	)
	_, err = repo.Record(ctx, wrongCurrentHintEvent)
	assertLogicErrorStatus(t, err, http.StatusConflict)

	prematureCompletion := newEvent(
		subjectOne,
		attemptOne,
		domains.EventTourCompleted,
		nil,
	)
	_, err = repo.Record(ctx, prematureCompletion)
	assertLogicErrorStatus(t, err, http.StatusConflict)

	recordCreated(newEvent(subjectOne, attemptOne, domains.EventHintShown, &firstHintId))
	recordCreated(newEvent(subjectOne, attemptOne, domains.EventHintCompleted, &firstHintId))

	var currentHintId *uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT current_hint_id FROM tour_progress
		WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`,
		appId,
		subjectOne,
		tourId,
	).Scan(&currentHintId))
	require.NotNil(t, currentHintId)
	assert.Equal(t, secondHintId, *currentHintId)

	recordCreated(newEvent(subjectOne, attemptOne, domains.EventHintShown, &secondHintId))
	recordCreated(newEvent(subjectOne, attemptOne, domains.EventHintCompleted, &secondHintId))
	completedOne := newEvent(subjectOne, attemptOne, domains.EventTourCompleted, nil)
	recordCreated(completedOne)

	var progressStatus string
	var finishedAt *time.Time
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT status,finished_at,current_hint_id FROM tour_progress
		WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`,
		appId,
		subjectOne,
		tourId,
	).Scan(&progressStatus, &finishedAt, &currentHintId))
	assert.Equal(t, string(domains.ProgressCompleted), progressStatus)
	assert.NotNil(t, finishedAt)
	assert.Nil(t, currentHintId)

	created, err = repo.Record(ctx, completedOne)
	require.NoError(t, err)
	assert.False(t, created)

	recordCreated(newEvent(subjectTwo, attemptTwo, domains.EventTourStarted, nil))
	recordCreated(newEvent(subjectTwo, attemptTwo, domains.EventHintShown, &firstHintId))
	recordCreated(newEvent(subjectTwo, attemptTwo, domains.EventTourDismissed, nil))

	analytics, err := repo.Analytics(ctx, tourId, versionId)
	require.NoError(t, err)
	assert.Equal(t, 2, analytics.Started)
	assert.Equal(t, 1, analytics.Completed)
	assert.Equal(t, 1, analytics.Dismissed)
	assert.Equal(t, 50.0, analytics.CompletionRate)
	require.Len(t, analytics.Steps, 2)
	assert.Equal(t, domains.StepAnalytics{
		HintId: firstHintId, Step: 1, Shown: 2, Completed: 1, DropOff: 1,
	}, analytics.Steps[0])
	assert.Equal(t, domains.StepAnalytics{
		HintId: secondHintId, Step: 2, Shown: 1, Completed: 1, DropOff: 0,
	}, analytics.Steps[1])

	invalidHintEvent := newEvent(subjectTwo, uuid.New(), domains.EventHintShown, &foreignHintId)
	_, err = repo.Record(ctx, invalidHintEvent)
	assertLogicErrorStatus(t, err, http.StatusBadRequest)

	foreignTourEvent := domains.TourEvent{
		AppId:         appId,
		EventId:       uuid.New(),
		AttemptId:     uuid.New(),
		SubjectId:     foreignSubject,
		TourId:        foreignTourId,
		TourVersionId: foreignVersionId,
		Type:          domains.EventTourStarted,
	}
	_, err = repo.Record(ctx, foreignTourEvent)
	assertLogicErrorStatus(t, err, http.StatusNotFound)
}

func assertLogicErrorStatus(t *testing.T, err error, want int) {
	t.Helper()
	var logicError *domains.LogicError
	require.ErrorAs(t, err, &logicError)
	assert.Equal(t, want, logicError.StatusCode())
}
