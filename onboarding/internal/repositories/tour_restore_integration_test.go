package repositories

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// Do not run it on a database containing production data.
type restoreFixture struct {
	pool          *pgxpool.Pool
	repo          *TourRepository
	appID         uuid.UUID
	tourID        uuid.UUID
	foreignTourID uuid.UUID
	archivedID    uuid.UUID
	publishedID   uuid.UUID
	draftID       uuid.UUID
}

type restoreHint struct {
	ID                                  uuid.UUID
	Step                                int
	Title, Content, Selector, Placement string
}

func openRestoreDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("MVP11_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MVP11_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))
	return pool
}

func newRestoreFixture(t *testing.T) *restoreFixture {
	t.Helper()
	ctx := context.Background()
	pool := openRestoreDB(t)
	f := &restoreFixture{
		pool:          pool,
		appID:         uuid.New(),
		tourID:        uuid.New(),
		foreignTourID: uuid.New(),
	}
	logger := zerolog.Nop()
	f.repo = NewTourRepository(pool, &logger)
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Cleanup only the fixtures created by this test.
		_, _ = pool.Exec(cleanCtx, `DELETE FROM hints WHERE tour_version_id IN (
            SELECT id FROM tour_versions WHERE tour_id=ANY($1::uuid[]))`,
			[]uuid.UUID{f.tourID, f.foreignTourID})
		_, _ = pool.Exec(cleanCtx, `DELETE FROM tour_versions WHERE tour_id=ANY($1::uuid[])`,
			[]uuid.UUID{f.tourID, f.foreignTourID})
		_, _ = pool.Exec(cleanCtx, `DELETE FROM tours WHERE id=ANY($1::uuid[])`,
			[]uuid.UUID{f.tourID, f.foreignTourID})
		_, _ = pool.Exec(cleanCtx, `DELETE FROM apps WHERE id=$1`, f.appID)
	})
	_, err := pool.Exec(ctx, `INSERT INTO apps(id,name,public_key) VALUES($1,'MVP11 restore test',$2)`,
		f.appID, "pk_"+f.appID.String())
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO tours(id,app_id,title,enabled) VALUES
        ($1,$3,'Restore fixture',true),($2,$3,'Foreign fixture',true)`,
		f.tourID, f.foreignTourID, f.appID)
	require.NoError(t, err)

	// Initial draft -> publish v1 -> edit v2 draft -> publish v2 -> v3 draft.
	initialDraft := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO tour_versions
        (id,tour_id,version,status,target_path,trigger_type,trigger_config,audience_rules,
         audience_show_once,audience_max_shows,audience_only_new)
        VALUES($1,$2,1,'draft','/first','delay',$3::jsonb,$4::jsonb,true,3,true)`,
		initialDraft, f.tourID, `{"delay_ms":1200}`, `[{"country":"LV"}]`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO hints(tour_version_id,step,title,content,selector,placement)
        VALUES ($1,1,'First','First content','#first','top'),
               ($1,2,'Second','Second content','#second','bottom')`, initialDraft)
	require.NoError(t, err)
	first, err := f.repo.Publish(ctx, f.tourID)
	require.NoError(t, err)
	require.Equal(t, initialDraft, first.PublishedVersionId)
	f.archivedID = first.PublishedVersionId

	_, err = pool.Exec(ctx, `UPDATE tour_versions SET
        target_path='/second',trigger_type='on_load',trigger_config='{}'::jsonb,
        audience_rules='[]'::jsonb,audience_show_once=false,
        audience_max_shows=0,audience_only_new=false WHERE id=$1`, first.NewDraftVersionId)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM hints WHERE tour_version_id=$1`, first.NewDraftVersionId)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO hints(tour_version_id,step,title,content,selector,placement)
        VALUES ($1,1,'New current','Current content','#current','right')`, first.NewDraftVersionId)
	require.NoError(t, err)
	second, err := f.repo.Publish(ctx, f.tourID)
	require.NoError(t, err)
	f.publishedID = second.PublishedVersionId
	f.draftID = second.NewDraftVersionId
	require.NotEqual(t, f.archivedID, f.publishedID)
	require.NotEqual(t, f.publishedID, f.draftID)
	return f
}

func restoreHints(t *testing.T, f *restoreFixture, versionID uuid.UUID) []restoreHint {
	t.Helper()
	rows, err := f.pool.Query(context.Background(), `SELECT id,step,title,content,selector,placement
        FROM hints WHERE tour_version_id=$1 ORDER BY step`, versionID)
	require.NoError(t, err)
	defer rows.Close()
	result := []restoreHint{}
	for rows.Next() {
		var h restoreHint
		require.NoError(t, rows.Scan(&h.ID, &h.Step, &h.Title, &h.Content, &h.Selector, &h.Placement))
		result = append(result, h)
	}
	require.NoError(t, rows.Err())
	return result
}

type restoreVersionSnapshot struct {
	ID                                   uuid.UUID
	Version                              int
	Status, Path, Trigger, Config, Rules string
	ShowOnce, OnlyNew                    bool
	MaxShows                             int
	PublishedAt                          *time.Time
	Hints                                []restoreHint
}

func restoreSnapshot(t *testing.T, f *restoreFixture, versionID uuid.UUID) restoreVersionSnapshot {
	t.Helper()
	var v restoreVersionSnapshot
	err := f.pool.QueryRow(context.Background(), `SELECT id,version,status,target_path,trigger_type,
        trigger_config::text,audience_rules::text,audience_show_once,audience_max_shows,
        audience_only_new,published_at FROM tour_versions WHERE id=$1`, versionID).
		Scan(&v.ID, &v.Version, &v.Status, &v.Path, &v.Trigger, &v.Config, &v.Rules,
			&v.ShowOnce, &v.MaxShows, &v.OnlyNew, &v.PublishedAt)
	require.NoError(t, err)
	v.Hints = restoreHints(t, f, versionID)
	return v
}

func requireRestoreStatus(t *testing.T, err error, want int) {
	t.Helper()
	require.Error(t, err)
	var logic *domains.LogicError
	require.ErrorAs(t, err, &logic)
	require.Equal(t, want, logic.Code)
}

func TestMVP11VersionHistoryAndRestoreIntegration(t *testing.T) {
	f := newRestoreFixture(t)
	ctx := context.Background()
	history, err := f.repo.ListVersions(ctx, f.tourID)
	require.NoError(t, err)
	require.Len(t, history, 3)
	require.Equal(t, []int{3, 2, 1}, []int{history[0].Version, history[1].Version, history[2].Version})
	require.Equal(t, []domains.TourStatus{domains.TourDraft, domains.TourPublished, domains.TourArchived},
		[]domains.TourStatus{history[0].Status, history[1].Status, history[2].Status})
	require.Equal(t, []uuid.UUID{f.draftID, f.publishedID, f.archivedID},
		[]uuid.UUID{history[0].Id, history[1].Id, history[2].Id})
	require.Equal(t, []int{1, 1, 2}, []int{history[0].StepsCount, history[1].StepsCount, history[2].StepsCount})
	require.Nil(t, history[0].PublishedAt)
	require.NotNil(t, history[1].PublishedAt)
	require.NotNil(t, history[2].PublishedAt)

	sourceBefore := restoreSnapshot(t, f, f.archivedID)
	publishedBefore := restoreSnapshot(t, f, f.publishedID)
	draftBefore := restoreSnapshot(t, f, f.draftID)
	require.NotEqual(t, sourceBefore.Path, draftBefore.Path)

	result, err := f.repo.RestoreVersion(ctx, f.tourID, f.archivedID)
	require.NoError(t, err)
	require.Equal(t, f.tourID, result.TourId)
	require.Equal(t, f.archivedID, result.RestoredFromVersionId)
	require.Equal(t, f.draftID, result.DraftVersionId)
	require.Equal(t, 3, result.DraftVersion)
	require.Equal(t, 2, result.StepsCount)

	restored := restoreSnapshot(t, f, f.draftID)
	require.Equal(t, draftBefore.ID, restored.ID)
	require.Equal(t, draftBefore.Version, restored.Version)
	require.Equal(t, "draft", restored.Status)
	require.Nil(t, restored.PublishedAt)
	require.Equal(t, sourceBefore.Path, restored.Path)
	require.Equal(t, sourceBefore.Trigger, restored.Trigger)
	require.JSONEq(t, sourceBefore.Config, restored.Config)
	require.JSONEq(t, sourceBefore.Rules, restored.Rules)
	require.Equal(t, sourceBefore.ShowOnce, restored.ShowOnce)
	require.Equal(t, sourceBefore.MaxShows, restored.MaxShows)
	require.Equal(t, sourceBefore.OnlyNew, restored.OnlyNew)
	require.Len(t, restored.Hints, len(sourceBefore.Hints))
	for i := range sourceBefore.Hints {
		want, got := sourceBefore.Hints[i], restored.Hints[i]
		require.Equal(t, want.Step, got.Step)
		require.Equal(t, want.Title, got.Title)
		require.Equal(t, want.Content, got.Content)
		require.Equal(t, want.Selector, got.Selector)
		require.Equal(t, want.Placement, got.Placement)
		require.NotEqual(t, want.ID, got.ID, "restored hints must get fresh UUIDs")
		for _, old := range draftBefore.Hints {
			require.NotEqual(t, old.ID, got.ID)
		}
	}
	require.Equal(t, sourceBefore, restoreSnapshot(t, f, f.archivedID), "archived version must remain unchanged")
	require.Equal(t, publishedBefore, restoreSnapshot(t, f, f.publishedID), "published version must remain unchanged")

	// Existing MVP-07 publish must publish the restored draft, then create v4 draft.
	publish, err := f.repo.Publish(ctx, f.tourID)
	require.NoError(t, err)
	require.Equal(t, f.draftID, publish.PublishedVersionId)
	require.Equal(t, 3, publish.Version)
	newDraft := restoreSnapshot(t, f, publish.NewDraftVersionId)
	require.Equal(t, 4, newDraft.Version)
	require.Equal(t, "draft", newDraft.Status)
	require.Equal(t, sourceBefore.Path, newDraft.Path)
	require.Len(t, newDraft.Hints, 2)
	require.Equal(t, sourceBefore, restoreSnapshot(t, f, f.archivedID))
	var previousStatus string
	require.NoError(t, f.pool.QueryRow(ctx, `SELECT status FROM tour_versions WHERE id=$1`, f.publishedID).Scan(&previousStatus))
	require.Equal(t, "archived", previousStatus)
}

func TestMVP11RestoreInvalidVersionsIntegration(t *testing.T) {
	f := newRestoreFixture(t)
	ctx := context.Background()
	foreignVersion := uuid.New()
	_, err := f.pool.Exec(ctx, `INSERT INTO tour_versions(id,tour_id,version,status,target_path)
        VALUES($1,$2,1,'draft','/foreign')`, foreignVersion, f.foreignTourID)
	require.NoError(t, err)

	_, err = f.repo.ListVersions(ctx, uuid.New())
	requireRestoreStatus(t, err, http.StatusNotFound)
	cases := []struct {
		name          string
		tour, version uuid.UUID
		status        int
	}{
		{"unknown tour", uuid.New(), f.archivedID, http.StatusNotFound},
		{"foreign version", f.tourID, foreignVersion, http.StatusNotFound},
		{"unknown version", f.tourID, uuid.New(), http.StatusNotFound},
		{"source is draft", f.tourID, f.draftID, http.StatusNotFound},
		{"source is current published", f.tourID, f.publishedID, http.StatusNotFound},
	}
	original := restoreSnapshot(t, f, f.draftID)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.repo.RestoreVersion(ctx, tc.tour, tc.version)
			requireRestoreStatus(t, err, tc.status)
			require.Equal(t, original, restoreSnapshot(t, f, f.draftID))
		})
	}
	_, err = f.pool.Exec(ctx, `DELETE FROM hints WHERE tour_version_id=$1`, f.draftID)
	require.NoError(t, err)
	_, err = f.pool.Exec(ctx, `DELETE FROM tour_versions WHERE id=$1`, f.draftID)
	require.NoError(t, err)
	_, err = f.repo.RestoreVersion(ctx, f.tourID, f.archivedID)
	requireRestoreStatus(t, err, http.StatusConflict)
}

func TestMVP11RestoreRollsBackOnHintCopyFailureIntegration(t *testing.T) {
	f := newRestoreFixture(t)
	ctx := context.Background()
	before := restoreSnapshot(t, f, f.draftID)
	archivedBefore := restoreSnapshot(t, f, f.archivedID)
	publishedBefore := restoreSnapshot(t, f, f.publishedID)

	// Force a failure during INSERT of the restored hints, after draft UPDATE and DELETE.
	// The trigger rejects only the fixture's draft and is removed by t.Cleanup.
	suffix := uuid.NewString()[:8]
	functionName := "mvp11_reject_hint_" + suffix
	triggerName := "mvp11_reject_hint_" + suffix
	_, err := f.pool.Exec(ctx, fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$
        BEGIN
          IF NEW.tour_version_id = '%s'::uuid THEN
             RAISE EXCEPTION 'MVP-11 forced hint insertion failure';
          END IF;
          RETURN NEW;
        END $$`, functionName, f.draftID))
	require.NoError(t, err, "test database must allow CREATE FUNCTION")
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = f.pool.Exec(cleanCtx, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON hints`, triggerName))
		_, _ = f.pool.Exec(cleanCtx, fmt.Sprintf(`DROP FUNCTION IF EXISTS %s()`, functionName))
	})
	_, err = f.pool.Exec(ctx, fmt.Sprintf(`CREATE TRIGGER %s BEFORE INSERT ON hints
        FOR EACH ROW EXECUTE FUNCTION %s()`, triggerName, functionName))
	require.NoError(t, err, "test database must allow CREATE TRIGGER")

	_, err = f.repo.RestoreVersion(ctx, f.tourID, f.archivedID)
	require.Error(t, err)
	require.Equal(t, before, restoreSnapshot(t, f, f.draftID), "draft must survive failed transaction intact")
	require.Equal(t, archivedBefore, restoreSnapshot(t, f, f.archivedID))
	require.Equal(t, publishedBefore, restoreSnapshot(t, f, f.publishedID))
}
