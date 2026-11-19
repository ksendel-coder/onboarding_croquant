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

func TestTourListDraftStateIntegration(t *testing.T) {
	dsn := os.Getenv("MVP08_TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("REQUIRE_TOUR_LIST_DB") == "1" {
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
	repo := NewTourRepository(pool, &logger)
	appID, otherAppID, tourID, firstDraftID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	otherTourID := uuid.New()
	exec := func(sql string, args ...any) {
		t.Helper()
		_, err := pool.Exec(ctx, sql, args...)
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		// Удаляем исключительно записи двух приложений, созданных тестом.
		for _, q := range []string{
			`DELETE FROM hints WHERE tour_version_id IN (SELECT v.id FROM tour_versions v JOIN tours t ON t.id=v.tour_id WHERE t.app_id=ANY($1::uuid[]))`,
			`DELETE FROM tour_versions WHERE tour_id IN (SELECT id FROM tours WHERE app_id=ANY($1::uuid[]))`,
			`DELETE FROM tours WHERE app_id=ANY($1::uuid[])`,
			`DELETE FROM apps WHERE id=ANY($1::uuid[])`,
		} {
			_, err := pool.Exec(context.Background(), q, []uuid.UUID{appID, otherAppID})
			if err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})
	exec(`INSERT INTO apps(id,name,public_key) VALUES ($1,'Draft list test',$2),($3,'Other app',$4)`, appID, "pk_"+appID.String(), otherAppID, "pk_"+otherAppID.String())
	exec(`INSERT INTO tours(id,app_id,title,enabled) VALUES ($1,$2,'Test tour',true),($3,$4,'Other tour',true)`, tourID, appID, otherTourID, otherAppID)
	exec(`INSERT INTO tour_versions(id,tour_id,version,status,target_path) VALUES ($1,$2,1,'draft','/home'),($3,$4,1,'draft','/home')`, firstDraftID, tourID, uuid.New(), otherTourID)
	addHint := func(versionID uuid.UUID, step int) {
		t.Helper()
		exec(`INSERT INTO hints(tour_version_id,step,title,content,selector,placement) VALUES ($1,$2,'Title','Original content','#target','bottom')`, versionID, step)
	}
	addHint(firstDraftID, 1)
	addHint(firstDraftID, 2)
	check := func(versionID uuid.UUID, status domains.TourStatus, published, changed bool, count int) {
		t.Helper()
		items, err := repo.List(ctx, appID)
		require.NoError(t, err)
		require.Len(t, items, 1, "list must be scoped to app")
		item := items[0]
		require.Equal(t, tourID, item.Id)
		require.Equal(t, versionID, item.VersionId)
		require.Equal(t, status, item.Status)
		require.Equal(t, published, item.HasPublished)
		require.Equal(t, changed, item.HasUnpublishedChanges)
		require.Equal(t, count, item.StepsCount)
	}
	runtime := func(versionID uuid.UUID, path string, count int) {
		t.Helper()
		items, err := repo.ResolveCandidates(ctx, appID, "new-visitor")
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, tourID, items[0].Id)
		require.Equal(t, versionID, items[0].VersionId)
		require.Equal(t, domains.TourPublished, items[0].Status)
		require.Equal(t, path, items[0].TargetPath)
		var n int
		require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM hints WHERE tour_version_id=$1`, versionID).Scan(&n))
		require.Equal(t, count, n)
	}
	// 1. Только draft.
	check(firstDraftID, domains.TourDraft, false, true, 2)
	candidates, err := repo.ResolveCandidates(ctx, appID, "new-visitor")
	require.NoError(t, err)
	require.Empty(t, candidates)

	// 2. Реальная публикация создаёт идентичный draft с другими UUID шагов.
	result, err := repo.Publish(ctx, tourID)
	require.NoError(t, err)
	require.Equal(t, firstDraftID, result.PublishedVersionId)
	draftID := result.NewDraftVersionId
	require.NotEqual(t, uuid.Nil, draftID)
	require.NotEqual(t, firstDraftID, draftID)
	check(draftID, domains.TourDraft, true, false, 2)
	runtime(firstDraftID, "/home", 2)

	// 3. Изменение текста/селектора без изменения количества.
	exec(`UPDATE hints SET content='Edited',selector='#new' WHERE tour_version_id=$1 AND step=1`, draftID)
	check(draftID, domains.TourDraft, true, true, 2)
	// 4–5. Добавление и удаление шага.
	addHint(draftID, 3)
	check(draftID, domains.TourDraft, true, true, 3)
	runtime(firstDraftID, "/home", 2)
	exec(`DELETE FROM hints WHERE tour_version_id=$1 AND step=3`, draftID)
	check(draftID, domains.TourDraft, true, true, 2)
	// 6–7. Возврат содержимого, затем изменение только настроек.
	exec(`UPDATE hints SET content='Original content',selector='#target' WHERE tour_version_id=$1 AND step=1`, draftID)
	check(draftID, domains.TourDraft, true, false, 2)
	path := "/profile"
	require.NoError(t, repo.UpdateSettings(ctx, tourID, domains.TourSettingsPatch{TargetPath: &path}))
	check(draftID, domains.TourDraft, true, true, 2)
	runtime(firstDraftID, "/home", 2)
	path = "/home"
	require.NoError(t, repo.UpdateSettings(ctx, tourID, domains.TourSettingsPatch{TargetPath: &path}))
	check(draftID, domains.TourDraft, true, false, 2)
	// Общие метаданные не являются версионируемыми изменениями.
	exec(`UPDATE tours SET title='Renamed' WHERE id=$1`, tourID)
	check(draftID, domains.TourDraft, true, false, 2)
	// Изменение заголовка шага.
	exec(`UPDATE hints SET title='Different' WHERE tour_version_id=$1 AND step=2`, draftID)
	check(draftID, domains.TourDraft, true, true, 2)
	// Повторная публикация: новый published, чистый draft, старая версия archived.
	addHint(draftID, 3)
	result, err = repo.Publish(ctx, tourID)
	require.NoError(t, err)
	require.Equal(t, draftID, result.PublishedVersionId)
	check(result.NewDraftVersionId, domains.TourDraft, true, false, 3)
	runtime(draftID, "/home", 3)
	var oldStatus string
	require.NoError(t, pool.QueryRow(ctx, `SELECT status FROM tour_versions WHERE id=$1`, firstDraftID).Scan(&oldStatus))
	require.Equal(t, "archived", oldStatus)
	// 8. Старые данные: только published, archived история не выбирается.
	exec(`DELETE FROM tour_versions WHERE id=$1`, result.NewDraftVersionId)
	check(draftID, domains.TourPublished, true, false, 3)
	// 9. Архивированный сценарий исключается из списка и runtime.
	exec(`UPDATE tours SET archived_at=now() WHERE id=$1`, tourID)
	items, err := repo.List(ctx, appID)
	require.NoError(t, err)
	require.Empty(t, items)
	candidates, err = repo.ResolveCandidates(ctx, appID, "new-visitor")
	require.NoError(t, err)
	require.Empty(t, candidates)
	// Проверяем, что List/ResolveCandidates сами не создали событий/прогресса.
	for _, q := range []string{`SELECT count(*) FROM tour_events WHERE app_id=$1`, `SELECT count(*) FROM tour_progress WHERE app_id=$1`} {
		var n int
		require.NoError(t, pool.QueryRow(ctx, q, appID).Scan(&n))
		require.Zero(t, n)
	}
}
