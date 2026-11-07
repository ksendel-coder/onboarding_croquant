package server

import (
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
	"github.com/stretchr/testify/require"
)

type analyticsAuthStub struct {
	user    *domains.User
	allowed bool
}

func (a *analyticsAuthStub) Register(_ context.Context, _ string, _ string) (*domains.User, string, error) {
	return nil, "", nil
}
func (a *analyticsAuthStub) Login(_ context.Context, _ string, _ string) (*domains.User, string, error) {
	return nil, "", nil
}
func (a *analyticsAuthStub) Authenticate(_ context.Context, token string) (*domains.User, error) {
	if token != "test-admin-token" {
		return nil, domains.ErrUnauthorized
	}
	return a.user, nil
}
func (a *analyticsAuthStub) Logout(_ context.Context, _ string) error { return nil }
func (a *analyticsAuthStub) UserHasApp(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return a.allowed, nil
}
func (a *analyticsAuthStub) UserHasTour(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return a.allowed, nil
}

type analyticsDomainStub struct {
	called    int
	tourId    uuid.UUID
	versionId uuid.UUID
}

func (d *analyticsDomainStub) Record(_ context.Context, _ domains.TourEvent) (bool, error) {
	return false, nil
}
func (d *analyticsDomainStub) Analytics(_ context.Context, tourId, versionId uuid.UUID) (*domains.TourAnalytics, error) {
	d.called++
	d.tourId = tourId
	d.versionId = versionId
	return &domains.TourAnalytics{
		TourId:         tourId,
		TourVersionId:  versionId,
		Started:        10,
		Completed:      7,
		Dismissed:      2,
		CompletionRate: 70,
		Steps: []domains.StepAnalytics{{
			HintId: uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"),
			Step:   1, Shown: 10, Completed: 9, DropOff: 1,
		}},
	}, nil
}

func newAnalyticsHTTPServer(t *testing.T, allowed bool) (http.Handler, *analyticsDomainStub) {
	t.Helper()
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logger := zerolog.Nop()
	auth := &analyticsAuthStub{
		user:    &domains.User{Id: uuid.New()},
		allowed: allowed,
	}
	analytics := &analyticsDomainStub{}
	srv := NewServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &CreateStruct{
		Logger:            &logger,
		AuthDomain:        auth,
		RuntimeController: controllers.NewRuntimeController(nil),
		TourController:    controllers.NewTourController(nil),
		HintController:    controllers.NewHintController(nil),
		AuthController:    controllers.NewAuthController(nil),
		EventController:   controllers.NewEventController(analytics),
	})
	return srv.Handler(), analytics
}

func requestAnalytics(handler http.Handler, tourId, versionId, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/tours/"+tourId+"/analytics?version_id="+versionId,
		nil,
	)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestAnalyticsRouteProtected(t *testing.T) {
	tourId := uuid.NewString()
	versionId := uuid.NewString()

	t.Run("without bearer returns 401", func(t *testing.T) {
		handler, domain := newAnalyticsHTTPServer(t, true)
		rec := requestAnalytics(handler, tourId, versionId, "")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, 0, domain.called)
	})

	t.Run("X-App-Key is not admin authorization", func(t *testing.T) {
		handler, domain := newAnalyticsHTTPServer(t, true)
		req := httptest.NewRequest(http.MethodGet,
			"/v1/tours/"+tourId+"/analytics?version_id="+versionId, nil)
		req.Header.Set("X-App-Key", "pk_test")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, 0, domain.called)
	})

	t.Run("invalid bearer returns 401", func(t *testing.T) {
		handler, domain := newAnalyticsHTTPServer(t, true)
		rec := requestAnalytics(handler, tourId, versionId, "Bearer invalid")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, 0, domain.called)
	})

	t.Run("foreign tour returns 403", func(t *testing.T) {
		handler, domain := newAnalyticsHTTPServer(t, false)
		rec := requestAnalytics(handler, tourId, versionId, "Bearer test-admin-token")
		require.Equal(t, http.StatusForbidden, rec.Code)
		require.Equal(t, 0, domain.called)
	})
}

func TestAnalyticsRouteValidationAndJSON(t *testing.T) {
	tourId := uuid.New()
	versionId := uuid.New()
	handler, domain := newAnalyticsHTTPServer(t, true)

	invalidTour := requestAnalytics(handler, "not-uuid", versionId.String(), "Bearer test-admin-token")
	require.Equal(t, http.StatusBadRequest, invalidTour.Code)

	invalidVersion := requestAnalytics(handler, tourId.String(), "not-uuid", "Bearer test-admin-token")
	require.Equal(t, http.StatusBadRequest, invalidVersion.Code)

	missingVersion := requestAnalytics(handler, tourId.String(), "", "Bearer test-admin-token")
	require.Equal(t, http.StatusBadRequest, missingVersion.Code)
	require.Equal(t, 0, domain.called)

	success := requestAnalytics(handler, tourId.String(), versionId.String(), "Bearer test-admin-token")
	require.Equal(t, http.StatusOK, success.Code)
	require.Equal(t, 1, domain.called)
	require.Equal(t, tourId, domain.tourId)
	require.Equal(t, versionId, domain.versionId)

	var body domains.TourAnalytics
	require.NoError(t, json.Unmarshal(success.Body.Bytes(), &body))
	require.Equal(t, 10, body.Started)
	require.Equal(t, 7, body.Completed)
	require.Equal(t, 2, body.Dismissed)
	require.Equal(t, 70.0, body.CompletionRate)
	require.Len(t, body.Steps, 1)
	require.Equal(t, 1, body.Steps[0].DropOff)
}
