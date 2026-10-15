package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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
		       v.audience_show_once, v.audience_max_shows, v.audience_only_new
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
			&showOnce, &maxShows, &onlyNew,
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
