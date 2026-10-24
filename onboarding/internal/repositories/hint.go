package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
	"github.com/rs/zerolog"
)

type HintRepository struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewHintRepository(
	pool *pgxpool.Pool,
	logger *zerolog.Logger,
) *HintRepository {
	return &HintRepository{
		pool:   pool,
		logger: logger,
	}
}

const hintColumns = `
	id,
	tour_version_id,
	step,
	title,
	content,
	selector,
	placement,
	created_at,
	updated_at
`

func (hr *HintRepository) Create(
	ctx context.Context,
	h *domains.Hint,
) error {
	err := hr.pool.QueryRow(ctx, `
		INSERT INTO hints (
			tour_version_id,
			step,
			title,
			content,
			selector,
			placement
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING
			id,
			created_at,
			updated_at`,
		h.TourVersionId,
		h.Step,
		h.Title,
		h.Content,
		h.Selector,
		h.Placement,
	).Scan(
		&h.Id,
		&h.CreatedAt,
		&h.UpdatedAt,
	)

	return wrap(
		err,
		domains.ErrHintNotFound,
		h.TourVersionId.String(),
		errs.Create,
	)
}

func (hr *HintRepository) GetById(
	ctx context.Context,
	id uuid.UUID,
) (*domains.Hint, error) {
	h, err := scanHint(
		hr.pool.QueryRow(
			ctx,
			`SELECT `+hintColumns+`
			 FROM hints
			 WHERE id = $1`,
			id,
		),
	)
	if err != nil {
		return nil, wrap(
			err,
			domains.ErrHintNotFound,
			id.String(),
			errs.Retrieve,
		)
	}

	return h, nil
}

func (hr *HintRepository) ListByVersion(
	ctx context.Context,
	versionId uuid.UUID,
) ([]domains.Hint, error) {
	rows, err := hr.pool.Query(
		ctx,
		`SELECT `+hintColumns+`
		 FROM hints
		 WHERE tour_version_id = $1
		 ORDER BY step`,
		versionId,
	)
	if err != nil {
		return nil, wrap(
			err,
			domains.ErrHintNotFound,
			versionId.String(),
			errs.Retrieve,
		)
	}
	defer rows.Close()

	hints := make(
		[]domains.Hint,
		0,
	)

	for rows.Next() {
		h, err := scanHint(rows)
		if err != nil {
			return nil, wrap(
				err,
				domains.ErrHintNotFound,
				versionId.String(),
				errs.Retrieve,
			)
		}

		hints = append(
			hints,
			*h,
		)
	}

	return hints, rows.Err()
}

func (hr *HintRepository) Update(
	ctx context.Context,
	h *domains.Hint,
) error {
	updated, err := scanHint(
		hr.pool.QueryRow(ctx, `
			UPDATE hints
			SET
				title = $2,
				content = $3,
				selector = $4,
				placement = $5
			WHERE id = $1
			RETURNING `+hintColumns,
			h.Id,
			h.Title,
			h.Content,
			h.Selector,
			h.Placement,
		),
	)
	if err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			h.Id.String(),
			errs.Update,
		)
	}

	*h = *updated

	return nil
}

func (hr *HintRepository) Delete(
	ctx context.Context,
	versionId uuid.UUID,
	id uuid.UUID,
) error {
	tx, err := hr.pool.Begin(ctx)
	if err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			id.String(),
			errs.Delete,
		)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
		ctx,
		`SET CONSTRAINTS ALL DEFERRED`,
	); err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			versionId.String(),
			errs.Update,
		)
	}

	var step int

	err = tx.QueryRow(
		ctx,
		`DELETE FROM hints
		 WHERE id = $1
		   AND tour_version_id = $2
		 RETURNING step`,
		id,
		versionId,
	).Scan(&step)
	if err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			id.String(),
			errs.Delete,
		)
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE hints
		 SET step = step - 1
		 WHERE tour_version_id = $1
		   AND step > $2`,
		versionId,
		step,
	); err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			versionId.String(),
			errs.Update,
		)
	}

	return wrap(
		tx.Commit(ctx),
		domains.ErrHintNotFound,
		id.String(),
		errs.Delete,
	)
}

func (hr *HintRepository) Reorder(
	ctx context.Context,
	versionId uuid.UUID,
	ids []uuid.UUID,
) error {
	tx, err := hr.pool.Begin(ctx)
	if err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			versionId.String(),
			errs.Update,
		)
	}
	defer tx.Rollback(ctx)

	var count int

	if err := tx.QueryRow(
		ctx,
		`SELECT count(*)
		 FROM hints
		 WHERE tour_version_id = $1`,
		versionId,
	).Scan(&count); err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			versionId.String(),
			errs.Retrieve,
		)
	}

	if count != len(ids) {
		return &errs.RepositoryError{
			Err:       domains.ErrReorderMismatch,
			EntityRef: versionId.String(),
			Operation: errs.Update,
			Reason:    errs.InvalidReference,
		}
	}

	if _, err := tx.Exec(
		ctx,
		`SET CONSTRAINTS ALL DEFERRED`,
	); err != nil {
		return wrap(
			err,
			domains.ErrHintNotFound,
			versionId.String(),
			errs.Update,
		)
	}

	for i, id := range ids {
		tag, err := tx.Exec(
			ctx,
			`UPDATE hints
			 SET step = $1
			 WHERE id = $2
			   AND tour_version_id = $3`,
			i+1,
			id,
			versionId,
		)
		if err != nil {
			return wrap(
				err,
				domains.ErrHintNotFound,
				id.String(),
				errs.Update,
			)
		}

		if tag.RowsAffected() == 0 {
			return &errs.RepositoryError{
				Err:       domains.ErrReorderMismatch,
				EntityRef: id.String(),
				Operation: errs.Update,
				Reason:    errs.InvalidReference,
			}
		}
	}

	return wrap(
		tx.Commit(ctx),
		domains.ErrHintNotFound,
		versionId.String(),
		errs.Update,
	)
}

func scanHint(
	row scanner,
) (*domains.Hint, error) {
	h := new(domains.Hint)

	err := row.Scan(
		&h.Id,
		&h.TourVersionId,
		&h.Step,
		&h.Title,
		&h.Content,
		&h.Selector,
		&h.Placement,
		&h.CreatedAt,
		&h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return h, nil
}
