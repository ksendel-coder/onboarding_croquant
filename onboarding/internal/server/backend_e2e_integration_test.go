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

	profilePath := "/v1/auth/me"
	profileBody := map[string]any{"first_name": "  Мария  ", "last_name": "Петрова"}
	expect(request(http.MethodPatch, profilePath, profileBody, "", "", ""), http.StatusUnauthorized)
	expect(request(http.MethodPatch, profilePath, map[string]any{"first_name": " ", "last_name": "Петрова"}, token, "", ""), http.StatusBadRequest)
	profileRes := request(http.MethodPatch, profilePath, profileBody, token, "", "")
	expect(profileRes, http.StatusOK)
	profile := decode(profileRes)
	require.Equal(t, userID.String(), profile["id"])
	require.Equal(t, "Мария", profile["first_name"])
	require.Equal(t, "Петрова", profile["last_name"])
	require.Equal(t, email, profile["email"])
	require.NotContains(t, profile, "password_hash")
	var savedFirst, savedLast string
	require.NoError(t, pool.QueryRow(ctx, `SELECT first_name, last_name FROM users WHERE id=$1`, userID).Scan(&savedFirst, &savedLast))
	require.Equal(t, "Мария", savedFirst)
	require.Equal(t, "Петрова", savedLast)
	profileMe := request(http.MethodGet, profilePath, nil, token, "", "")
	expect(profileMe, http.StatusOK)
	require.Equal(t, profile, decode(profileMe))
	profileLogin := request(http.MethodPost, "/v1/auth/login", map[string]any{
		"email": email, "password": "Mvp12TestPassword!",
	}, "", "", "")
	expect(profileLogin, http.StatusOK)
	require.Equal(t, profile, decode(profileLogin)["user"])

	origin := "http://localhost:5173"
	appRes := request(http.MethodPost, "/v1/apps", map[string]any{
		"name": "mvp12-e2e-" + uuid.NewString(), "allowed_origins": []string{origin},
	}, token, "", "")
	expect(appRes, http.StatusCreated)
	app := decode(appRes)
	appID = uuid.MustParse(stringField(app, "id"))
	key := stringField(app, "public_key")
	require.Equal(t, map[string]any{"primary_color": "#696CF1", "background_color": "#FFFFFF", "text_color": "#181818"}, app["theme"])
	theme := map[string]any{"primary_color": "#123ABC", "background_color": "#F0F0F0", "text_color": "#202020"}
	appPath := "/v1/apps/" + appID.String()
	themeBody := map[string]any{"theme": theme}
	expect(request(http.MethodPatch, appPath, themeBody, "", key, origin), http.StatusUnauthorized)
	expect(request(http.MethodPatch, appPath, map[string]any{"theme": map[string]any{"primary_color": "red"}}, token, "", ""), http.StatusBadRequest)
	patched := request(http.MethodPatch, appPath, themeBody, token, "", "")
	expect(patched, http.StatusOK)
	require.Equal(t, theme, decode(patched)["theme"])

	tourRes := request(http.MethodPost, "/v1/tours?appId="+appID.String(), map[string]any{
		"title": "MVP12 E2E", "description": "end to end", "enabled": true,
		"priority": 100, "target_path": "/mvp12",
	}, token, "", "")
	expect(tourRes, http.StatusCreated)
	tourID = uuid.MustParse(stringField(decode(tourRes), "id"))

	for _, step := range []int{1, 2} {
		pagePath := ""
		if step == 2 {
			pagePath = "/mvp12/profile"
		}
		res := request(http.MethodPost, "/v1/tours/"+tourID.String()+"/hints", map[string]any{
			"title": fmt.Sprintf("Step %d", step), "content": "Click next",
			"selector": fmt.Sprintf("#step-%d", step), "placement": "bottom", "page_path": pagePath,
		}, token, "", "")
		expect(res, http.StatusCreated)
	}
	expect(request(http.MethodPatch, "/v1/tours/"+tourID.String()+"/settings", map[string]any{
		"trigger_type": "delay", "trigger_config": map[string]any{"delay_ms": 2000},
	}, token, "", ""), http.StatusNoContent)
	publish := request(http.MethodPost, "/v1/tours/"+tourID.String()+"/publish", nil, token, "", "")
	expect(publish, http.StatusOK)
	draftHints := request(http.MethodGet, "/v1/tours/"+tourID.String()+"/hints", nil, token, "", "")
	expect(draftHints, http.StatusOK)
	var draftSteps []map[string]any
	require.NoError(t, json.Unmarshal(draftHints.Body.Bytes(), &draftSteps))
	require.Len(t, draftSteps, 2)
	require.Equal(t, "/mvp12/profile", draftSteps[1]["page_path"])

	// 2. Runtime: authorization, Origin, matching and idempotent resolve.
	resolvePath := "/v1/resolve"
	subject := "mvp12-subject-" + uuid.NewString()
	resolveBody := map[string]any{
		"url": origin + "/mvp12", "subject_id": subject,
		"session_id": "mvp12-session", "props": map[string]any{},
	}
	expect(request(http.MethodPost, resolvePath, resolveBody, "", "", origin), http.StatusUnauthorized)
	expect(request(http.MethodPost, resolvePath, resolveBody, "", key, "https://untrusted.example"), http.StatusForbidden)
	resolveBody["defer_start"] = true
	resolveBody["start_version_id"] = uuid.New()
	candidate := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	expect(candidate, http.StatusOK)
	pending := decode(candidate)
	require.Equal(t, true, pending["awaiting_trigger"])
	require.NotContains(t, pending, "attempt_id")
	var countBeforeTrigger int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM tour_progress WHERE app_id=$1 AND subject_id=$2", appID, subject).Scan(&countBeforeTrigger))
	require.Zero(t, countBeforeTrigger)
	resolveBody["start_version_id"] = pending["tour_version_id"]
	first := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	delete(resolveBody, "start_version_id")
	expect(first, http.StatusOK)
	resolved := decode(first)
	require.Equal(t, theme, resolved["theme"])
	versionID := uuid.MustParse(stringField(resolved, "tour_version_id"))
	initialAttempt := stringField(resolved, "attempt_id")
	tourJSON, ok := resolved["tour"].(map[string]any)
	require.True(t, ok)
	hints, ok := tourJSON["hints"].([]any)
	require.True(t, ok)
	require.Len(t, hints, 2)
	require.Equal(t, "/mvp12/profile", hints[1].(map[string]any)["page_path"])
	firstHint := uuid.MustParse(stringField(hints[0].(map[string]any), "id"))
	secondHint := uuid.MustParse(stringField(hints[1].(map[string]any), "id"))
	require.Equal(t, firstHint.String(), stringField(resolved, "current_hint_id"))
	again := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	delete(resolveBody, "defer_start")
	expect(again, http.StatusOK)
	require.Equal(t, versionID.String(), stringField(decode(again), "tour_version_id"))
	require.Equal(t, initialAttempt, stringField(decode(again), "attempt_id"))

	var shows int
	require.NoError(t, pool.QueryRow(ctx, `SELECT shows_count FROM tour_progress
        WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`, appID, subject, tourID).Scan(&shows))
	require.Equal(t, 1, shows, "repeat resolve must not increment shows_count")

	noMatch := map[string]any{"url": origin + "/not-this-path", "subject_id": uuid.NewString(), "session_id": "s", "props": map[string]any{}}
	expect(request(http.MethodPost, resolvePath, noMatch, "", key, origin), http.StatusNoContent)

	// 3. Real event ingestion, deduplication, progress updates and analytics.
	attemptID := uuid.MustParse(initialAttempt)
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
	resolveBody["session_id"] = "reopened-tab"
	resumed := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	expect(resumed, http.StatusOK)
	require.Equal(t, attemptID.String(), stringField(decode(resumed), "attempt_id"))
	destination := map[string]any{"url": origin + "/mvp12/profile", "subject_id": subject, "session_id": "destination"}
	expect(request(http.MethodPost, resolvePath, destination, "", key, origin), http.StatusNoContent)
	expect(request(http.MethodPost, eventsPath, event("hint_shown", &firstHint), "", key, origin), http.StatusAccepted)
	expect(request(http.MethodPost, eventsPath, event("hint_completed", &firstHint), "", key, origin), http.StatusAccepted)
	onDestination := request(http.MethodPost, resolvePath, destination, "", key, origin)
	expect(onDestination, http.StatusOK)
	require.Equal(t, secondHint.String(), stringField(decode(onDestination), "current_hint_id"))
	require.Equal(t, attemptID.String(), stringField(decode(onDestination), "attempt_id"))
	for _, body := range []map[string]any{
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

	skipSubject := uuid.NewString()
	skipResolve := request(http.MethodPost, resolvePath, map[string]any{
		"url": origin + "/mvp12", "subject_id": skipSubject, "session_id": "skip-test",
	}, "", key, origin)
	expect(skipResolve, http.StatusOK)
	skipAttempt := stringField(decode(skipResolve), "attempt_id")
	recordSkipAttempt := func(body map[string]any, status int) {
		body["subject_id"], body["attempt_id"] = skipSubject, skipAttempt
		expect(request(http.MethodPost, eventsPath, body, "", key, origin), status)
	}
	recordSkipAttempt(event("tour_started", nil), http.StatusAccepted)
	skipped := event("hint_skipped", &firstHint)
	recordSkipAttempt(skipped, http.StatusAccepted)
	recordSkipAttempt(skipped, http.StatusOK)
	skipResume := request(http.MethodPost, resolvePath, map[string]any{
		"url": origin + "/mvp12", "subject_id": skipSubject, "session_id": "skip-resume",
	}, "", key, origin)
	expect(skipResume, http.StatusOK)
	require.Equal(t, secondHint.String(), stringField(decode(skipResume), "current_hint_id"))
	recordSkipAttempt(event("hint_skipped", &secondHint), http.StatusAccepted)
	recordSkipAttempt(event("tour_completed", nil), http.StatusAccepted)
	skipAnalytics := request(http.MethodGet, analyticsPath, nil, token, "", "")
	expect(skipAnalytics, http.StatusOK)
	skipMetrics := decode(skipAnalytics)
	require.EqualValues(t, 2, skipMetrics["started"])
	require.EqualValues(t, 1, skipMetrics["completed"])
	require.EqualValues(t, 0, skipMetrics["dismissed"])
	require.EqualValues(t, 50, skipMetrics["completion_rate"])
	firstStep := skipMetrics["steps"].([]any)[0].(map[string]any)
	require.EqualValues(t, 1, firstStep["shown"])
	require.EqualValues(t, 1, firstStep["completed"])
	require.EqualValues(t, 1, firstStep["skipped"])
	require.EqualValues(t, 1, firstStep["not_completed"])
	require.NotEmpty(t, stringField(firstStep, "last_skipped_at"))

	repeat := request(http.MethodPost, resolvePath, resolveBody, "", key, origin)
	expect(repeat, http.StatusOK)
	repeatBody := decode(repeat)
	repeatAttempt := stringField(repeatBody, "attempt_id")
	require.NotEqual(t, initialAttempt, repeatAttempt)
	require.Equal(t, versionID.String(), stringField(repeatBody, "tour_version_id"))
	require.Equal(t, firstHint.String(), stringField(repeatBody, "current_hint_id"))
	expect(request(http.MethodPost, eventsPath, event("tour_dismissed", nil), "", key, origin), http.StatusConflict)
	repeatStart := event("tour_started", nil)
	repeatStart["attempt_id"] = repeatAttempt
	expect(request(http.MethodPost, eventsPath, repeatStart, "", key, origin), http.StatusAccepted)
	repeatDismiss := event("tour_dismissed", nil)
	repeatDismiss["attempt_id"] = repeatAttempt
	expect(request(http.MethodPost, eventsPath, repeatDismiss, "", key, origin), http.StatusAccepted)

	// 4. Republish different audience settings; old active progress is not reassigned.
	oldSubject := "mvp12-active-" + uuid.NewString()
	oldBody := map[string]any{"url": origin + "/mvp12", "subject_id": oldSubject,
		"session_id": "old-session", "props": map[string]any{}}
	oldActive := request(http.MethodPost, resolvePath, oldBody, "", key, origin)
	expect(oldActive, http.StatusOK)
	require.Equal(t, versionID.String(), stringField(decode(oldActive), "tour_version_id"))
	oldAttempt := stringField(decode(oldActive), "attempt_id")
	postOldEvent := func(kind string, hint *uuid.UUID) {
		body := event(kind, hint)
		body["subject_id"], body["attempt_id"] = oldSubject, oldAttempt
		expect(request(http.MethodPost, eventsPath, body, "", key, origin), http.StatusAccepted)
	}
	postOldEvent("tour_started", nil)
	postOldEvent("hint_shown", &firstHint)
	postOldEvent("hint_completed", &firstHint)

	settings := request(http.MethodPatch, "/v1/tours/"+tourID.String()+"/settings", map[string]any{
		"audience_show_once": true, "audience_max_shows": 1, "audience_only_new": true,
	}, token, "", "")
	expect(settings, http.StatusNoContent)
	expect(request(http.MethodPost, "/v1/tours/"+tourID.String()+"/publish", nil, token, "", ""), http.StatusOK)

	// The new version's audience restrictions still apply to finished visitors.
	expect(request(http.MethodPost, resolvePath, resolveBody, "", key, origin), http.StatusNoContent)
	oldBody["url"] = origin + "/mvp12/profile"
	oldBody["session_id"] = "reloaded-after-publication"
	archivedResume := request(http.MethodPost, resolvePath, oldBody, "", key, origin)
	expect(archivedResume, http.StatusOK)
	archived := decode(archivedResume)
	require.Equal(t, versionID.String(), stringField(archived, "tour_version_id"))
	require.Equal(t, oldAttempt, stringField(archived, "attempt_id"))
	require.Equal(t, secondHint.String(), stringField(archived, "current_hint_id"))
	var savedVersion uuid.UUID
	require.NoError(t, pool.QueryRow(ctx, `SELECT tour_version_id FROM tour_progress
        WHERE app_id=$1 AND subject_id=$2 AND tour_id=$3`, appID, oldSubject, tourID).Scan(&savedVersion))
	require.Equal(t, versionID, savedVersion)

	postOldEvent("hint_shown", &secondHint)

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
