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
	"github.com/stretchr/testify/require"
)

type restoreHTTPDomainStub struct {
	domains.TourDomainInterface
	listCalls, restoreCalls int
	tourID, sourceID        uuid.UUID
}

func (d *restoreHTTPDomainStub) ListVersions(_ context.Context, tourID uuid.UUID) ([]domains.TourVersionHistoryItem, error) {
	d.listCalls++
	d.tourID = tourID
	return []domains.TourVersionHistoryItem{{Id: uuid.New(), Version: 3, Status: domains.TourDraft}}, nil
}

func (d *restoreHTTPDomainStub) RestoreVersion(_ context.Context, tourID, sourceID uuid.UUID) (*domains.RestoreVersionResult, error) {
	d.restoreCalls++
	d.tourID = tourID
	d.sourceID = sourceID
	return &domains.RestoreVersionResult{TourId: tourID, RestoredFromVersionId: sourceID,
		DraftVersionId: uuid.New(), DraftVersion: 3, StepsCount: 2}, nil
}

func newRestoreHTTPServer(t *testing.T, allowed bool) (http.Handler, *restoreHTTPDomainStub) {
	t.Helper()
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logger := zerolog.Nop()
	auth := &analyticsAuthStub{user: &domains.User{Id: uuid.New()}, allowed: allowed}
	domain := &restoreHTTPDomainStub{}
	srv := NewServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &CreateStruct{
		Logger: &logger, AuthDomain: auth,
		RuntimeController: controllers.NewRuntimeController(nil),
		TourController:    controllers.NewTourController(domain),
		HintController:    controllers.NewHintController(nil),
		AuthController:    controllers.NewAuthController(nil),
		EventController:   controllers.NewEventController(nil),
	})
	return srv.Handler(), domain
}

func restoreHTTPRequest(h http.Handler, method, path, bearer, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestMVP11RestoreAndVersionRoutesProtected(t *testing.T) {
	tourID := uuid.NewString()
	sourceID := uuid.NewString()
	paths := []struct{ method, path, body string }{
		{http.MethodGet, "/v1/tours/" + tourID + "/versions", ""},
		{http.MethodPost, "/v1/tours/" + tourID + "/versions/" + sourceID + "/restore", `{"confirm_replace_draft":true}`},
	}
	for _, p := range paths {
		t.Run(p.method+" "+p.path, func(t *testing.T) {
			h, d := newRestoreHTTPServer(t, true)
			r := restoreHTTPRequest(h, p.method, p.path, "", p.body)
			require.Equal(t, http.StatusUnauthorized, r.Code)
			r = restoreHTTPRequest(h, p.method, p.path, "Bearer invalid", p.body)
			require.Equal(t, http.StatusUnauthorized, r.Code)
			require.Equal(t, 0, d.listCalls+d.restoreCalls)

			h, d = newRestoreHTTPServer(t, false)
			r = restoreHTTPRequest(h, p.method, p.path, "Bearer test-admin-token", p.body)
			require.Equal(t, http.StatusForbidden, r.Code)
			require.Equal(t, 0, d.listCalls+d.restoreCalls)
		})
	}
}

func TestMVP11RestoreConfirmationAndValidation(t *testing.T) {
	h, d := newRestoreHTTPServer(t, true)
	tourID := uuid.New()
	sourceID := uuid.New()
	path := "/v1/tours/" + tourID.String() + "/versions/" + sourceID.String() + "/restore"
	authorized := "Bearer test-admin-token"
	for _, body := range []string{`{}`, `{"confirm_replace_draft":false}`, `{"wrong_field":true}`} {
		rec := restoreHTTPRequest(h, http.MethodPost, path, authorized, body)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Zero(t, d.restoreCalls, "rejected request must not change data")
	}
	missingBody := restoreHTTPRequest(h, http.MethodPost, path, authorized, "")
	require.Equal(t, http.StatusBadRequest, missingBody.Code)
	require.Zero(t, d.restoreCalls)

	invalidTour := restoreHTTPRequest(h, http.MethodPost,
		"/v1/tours/not-uuid/versions/"+sourceID.String()+"/restore", authorized, `{"confirm_replace_draft":true}`)
	require.Equal(t, http.StatusBadRequest, invalidTour.Code)
	invalidVersion := restoreHTTPRequest(h, http.MethodPost,
		"/v1/tours/"+tourID.String()+"/versions/not-uuid/restore", authorized, `{"confirm_replace_draft":true}`)
	require.Equal(t, http.StatusBadRequest, invalidVersion.Code)
	require.Zero(t, d.restoreCalls)

	success := restoreHTTPRequest(h, http.MethodPost, path, authorized, `{"confirm_replace_draft":true}`)
	require.Equal(t, http.StatusOK, success.Code, success.Body.String())
	require.Equal(t, 1, d.restoreCalls)
	require.Equal(t, tourID, d.tourID)
	require.Equal(t, sourceID, d.sourceID)
	var result domains.RestoreVersionResult
	require.NoError(t, json.Unmarshal(success.Body.Bytes(), &result))
	require.Equal(t, 3, result.DraftVersion)
	require.Equal(t, 2, result.StepsCount)
}

func TestMVP11VersionListHTTP(t *testing.T) {
	h, d := newRestoreHTTPServer(t, true)
	tourID := uuid.New()
	path := "/v1/tours/" + tourID.String() + "/versions"
	rec := restoreHTTPRequest(h, http.MethodGet, path, "Bearer test-admin-token", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, 1, d.listCalls)
	require.Equal(t, tourID, d.tourID)
	var result []domains.TourVersionHistoryItem
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Len(t, result, 1)
	require.Equal(t, 3, result[0].Version)
}
