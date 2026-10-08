package repositories

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func wrap(
	err error,
	sentinel error,
	ref string,
	op errs.RepoOperation,
) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NotFound(sentinel, ref)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return &errs.RepositoryError{
				Err:       err,
				EntityRef: ref,
				Operation: op,
				Reason:    errs.AlreadyExists,
			}
		case pgForeignKeyViolation:
			return &errs.RepositoryError{
				Err:       err,
				EntityRef: ref,
				Operation: op,
				Reason:    errs.InvalidReference,
			}
		}
	}

	reason := errs.QueryError
	if op != errs.Retrieve {
		reason = errs.ExecError
	}

	return &errs.RepositoryError{
		Err:       err,
		EntityRef: ref,
		Operation: op,
		Reason:    reason,
	}
}
