package core

import (
	"errors"
	"fmt"
)

var (
	// ErrNotImplemented indicates something hasn't been implemented yet
	ErrNotImplemented = errors.New("not implemented")
	// ErrTODO is like ErrNotImplemented but used especially to
	// indicate something needs to be implemented
	ErrTODO = Wrap(ErrNotImplemented, "TODO")
	// ErrExists indicates something already exists
	ErrExists = errors.New("already exists")
	// ErrNotExists indicates something doesn't exist
	ErrNotExists = errors.New("does not exist")
	// ErrInvalid indicates an argument isn't valid
	ErrInvalid = errors.New("invalid argument")
	// ErrUnknown indicates something isn't recognized
	ErrUnknown = errors.New("unknown")
)

var (
	_ Unwrappable = (*WrappedError)(nil)
)

// Unwrappable represents an error that can be Unwrap() to get the cause
type Unwrappable interface {
	Error() string
	Unwrap() error
}

// Wrap annotates an error, optionally with a formatted string.
// if %w is used the argument will be unwrapped
func Wrap(err error, format string, args ...any) error {
	var note string

	if err == nil {
		return nil
	}

	if len(args) > 0 {
		note = fmt.Errorf(format, args...).Error()
	} else {
		note = format
	}

	if len(note) == 0 {
		return err
	}

	return &WrappedError{
		cause: err,
		note:  note,
	}
}

// WrappedError is an annotated error that can be Unwrapped
type WrappedError struct {
	cause error
	note  string
}

func (w *WrappedError) Error() string {
	s := w.cause.Error()
	if len(s) == 0 {
		return w.note
	}

	return fmt.Sprintf("%s: %s", w.note, s)
}

func (w *WrappedError) Unwrap() error {
	return w.cause
}

// CoalesceError returns the first non-nil error argument.
// error isn't compatible with Coalesce's comparable generic
// type.
func CoalesceError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// Unwrap unwraps one layer of a compound error,
// ensuring there are no nil entries.
func Unwrap(err error) []error {
	var errs []error

	if err == nil {
		return nil
	}

	switch w := err.(type) {
	case interface {
		Unwrap() []error
	}:
		errs = w.Unwrap()
	case interface {
		Errors() []error
	}:
		errs = w.Errors()
	case interface {
		Unwrap() error
	}:
		errs = append(errs, w.Unwrap())
	}

	return SliceReplaceFn(errs, func(_ []error, err error) (error, bool) {
		return err, err != nil
	})
}
