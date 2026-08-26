// Package apperr provides a typed error taxonomy for the application.
// Every error that crosses domain / adapter / transport boundaries must use this type
// so the transport layer can map errors to the correct HTTP status without string-matching.
package apperr

import "fmt"

// Code is a machine-readable error classification used to derive HTTP status codes.
type Code string

const (
	CodeInvalid  Code = "INVALID_INPUT"
	CodeNotFound Code = "NOT_FOUND"
	CodeConflict Code = "CONFLICT"
	CodeUpstream Code = "UPSTREAM_ERROR"
	CodeInternal Code = "INTERNAL"
)

// Error is the single error type that crosses domain / adapter / transport boundaries.
type Error struct {
	Code    Code
	Message string
	Err     error // underlying cause, never exposed to clients
}

// New creates an *Error with no underlying cause.
func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap creates an *Error that wraps an underlying cause.
func Wrap(code Code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause, supporting errors.Is / errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the authoritative HTTP status code for this error Code.
// This is the single source of truth for Code → HTTP mapping.
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeInvalid:
		return 400
	case CodeNotFound:
		return 404
	case CodeConflict:
		return 409
	case CodeUpstream:
		return 502
	case CodeInternal:
		return 500
	default:
		return 500
	}
}
