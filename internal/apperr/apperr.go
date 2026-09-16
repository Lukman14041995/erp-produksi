// Package apperr defines typed application errors that carry an HTTP status,
// so handlers can translate service-layer failures into the standard
// response envelope without re-deriving status codes ad hoc.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Kind string

const (
	KindValidation Kind = "VALIDATION"
	KindNotFound   Kind = "NOT_FOUND"
	KindConflict   Kind = "CONFLICT"
	KindForbidden  Kind = "FORBIDDEN"
	KindInternal   Kind = "INTERNAL"
)

type Error struct {
	Kind    Kind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func (e *Error) HTTPStatus() int {
	switch e.Kind {
	case KindValidation:
		return http.StatusUnprocessableEntity
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func Validation(msg string) *Error { return &Error{Kind: KindValidation, Message: msg} }
func NotFound(msg string) *Error   { return &Error{Kind: KindNotFound, Message: msg} }
func Conflict(msg string) *Error   { return &Error{Kind: KindConflict, Message: msg} }
func Forbidden(msg string) *Error  { return &Error{Kind: KindForbidden, Message: msg} }

func Internal(msg string, err error) *Error {
	return &Error{Kind: KindInternal, Message: msg, Err: err}
}

func Wrapf(err error, format string, args ...any) *Error {
	var ae *Error
	if errors.As(err, &ae) {
		return &Error{Kind: ae.Kind, Message: fmt.Sprintf(format, args...) + ": " + ae.Message, Err: ae.Err}
	}
	return Internal(fmt.Sprintf(format, args...), err)
}

func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
