package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
	"github.com/rs/zerolog"
)

type TourRepository struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewTourRepository(
	pool *pgxpool.Pool,
	logger *zerolog.Logger,
) *TourRepository {
	return &TourRepository{
		pool:   pool,
		logger: logger,
	}
}

func (tr *TourRepository) CreateWithDraft(ctx context.Context, t *domains.Tour) (*domains.Tour, error) {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, "", errs.Create)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
			INSERT INTO tours (app_id, title, description, enabled, priority)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at`,
		t.AppId, t.Title, t.Description, t.Enabled, t.Priority,
	).Scan(&t.Id, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, t.Title, errs.Create)
	}

	triggerConfigBytes, err := marshalTriggerConfig(t.TriggerConfig)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, t.Id.String(), errs.Create)
	}

	var audienceRulesBytes []byte
	if t.Audience != nil && len(t.Audience.Rules) > 0 {
		audienceRulesBytes, err = marshalAudienceRules(t.Audience.Rules)
		if err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, t.Id.String(), errs.Create)
		}
	}

	err = tx.QueryRow(ctx, `
			INSERT INTO tour_versions (
				tour_id, version, status, trigger_type, target_path,
				audience_show_once, audience_max_shows, audience_only_new,
				trigger_config, audience_rules)
			VALUES ($1, 1, 'draft', $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, version`,
		t.Id, t.TriggerType, t.TargetPath,
		t.Audience.ShowOnce, t.Audience.MaxShows, t.Audience.OnlyNew,
		triggerConfigBytes, audienceRulesBytes,
	).Scan(&t.VersionId, &t.Version)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, t.Id.String(), errs.Create)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, t.Id.String(), errs.Create)
	}

	t.Status = domains.TourDraft
	return t, nil
}

func (tr *TourRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.Tour, error) {
	t := new(domains.Tour)
	err := tr.pool.QueryRow(ctx, `
			SELECT id, app_id, title, description, enabled, priority, created_at, updated_at
			FROM tours
			WHERE id = $1 AND archived_at IS NULL`, id,
	).Scan(&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, id.String(), errs.Retrieve)
	}
	return t, nil
}

func (tr *TourRepository) List(ctx context.Context, appId uuid.UUID) ([]*domains.Tour, error) {
	rows, err := tr.pool.Query(ctx, `
			SELECT t.id, t.app_id, t.title, t.description, t.enabled, t.priority,
				t.created_at, t.updated_at,
				v.id, v.version, v.status, v.trigger_type, v.target_path,
				v.audience_show_once, v.audience_max_shows, v.audience_only_new, 
				(SELECT COUNT(*)
   				FROM hints h
    			WHERE h.tour_version_id = v.id) AS steps_count
			FROM tours t
			LEFT JOIN LATERAL (
				SELECT * FROM tour_versions tv
				WHERE tv.tour_id = t.id
				ORDER BY (tv.status = 'published') DESC, tv.version DESC
				LIMIT 1
			) v ON true
			WHERE t.app_id = $1 AND t.archived_at IS NULL
			ORDER BY t.priority DESC, t.created_at ASC`, appId)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, appId.String(), errs.Retrieve)
	}
	defer rows.Close()

	tours := make([]*domains.Tour, 0)
	for rows.Next() {
		t := new(domains.Tour)
		var (
			versionId   *uuid.UUID
			version     *int
			status      *string
			triggerType *string
			targetPath  *string
			showOnce    *bool
			maxShows    *int
			onlyNew     *bool
		)
		if err := rows.Scan(
			&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority,
			&t.CreatedAt, &t.UpdatedAt,
			&versionId, &version, &status, &triggerType, &targetPath,
			&showOnce, &maxShows, &onlyNew, &t.StepsCount,
		); err != nil {
			return nil, wrap(err, domains.ErrTourNotFound, appId.String(), errs.Retrieve)
		}

		if versionId != nil {
			t.VersionId = *versionId
			t.Version = *version
			t.Status = domains.TourStatus(*status)
			t.TriggerType = domains.TriggerType(*triggerType)
			t.TargetPath = *targetPath
			t.Audience = &domains.Audience{ShowOnce: *showOnce, MaxShows: *maxShows, OnlyNew: *onlyNew}
		}
		tours = append(tours, t)
	}
	return tours, rows.Err()
}

func (tr *TourRepository) UpdateMeta(ctx context.Context, id uuid.UUID, p domains.TourMetaPatch) (*domains.Tour, error) {
	set := make([]string, 0, 4)
	args := make([]any, 0, 5)

	add := func(column string, value any) {
		args = append(args, value)
		set = append(set, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if p.Title != nil {
		add("title", *p.Title)
	}
	if p.Description != nil {
		add("description", *p.Description)
	}
	if p.Enabled != nil {
		add("enabled", *p.Enabled)
	}
	if p.Priority != nil {
		add("priority", *p.Priority)
	}
	if len(set) == 0 {
		return tr.GetById(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
			UPDATE tours SET %s
			WHERE id = $%d AND archived_at IS NULL
			RETURNING id, app_id, title, description, enabled, priority, created_at, updated_at`,
		strings.Join(set, ", "), len(args))

	t := new(domains.Tour)
	err := tr.pool.QueryRow(ctx, query, args...).Scan(
		&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, id.String(), errs.Update)
	}
	return t, nil
}

func (tr *TourRepository) Archive(ctx context.Context, id uuid.UUID) error {
	tag, err := tr.pool.Exec(ctx,
		`UPDATE tours SET archived_at = now() WHERE id = $1 AND archived_at IS NULL`, id)
	if err != nil {
		return wrap(err, domains.ErrTourNotFound, id.String(), errs.Delete)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound(domains.ErrTourNotFound, id.String())
	}
	return nil
}

func (tr *TourRepository) ResolveCandidates(ctx context.Context, appId uuid.UUID) ([]*domains.Tour, error) {
	rows, err := tr.pool.Query(ctx, `
			SELECT t.id, t.app_id, t.title, t.description, t.enabled, t.priority,
				t.created_at, t.updated_at,
				v.id, v.version, v.trigger_type, v.target_path,
				v.audience_show_once, v.audience_max_shows, v.audience_only_new,
				v.audience_rules, v.trigger_config
			FROM tours t
			JOIN tour_versions v ON v.tour_id = t.id AND v.status = 'published'
			WHERE t.app_id = $1 AND t.enabled AND t.archived_at IS NULL
			ORDER BY t.priority DESC, t.created_at ASC`, appId)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, appId.String(), errs.Retrieve)
	}
	defer rows.Close()

	tours := make([]*domains.Tour, 0)
	for rows.Next() {
		t := new(domains.Tour)
		t.Audience = new(domains.Audience)

		var audienceRulesBytes []byte
		var triggerConfigBytes []byte

		if err := rows.Scan(
			&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority,
			&t.CreatedAt, &t.UpdatedAt,
			&t.VersionId, &t.Version, &t.TriggerType, &t.TargetPath,
			&t.Audience.ShowOnce, &t.Audience.MaxShows, &t.Audience.OnlyNew,
			&audienceRulesBytes, &triggerConfigBytes,
		); err != nil {
			return nil, wrap(err, domains.ErrTourNotFound, appId.String(), errs.Retrieve)
		}

		if err := unmarshalAudienceRules(audienceRulesBytes, &t.Audience.Rules); err != nil {
			tr.logger.Error().
				Err(err).
				Str("tour_id", t.Id.String()).
				Str("version_id", t.VersionId.String()).
				Str("app_id", appId.String()).
				Msg("corrupted audience_rules in published tour, skipping tour")
			continue
		}

		if err := unmarshalTriggerConfig(triggerConfigBytes, &t.TriggerConfig); err != nil {
			tr.logger.Error().
				Err(err).
				Str("tour_id", t.Id.String()).
				Str("version_id", t.VersionId.String()).
				Str("app_id", appId.String()).
				Msg("corrupted trigger_config in published tour, skipping tour")
			continue
		}

		t.Status = domains.TourPublished
		tours = append(tours, t)
	}

	return tours, rows.Err()
}

func (tr *TourRepository) VersionByStatus(ctx context.Context, tourId uuid.UUID, status domains.TourStatus) (*domains.TourVersion, error) {
	version := new(domains.TourVersion)

	err := tr.pool.QueryRow(ctx, `
			SELECT id, tour_id, version, status
			FROM tour_versions
			WHERE tour_id = $1 AND status = $2
		`, tourId, status).Scan(
		&version.Id,
		&version.TourId,
		&version.Version,
		&version.Status,
	)
	if err != nil {
		return nil, wrap(
			err,
			domains.ErrVersionNotFound,
			tourId.String()+"/"+string(status),
			errs.Retrieve,
		)
	}

	return version, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func marshalJSON(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Pointer,
		reflect.Interface,
		reflect.Map,
		reflect.Slice,
		reflect.Chan,
		reflect.Func:

		if rv.IsNil() {
			return nil, nil
		}
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf(
			"json marshal: %w",
			err,
		)
	}

	return b, nil
}

func unmarshalJSONB(data []byte, dst any) error {
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}
	return nil
}

func (tr *TourRepository) Publish(ctx context.Context, tourId uuid.UUID) (*domains.PublishResult, error) {
	fail := func(message string, status int) error {
		return &domains.LogicError{
			Err:   errors.New(message),
			Stage: "tour publish",
			Code:  status,
		}
	}

	tx, err := tr.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Блокируем сценарий, чтобы две публикации не прошли одновременно.
	var lockedTourId uuid.UUID
	err = tx.QueryRow(ctx, `
        SELECT id FROM tours
        WHERE id = $1 AND archived_at IS NULL
        FOR UPDATE`, tourId,
	).Scan(&lockedTourId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fail("tour not found", http.StatusNotFound)
	}
	if err != nil {
		return nil, err
	}

	// 2. Ищем draft.
	var draftId uuid.UUID
	var draftVersion int
	err = tx.QueryRow(ctx, `
        SELECT id, version FROM tour_versions
        WHERE tour_id = $1 AND status = 'draft'
        FOR UPDATE`, tourId,
	).Scan(&draftId, &draftVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fail("draft version not found", http.StatusConflict)
	}
	if err != nil {
		return nil, err
	}

	// 3. Проверяем, что шаги заполнены и пронумерованы от 1 до N.
	rows, err := tx.Query(ctx, `
        SELECT step, title, content, selector, placement
        FROM hints WHERE tour_version_id = $1
        ORDER BY step`, draftId,
	)
	if err != nil {
		return nil, err
	}

	count := 0
	invalid := false
	for rows.Next() {
		var step int
		var title, content, selector, placement string
		if err := rows.Scan(&step, &title, &content, &selector, &placement); err != nil {
			rows.Close()
			return nil, err
		}
		count++
		validPlacement := domains.Placement(placement).Valid()
		if step != count || strings.TrimSpace(title) == "" ||
			strings.TrimSpace(content) == "" || strings.TrimSpace(selector) == "" ||
			!validPlacement {
			invalid = true
		}
	}
	scanErr := rows.Err()
	rows.Close()
	if scanErr != nil {
		return nil, scanErr
	}
	if count == 0 || invalid {
		return nil, fail("draft must contain valid, consecutively numbered hints", http.StatusConflict)
	}

	// 4. Проверяем, отличается ли draft от текущей опубликованной версии.
	var oldPublishedId *uuid.UUID
	err = tx.QueryRow(ctx, `
        SELECT id FROM tour_versions
        WHERE tour_id = $1 AND status = 'published'`, tourId,
	).Scan(&oldPublishedId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if oldPublishedId != nil {
		var same bool
		err = tx.QueryRow(ctx, `
            SELECT (
                d.trigger_type IS NOT DISTINCT FROM p.trigger_type
                AND d.target_path IS NOT DISTINCT FROM p.target_path
                AND d.audience_show_once IS NOT DISTINCT FROM p.audience_show_once
                AND d.audience_max_shows IS NOT DISTINCT FROM p.audience_max_shows
                AND d.audience_only_new IS NOT DISTINCT FROM p.audience_only_new
                AND d.trigger_config IS NOT DISTINCT FROM p.trigger_config
                AND d.audience_rules IS NOT DISTINCT FROM p.audience_rules
                AND (
                    SELECT COALESCE(jsonb_agg(jsonb_build_array(step, title, content, selector, placement)
                                              ORDER BY step), '[]'::jsonb)
                    FROM hints WHERE tour_version_id = d.id
                ) = (
                    SELECT COALESCE(jsonb_agg(jsonb_build_array(step, title, content, selector, placement)
                                              ORDER BY step), '[]'::jsonb)
                    FROM hints WHERE tour_version_id = p.id
                )
            )
            FROM tour_versions d
            JOIN tour_versions p ON p.id = $2
            WHERE d.id = $1`, draftId, *oldPublishedId,
		).Scan(&same)
		if err != nil {
			return nil, err
		}
		if same {
			return nil, fail("no changes to publish", http.StatusConflict)
		}

		// 5. Архивируем старую версию, не удаляя её или её hints.
		_, err = tx.Exec(ctx, `
            UPDATE tour_versions
            SET status = 'archived', archived_at = now()
            WHERE id = $1 AND status = 'published'`, *oldPublishedId,
		)
		if err != nil {
			return nil, err
		}
	}

	// 6. Публикуем draft.
	_, err = tx.Exec(ctx, `
        UPDATE tour_versions
        SET status = 'published', published_at = now()
        WHERE id = $1 AND status = 'draft'`, draftId,
	)
	if err != nil {
		return nil, err
	}

	// 7. Создаём новый draft-клон опубликованной версии.
	var newDraftId uuid.UUID
	err = tx.QueryRow(ctx, `
        INSERT INTO tour_versions (
            tour_id, version, status, trigger_type, target_path,
            audience_show_once, audience_max_shows, audience_only_new,
            trigger_config, audience_rules, created_by
        )
        SELECT
            tv.tour_id,
            (SELECT COALESCE(MAX(version), 0) + 1 FROM tour_versions WHERE tour_id = $2),
            'draft', tv.trigger_type, tv.target_path,
            tv.audience_show_once, tv.audience_max_shows, tv.audience_only_new,
            tv.trigger_config, tv.audience_rules, tv.created_by
        FROM tour_versions tv WHERE tv.id = $1
        RETURNING id`, draftId, tourId,
	).Scan(&newDraftId)
	if err != nil {
		return nil, err
	}

	// 8. Копируем hints. UUID и временные поля назначаются БД заново.
	_, err = tx.Exec(ctx, `
        INSERT INTO hints (tour_version_id, step, title, content, selector, placement)
        SELECT $1, step, title, content, selector, placement
        FROM hints
        WHERE tour_version_id = $2
        ORDER BY step`, newDraftId, draftId,
	)
	if err != nil {
		return nil, err
	}

	// 9. Всё фиксируем вместе.
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &domains.PublishResult{
		TourId:             tourId,
		PublishedVersionId: draftId,
		Version:            draftVersion,
		NewDraftVersionId:  newDraftId,
	}, nil
}

func (tr *TourRepository) UpdateSettings(
	ctx context.Context,
	tourId uuid.UUID,
	p domains.TourSettingsPatch,
) error {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var draftId uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT v.id
		FROM tours t
		JOIN tour_versions v ON v.tour_id=t.id AND v.status='draft'
		WHERE t.id=$1 AND t.archived_at IS NULL
		FOR UPDATE OF t, v`,
		tourId,
	).Scan(&draftId)
	if errors.Is(err, pgx.ErrNoRows) {
		return &domains.LogicError{
			Err: errors.New("draft not found"), Stage: "tour settings", Code: http.StatusConflict,
		}
	}
	if err != nil {
		return err
	}

	set := make([]string, 0, 6)
	args := make([]any, 0, 7)
	add := func(column string, value any) {
		args = append(args, value)
		set = append(set, fmt.Sprintf("%s=$%d", column, len(args)))
	}
	if p.TargetPath != nil {
		add("target_path", strings.TrimSpace(*p.TargetPath))
	}
	if p.TriggerType != nil {
		add("trigger_type", string(*p.TriggerType))
	}
	if p.AudienceShowOnce != nil {
		add("audience_show_once", *p.AudienceShowOnce)
	}
	if p.AudienceMaxShows != nil {
		add("audience_max_shows", *p.AudienceMaxShows)
	}
	if p.AudienceOnlyNew != nil {
		add("audience_only_new", *p.AudienceOnlyNew)
	}
	if p.TriggerConfig != nil {
		bytes, err := marshalTriggerConfig(p.TriggerConfig)
		if err != nil {
			return err
		}
		add("trigger_config", bytes)
	}

	if len(set) > 0 {
		args = append(args, draftId)
		query := fmt.Sprintf(
			`UPDATE tour_versions SET %s WHERE id=$%d AND status='draft'`,
			strings.Join(set, ", "),
			len(args),
		)
		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return &domains.LogicError{
				Err: errors.New("draft not found"), Stage: "tour settings", Code: http.StatusConflict,
			}
		}
	}

	return tx.Commit(ctx)
}
