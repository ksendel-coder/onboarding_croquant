package repositories

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTourSettingsIntegration(t *testing.T) {
	dsn := os.Getenv("MVP08_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MVP08_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	appId := uuid.New()
	tourId := uuid.New()
	noDraftTourId := uuid.New()
	publishedId := uuid.New()
	draftId := uuid.New()
	noDraftPublishedId := uuid.New()

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM hints WHERE tour_version_id IN (
				SELECT id FROM tour_versions WHERE tour_id=ANY($1::uuid[])
			)`, []uuid.UUID{tourId, noDraftTourId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM tour_versions WHERE tour_id=ANY($1::uuid[])`, []uuid.UUID{tourId, noDraftTourId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM tours WHERE id=ANY($1::uuid[])`, []uuid.UUID{tourId, noDraftTourId})
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appId)
	})

	_, err = pool.Exec(ctx, `INSERT INTO apps(id,name,public_key) VALUES($1,'MVP-08 test app',$2)`, appId, "pk_"+appId.String())
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tours(id,app_id,title,enabled) VALUES
			($1,$2,'Settings tour',true),
			($3,$2,'No draft tour',true)`, tourId, appId, noDraftTourId)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_versions(id,tour_id,version,status,target_path,published_at) VALUES
			($1,$2,1,'published','/old',now()),
			($3,$2,2,'draft','/draft',NULL),
			($4,$5,1,'published','/unchanged',now())`,
		publishedId, tourId, draftId, noDraftPublishedId, noDraftTourId)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO hints(tour_version_id,step,title,content,selector,placement) VALUES
			($1,1,'Published step','Old content','#old','bottom'),
			($2,1,'Draft step','New content','#welcome','bottom')`, publishedId, draftId)
	require.NoError(t, err)

	logger := zerolog.Nop()
	repo := NewTourRepository(pool, &logger)
	targetPath := "/dashboard"
	triggerType := domains.TriggerDelay
	delayMs := 1500
	scrollDepth := 60
	inactivitySecs := 30
	elementSelector := "#welcome"
	showOnce := true
	maxShows := 2
	onlyNew := false
	patch := domains.TourSettingsPatch{
		TargetPath:  &targetPath,
		TriggerType: &triggerType,
		TriggerConfig: &domains.TriggerConfig{
			DelayMs:         &delayMs,
			ScrollDepth:     &scrollDepth,
			InactivitySecs:  &inactivitySecs,
			ElementSelector: &elementSelector,
		},
		AudienceShowOnce: &showOnce,
		AudienceMaxShows: &maxShows,
		AudienceOnlyNew:  &onlyNew,
	}

	require.NoError(t, repo.UpdateSettings(ctx, tourId, patch))
	require.NoError(t, repo.UpdateSettings(ctx, tourId, domains.TourSettingsPatch{}))

	var publishedPath string
	var publishedTrigger string
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT target_path,trigger_type FROM tour_versions WHERE id=$1`, publishedId,
	).Scan(&publishedPath, &publishedTrigger))
	assert.Equal(t, "/old", publishedPath)
	assert.Equal(t, "on_load", publishedTrigger)

	var draftPath, draftTrigger, triggerConfig string
	var draftShowOnce, draftOnlyNew bool
	var draftMaxShows int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT target_path,trigger_type,trigger_config::text,
			audience_show_once,audience_max_shows,audience_only_new
		FROM tour_versions WHERE id=$1`, draftId,
	).Scan(&draftPath, &draftTrigger, &triggerConfig, &draftShowOnce, &draftMaxShows, &draftOnlyNew))
	assert.Equal(t, "/dashboard", draftPath)
	assert.Equal(t, "delay", draftTrigger)
	assert.JSONEq(t, `{"delay_ms":1500,"scroll_depth":60,"inactivity_secs":30,"element_selector":"#welcome"}`, triggerConfig)
	assert.True(t, draftShowOnce)
	assert.Equal(t, 2, draftMaxShows)
	assert.False(t, draftOnlyNew)

	err = repo.UpdateSettings(ctx, noDraftTourId, patch)
	var logicError *domains.LogicError
	require.ErrorAs(t, err, &logicError)
	assert.Equal(t, http.StatusConflict, logicError.Code)
	var unchangedPath string
	require.NoError(t, pool.QueryRow(ctx, `SELECT target_path FROM tour_versions WHERE id=$1`, noDraftPublishedId).Scan(&unchangedPath))
	assert.Equal(t, "/unchanged", unchangedPath)

	result, err := repo.Publish(ctx, tourId)
	require.NoError(t, err)
	assert.Equal(t, draftId, result.PublishedVersionId)

	candidates, err := repo.ResolveCandidates(ctx, appId)
	require.NoError(t, err)
	var resolved *domains.Tour
	for _, candidate := range candidates {
		if candidate.Id == tourId {
			resolved = candidate
			break
		}
	}
	require.NotNil(t, resolved)
	assert.Equal(t, draftId, resolved.VersionId)
	assert.Equal(t, "/dashboard", resolved.TargetPath)
	assert.Equal(t, domains.TriggerDelay, resolved.TriggerType)
	require.NotNil(t, resolved.TriggerConfig)
	require.NotNil(t, resolved.TriggerConfig.DelayMs)
	require.NotNil(t, resolved.TriggerConfig.ScrollDepth)
	require.NotNil(t, resolved.TriggerConfig.InactivitySecs)
	require.NotNil(t, resolved.TriggerConfig.ElementSelector)
	assert.Equal(t, 1500, *resolved.TriggerConfig.DelayMs)
	assert.Equal(t, 60, *resolved.TriggerConfig.ScrollDepth)
	assert.Equal(t, 30, *resolved.TriggerConfig.InactivitySecs)
	assert.Equal(t, "#welcome", *resolved.TriggerConfig.ElementSelector)
	require.NotNil(t, resolved.Audience)
	assert.True(t, resolved.Audience.ShowOnce)
	assert.Equal(t, 2, resolved.Audience.MaxShows)
	assert.False(t, resolved.Audience.OnlyNew)
}
