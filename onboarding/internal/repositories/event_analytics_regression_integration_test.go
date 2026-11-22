package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestEventAnalyticsRegressionIntegration(t *testing.T) {
	dsn := os.Getenv("EVENT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("EVENT_TEST_DATABASE_URL is required: run against an isolated migrated PostgreSQL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	appID := uuid.New()
	tourID := uuid.New()
	v1, v2 := uuid.New(), uuid.New()
	h1, h2, hV2 := uuid.New(), uuid.New(), uuid.New()
	attemptOK, attemptSkipped, attemptDismissed, attemptV2 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	t.Cleanup(func() {
		cleanup := context.Background()
		// Удалять дочерние записи прежде родительских; никаких TRUNCATE.
		if _, err := pool.Exec(cleanup, `DELETE FROM tour_events WHERE app_id=$1`, appID); err != nil {
			t.Error(err)
		}
		if _, err := pool.Exec(cleanup, `DELETE FROM hints WHERE tour_version_id IN ($1,$2)`, v1, v2); err != nil {
			t.Error(err)
		}
		if _, err := pool.Exec(cleanup, `DELETE FROM tour_versions WHERE id IN ($1,$2)`, v1, v2); err != nil {
			t.Error(err)
		}
		if _, err := pool.Exec(cleanup, `DELETE FROM tours WHERE id=$1`, tourID); err != nil {
			t.Error(err)
		}
		if _, err := pool.Exec(cleanup, `DELETE FROM apps WHERE id=$1`, appID); err != nil {
			t.Error(err)
		}
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO apps(id,name,public_key) VALUES ($1,'Analytics regression',$2)`,
		appID, "pk_"+appID.String())
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tours(id,app_id,title,enabled) VALUES ($1,$2,'Analytics regression',true)`,
		tourID, appID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_versions(id,tour_id,version,status,target_path,published_at,archived_at)
		VALUES ($1,$3,1,'archived','/demo',now(),now()),
		       ($2,$3,2,'published','/demo',now(),NULL)`,
		v1, v2, tourID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO hints(id,tour_version_id,step,title,content,placement)
		VALUES ($1,$4,1,'Step 1','Text','bottom'),
		       ($2,$4,2,'Step 2','Text','bottom'),
		       ($3,$5,1,'Step v2','Text','bottom')`,
		h1, h2, hV2, v1, v2)
	require.NoError(t, err)

	// Пишем реальные строки событий с валидными FK и типами.
	// attemptOK: обычное завершение двух шагов.
	// attemptSkipped: оба шага автоматически пропущены, tour_completed не засчитывается как completed.
	// attemptDismissed: показ первого шага и закрытие.
	// attemptV2: события другой версии, не должны попадать в v1.
	put := func(subject string, attempt, version uuid.UUID, typ string, hint *uuid.UUID) uuid.UUID {
		t.Helper()
		eventID := uuid.New()
		_, insertErr := pool.Exec(ctx, `
			INSERT INTO tour_events(
			  app_id,event_id,attempt_id,subject_id,tour_id,tour_version_id,hint_id,event_type
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			appID, eventID, attempt, subject, tourID, version, hint, typ)
		require.NoError(t, insertErr)
		return eventID
	}
	ptr := func(id uuid.UUID) *uuid.UUID { return &id }

	put("ok", attemptOK, v1, "tour_started", nil)
	put("ok", attemptOK, v1, "hint_shown", ptr(h1))
	put("ok", attemptOK, v1, "hint_completed", ptr(h1))
	put("ok", attemptOK, v1, "hint_shown", ptr(h2))
	put("ok", attemptOK, v1, "hint_completed", ptr(h2))
	put("ok", attemptOK, v1, "tour_completed", nil)

	put("skipped", attemptSkipped, v1, "tour_started", nil)
	put("skipped", attemptSkipped, v1, "hint_skipped", ptr(h1))
	put("skipped", attemptSkipped, v1, "hint_skipped", ptr(h2))
	put("skipped", attemptSkipped, v1, "tour_completed", nil)

	put("dismissed", attemptDismissed, v1, "tour_started", nil)
	put("dismissed", attemptDismissed, v1, "hint_shown", ptr(h1))
	put("dismissed", attemptDismissed, v1, "tour_dismissed", nil)

	put("v2", attemptV2, v2, "tour_started", nil)
	put("v2", attemptV2, v2, "hint_shown", ptr(hV2))

	// SQL-ограничение UNIQUE(app_id,event_id) защищает от повторного event_id.
	duplicateEventID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_events(app_id,event_id,attempt_id,subject_id,tour_id,tour_version_id,event_type)
		VALUES ($1,$2,$3,'ok',$4,$5,'tour_started')`,
		appID, duplicateEventID, attemptOK, tourID, v1)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_events(app_id,event_id,attempt_id,subject_id,tour_id,tour_version_id,event_type)
		VALUES ($1,$2,$3,'duplicate-again',$4,$5,'tour_started')`,
		appID, duplicateEventID, uuid.New(), tourID, v1)
	requireSQLState(t, err, "23505") // unique_violation

	// CHECK: неизвестный тип события отклоняется.
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_events(app_id,event_id,attempt_id,subject_id,tour_id,tour_version_id,event_type)
		VALUES ($1,$2,$3,'invalid',$4,$5,'not_an_event')`,
		appID, uuid.New(), uuid.New(), tourID, v1)
	requireSQLState(t, err, "23514") // check_violation

	// CHECK: для hint_shown необходим hint_id.
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_events(app_id,event_id,attempt_id,subject_id,tour_id,tour_version_id,event_type)
		VALUES ($1,$2,$3,'invalid',$4,$5,'hint_shown')`,
		appID, uuid.New(), uuid.New(), tourID, v1)
	requireSQLState(t, err, "23514")

	// CHECK: у tour_completed не должно быть hint_id.
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_events(app_id,event_id,attempt_id,subject_id,tour_id,tour_version_id,hint_id,event_type)
		VALUES ($1,$2,$3,'invalid',$4,$5,$6,'tour_completed')`,
		appID, uuid.New(), uuid.New(), tourID, v1, h1)
	requireSQLState(t, err, "23514")

	repo := NewEventRepository(pool)
	a1, err := repo.Analytics(ctx, tourID, v1)
	require.NoError(t, err)
	require.Equal(t, 3, a1.Started)
	require.Equal(t, 1, a1.Completed)
	require.Equal(t, 1, a1.Dismissed)
	require.InDelta(t, 100.0/3.0, a1.CompletionRate, 0.0001)
	require.Len(t, a1.Steps, 2)
	require.Equal(t, h1, a1.Steps[0].HintId)
	require.Equal(t, 2, a1.Steps[0].Shown)
	require.Equal(t, 1, a1.Steps[0].Completed)
	require.Equal(t, 1, a1.Steps[0].Skipped)
	require.Equal(t, 1, a1.Steps[0].DropOff)
	require.Equal(t, h2, a1.Steps[1].HintId)
	require.Equal(t, 1, a1.Steps[1].Shown)
	require.Equal(t, 1, a1.Steps[1].Completed)
	require.Equal(t, 1, a1.Steps[1].Skipped)
	require.Equal(t, 0, a1.Steps[1].DropOff)

	a2, err := repo.Analytics(ctx, tourID, v2)
	require.NoError(t, err)
	require.Equal(t, 1, a2.Started)
	require.Equal(t, 0, a2.Completed)
	require.Equal(t, 0, a2.Dismissed)
	require.Equal(t, 0.0, a2.CompletionRate)
	require.Len(t, a2.Steps, 1)
	require.Equal(t, 1, a2.Steps[0].Shown)

	_, err = repo.Analytics(ctx, tourID, uuid.New())
	require.Error(t, err, "version_id of a different/nonexistent version must be rejected")
}

func requireSQLState(t *testing.T, err error, want string) {
	t.Helper()
	require.Error(t, err)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, want, pgErr.Code)
}
