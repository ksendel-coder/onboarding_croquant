package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/controllers"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/http/response"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/repositories"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const previewTestToken = "preview-admin-token"

type previewAuthStub struct {
	domains.AuthDomainInterface
	user    *domains.User
	allowed map[uuid.UUID]bool
}

func (a *previewAuthStub) Authenticate(_ context.Context, token string) (*domains.User, error) {
	if token != previewTestToken {
		return nil, domains.ErrUnauthorized
	}
	return a.user, nil
}

func (a *previewAuthStub) UserHasTour(_ context.Context, _ uuid.UUID, tourId uuid.UUID) (bool, error) {
	return a.allowed[tourId], nil
}

func previewRequest(handler http.Handler, tourId uuid.UUID, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/tours/"+tourId.String()+"/preview", nil)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestTourPreviewPostgreSQLIntegration(t *testing.T) {
	dsn := os.Getenv("MVP10_TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("MVP10_TEST_DATABASE_URL must be set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	appId := uuid.New()
	tourId := uuid.New()
	noDraftTourId := uuid.New()
	archivedTourId := uuid.New()
	emptyTourId := uuid.New()
	tourIds := []uuid.UUID{tourId, noDraftTourId, archivedTourId, emptyTourId}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM tour_events WHERE app_id=$1`, appId)
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM tour_progress WHERE app_id=$1`, appId)
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM hints WHERE tour_version_id IN (
				SELECT id FROM tour_versions WHERE tour_id=ANY($1::uuid[])
			)`, tourIds)
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM tour_versions WHERE tour_id=ANY($1::uuid[])`, tourIds)
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM tours WHERE id=ANY($1::uuid[])`, tourIds)
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appId)
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO apps(id,name,public_key)
		VALUES($1,'MVP-10 preview app',$2)`, appId, "pk_"+appId.String())
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO tours(id,app_id,title,description,enabled,priority,archived_at) VALUES
			($1,$5,'Preview tour','draft preview',true,10,NULL),
			($2,$5,'No draft tour','published only',true,5,NULL),
			($3,$5,'Archived tour','archived',true,1,now()),
			($4,$5,'Empty draft tour','no hints yet',false,0,NULL)`,
		tourId, noDraftTourId, archivedTourId, emptyTourId, appId)
	require.NoError(t, err)

	publishedId := uuid.New()
	draftId := uuid.New()
	noDraftPublishedId := uuid.New()
	archivedDraftId := uuid.New()
	emptyDraftId := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO tour_versions(
			id,tour_id,version,status,trigger_type,target_path,
			audience_show_once,audience_max_shows,audience_only_new,
			trigger_config,audience_rules,published_at
		) VALUES
			($1,$6,1,'published','on_load','/published',false,0,false,'{}','[]',now()),
			($2,$6,2,'draft','delay','/dashboard',true,3,true,
			 '{"delay_ms":1500,"scroll_depth":60,"inactivity_secs":30,"element_selector":"#welcome"}',
			 '[{"Type":"segment","Key":"plan","Operator":"eq","Value":"pro","Timeframe":""}]',NULL),
			($3,$7,1,'published','on_load','/published-only',false,0,false,'{}','[]',now()),
			($4,$8,1,'draft','manual','/archived',false,0,false,'{}','[]',NULL),
			($5,$9,1,'draft','on_load','/empty',false,0,false,'{}','[]',NULL)`,
		publishedId, draftId, noDraftPublishedId, archivedDraftId, emptyDraftId,
		tourId, noDraftTourId, archivedTourId, emptyTourId)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO hints(tour_version_id,step,title,content,selector,placement) VALUES
			($1,1,'Published step','Published content','#published','bottom'),
			($2,2,'Second step','Second draft content','#second','right'),
			($2,1,'First step','First draft content','#first','bottom')`,
		publishedId, draftId)
	require.NoError(t, err)

	logger := zerolog.Nop()
	repo := repositories.NewTourRepository(pool, &logger)
	domain := domains.NewTourDomain(repo)
	auth := &previewAuthStub{
		user: &domains.User{Id: uuid.New()},
		allowed: map[uuid.UUID]bool{
			tourId:         true,
			noDraftTourId:  true,
			archivedTourId: true,
			emptyTourId:    true,
		},
	}
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	srv := NewServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &CreateStruct{
		Logger:            &logger,
		RuntimeController: controllers.NewRuntimeController(nil),
		TourController:    controllers.NewTourController(domain),
		HintController:    controllers.NewHintController(nil),
		AuthDomain:        auth,
		AuthController:    controllers.NewAuthController(nil),
		EventController:   controllers.NewEventController(nil),
	})
	handler := srv.Handler()
	authorized := "Bearer " + previewTestToken

	withoutToken := previewRequest(handler, tourId, "")
	require.Equal(t, http.StatusUnauthorized, withoutToken.Code)
	foreign := previewRequest(handler, uuid.New(), authorized)
	require.Equal(t, http.StatusForbidden, foreign.Code)

	var progressBefore, eventsBefore int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tour_progress WHERE app_id=$1`, appId).Scan(&progressBefore))
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tour_events WHERE app_id=$1`, appId).Scan(&eventsBefore))

	first := previewRequest(handler, tourId, authorized)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	var firstPreview response.Preview
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstPreview))
	require.True(t, firstPreview.Preview)
	require.NotNil(t, firstPreview.Tour)
	assert.Equal(t, tourId, firstPreview.Tour.Id)
	assert.Equal(t, draftId, firstPreview.Tour.VersionId)
	assert.Equal(t, 2, firstPreview.Tour.Version)
	assert.Equal(t, string(domains.TourDraft), firstPreview.Tour.Status)
	assert.Equal(t, "/dashboard", firstPreview.Tour.TargetPath)
	assert.Equal(t, 2, firstPreview.Tour.StepsCount)
	require.Len(t, firstPreview.Tour.Hints, 2)
	assert.Equal(t, 1, firstPreview.Tour.Hints[0].Step)
	assert.Equal(t, 2, firstPreview.Tour.Hints[1].Step)
	assert.Equal(t, "delay", firstPreview.TriggerType)
	require.NotNil(t, firstPreview.TriggerConfig)
	require.NotNil(t, firstPreview.TriggerConfig.DelayMs)
	assert.Equal(t, 1500, *firstPreview.TriggerConfig.DelayMs)
	require.NotNil(t, firstPreview.Audience)
	assert.True(t, firstPreview.Audience.ShowOnce)
	assert.Equal(t, 3, firstPreview.Audience.MaxShows)
	assert.True(t, firstPreview.Audience.OnlyNew)
	require.Len(t, firstPreview.Audience.Rules, 1)

	repeat := previewRequest(handler, tourId, authorized)
	require.Equal(t, http.StatusOK, repeat.Code, repeat.Body.String())

	empty := previewRequest(handler, emptyTourId, authorized)
	require.Equal(t, http.StatusOK, empty.Code, empty.Body.String())
	var emptyPreview response.Preview
	require.NoError(t, json.Unmarshal(empty.Body.Bytes(), &emptyPreview))
	require.NotNil(t, emptyPreview.Tour.Hints)
	assert.Empty(t, emptyPreview.Tour.Hints)
	assert.Contains(t, empty.Body.String(), `"hints":[]`)

	noDraft := previewRequest(handler, noDraftTourId, authorized)
	require.Equal(t, http.StatusConflict, noDraft.Code, noDraft.Body.String())
	archived := previewRequest(handler, archivedTourId, authorized)
	require.Equal(t, http.StatusNotFound, archived.Code, archived.Body.String())

	published, err := repo.Publish(ctx, tourId)
	require.NoError(t, err)
	require.Equal(t, draftId, published.PublishedVersionId)
	_, err = pool.Exec(ctx, `
		UPDATE hints SET content='Changed content in new draft'
		WHERE tour_version_id=$1 AND step=1`, published.NewDraftVersionId)
	require.NoError(t, err)

	afterPublish := previewRequest(handler, tourId, authorized)
	require.Equal(t, http.StatusOK, afterPublish.Code, afterPublish.Body.String())
	var newDraft response.Preview
	require.NoError(t, json.Unmarshal(afterPublish.Body.Bytes(), &newDraft))
	assert.Equal(t, published.NewDraftVersionId, newDraft.Tour.VersionId)
	assert.NotEqual(t, published.PublishedVersionId, newDraft.Tour.VersionId)
	assert.Equal(t, string(domains.TourDraft), newDraft.Tour.Status)
	require.Len(t, newDraft.Tour.Hints, 2)
	assert.Equal(t, "Changed content in new draft", newDraft.Tour.Hints[0].Content)

	var progressAfter, eventsAfter int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tour_progress WHERE app_id=$1`, appId).Scan(&progressAfter))
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tour_events WHERE app_id=$1`, appId).Scan(&eventsAfter))
	assert.Equal(t, progressBefore, progressAfter)
	assert.Equal(t, eventsBefore, eventsAfter)
}
