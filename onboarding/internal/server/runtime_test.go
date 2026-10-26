package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/controllers"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runtimeDomainStub struct {
	domains.RuntimeDomainInterface
	app     *domains.App
	result  *domains.ResolveResult
	lastReq domains.ResolveRequest
}

func (s *runtimeDomainStub) AppByKey(
	_ context.Context,
	key string,
) (*domains.App, error) {
	if s.app == nil || key != s.app.PublicKey {
		return nil, domains.ErrAppNotFound
	}
	return s.app, nil
}

func (s *runtimeDomainStub) Resolve(
	_ context.Context,
	req domains.ResolveRequest,
) (*domains.ResolveResult, error) {
	s.lastReq = req
	return s.result, nil
}

func newRuntimeTestServer(
	t *testing.T,
	domain *runtimeDomainStub,
) http.Handler {
	t.Helper()

	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logger := zerolog.Nop()

	srv := NewServer(
		&config.ServerConfig{Address: "127.0.0.1:0"},
		&CreateStruct{
			Logger:            &logger,
			RuntimeDomain:     domain,
			RuntimeController: controllers.NewRuntimeController(domain),
			TourController:    controllers.NewTourController(nil),
			HintController:    controllers.NewHintController(nil),
		},
	)

	return srv.Handler()
}

func resolveRequest(
	t *testing.T,
	handler http.Handler,
	key string,
	origin string,
	path string,
) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"url":        path,
		"subject_id": "user-1",
		"session_id": "session-1",
		"props":      map[string]any{},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/resolve",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
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

func TestResolveRouteReturnsScenarioHints(t *testing.T) {
	appId := uuid.New()
	versionId := uuid.New()
	hintId := uuid.New()
	domain := &runtimeDomainStub{
		app: &domains.App{
			Id:             appId,
			PublicKey:      "pk_test",
			AllowedOrigins: []string{"http://localhost:5173"},
		},
		result: &domains.ResolveResult{
			Tour: &domains.Tour{
				Id:        uuid.New(),
				AppId:     appId,
				VersionId: versionId,
				Version:   1,
				Status:    domains.TourPublished,
				Enabled:   true,
				Hints: []domains.Hint{{
					Id:            hintId,
					TourVersionId: versionId,
					Step:          1,
				}},
			},
			TourVersionId: versionId,
			Version:       1,
			CurrentHintId: &hintId,
		},
	}
	handler := newRuntimeTestServer(t, domain)

	response := resolveRequest(
		t,
		handler,
		"pk_test",
		"http://localhost:5173",
		"http://localhost:5173/dashboard",
	)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, appId, domain.lastReq.AppId)
	assert.JSONEq(t, `{
		"tour": {
			"id": "`+domain.result.Tour.Id.String()+`",
			"app_id": "`+appId.String()+`",
			"title": "",
			"description": "",
			"enabled": true,
			"priority": 0,
			"created_at": "0001-01-01T00:00:00Z",
			"updated_at": "0001-01-01T00:00:00Z",
			"version_id": "`+versionId.String()+`",
			"version": 1,
			"status": "published",
			"hints": [{
				"id": "`+hintId.String()+`",
				"tour_version_id": "`+versionId.String()+`",
				"step": 1,
				"title": "",
				"content": "",
				"selector": "",
				"placement": "",
				"created_at": "0001-01-01T00:00:00Z",
				"updated_at": "0001-01-01T00:00:00Z"
			}],
			"steps_count": 0
		},
		"tour_version_id": "`+versionId.String()+`",
		"version": 1,
		"current_hint_id": "`+hintId.String()+`"
	}`, response.Body.String())
}

func TestResolveRouteReturnsNoContentWithoutScenario(t *testing.T) {
	domain := &runtimeDomainStub{
		app: &domains.App{
			Id:        uuid.New(),
			PublicKey: "pk_test",
		},
	}
	handler := newRuntimeTestServer(t, domain)

	response := resolveRequest(t, handler, "pk_test", "", "/missing")

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())
}

func TestResolveRouteAuthorization(t *testing.T) {
	domain := &runtimeDomainStub{
		app: &domains.App{
			Id:             uuid.New(),
			PublicKey:      "pk_test",
			AllowedOrigins: []string{"http://localhost:5173"},
		},
	}
	handler := newRuntimeTestServer(t, domain)

	tests := []struct {
		name   string
		key    string
		origin string
		want   int
	}{
		{name: "missing app key", origin: "http://localhost:5173", want: http.StatusUnauthorized},
		{name: "invalid app key", key: "pk_unknown", origin: "http://localhost:5173", want: http.StatusUnauthorized},
		{name: "forbidden origin", key: "pk_test", origin: "http://evil.example.com", want: http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := resolveRequest(t, handler, tc.key, tc.origin, "/dashboard")
			assert.Equal(t, tc.want, response.Code)
		})
	}
}
