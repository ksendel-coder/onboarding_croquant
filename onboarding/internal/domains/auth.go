package domains

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
	"golang.org/x/crypto/bcrypt"
)

const (
	adminSessionTTL = 24 * time.Hour
	minPasswordLen  = 8
)

type AuthDomain struct {
	auth AuthRepositoryInterface
}

type AuthDomainInterface interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (*User, string, error)

	Login(
		ctx context.Context,
		email string,
		password string,
	) (*User, string, error)

	Authenticate(
		ctx context.Context,
		token string,
	) (*User, error)

	Logout(
		ctx context.Context,
		token string,
	) error

	UserHasApp(
		ctx context.Context,
		userId uuid.UUID,
		appId uuid.UUID,
	) (bool, error)

	UserHasTour(
		ctx context.Context,
		userId uuid.UUID,
		tourId uuid.UUID,
	) (bool, error)
}

type AuthRepositoryInterface interface {
	CreateUser(
		ctx context.Context,
		user *User,
	) error

	UserByEmail(
		ctx context.Context,
		email string,
	) (*User, error)

	UserById(
		ctx context.Context,
		userId uuid.UUID,
	) (*User, error)

	CreateSession(
		ctx context.Context,
		session *AdminSession,
	) error

	SessionByTokenHash(
		ctx context.Context,
		tokenHash string,
	) (*AdminSession, *User, error)

	DeleteSessionByTokenHash(
		ctx context.Context,
		tokenHash string,
	) error

	UserHasApp(
		ctx context.Context,
		userId uuid.UUID,
		appId uuid.UUID,
	) (bool, error)

	UserHasTour(
		ctx context.Context,
		userId uuid.UUID,
		tourId uuid.UUID,
	) (bool, error)
}

func NewAuthDomain(
	auth AuthRepositoryInterface,
) *AuthDomain {
	return &AuthDomain{
		auth: auth,
	}
}

func (ad *AuthDomain) Register(
	ctx context.Context,
	email string,
	password string,
) (*User, string, error) {
	normalizedEmail, err := validateCredentials(
		email,
		password,
	)
	if err != nil {
		return nil, "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, "", logicErr(
			err,
			"password hashing",
			http.StatusInternalServerError,
		)
	}

	user := &User{
		Email:        normalizedEmail,
		PasswordHash: string(passwordHash),
	}

	if err := ad.auth.CreateUser(
		ctx,
		user,
	); err != nil {
		return nil, "", err
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return nil, "", logicErr(
			err,
			"session token generation",
			http.StatusInternalServerError,
		)
	}

	session := &AdminSession{
		UserId:    user.Id,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(adminSessionTTL),
	}

	if err := ad.auth.CreateSession(
		ctx,
		session,
	); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (ad *AuthDomain) Login(
	ctx context.Context,
	email string,
	password string,
) (*User, string, error) {
	normalizedEmail := strings.ToLower(
		strings.TrimSpace(email),
	)

	user, err := ad.auth.UserByEmail(
		ctx,
		normalizedEmail,
	)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, "", logicErr(
				ErrInvalidCredentials,
				"login",
				http.StatusUnauthorized,
			)
		}

		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, "", logicErr(
			ErrInvalidCredentials,
			"login",
			http.StatusUnauthorized,
		)
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return nil, "", logicErr(
			err,
			"session token generation",
			http.StatusInternalServerError,
		)
	}

	session := &AdminSession{
		UserId:    user.Id,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(adminSessionTTL),
	}

	if err := ad.auth.CreateSession(
		ctx,
		session,
	); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (ad *AuthDomain) Authenticate(
	ctx context.Context,
	token string,
) (*User, error) {
	if strings.TrimSpace(token) == "" {
		return nil, logicErr(
			ErrUnauthorized,
			"authentication",
			http.StatusUnauthorized,
		)
	}

	_, user, err := ad.auth.SessionByTokenHash(
		ctx,
		hashToken(token),
	)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, logicErr(
				ErrUnauthorized,
				"authentication",
				http.StatusUnauthorized,
			)
		}

		return nil, err
	}

	return user, nil
}

func (ad *AuthDomain) Logout(
	ctx context.Context,
	token string,
) error {
	if strings.TrimSpace(token) == "" {
		return logicErr(
			ErrUnauthorized,
			"logout",
			http.StatusUnauthorized,
		)
	}

	err := ad.auth.DeleteSessionByTokenHash(
		ctx,
		hashToken(token),
	)
	if err != nil && errs.IsNotFound(err) {
		return logicErr(
			ErrUnauthorized,
			"logout",
			http.StatusUnauthorized,
		)
	}

	return err
}

func (ad *AuthDomain) UserHasApp(
	ctx context.Context,
	userId uuid.UUID,
	appId uuid.UUID,
) (bool, error) {
	return ad.auth.UserHasApp(
		ctx,
		userId,
		appId,
	)
}

func (ad *AuthDomain) UserHasTour(
	ctx context.Context,
	userId uuid.UUID,
	tourId uuid.UUID,
) (bool, error) {
	return ad.auth.UserHasTour(
		ctx,
		userId,
		tourId,
	)
}

func validateCredentials(
	email string,
	password string,
) (string, error) {
	normalizedEmail := strings.ToLower(
		strings.TrimSpace(email),
	)

	if normalizedEmail == "" {
		return "", logicErr(
			ErrEmailRequired,
			"registration",
			http.StatusBadRequest,
		)
	}

	parsed, err := mail.ParseAddress(
		normalizedEmail,
	)
	if err != nil ||
		parsed.Address != normalizedEmail {
		return "", logicErr(
			ErrInvalidEmail,
			"registration",
			http.StatusBadRequest,
		)
	}

	if len(password) < minPasswordLen {
		return "", logicErr(
			ErrPasswordTooShort,
			"registration",
			http.StatusBadRequest,
		)
	}

	return normalizedEmail, nil
}

func newSessionToken() (
	string,
	string,
	error,
) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)

	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256(
		[]byte(token),
	)

	return hex.EncodeToString(
		sum[:],
	)
}
