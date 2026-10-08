package errs

import "fmt"

// ----------------------------- //
// ----- Permissions Error ----- //
// ----------------------------- //

type (
	PermissionError struct {
		Err  error
		Code int
	}
)

func (pe *PermissionError) Error() string {
	return fmt.Sprintf("access denied: %v", pe.Err)
}

func (pe *PermissionError) Message() string {
	return "Access denied"
}

func (pe *PermissionError) ToHTTPError() *HTTPError {
	return &HTTPError{Message: pe.Message(), Error: "access denied"}
}

func (pe *PermissionError) Unwrap() error {
	return pe.Err
}

func (pe *PermissionError) StatusCode() int {
	return pe.Code
}
