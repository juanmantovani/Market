package domain

import "errors"

// Kind classifies a business error by the kind of failure it represents,
// independent of any delivery mechanism. It's a string so it reads directly
// in logs.
type Kind string

const (
	KindInvalid     Kind = "invalid"
	KindForbidden   Kind = "forbidden"
	KindConflict    Kind = "conflict"
	KindNotFound    Kind = "not_found"
	KindUnavailable Kind = "unavailable"
)

var (
	ErrStorageUnavailable = &Error{Kind: KindUnavailable, Code: "storage_unavailable", Message: "El almacenamiento no está disponible"}
	ErrForbidden          = &Error{Kind: KindForbidden, Code: "forbidden", Message: "No hay permiso para realizar esta acción"}
)

// Error is a business error with a stable Code and a user-facing Message,
// classified by Kind so adapters can map it to their own representation
// (e.g. an HTTP status). Cause, when set, is the underlying failure: adapters
// may log it, but only Message is safe to show to the caller.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Cause   error
}

func (err *Error) Error() string {
	return err.Message
}

func (err *Error) Unwrap() error {
	return err.Cause
}

// Is reports errors with the same Code as equivalent, so errors.Is matches a
// sentinel against the copy WithCause returns.
func (err *Error) Is(target error) bool {
	other, ok := target.(*Error)

	return ok && other.Code == err.Code
}

// WithCause returns a copy of err carrying cause, leaving the receiver
// untouched so package-level sentinels stay safe to share.
func (err *Error) WithCause(cause error) *Error {
	withCause := *err
	withCause.Cause = cause

	return &withCause
}

// FromStorage is what a use case returns for an error coming from one of its
// adapters: a business error travels as is, anything else is a storage
// failure that isn't the caller's fault.
func FromStorage(err error) error {
	if err == nil {
		return nil
	}

	var domainErr *Error
	if errors.As(err, &domainErr) {
		return err
	}

	return ErrStorageUnavailable.WithCause(err)
}
