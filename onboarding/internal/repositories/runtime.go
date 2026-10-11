package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

type RuntimeRepository struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewRuntimeRepository(pool *pgxpool.Pool, logger *zerolog.Logger) *RuntimeRepository {
	return &RuntimeRepository{
		pool:   pool,
		logger: logger,
	}
}

func (rr *RuntimeRepository) Ping(ctx context.Context) error {
	return rr.pool.Ping(ctx)
}

func (rr *RuntimeRepository) CreateApp(ctx context.Context, a *domains.App) error {
	err := rr.pool.QueryRow(ctx, `
		INSERT INTO apps (name, public_key, allowed_origins)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		a.Name,
		a.PublicKey,
		a.AllowedOrigins,
	).Scan(&a.Id, &a.CreatedAt)

	return wrap(err, domains.ErrAppNotFound, a.PublicKey, errs.Create)
}

func (rr *RuntimeRepository) AppByKey(ctx context.Context, key string) (*domains.App, error) {
	a := new(domains.App)

	err := rr.pool.QueryRow(ctx, `
		SELECT id, name, public_key, allowed_origins, created_at, archived_at
		FROM apps
		WHERE public_key = $1 AND archived_at IS NULL`,
		key,
	).Scan(
		&a.Id,
		&a.Name,
		&a.PublicKey,
		&a.AllowedOrigins,
		&a.CreatedAt,
		&a.ArchivedAt,
	)

	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, key, errs.Retrieve)
	}

	return a, nil
}

func (rr *RuntimeRepository) UpdateApp(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	allowedOrigins []string,
) (*domains.App, error) {
	set := make([]string, 0, 2)
	args := make([]any, 0, 3)

	if name != nil {
		args = append(args, *name)
		set = append(set, fmt.Sprintf("name = $%d", len(args)))
	}

	if allowedOrigins != nil {
		args = append(args, allowedOrigins)
		set = append(set, fmt.Sprintf("allowed_origins = $%d", len(args)))
	}

	if len(set) == 0 {
		return rr.AppByKey(ctx, "")
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE apps SET %s
		WHERE id = $%d AND archived_at IS NULL
		RETURNING id, name, public_key, allowed_origins, created_at, archived_at`,
		strings.Join(set, ", "),
		len(args),
	)

	a := new(domains.App)

	err := rr.pool.QueryRow(ctx, query, args...).Scan(
		&a.Id,
		&a.Name,
		&a.PublicKey,
		&a.AllowedOrigins,
		&a.CreatedAt,
		&a.ArchivedAt,
	)

	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, id.String(), errs.Update)
	}

	return a, nil
}

func (rr *RuntimeRepository) ListApps(ctx context.Context) ([]*domains.App, error) {
	rows, err := rr.pool.Query(ctx, `
		SELECT id, name, public_key, allowed_origins, created_at, archived_at
		FROM apps
		WHERE archived_at IS NULL
		ORDER BY created_at`)

	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, "", errs.Retrieve)
	}

	defer rows.Close()

	apps := make([]*domains.App, 0)

	for rows.Next() {
		a := new(domains.App)

		if err := rows.Scan(
			&a.Id,
			&a.Name,
			&a.PublicKey,
			&a.AllowedOrigins,
			&a.CreatedAt,
			&a.ArchivedAt,
		); err != nil {
			return nil, wrap(err, domains.ErrAppNotFound, "", errs.Retrieve)
		}

		apps = append(apps, a)
	}

	return apps, rows.Err()
}
