package middlewares

import "github.com/ksendel-coder/onboarding_croquant/platform/errs"

// AuthHeaderError is also used by the SDK app-key middleware for 401/403
// responses. The OIDC login system was removed; X-App-Key remains because it
// identifies which application's onboarding scenarios should be resolved.
type AuthHeaderError struct {
	Code int
	msg  string
	Err  error
}

func NewAuthHeaderError(code int, msg string, err error) *AuthHeaderError {
	return &AuthHeaderError{Code: code, msg: msg, Err: err}
}

func (iah *AuthHeaderError) Error() string   { return iah.Err.Error() }
func (iah *AuthHeaderError) Message() string { return iah.msg }
func (iah *AuthHeaderError) Unwrap() error   { return iah.Err }
func (iah *AuthHeaderError) StatusCode() int { return iah.Code }
func (iah *AuthHeaderError) ToHTTPError() *errs.HTTPError {
	return &errs.HTTPError{Error: iah.Err.Error(), Message: iah.msg}
}
