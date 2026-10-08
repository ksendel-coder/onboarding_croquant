package errs

// BaseError is the contract the HTTP error middleware type-switches on. Every
// error a handler hands to gin is expected to satisfy it, which is what lets a
// single ErrorHandler turn any layer's failure into a response body.
//
// It lives here rather than next to the middleware because both services and
// all three layers implement it: the interface is the shared vocabulary, not a
// detail of one consumer.
type BaseError interface {
	Error() string
	Unwrap() error
	Message() string
	ToHTTPError() *HTTPError
	StatusCode() int
}

// Capitalise upper-cases the first rune so a lowercase sentinel reads as a
// sentence when it is surfaced to a client.
func Capitalise(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
}
