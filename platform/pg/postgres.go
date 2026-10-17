package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ksendel-coder/onboarding_croquant/migrations"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func Connect(ctx context.Context, dsn string, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for attempt := 1; ; attempt++ {
		err = pool.Ping(ctx)
		if err == nil {
			logger.Info().Int("attempt", attempt).Msg("Database is up")
			return pool, nil
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			pool.Close()
			return nil, fmt.Errorf("database unreachable after %d attempts: %w", attempt, err)
		}
		logger.Debug().Err(err).Int("attempt", attempt).Msg("Database not ready, retrying")

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func Migrate(dsn *string, logger *zerolog.Logger) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer db.Close()

	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("build migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("build migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("read schema version: %w", err)
	}

	logger.Info().
		Uint("version", version).
		Bool("dirty", dirty).
		Msg("Schema is up to date")

	return nil
}
