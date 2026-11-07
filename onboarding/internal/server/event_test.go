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

type eventDomainStub struct {
	created bool
	last    domains.TourEvent
}

func (s *eventDomainStub) Record(
	_ context.Context,
	event domains.TourEvent,
) (bool, error) {
	s.last = event
	return s.created, nil
}

func (s *eventDomainStub) Analytics(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) (*domains.TourAnalytics, error) {
	return nil, nil
}

func newEventTestServer(
	t *testing.T,
	runtimeDomain *runtimeDomainStub,
	eventDomain *eventDomainStub,
) http.Handler {
	t.Helper()

	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logger := zerolog.Nop()

	srv := NewServer(
		&config.ServerConfig{Address: "127.0.0.1:0"},
		&CreateStruct{
			Logger:            &logger,
			RuntimeDomain:     runtimeDomain,
			RuntimeController: controllers.NewRuntimeController(runtimeDomain),
			TourController:    controllers.NewTourController(nil),
			HintController:    controllers.NewHintController(nil),
			AuthController:    controllers.NewAuthController(nil),
			EventController:   controllers.NewEventController(eventDomain),
		},
	)

	return srv.Handler()
}

func eventRequest(
	t *testing.T,
	handler http.Handler,
	key string,
	origin string,
	body map[string]any,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewReader(payload))
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

func TestEventRouteRecordsEvent(t *testing.T) {
	appId := uuid.New()
	hintId := uuid.New()
	runtimeDomain := &runtimeDomainStub{app: &domains.App{
		Id:             appId,
		PublicKey:      "pk_test",
		AllowedOrigins: []string{"http://localhost:5173"},
	}}
	eventDomain := &eventDomainStub{created: true}
	handler := newEventTestServer(t, runtimeDomain, eventDomain)
	body := map[string]any{
		"event_id":        uuid.New(),
		"attempt_id":      uuid.New(),
		"subject_id":      "visitor-123",
		"tour_id":         uuid.New(),
		"tour_version_id": uuid.New(),
		"hint_id":         hintId,
		"type":            domains.EventHintShown,
	}

	response := eventRequest(
		t,
		handler,
		"pk_test",
		"http://localhost:5173",
		body,
	)

	assert.Equal(t, http.StatusAccepted, response.Code)
	assert.Equal(t, appId, eventDomain.last.AppId)
	assert.Equal(t, "visitor-123", eventDomain.last.SubjectId)
	assert.Equal(t, hintId, *eventDomain.last.HintId)
}

func TestEventRouteReturnsOKForDuplicate(t *testing.T) {
	runtimeDomain := &runtimeDomainStub{app: &domains.App{
		Id:        uuid.New(),
		PublicKey: "pk_test",
	}}
	handler := newEventTestServer(t, runtimeDomain, &eventDomainStub{})

	response := eventRequest(t, handler, "pk_test", "", map[string]any{
		"event_id":        uuid.New(),
		"attempt_id":      uuid.New(),
		"subject_id":      "visitor-123",
		"tour_id":         uuid.New(),
		"tour_version_id": uuid.New(),
		"type":            domains.EventTourStarted,
	})

	assert.Equal(t, http.StatusOK, response.Code)
}

func TestEventRouteAuthorizationAndValidation(t *testing.T) {
	runtimeDomain := &runtimeDomainStub{app: &domains.App{
		Id:             uuid.New(),
		PublicKey:      "pk_test",
		AllowedOrigins: []string{"http://localhost:5173"},
	}}
	handler := newEventTestServer(t, runtimeDomain, &eventDomainStub{})
	validBody := map[string]any{
		"event_id":        uuid.New(),
		"attempt_id":      uuid.New(),
		"subject_id":      "visitor-123",
		"tour_id":         uuid.New(),
		"tour_version_id": uuid.New(),
		"type":            domains.EventTourStarted,
	}

	tests := []struct {
		name   string
		key    string
		origin string
		body   map[string]any
		want   int
	}{
		{name: "missing app key", origin: "http://localhost:5173", body: validBody, want: http.StatusUnauthorized},
		{name: "invalid app key", key: "pk_unknown", origin: "http://localhost:5173", body: validBody, want: http.StatusUnauthorized},
		{name: "forbidden origin", key: "pk_test", origin: "http://evil.example.com", body: validBody, want: http.StatusForbidden},
		{name: "invalid body", key: "pk_test", body: map[string]any{"type": domains.EventTourStarted}, want: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := eventRequest(t, handler, tc.key, tc.origin, tc.body)
			assert.Equal(t, tc.want, response.Code)
		})
	}
}

func TestAnalyticsIsNotExposedThroughPublicAppKeyRoutes(t *testing.T) {
	runtimeDomain := &runtimeDomainStub{app: &domains.App{
		Id:        uuid.New(),
		PublicKey: "pk_test",
	}}
	handler := newEventTestServer(t, runtimeDomain, &eventDomainStub{})
	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/tours/"+uuid.NewString()+"/analytics?version_id="+uuid.NewString(),
		nil,
	)
	req.Header.Set("X-App-Key", "pk_test")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
