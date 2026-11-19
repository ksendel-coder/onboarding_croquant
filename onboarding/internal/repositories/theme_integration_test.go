package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestAppThemeIntegration(t *testing.T) {
	dsn := os.Getenv("MVP08_TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("REQUIRE_THEME_DB") == "1" {
			t.Fatal("MVP08_TEST_DATABASE_URL is required")
		}
		t.Skip("MVP08_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))
	logger := zerolog.Nop()
	repo := NewRuntimeRepository(pool, &logger)
	domain := domains.NewRuntimeDomain(nil, nil, repo)
	owner, other := uuid.New(), uuid.New()
	a, b := &domains.App{Name: "Theme A", AllowedOrigins: []string{"http://localhost:5173"}}, &domains.App{Name: "Theme B", AllowedOrigins: []string{"http://localhost:5173"}}
	t.Cleanup(func() {
		if _, err := pool.Exec(ctx, `DELETE FROM apps WHERE id=$1 OR id=$2`, a.Id, b.Id); err != nil {
			t.Error(err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM users WHERE id=$1 OR id=$2`, owner, other); err != nil {
			t.Error(err)
		}
	})
	_, err = pool.Exec(ctx, `INSERT INTO users(id,email,password_hash) VALUES ($1,$2,'test'),($3,$4,'test')`, owner, owner.String()+"@example.test", other, other.String()+"@example.test")
	require.NoError(t, err)
	_, err = domain.CreateApp(ctx, owner, a)
	require.NoError(t, err)
	_, err = domain.CreateApp(ctx, other, b)
	require.NoError(t, err)
	defaults := domains.Theme{PrimaryColor: "#696CF1", BackgroundColor: "#FFFFFF", TextColor: "#181818"}
	require.Equal(t, defaults, a.Theme)
	primary, bg, fg := "#aabbcc", "#101010", "#eeeeee"
	updated, err := domain.UpdateApp(ctx, owner, a.Id, nil, nil, domains.ThemePatch{PrimaryColor: &primary, BackgroundColor: &bg, TextColor: &fg})
	require.NoError(t, err)
	want := domains.Theme{PrimaryColor: "#AABBCC", BackgroundColor: "#101010", TextColor: "#EEEEEE"}
	require.Equal(t, want, updated.Theme)
	list, err := domain.ListApps(ctx, owner)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, want, list[0].Theme)
	public, err := domain.AppByKey(ctx, a.PublicKey)
	require.NoError(t, err)
	require.Equal(t, want, public.Theme)
	_, err = domain.UpdateApp(ctx, other, a.Id, nil, nil, domains.ThemePatch{PrimaryColor: &fg})
	require.Error(t, err)
	invalid := "red"
	_, err = domain.UpdateApp(ctx, owner, a.Id, nil, nil, domains.ThemePatch{PrimaryColor: &invalid})
	require.Error(t, err)
	public, err = domain.AppByKey(ctx, a.PublicKey)
	require.NoError(t, err)
	require.Equal(t, want, public.Theme)
	name := "Renamed"
	updated, err = domain.UpdateApp(ctx, owner, a.Id, &name, nil, domains.ThemePatch{})
	require.NoError(t, err)
	require.Equal(t, want, updated.Theme)
	require.Equal(t, name, updated.Name)
	primary = "#696CF1"
	updated, err = domain.UpdateApp(ctx, owner, a.Id, nil, nil, domains.ThemePatch{PrimaryColor: &primary})
	require.NoError(t, err)
	want.PrimaryColor = primary
	require.Equal(t, want, updated.Theme)
	unchanged, err := domain.AppByKey(ctx, b.PublicKey)
	require.NoError(t, err)
	require.Equal(t, defaults, unchanged.Theme)
	_, err = pool.Exec(ctx, `UPDATE apps SET primary_color='bad' WHERE id=$1`, a.Id)
	require.Error(t, err)
}
