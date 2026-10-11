package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// ---------------------------- //
// ----- Repository Error ----- //
// ---------------------------- //

type (
	RepoOperation string

	RepoErrorReason string

	RepositoryErrorInterface interface {
		Error() string
		Unwrap() error
		Message() string
		ToHTTPError() *HTTPError
		StatusCode() int
	}

	RepositoryError struct {
		Err       error
		EntityRef string
		Operation RepoOperation
		Reason    RepoErrorReason
	}
)

// Repository operations
const (
	Create   RepoOperation = "create"
	Retrieve RepoOperation = "retrieve"
	Update   RepoOperation = "update"
	Delete   RepoOperation = "delete"
)

// Repository error reasons
const (
	Unknown          RepoErrorReason = "unknown"
	NoRecord         RepoErrorReason = "norecord"
	AlreadyExists    RepoErrorReason = "exists"
	CacheMiss        RepoErrorReason = "cachemmiss"
	QueryError       RepoErrorReason = "query"
	ExecError        RepoErrorReason = "exec"
	InvalidReference RepoErrorReason = "invref"
)

func (re *RepositoryError) Error() string {
	return fmt.Sprintf("failed to %s entity %s: %v", re.Operation, re.EntityRef, re.Err)
}

func (re *RepositoryError) Message() string {
	if re.Reason == NoRecord && re.Err != nil {
		return Capitalise(re.Err.Error())
	}
	return fmt.Sprintf("Unable to %s the requested entity", re.Operation)
}

func (re *RepositoryError) Unwrap() error {
	return re.Err
}

func (re *RepositoryError) StatusCode() int {
	switch re.Reason {
	case NoRecord:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case InvalidReference:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (re *RepositoryError) ToHTTPError() *HTTPError {
	return &HTTPError{
		Error:   fmt.Sprintf("failed to %s entity %s: %s", re.Operation, re.EntityRef, re.Reason),
		Message: re.Message(),
	}
}

func NotFound(sentinel error, ref string) error {
	return &RepositoryError{
		Err:       sentinel,
		EntityRef: ref,
		Operation: Retrieve,
		Reason:    NoRecord,
	}
}

func IsNotFound(err error) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Reason == NoRecord
	}
	return false
}
