package server

import (
	"bytes"
	"context"
	"errors"
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

type settingsAuthStub struct {
	domains.AuthDomainInterface
	user        *domains.User
	allowedTour uuid.UUID
}

func (s *settingsAuthStub) Authenticate(_ context.Context, token string) (*domains.User, error) {
	if token != "session-ok" {
		return nil, errors.New("invalid session")
	}
	return s.user, nil
}

func (s *settingsAuthStub) UserHasTour(_ context.Context, _ uuid.UUID, tourId uuid.UUID) (bool, error) {
	return tourId == s.allowedTour, nil
}

type settingsTourDomainStub struct {
	domains.TourDomainInterface
	called bool
	patch  domains.TourSettingsPatch
}

func (s *settingsTourDomainStub) UpdateSettings(_ context.Context, _ uuid.UUID, patch domains.TourSettingsPatch) error {
	s.called = true
	s.patch = patch
	return nil
}

func newSettingsTestServer(t *testing.T, tourId uuid.UUID) (http.Handler, *settingsTourDomainStub) {
	t.Helper()
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logger := zerolog.Nop()
	tourDomain := &settingsTourDomainStub{}
	authDomain := &settingsAuthStub{
		user:        &domains.User{Id: uuid.New()},
		allowedTour: tourId,
	}

	srv := NewServer(
		&config.ServerConfig{Address: "127.0.0.1:0"},
		&CreateStruct{
			Logger:            &logger,
			RuntimeController: controllers.NewRuntimeController(nil),
			TourController:    controllers.NewTourController(tourDomain),
			HintController:    controllers.NewHintController(nil),
			AuthDomain:        authDomain,
			AuthController:    controllers.NewAuthController(nil),
			EventController:   controllers.NewEventController(nil),
		},
	)

	return srv.Handler(), tourDomain
}

func patchSettings(handler http.Handler, tourId uuid.UUID, token string) *httptest.ResponseRecorder {
	body := []byte(`{
		"target_path":"/dashboard",
		"trigger_type":"delay",
		"trigger_config":{"delay_ms":1500,"scroll_depth":60,"element_selector":"#welcome"},
		"audience_show_once":true,
		"audience_max_shows":2,
		"audience_only_new":false
	}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/tours/"+tourId.String()+"/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestTourSettingsRouteRequiresAdminAndTourAccess(t *testing.T) {
	ownedTour := uuid.New()
	handler, domain := newSettingsTestServer(t, ownedTour)

	withoutSession := patchSettings(handler, ownedTour, "")
	assert.Equal(t, http.StatusUnauthorized, withoutSession.Code)

	foreignTour := patchSettings(handler, uuid.New(), "session-ok")
	assert.Equal(t, http.StatusForbidden, foreignTour.Code)
	assert.False(t, domain.called)

	allowed := patchSettings(handler, ownedTour, "session-ok")
	require.Equal(t, http.StatusNoContent, allowed.Code)
	require.True(t, domain.called)
	require.NotNil(t, domain.patch.TriggerConfig)
	require.NotNil(t, domain.patch.TriggerConfig.DelayMs)
	require.NotNil(t, domain.patch.TriggerConfig.ScrollDepth)
	require.NotNil(t, domain.patch.TriggerConfig.ElementSelector)
	assert.Equal(t, 1500, *domain.patch.TriggerConfig.DelayMs)
	assert.Equal(t, 60, *domain.patch.TriggerConfig.ScrollDepth)
	assert.Equal(t, "#welcome", *domain.patch.TriggerConfig.ElementSelector)
}
