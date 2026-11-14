package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/repositories"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// Requires an ISOLATED PostgreSQL database with all migrations applied.
// Do not use a production or shared development DB.
func TestMVP12BackendE2ELifecycle(t *testing.T) {
	dsn := os.Getenv("MVP12_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MVP12_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	logger := zerolog.Nop()
	authRepo := repositories.NewAuthRepository(pool, &logger)
	tourRepo := repositories.NewTourRepository(pool, &logger)
	hintRepo := repositories.NewHintRepository(pool, &logger)
	runtimeRepo := repositories.NewRuntimeRepository(pool, &logger)
	eventRepo := repositories.NewEventRepository(pool)

	authDomain := domains.NewAuthDomain(authRepo)
	tourDomain := domains.NewTourDomain(tourRepo)
	hintDomain := domains.NewHintDomain(tourRepo, hintRepo)
	runtimeDomain := domains.NewRuntimeDomain(tourRepo, hintRepo, runtimeRepo)
	eventDomain := domains.NewEventDomain(eventRepo)

	gin.DefaultWriter, gin.DefaultErrorWriter = io.Discard, io.Discard
	handler := NewServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &CreateStruct{
		Logger:     &logger,
		AuthDomain: authDomain, AuthController: controllers.NewAuthController(authDomain),
		RuntimeDomain: runtimeDomain, RuntimeController: controllers.NewRuntimeController(runtimeDomain),
		TourController:  controllers.NewTourController(tourDomain),
		HintController:  controllers.NewHintController(hintDomain),
		EventController: controllers.NewEventController(eventDomain),
	}).Handler()

	request := func(method, path string, body any, token, key, origin string) *httptest.ResponseRecorder {
		t.Helper()
		var data []byte
		if body != nil {
			var marshalErr error
			data, marshalErr = json.Marshal(body)
			require.NoError(t, marshalErr)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if key != "" {
			req.Header.Set("X-App-Key", key)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}
	expect := func(rec *httptest.ResponseRecorder, code int) {
		t.Helper()
		require.Equal(t, code, rec.Code, "body: %s", rec.Body.String())
	}
	decode := func(rec *httptest.ResponseRecorder) map[string]any {
		t.Helper()
		var obj map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &obj), "body: %s", rec.Body.String())
		return obj
	}
	stringField := func(obj map[string]any, key string) string {
		t.Helper()
		value, ok := obj[key].(string)
		require.True(t, ok && value != "", "missing %s: %#v", key, obj)
		return value
	}

	// 1. Create an isolated user, app, draft, hints and published version through real HTTP routes.
	email := "mvp12-" + uuid.NewString() + "@example.test"
	registration := request(http.MethodPost, "/v1/auth/register", map[string]any{
		"email": email, "password": "Mvp12TestPassword!",
		"first_name": "MVP", "last_name": "Twelve",
	}, "", "", "")
	expect(registration, http.StatusCreated)
	auth := decode(registration)
	token := stringField(auth, "token")
	user := auth["user"].(map[string]any)
	userID := uuid.MustParse(stringField(user, "id"))

	appID := uuid.Nil
	tourID := uuid.Nil
	// Cleanup is scoped to the UUIDs created by this test; do not truncate tables.
	t.Cleanup(func() {
		cleanup := context.Background()
		if tourID != uuid.Nil {
			_, _ = pool.Exec(cleanup, "DELETE FROM tour_events WHERE tour_id=$1", tourID)
			_, _ = pool.Exec(cleanup, "DELETE FROM tour_progress WHERE tour_id=$1", tourID)
			_, _ = pool.Exec(cleanup, "DELETE FROM hints WHERE tour_version_id IN (SELECT id FROM tour_versions WHERE tour_id=$1)", tourID)
			_, _ = pool.Exec(cleanup, "DELETE FROM tour_versions WHERE tour_id=$1", tourID)
			_, _ = pool.Exec(cleanup, "DELETE FROM tours WHERE id=$1", tourID)
		}
		if appID != uuid.Nil {
			_, _ = pool.Exec(cleanup, "DELETE FROM apps WHERE id=$1", appID)
		}
		_, _ = pool.Exec(cleanup, "DELETE FROM users WHERE id=$1", userID)
	})

	origin := "http://localhost:5173"
	appRes := request(http.MethodPost, "/v1/apps", map[string]any{
		"name": "mvp12-e2e-" + uuid.NewString(), "allowed_origins": []string{origin},
	}, token, "", "")
	expect(appRes, http.StatusCreated)
	app := decode(appRes)
	appID = uuid.MustParse(stringField(app, "id"))
	key := stringField(app, "public_key")

	tourRes := request(http.MethodPost, "/v1/tours?appId="+appID.String(), map[string]any{
		"title": "MVP12 E2E", "description": "end to end", "enabled": true,
		"priority": 100, "target_path": "/mvp12",
	}, token, "", "")
	expect(tourRes, http.StatusCreated)
	tourID = uuid.MustParse(stringField(decode(tourRes), "id"))

	for _, step := range []int{1, 2} {
		res := request(http.MethodPost, "/v1/tours/"+tourID.String()+"/hints", map[string]any{
			"title": fmt.Sprintf("Step %d", step), "content": "Click next",
			"selector": fmt.Sprintf("#step-%d", step), "placement": "bottom",
		}, token, "", "")
		expect(res, http.StatusCreated)
	}
	publish := request(http.MethodPost, "/v1/tours/"+tourID.String()+"/publish", nil, token, "", "")
	expect(publish, http.StatusOK)

	// 2. Runtime: authorization, Origin, matching and idempotent resolve.
	resolvePath := "/v1/resolve"
	subject := "mvp12-subject-" + uuid.NewString()
	resolveBody := map[string]any{
		"url": origin + "/mvp12", "subject_id": subject,
		"session_id": "mvp12-session", "props": map[string]any{},
	}
	expect(request(http.MethodPost, resolvePath, resolveBody, "", "", origin), http.StatusUnauthorized)
	expect(request(http.MethodPost, resolvePath, resolveBody, "", key, "https://untrusted.example"), http.StatusForbidden)
	first := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	expect(first, http.StatusOK)
	resolved := decode(first)
	versionID := uuid.MustParse(stringField(resolved, "tour_version_id"))
	tourJSON, ok := resolved["tour"].(map[string]any)
	require.True(t, ok)
	hints, ok := tourJSON["hints"].([]any)
	require.True(t, ok)
	require.Len(t, hints, 2)
	firstHint := uuid.MustParse(stringField(hints[0].(map[string]any), "id"))
	secondHint := uuid.MustParse(stringField(hints[1].(map[string]any), "id"))
	require.Equal(t, firstHint.String(), stringField(resolved, "current_hint_id"))
	again := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	expect(again, http.StatusOK)
	require.Equal(t, versionID.String(), stringField(decode(again), "tour_version_id"))

	var shows int
	require.NoError(t, pool.QueryRow(ctx, `SELECT shows_count FROM tour_progress
        WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`, appID, subject, tourID).Scan(&shows))
	require.Equal(t, 1, shows, "repeat resolve must not increment shows_count")

	noMatch := map[string]any{"url": origin + "/not-this-path", "subject_id": uuid.NewString(), "session_id": "s", "props": map[string]any{}}
	expect(request(http.MethodPost, resolvePath, noMatch, "", key, origin), http.StatusNoContent)

	// 3. Real event ingestion, deduplication, progress updates and analytics.
	attemptID := uuid.New()
	event := func(typ string, hintID *uuid.UUID) map[string]any {
		body := map[string]any{
			"event_id": uuid.New(), "attempt_id": attemptID,
			"subject_id": subject, "tour_id": tourID, "tour_version_id": versionID,
			"type": typ,
		}
		if hintID != nil {
			body["hint_id"] = *hintID
		}
		return body
	}
	eventsPath := "/v1/events"
	started := event("tour_started", nil)
	expect(request(http.MethodPost, eventsPath, started, "", key, origin), http.StatusAccepted)
	expect(request(http.MethodPost, eventsPath, started, "", key, origin), http.StatusOK)
	for _, body := range []map[string]any{
		event("hint_shown", &firstHint),
		event("hint_completed", &firstHint),
		event("hint_shown", &secondHint),
		event("hint_completed", &secondHint),
		event("tour_completed", nil),
	} {
		expect(request(http.MethodPost, eventsPath, body, "", key, origin), http.StatusAccepted)
	}
	var status string
	require.NoError(t, pool.QueryRow(ctx, `SELECT status, shows_count FROM tour_progress
        WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`, appID, subject, tourID).Scan(&status, &shows))
	require.Equal(t, "completed", status)
	require.Equal(t, 1, shows)
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM tour_events WHERE app_id=$1 AND tour_id=$2 AND subject_id=$3`,
		appID, tourID, subject).Scan(&count))
	require.Equal(t, 6, count, "repeated event_id must not create a duplicate")

	analyticsPath := "/v1/tours/" + tourID.String() + "/analytics?version_id=" + versionID.String()
	expect(request(http.MethodGet, analyticsPath, nil, "", key, ""), http.StatusUnauthorized)
	analysis := request(http.MethodGet, analyticsPath, nil, token, "", "")
	expect(analysis, http.StatusOK)
	metrics := decode(analysis)
	require.EqualValues(t, 1, metrics["started"])
	require.EqualValues(t, 1, metrics["completed"])

	// 4. Republish different audience settings; old active progress is not reassigned.
	oldSubject := "mvp12-active-" + uuid.NewString()
	oldBody := map[string]any{"url": origin + "/mvp12", "subject_id": oldSubject,
		"session_id": "old-session", "props": map[string]any{}}
	oldActive := request(http.MethodPost, resolvePath, oldBody, "", key, origin)
	expect(oldActive, http.StatusOK)
	require.Equal(t, versionID.String(), stringField(decode(oldActive), "tour_version_id"))

	settings := request(http.MethodPatch, "/v1/tours/"+tourID.String()+"/settings", map[string]any{
		"audience_show_once": true, "audience_max_shows": 1, "audience_only_new": true,
	}, token, "", "")
	expect(settings, http.StatusNoContent)
	expect(request(http.MethodPost, "/v1/tours/"+tourID.String()+"/publish", nil, token, "", ""), http.StatusOK)

	// The completed visitor cannot restart and the active old version is not overwritten.
	expect(request(http.MethodPost, resolvePath, resolveBody, "", key, origin), http.StatusNoContent)
	expect(request(http.MethodPost, resolvePath, oldBody, "", key, origin), http.StatusNoContent)
	var savedVersion uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `SELECT tour_version_id FROM tour_progress
        WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`, appID, oldSubject, tourID).Scan(&savedVersion))
	require.Equal(t, versionID, savedVersion)

	// Events for the previous (now archived) version must still be accepted.
	oldStarted := map[string]any{
		"event_id": uuid.New(), "attempt_id": uuid.New(), "subject_id": oldSubject,
		"tour_id": tourID, "tour_version_id": versionID, "type": "tour_started",
	}
	expect(request(http.MethodPost, eventsPath, oldStarted, "", key, origin), http.StatusAccepted)

	newSubject := "mvp12-new-" + uuid.NewString()
	newBody := map[string]any{"url": origin + "/mvp12", "subject_id": newSubject,
		"session_id": "new-session", "props": map[string]any{}}
	newResolve := request(http.MethodPost, resolvePath, newBody, "", key, origin)
	expect(newResolve, http.StatusOK)
	newVersion := stringField(decode(newResolve), "tour_version_id")
	require.NotEqual(t, versionID.String(), newVersion)
	expect(request(http.MethodPost, resolvePath, newBody, "", key, origin), http.StatusOK)
	require.NoError(t, pool.QueryRow(ctx, `SELECT shows_count FROM tour_progress
        WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`, appID, newSubject, tourID).Scan(&shows))
	require.Equal(t, 1, shows)

	// Analytics of the archived version is still queryable and isolated from v2.
	oldAnalytics := request(http.MethodGet, analyticsPath, nil, token, "", "")
	expect(oldAnalytics, http.StatusOK)
	require.EqualValues(t, 1, decode(oldAnalytics)["completed"])
	newAnalytics := request(http.MethodGet,
		"/v1/tours/"+tourID.String()+"/analytics?version_id="+newVersion,
		nil, token, "", "")
	expect(newAnalytics, http.StatusOK)
	require.EqualValues(t, 0, decode(newAnalytics)["completed"])
}
