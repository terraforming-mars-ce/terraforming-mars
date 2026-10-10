package save

import "errors"

// Error carries a stable failure code and a safe message independently of its cause.
type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap preserves diagnostic causes for server-side inspection.
func (e *Error) Unwrap() error { return e.Cause }

// Failure constructs a player-facing failure without exposing its diagnostic cause.
func Failure(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// PublicError returns only deliberately classified messages; unknown failures stay private.
func PublicError(err error, fallback string) *Error {
	var known *Error
	if errors.As(err, &known) {
		return known
	}
	return &Error{Code: "internal_error", Message: fallback, Cause: err}
}
