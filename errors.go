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
func Wrap(err error, format string, args ...any) error {
	return doWrap(err, false, format, args...)
}

// QuietWrap replaces the text of the error it's wrapping.
// if %w is used the argument will be unwrapped.
func QuietWrap(err error, format string, args ...any) error {
	return doWrap(err, true, format, args...)
}

func doWrap(err error, quiet bool, format string, args ...any) error {
	var note string

	switch {
	case err == nil:
		return nil
	case len(args) > 0:
		note = fmt.Sprintf(format, args...)
	default:
		note = format
	}

	if len(note) == 0 {
		return err
	}

	return &WrappedError{
		cause: err,
		note:  note,
		quiet: quiet,
	}
}

// WrappedError is an annotated error that can be Unwrapped
type WrappedError struct {
	cause error
	note  string
	quiet bool
}

func (w *WrappedError) Error() string {
	switch {
	case w == nil:
		return ""
	case w.cause == nil, w.quiet:
		return w.note
	}

	s := w.cause.Error()
	if len(s) == 0 {
		return w.note
	}

	return fmt.Sprintf("%s: %s", w.note, s)
}

func (w *WrappedError) Unwrap() error {
	if w == nil {
		return nil
	}
	return w.cause
}

// TemporaryError is an error wrapper that satisfies IsTimeout()
// and IsTemporary()
type TemporaryError struct {
	cause   error
	timeout bool
}

func (w *TemporaryError) Error() string {
	var cause string

	switch {
	case w == nil:
		return ""
	case w.cause != nil:
		cause = w.cause.Error()
	}

	switch {
	case !w.timeout:
		return cause
	case cause == "":
		return "time-out"
	default:
		return fmt.Sprintf("%s: %s", "time-out", cause)
	}
}

// IsTemporary tells this error is temporary.
func (*TemporaryError) IsTemporary() bool { return true }

// IsTimeout tells if this error is a time-out or not.
func (w *TemporaryError) IsTimeout() bool {
	if w != nil {
		return w.timeout
	}
	return false
}

// Temporary tells this error is temporary.
func (*TemporaryError) Temporary() bool { return true }

// Timeout tells if this error is a time-out or not.
func (w *TemporaryError) Timeout() bool { return w.IsTimeout() }

// NewTimeoutError returns an error that returns true
// to IsTimeout() and IsTemporary()
func NewTimeoutError(err error) error {
	return &TemporaryError{
		cause:   err,
		timeout: true,
	}
}

// NewTemporaryError returns an error that returns false
// to IsTimeout() and true to IsTemporary()
func NewTemporaryError(err error) error {
	return &TemporaryError{
		cause: err,
	}
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

// IsError recursively check if the given error is in in the given list,
// or just non-nil if no options to check are given.
func IsError(err error, errs ...error) bool {
	switch {
	case err == nil:
		return false
	case len(errs) == 0:
		return true
	}

	fn := func(err error) bool {
		for _, e := range errs {
			if err == e {
				return true
			}
		}
		return false
	}

	return IsErrorFn(fn, err)
}

// IsErrorFn recursively checks if any of the given errors satisfies
// the specified check function.
//
// revive:disable:cognitive-complexity
func IsErrorFn(check func(error) bool, errs ...error) bool {
	// revive:enable:cognitive-complexity
	if check == nil || len(errs) == 0 {
		return false
	}

	// direct match first
	for _, e := range errs {
		if e != nil && check(e) {
			return true
		}
	}

	// and unwrapping
	for _, e := range errs {
		if errs := Unwrap(e); len(errs) > 0 {
			if IsErrorFn(check, errs...) {
				return true
			}
		}
	}

	return false
}

// IsErrorFn2 recursively checks if any of the given errors gets a
// certain answer from the check function.
// As opposed to IsErrorFn, IsErrorFn2 will stop when it has certainty
// of a false result.
//
// revive:disable:cognitive-complexity
func IsErrorFn2(check func(error) (bool, bool), errs ...error) (is bool, known bool) {
	// revive:enable:cognitive-complexity
	if check == nil || len(errs) == 0 {
		return false, true
	}

	// direct match first
	for _, e := range errs {
		if e != nil {
			if is, known = check(e); known {
				return is, true
			}
		}
	}

	// and unwrapping
	for _, e := range errs {
		if errs := Unwrap(e); len(errs) > 0 {
			if is, known = IsErrorFn2(check, errs...); known {
				return is, true
			}
		}
	}

	// unknown
	return false, false
}

// CheckIsTemporary tests an error for Temporary(), IsTemporary(),
// Timeout() and IsTimeout() without unwrapping.
func CheckIsTemporary(err error) (is, known bool) {
	switch e := err.(type) {
	case nil:
		return false, true
	case interface {
		Temporary() bool
	}:
		return e.Temporary(), true
	case interface {
		IsTemporary() bool
	}:
		return e.IsTemporary(), true
	default:
		return CheckIsTimeout(err)
	}
}

// IsTemporary tests an error for Temporary(), IsTemporary(),
// Timeout() and IsTimeout() recursively.
func IsTemporary(err error) bool {
	is, _ := IsErrorFn2(CheckIsTemporary, err)
	return is
}

// CheckIsTimeout tests an error for Timeout() and IsTimeout()
// without unwrapping.
func CheckIsTimeout(err error) (is, known bool) {
	switch e := err.(type) {
	case nil:
		return false, true
	case interface {
		Timeout() bool
	}:
		return e.Timeout(), true
	case interface {
		IsTimeout() bool
	}:
		return e.IsTimeout(), true
	default:
		return false, false
	}
}

// IsTimeout tests an error for Timeout() and IsTimeout()
// recursively.
func IsTimeout(err error) bool {
	is, _ := IsErrorFn2(CheckIsTimeout, err)
	return is
}
