package domains

import (
	"errors"
	"fmt"

	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

// Business Logic Error

type LogicError struct {
	Err   error
	Stage string
	Code  int
}

func (le *LogicError) Error() string {
	return fmt.Sprintf("failure during %s: %v", le.Stage, le.Err)
}

func (le *LogicError) Message() string {
	if le.Err == nil {
		return fmt.Sprintf("An error occurred during %s", le.Stage)
	}
	return errs.Capitalise(le.Err.Error())
}

func (le *LogicError) ToHTTPError() *errs.HTTPError {
	return &errs.HTTPError{
		Message: le.Message(),
		Error:   fmt.Sprintf("failure during %s: %s", le.Stage, le.Err),
	}
}

func (le *LogicError) Unwrap() error {
	return le.Err
}

func (le *LogicError) StatusCode() int {
	return le.Code
}

// Onboarding Reasons

var (
	ErrTitleRequired      = errors.New("title is required")
	ErrOriginRequired     = errors.New("allowed_origins must not be empty")
	ErrInvalidOrigin      = errors.New("invalid origin format")
	ErrAppNotFound        = errors.New("app not found")
	ErrAppKeyRequired     = errors.New("X-App-Key header is required")
	ErrOriginNotAllowed   = errors.New("origin is not in the allowed list")
	ErrSubjectRequired    = errors.New("subjectId is required")
	ErrTargetPathRequired = errors.New("target_path is required")
	ErrPriorityNegative   = errors.New("priority must be >= 0")
	ErrTourNotFound       = errors.New("tour not found")
	ErrVersionNotFound    = errors.New("version not found")
)

// Helpers

func logicErr(err error, stage string, code int) error {
	return &LogicError{Err: err, Stage: stage, Code: code}
}
