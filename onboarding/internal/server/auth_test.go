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

// Embed the actual AuthDomainInterface; override only the methods exercised here.
type authNamesHTTPStub struct {
	domains.AuthDomainInterface
	user                         *domains.User
	first, last, email, password string
	registerCalls                int
	loginCalls                   int
	logoutToken                  string
}

func (s *authNamesHTTPStub) Register(_ context.Context, first, last, email, password string) (*domains.User, string, error) {
	s.registerCalls++
	s.first, s.last, s.email, s.password = first, last, email, password
	return s.user, "registered-token", nil
}

func (s *authNamesHTTPStub) Login(_ context.Context, email, password string) (*domains.User, string, error) {
	s.loginCalls++
	s.email, s.password = email, password
	return s.user, "logged-in-token", nil
}

func (s *authNamesHTTPStub) Authenticate(_ context.Context, token string) (*domains.User, error) {
	if token != "valid-token" {
		return nil, domains.ErrUnauthorized
	}
	return s.user, nil
}

func (s *authNamesHTTPStub) Logout(_ context.Context, token string) error {
	s.logoutToken = token
	return nil
}

func newAuthNamesTestServer(t *testing.T) (http.Handler, *authNamesHTTPStub) {
	t.Helper()
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logger := zerolog.Nop()
	auth := &authNamesHTTPStub{user: &domains.User{
		Id: uuid.New(), FirstName: "Анна", LastName: "Иванова", Email: "anna@example.com",
	}}
	server := NewServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &CreateStruct{
		Logger:            &logger,
		AuthDomain:        auth,
		AuthController:    controllers.NewAuthController(auth),
		RuntimeController: controllers.NewRuntimeController(nil),
		TourController:    controllers.NewTourController(nil),
		HintController:    controllers.NewHintController(nil),
		EventController:   controllers.NewEventController(nil),
	})
	return server.Handler(), auth
}

func authNamesRequest(handler http.Handler, method, path, body, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestAuthNamesRegisterHTTP(t *testing.T) {
	handler, auth := newAuthNamesTestServer(t)
	rec := authNamesRequest(handler, http.MethodPost, "/v1/auth/register", `{"first_name":"Анна","last_name":"Иванова","email":"anna@example.com","password":"Password123!"}`, "")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.Equal(t, 1, auth.registerCalls)
	require.Equal(t, "Анна", auth.first)
	require.Equal(t, "Иванова", auth.last)
	require.Equal(t, "anna@example.com", auth.email)
	require.Equal(t, "Password123!", auth.password)
	var payload struct {
		Token string `json:"token"`
		User  struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, "registered-token", payload.Token)
	require.Equal(t, "Анна", payload.User.FirstName)
	require.Equal(t, "Иванова", payload.User.LastName)
	require.Equal(t, "anna@example.com", payload.User.Email)
	require.NotContains(t, rec.Body.String(), "password_hash")
}

func TestAuthNamesRegisterRejectsMissingNameHTTP(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"missing first_name", `{"last_name":"Иванова","email":"anna@example.com","password":"Password123!"}`},
		{"missing last_name", `{"first_name":"Анна","email":"anna@example.com","password":"Password123!"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler, auth := newAuthNamesTestServer(t)
			rec := authNamesRequest(handler, http.MethodPost, "/v1/auth/register", tc.body, "")
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			require.Equal(t, 0, auth.registerCalls)
		})
	}
}

func TestAuthNamesLoginHTTP(t *testing.T) {
	handler, auth := newAuthNamesTestServer(t)
	rec := authNamesRequest(handler, http.MethodPost, "/v1/auth/login", `{"email":"anna@example.com","password":"Password123!"}`, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, 1, auth.loginCalls)
	var payload struct {
		User struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, "Анна", payload.User.FirstName)
	require.Equal(t, "Иванова", payload.User.LastName)
}

func TestAuthNamesMeRequiresSessionAndReturnsNames(t *testing.T) {
	handler, _ := newAuthNamesTestServer(t)
	unauthorized := authNamesRequest(handler, http.MethodGet, "/v1/auth/me", "", "")
	require.Equal(t, http.StatusUnauthorized, unauthorized.Code, unauthorized.Body.String())
	invalid := authNamesRequest(handler, http.MethodGet, "/v1/auth/me", "", "invalid-token")
	require.Equal(t, http.StatusUnauthorized, invalid.Code, invalid.Body.String())
	rec := authNamesRequest(handler, http.MethodGet, "/v1/auth/me", "", "valid-token")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var user struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &user))
	require.Equal(t, "Анна", user.FirstName)
	require.Equal(t, "Иванова", user.LastName)
}

func TestAuthNamesLogoutHTTP(t *testing.T) {
	handler, auth := newAuthNamesTestServer(t)
	rec := authNamesRequest(handler, http.MethodPost, "/v1/auth/logout", "", "valid-token")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	require.Equal(t, "valid-token", auth.logoutToken)
}
