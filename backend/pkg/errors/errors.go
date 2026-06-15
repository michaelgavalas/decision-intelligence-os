// Package errors provides a typed error kernel for the backend. Each error
// carries a Kind (for transport-level mapping), a stable machine Code, a human
// readable Message, and an optional wrapped cause.
package errors

import (
	stderrors "errors"
	"fmt"
)

// Kind classifies an error so that transport layers (such as GraphQL) can map
// it to an appropriate response. The zero value is KindInternal, which is the
// safe default for an unclassified failure.
type Kind int

const (
	// KindInternal is an unexpected server-side failure.
	KindInternal Kind = iota
	// KindValidation is a malformed or invalid client request.
	KindValidation
	// KindNotFound is a missing resource.
	KindNotFound
	// KindConflict is a state conflict, such as a uniqueness violation.
	KindConflict
	// KindUnauthenticated indicates the caller is not authenticated.
	KindUnauthenticated
	// KindForbidden indicates the caller is authenticated but not authorized.
	KindForbidden
)

// Error is the typed error used throughout the backend.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	err     error
}

// Error implements the error interface. It includes the Code and Message, and
// appends the wrapped cause when present.
func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause, allowing errors.Is and errors.As to walk
// the chain.
func (e *Error) Unwrap() error {
	return e.err
}

// NotFound builds a KindNotFound error.
func NotFound(code, msg string) *Error {
	return &Error{Kind: KindNotFound, Code: code, Message: msg}
}

// Validation builds a KindValidation error.
func Validation(code, msg string) *Error {
	return &Error{Kind: KindValidation, Code: code, Message: msg}
}

// Conflict builds a KindConflict error.
func Conflict(code, msg string) *Error {
	return &Error{Kind: KindConflict, Code: code, Message: msg}
}

// Unauthenticated builds a KindUnauthenticated error.
func Unauthenticated(code, msg string) *Error {
	return &Error{Kind: KindUnauthenticated, Code: code, Message: msg}
}

// Forbidden builds a KindForbidden error.
func Forbidden(code, msg string) *Error {
	return &Error{Kind: KindForbidden, Code: code, Message: msg}
}

// Internal builds a KindInternal error.
func Internal(code, msg string) *Error {
	return &Error{Kind: KindInternal, Code: code, Message: msg}
}

// Wrap wraps an existing error with the given kind, code, and message while
// preserving the underlying error for the Unwrap chain.
func Wrap(err error, kind Kind, code, msg string) *Error {
	return &Error{Kind: kind, Code: code, Message: msg, err: err}
}

// KindOf walks the error chain and returns the Kind of the first *Error found.
// It returns KindInternal when no *Error is present.
func KindOf(err error) Kind {
	var e *Error
	if stderrors.As(err, &e) {
		return e.Kind
	}
	return KindInternal
}

// CodeOf walks the error chain and returns the Code of the first *Error found.
// It returns the empty string when no *Error is present.
func CodeOf(err error) string {
	var e *Error
	if stderrors.As(err, &e) {
		return e.Code
	}
	return ""
}
