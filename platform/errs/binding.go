package errs

import "fmt"

// ------------------------- //
// ----- Binding Error ----- //
// ------------------------- //

type (
	BindingZone string

	BindingError struct {
		Err  error
		Zone BindingZone
		Code int
		// Detail carries a human-readable, safe-to-expose explanation of the
		// failure. When set, it is surfaced in Message() instead of the generic
		// text, and Error() stays generic so verbose binder internals never
		// reach the client.
		Detail string
	}
)

const (
	URI    BindingZone = "URI"
	Body   BindingZone = "body"
	JSON   BindingZone = "JSON"
	Header BindingZone = "header"
	Query  BindingZone = "query"
)

func (be *BindingError) Error() string {
	if be.Detail != "" {
		return fmt.Sprintf("failed to bind %s", be.Zone)
	}
	return fmt.Sprintf("failed to bind %s: %v", be.Zone, be.Err)
}

func (be *BindingError) Message() string {
	if be.Detail != "" {
		return "Failed to parse: " + be.Detail
	}
	return fmt.Sprintf("Failed to parse: invalid key/value in %s", be.Zone)
}

func (be *BindingError) ToHTTPError() *HTTPError {
	return &HTTPError{Message: be.Message(), Error: be.Error()}
}

func (be *BindingError) Unwrap() error {
	return be.Err
}

func (be *BindingError) StatusCode() int {
	return be.Code
}
