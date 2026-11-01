package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
	"github.com/rs/zerolog"
)

type AuthRepository struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewAuthRepository(
	pool *pgxpool.Pool,
	logger *zerolog.Logger,
) *AuthRepository {
	return &AuthRepository{
		pool:   pool,
		logger: logger,
	}
}

func (ar *AuthRepository) CreateUser(
	ctx context.Context,
	user *domains.User,
) error {
	err := ar.pool.QueryRow(ctx, `
		INSERT INTO users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING id, created_at`,
		user.Email,
		user.PasswordHash,
	).Scan(
		&user.Id,
		&user.CreatedAt,
	)

	return wrap(
		err,
		domains.ErrUserNotFound,
		user.Email,
		errs.Create,
	)
}

func (ar *AuthRepository) UserByEmail(
	ctx context.Context,
	email string,
) (*domains.User, error) {
	user := new(domains.User)

	err := ar.pool.QueryRow(ctx, `
		SELECT
			id,
			email,
			password_hash,
			created_at
		FROM users
		WHERE email = $1`,
		email,
	).Scan(
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, wrap(
			err,
			domains.ErrUserNotFound,
			email,
			errs.Retrieve,
		)
	}

	return user, nil
}

func (ar *AuthRepository) UserById(
	ctx context.Context,
	userId uuid.UUID,
) (*domains.User, error) {
	user := new(domains.User)

	err := ar.pool.QueryRow(ctx, `
		SELECT
			id,
			email,
			password_hash,
			created_at
		FROM users
		WHERE id = $1`,
		userId,
	).Scan(
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, wrap(
			err,
			domains.ErrUserNotFound,
			userId.String(),
			errs.Retrieve,
		)
	}

	return user, nil
}

func (ar *AuthRepository) CreateSession(
	ctx context.Context,
	session *domains.AdminSession,
) error {
	err := ar.pool.QueryRow(ctx, `
		INSERT INTO admin_sessions (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		session.UserId,
		session.TokenHash,
		session.ExpiresAt,
	).Scan(
		&session.Id,
		&session.CreatedAt,
	)

	return wrap(
		err,
		domains.ErrSessionNotFound,
		session.UserId.String(),
		errs.Create,
	)
}

func (ar *AuthRepository) SessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) (
	*domains.AdminSession,
	*domains.User,
	error,
) {
	session := new(domains.AdminSession)
	user := new(domains.User)

	err := ar.pool.QueryRow(ctx, `
		SELECT
			s.id,
			s.user_id,
			s.token_hash,
			s.expires_at,
			s.created_at,
			u.id,
			u.email,
			u.password_hash,
			u.created_at
		FROM admin_sessions s
		JOIN users u
		  ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.expires_at > now()`,
		tokenHash,
	).Scan(
		&session.Id,
		&session.UserId,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, nil, wrap(
			err,
			domains.ErrSessionNotFound,
			tokenHash,
			errs.Retrieve,
		)
	}

	return session, user, nil
}

func (ar *AuthRepository) DeleteSessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) error {
	tag, err := ar.pool.Exec(ctx, `
		DELETE FROM admin_sessions
		WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return wrap(
			err,
			domains.ErrSessionNotFound,
			tokenHash,
			errs.Delete,
		)
	}

	if tag.RowsAffected() == 0 {
		return errs.NotFound(
			domains.ErrSessionNotFound,
			tokenHash,
		)
	}

	return nil
}

func (ar *AuthRepository) UserHasApp(
	ctx context.Context,
	userId uuid.UUID,
	appId uuid.UUID,
) (bool, error) {
	var allowed bool

	err := ar.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_apps
			WHERE user_id = $1
			  AND app_id = $2
		)`,
		userId,
		appId,
	).Scan(&allowed)

	if err != nil {
		return false, wrap(
			err,
			domains.ErrAppNotFound,
			appId.String(),
			errs.Retrieve,
		)
	}

	return allowed, nil
}

func (ar *AuthRepository) UserHasTour(
	ctx context.Context,
	userId uuid.UUID,
	tourId uuid.UUID,
) (bool, error) {
	var allowed bool

	err := ar.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM tours t
			JOIN user_apps ua
			  ON ua.app_id = t.app_id
			WHERE ua.user_id = $1
			  AND t.id = $2
			  AND t.archived_at IS NULL
		)`,
		userId,
		tourId,
	).Scan(&allowed)

	if err != nil {
		return false, wrap(
			err,
			domains.ErrTourNotFound,
			tourId.String(),
			errs.Retrieve,
		)
	}

	return allowed, nil
}
