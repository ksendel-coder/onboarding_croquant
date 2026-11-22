package domains

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type authUnitRepo struct {
	AuthRepositoryInterface
	user               *User
	session            *AdminSession
	createUserCalls    int
	createSessionCalls int
	lookupEmail        string
	lookupHash         string
	deletedHash        string
	lookupErr          error
	deleteErr          error
	createUserErr      error
	createSessionErr   error
	permissionApp      bool
	permissionTour     bool
	permissionErr      error
}

func (s *authUnitRepo) CreateUser(_ context.Context, u *User) error {
	s.createUserCalls++
	if s.createUserErr != nil {
		return s.createUserErr
	}
	u.Id = uuid.New()
	s.user = u
	return nil
}
func (s *authUnitRepo) CreateSession(_ context.Context, session *AdminSession) error {
	s.createSessionCalls++
	if s.createSessionErr != nil {
		return s.createSessionErr
	}
	s.session = session
	return nil
}
func (s *authUnitRepo) UserByEmail(_ context.Context, email string) (*User, error) {
	s.lookupEmail = email
	return s.user, s.lookupErr
}
func (s *authUnitRepo) SessionByTokenHash(_ context.Context, hash string) (*AdminSession, *User, error) {
	s.lookupHash = hash
	return s.session, s.user, s.lookupErr
}
func (s *authUnitRepo) DeleteSessionByTokenHash(_ context.Context, hash string) error {
	s.deletedHash = hash
	return s.deleteErr
}
func (s *authUnitRepo) UserHasApp(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return s.permissionApp, s.permissionErr
}
func (s *authUnitRepo) UserHasTour(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return s.permissionTour, s.permissionErr
}

func requireAuthHTTPCode(t *testing.T, err error, code int) {
	t.Helper()
	var logic *LogicError
	require.ErrorAs(t, err, &logic)
	require.Equal(t, code, logic.Code)
}

func TestAuthUnitRegisterValidationAndPersistence(t *testing.T) {
	tests := []struct{ name, first, last, email, password string }{
		{"blank first", " ", "Иванова", "a@example.com", "password123"},
		{"blank last", "Анна", "  ", "a@example.com", "password123"},
		{"invalid email", "Анна", "Иванова", "invalid", "password123"},
		{"short password", "Анна", "Иванова", "a@example.com", "short"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &authUnitRepo{}
			user, token, err := NewAuthDomain(repo).Register(context.Background(), tc.first, tc.last, tc.email, tc.password)
			requireAuthHTTPCode(t, err, http.StatusBadRequest)
			require.Nil(t, user)
			require.Empty(t, token)
			require.Zero(t, repo.createUserCalls)
			require.Zero(t, repo.createSessionCalls)
		})
	}
	repo := &authUnitRepo{}
	before := time.Now().UTC()
	user, token, err := NewAuthDomain(repo).Register(context.Background(), " Анна ", " Иванова ", " ANNA@EXAMPLE.COM ", "password123")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, "Анна", user.FirstName)
	require.Equal(t, "Иванова", user.LastName)
	require.Equal(t, "anna@example.com", user.Email)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")))
	require.Equal(t, 1, repo.createUserCalls)
	require.Equal(t, 1, repo.createSessionCalls)
	require.Equal(t, user.Id, repo.session.UserId)
	require.Equal(t, hashToken(token), repo.session.TokenHash)
	require.WithinDuration(t, before.Add(adminSessionTTL), repo.session.ExpiresAt, 5*time.Second)
}

func TestAuthUnitRegisterStopsAfterRepositoryError(t *testing.T) {
	sentinel := errors.New("create user failed")
	repo := &authUnitRepo{createUserErr: sentinel}
	user, token, err := NewAuthDomain(repo).Register(context.Background(), "Анна", "Иванова", "a@example.com", "password123")
	require.ErrorIs(t, err, sentinel)
	require.Nil(t, user)
	require.Empty(t, token)
	require.Equal(t, 1, repo.createUserCalls)
	require.Zero(t, repo.createSessionCalls)
}

func TestAuthUnitLoginRejectsIncorrectPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	require.NoError(t, err)
	repo := &authUnitRepo{user: &User{Id: uuid.New(), PasswordHash: string(passwordHash)}}
	user, token, err := NewAuthDomain(repo).Login(context.Background(), " SOME@EXAMPLE.COM ", "incorrectpassword")
	requireAuthHTTPCode(t, err, http.StatusUnauthorized)
	require.Nil(t, user)
	require.Empty(t, token)
	require.Equal(t, "some@example.com", repo.lookupEmail)
	require.Zero(t, repo.createSessionCalls)
}

func TestAuthUnitAuthenticateAndLogoutUseHashedToken(t *testing.T) {
	token := "raw-admin-session-token"
	repo := &authUnitRepo{user: &User{Id: uuid.New()}, session: &AdminSession{}}
	auth := NewAuthDomain(repo)
	actual, err := auth.Authenticate(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, repo.user.Id, actual.Id)
	require.Equal(t, hashToken(token), repo.lookupHash)
	require.NotEqual(t, token, repo.lookupHash)
	require.NoError(t, auth.Logout(context.Background(), token))
	require.Equal(t, hashToken(token), repo.deletedHash)
	for _, action := range []struct {
		name string
		call func() error
	}{
		{"authenticate empty", func() error { _, err := auth.Authenticate(context.Background(), " "); return err }},
		{"logout empty", func() error { return auth.Logout(context.Background(), " ") }},
	} {
		t.Run(action.name, func(t *testing.T) { requireAuthHTTPCode(t, action.call(), http.StatusUnauthorized) })
	}
}

func TestAuthUnitPermissionsPassThrough(t *testing.T) {
	repo := &authUnitRepo{permissionApp: true, permissionTour: false}
	auth := NewAuthDomain(repo)
	allowed, err := auth.UserHasApp(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = auth.UserHasTour(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.False(t, allowed)
	sentinel := errors.New("db unavailable")
	repo.permissionErr = sentinel
	_, err = auth.UserHasApp(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, sentinel)
	_, err = auth.UserHasTour(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, sentinel)
}
