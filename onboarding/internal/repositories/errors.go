package repositories

import (
	"errors"

	"github.com/Vanady39/cluer/platform/errs"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation     = "23505"
	pgRestrictViolation   = "23001"
	pgForeignKeyViolation = "23503"
)

func wrap(err error, sentinel error, ref string, op errs.RepoOperation) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NotFound(sentinel, ref)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgRestrictViolation:
			return &domains.LogicError{
				Err:   domains.ErrVersionImmutable,
				Stage: "immutability guard",
				Code:  409,
			}
		case pgUniqueViolation:
			return &errs.RepositoryError{
				Err: err, EntityRef: ref, Operation: op, Reason: errs.AlreadyExists,
			}
		case pgForeignKeyViolation:
			return &errs.RepositoryError{
				Err: err, EntityRef: ref, Operation: op, Reason: errs.InvalidReference,
			}
		}
	}

	reason := errs.QueryError
	if op != errs.Retrieve {
		reason = errs.ExecError
	}
	return &errs.RepositoryError{Err: err, EntityRef: ref, Operation: op, Reason: reason}
}
