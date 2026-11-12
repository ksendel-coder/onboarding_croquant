package domains

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Uses the existing AuthRepositoryInterface instead of inventing a new interface.
type authNamesRepositoryStub struct {
	AuthRepositoryInterface
	createdUser *User
	session     *AdminSession
	userByEmail *User
}

func (s *authNamesRepositoryStub) CreateUser(_ context.Context, user *User) error {
	user.Id = uuid.New()
	user.CreatedAt = time.Now().UTC()
	copy := *user
	s.createdUser = &copy
	return nil
}

func (s *authNamesRepositoryStub) CreateSession(_ context.Context, session *AdminSession) error {
	session.Id = uuid.New()
	copy := *session
	s.session = &copy
	return nil
}

func (s *authNamesRepositoryStub) UserByEmail(_ context.Context, _ string) (*User, error) {
	return s.userByEmail, nil
}

func TestAuthRegisterSavesNamesAndCreatesSession(t *testing.T) {
	repo := &authNamesRepositoryStub{}
	auth := NewAuthDomain(repo)

	user, token, err := auth.Register(context.Background(), "  Анна  ", "  Иванова  ", "  ANNA@EXAMPLE.COM  ", "Password123!")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, user)
	require.NotNil(t, repo.createdUser)
	require.Equal(t, "Анна", user.FirstName)
	require.Equal(t, "Иванова", user.LastName)
	require.Equal(t, "anna@example.com", user.Email)
	require.Equal(t, user.FirstName, repo.createdUser.FirstName)
	require.Equal(t, user.LastName, repo.createdUser.LastName)
	require.NotEqual(t, "Password123!", repo.createdUser.PasswordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.createdUser.PasswordHash), []byte("Password123!")))
	require.NotNil(t, repo.session)
	require.Equal(t, user.Id, repo.session.UserId)
	require.Equal(t, hashToken(token), repo.session.TokenHash)
}

func TestAuthRegisterRejectsEmptyProfileNames(t *testing.T) {
	cases := []struct{ name, first, last string }{
		{"missing first name", "  ", "Иванова"},
		{"missing last name", "Анна", "  "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &authNamesRepositoryStub{}
			user, token, err := NewAuthDomain(repo).Register(context.Background(), tc.first, tc.last, "anna@example.com", "Password123!")
			require.Error(t, err)
			require.Nil(t, user)
			require.Empty(t, token)
			require.Nil(t, repo.createdUser, "invalid names must not create a user")
			require.Nil(t, repo.session, "invalid names must not create a session")
		})
	}
}

func TestAuthLoginPreservesProfileNames(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.MinCost)
	require.NoError(t, err)
	stored := &User{
		Id: uuid.New(), FirstName: "Анна", LastName: "Иванова",
		Email: "anna@example.com", PasswordHash: string(hash),
	}
	repo := &authNamesRepositoryStub{userByEmail: stored}
	user, token, err := NewAuthDomain(repo).Login(context.Background(), "ANNA@EXAMPLE.COM", "Password123!")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, "Анна", user.FirstName)
	require.Equal(t, "Иванова", user.LastName)
	require.Equal(t, stored.Id, repo.session.UserId)
}
