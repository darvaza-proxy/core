package core

import (
	"fmt"
	"io/fs"
	"slices"
)

// Sentinel errors for common conditions. Except for [ErrTODO] and
// [ErrInvalid], these are [StringError] constants: they can't be
// reassigned and match by value.
const (
	// ErrNotImplemented indicates something hasn't been implemented yet
	ErrNotImplemented StringError = "not implemented"
	// ErrExists indicates something already exists. It is deliberately
	// general, distinct from the file-centric [fs.ErrExist].
	ErrExists StringError = "already exists"
	// ErrNotExists indicates something doesn't exist. It is deliberately
	// general, distinct from the file-centric [fs.ErrNotExist].
	ErrNotExists StringError = "does not exist"
	// ErrUnknown indicates something isn't recognized
	ErrUnknown StringError = "unknown"
	// ErrNilReceiver indicates a method was called over a nil instance
	ErrNilReceiver StringError = "nil receiver"
	// ErrUnreachable indicates something impossible happened
	ErrUnreachable StringError = "unreachable"
)

var (
	// ErrTODO is like ErrNotImplemented but used especially to
	// indicate something needs to be implemented
	ErrTODO = Wrap(ErrNotImplemented, "TODO")
	// ErrInvalid indicates an argument isn't valid. It's an alias of
	// [fs.ErrInvalid] so errors.Is matches across the boundary.
	ErrInvalid = fs.ErrInvalid
)

var (
	_ Unwrappable = (*WrappedError)(nil)
)

// Unwrappable represents an error that can be Unwrap() to get the cause
type Unwrappable interface {
	Error() string
	Unwrap() error
}

// Wrap annotates an error with a single string.
func Wrap(err error, msg string) error {
	return doWrap(err, false, "%s", msg)
}

// Wrapf annotates an error with a formatted string.
func Wrapf(err error, format string, args ...any) error {
	return doWrap(err, false, format, args...)
}

// QuietWrap replaces the text of the error it's wrapping.
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

	if note == "" {
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
	if s == "" {
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
	default:
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
	default:
	}

	return SliceReplaceFn(errs, func(_ []error, err error) (error, bool) {
		return err, err != nil
	})
}

// IsError recursively checks if the given error is in the given list,
// or just non-nil if no options to check are given. An error matches
// a target by identity or via the error's own Is(error) bool method —
// the same per-node test [errors.Is] applies — never the target's.
// Recursion covers Unwrap() error, Unwrap() []error and
// Errors() []error layers.
func IsError(err error, errs ...error) bool {
	switch {
	case err == nil:
		return false
	case len(errs) == 0:
		return true
	default:
		return IsErrorFn(NewCheckErrorIsIn(errs), err)
	}
}

// IsErrorFn recursively checks if any of the given errors satisfies
// the specified check function, testing the errors themselves before
// descending into their unwrapped layers.
func IsErrorFn(check func(error) bool, errs ...error) bool {
	switch {
	case check == nil, len(errs) == 0:
		return false
	case checkIsErrorFnShallow(check, errs):
		return true
	case checkIsErrorFnUnwrapped(check, errs):
		return true
	default:
		return false
	}
}

// checkIsErrorFnShallow tests the errors themselves against the
// check function, without unwrapping.
func checkIsErrorFnShallow(check func(error) bool, errs []error) bool {
	for _, e := range errs {
		if e != nil && check(e) {
			return true
		}
	}
	return false
}

// checkIsErrorFnUnwrapped tests the unwrapped layers of each error,
// recursing via IsErrorFn.
func checkIsErrorFnUnwrapped(check func(error) bool, errs []error) bool {
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
// certain answer from the check function, testing the errors
// themselves before descending into their unwrapped layers.
// As opposed to IsErrorFn, IsErrorFn2 will stop when it has certainty
// of a false result.
func IsErrorFn2(check func(error) (bool, bool), errs ...error) (is, known bool) {
	if check == nil || len(errs) == 0 {
		return false, true
	} else if is, known := checkIsErrorFn2Shallow(check, errs); known {
		return is, true
	} else if is, known := checkIsErrorFn2Unwrapped(check, errs); known {
		return is, true
	}

	// unknown
	return false, false
}

// checkIsErrorFn2Shallow tests the errors themselves against the
// check function, without unwrapping, stopping at the first known
// answer.
func checkIsErrorFn2Shallow(check func(error) (bool, bool), errs []error) (is, known bool) {
	for _, e := range errs {
		if e != nil {
			if is, known = check(e); known {
				return is, true
			}
		}
	}

	// unknown
	return false, false
}

// checkIsErrorFn2Unwrapped tests the unwrapped layers of each error,
// recursing via IsErrorFn2, stopping at the first known answer.
func checkIsErrorFn2Unwrapped(check func(error) (bool, bool), errs []error) (is, known bool) {
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

// NewCheckErrorIsIn builds the [IsErrorFn] predicate for IsError: the
// returned closure reports whether a chain node matches any of the
// targets, by identity or the node's own Is(error) bool method.
func NewCheckErrorIsIn(targets []error) func(error) bool {
	return func(err error) bool {
		return slices.ContainsFunc(targets, func(target error) bool {
			is, _ := errorIs(target, err)
			return is
		})
	}
}

// NewCheckErrorIsIn2 is the [IsErrorFn2] counterpart of
// [NewCheckErrorIsIn]: the returned closure reports whether a chain
// node matches any of the targets, and whether that answer is final.
// A match is final; a miss is final only when no target could match
// deeper nodes either — e.g. every target is nil — so IsErrorFn2 can
// stop early without missing a wrapped match.
func NewCheckErrorIsIn2(targets []error) func(error) (bool, bool) {
	return func(err error) (is, known bool) {
		known = true
		for _, target := range targets {
			is2, known2 := errorIs(target, err)
			if is2 {
				return true, true
			}
			known = known && known2
		}
		return false, known
	}
}

// errorIs tests a single node against a target: identity first via
// errorEq, then the node's own Is(error) bool method — the same
// per-node test [errors.Is] applies. Unwrapping is the caller's
// concern. known follows the [IsErrorFn2] convention, reporting
// whether the answer is final for the whole chain: a match is final,
// and so is a nil operand — nil matches nothing deeper either. A miss
// stays unknown because a wrapped error may still match.
func errorIs(target, err error) (is, known bool) {
	if is, known := errorEq(target, err); known {
		return is, true
	}

	// try the Is(error) bool interface, if implemented
	if e, ok := err.(interface{ Is(error) bool }); ok && e.Is(target) {
		return true, true
	}

	// no match here; deeper nodes may still match
	return false, false
}

// errorEq compares two errors by identity. known is true only when
// identity settles the match question: the errors are equal, or one
// is nil — nil only ever matches nil. Inequality stays unknown, as an
// Is method may still match, and so does a panicking comparison of a
// shared non-comparable dynamic type — the case [errors.Is] sidesteps
// via reflection.
func errorEq(a, b error) (is, known bool) {
	// a panicking comparison leaves the zero (false, false)
	defer func() {
		_ = recover()
	}()

	switch {
	case a == nil, b == nil:
		return a == b, true
	case a == b:
		return true, true
	default:
		return false, false
	}
}

// CheckIsTemporary tests an error for temporary conditions without unwrapping.
// It checks if the error implements Temporary() bool or IsTemporary() bool
// interfaces directly, without traversing wrapped error chains.
//
// The function examines the error in the following priority order:
//   - Temporary() bool interface (legacy net.Error style)
//   - IsTemporary() bool interface (modern style)
//   - Falls back to CheckIsTimeout for timeout-based temporary errors
//
// Returns:
//   - is: true if the error indicates a temporary condition
//   - known: true if the error type implements a recognized interface
//
// For nil errors, returns (false, true) indicating definitively not temporary.
// For errors with no recognized interface, returns result from CheckIsTimeout.
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

// IsTemporary tests an error chain for temporary conditions recursively.
// It traverses wrapped error chains using IsErrorFn2 to find any error
// that implements temporary condition interfaces.
//
// This function provides comprehensive temporary error detection by:
//   - Checking each error in the unwrapping chain via CheckIsTemporary
//   - Following both Unwrap() error and Unwrap() []error patterns
//   - Detecting legacy net.Error.Temporary() implementations
//   - Detecting modern IsTemporary() bool implementations
//   - Detecting timeout errors (which are also considered temporary)
//
// Returns true if any error in the chain indicates a temporary condition.
// Returns false for nil errors or chains with no temporary indicators.
//
// Use this function when you need to determine if an operation should be
// retried based on the error's temporary nature.
func IsTemporary(err error) bool {
	is, _ := IsErrorFn2(CheckIsTemporary, err)
	return is
}

// CheckIsTimeout tests an error for timeout conditions without unwrapping.
// It checks if the error implements Timeout() bool or IsTimeout() bool
// interfaces directly, without traversing wrapped error chains.
//
// The function examines the error in the following priority order:
//   - Timeout() bool interface (legacy net.Error style)
//   - IsTimeout() bool interface (modern style)
//
// Returns:
//   - is: true if the error indicates a timeout condition
//   - known: true if the error type implements a recognized timeout interface
//
// For nil errors, returns (false, true) indicating definitively not a timeout.
// For errors with no recognized timeout interface, returns (false, false).
//
// Note that timeout errors are typically also considered temporary conditions,
// but this function specifically tests for timeout semantics only.
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

// IsTimeout tests an error chain for timeout conditions recursively.
// It traverses wrapped error chains using IsErrorFn2 to find any error
// that implements timeout condition interfaces.
//
// This function provides comprehensive timeout error detection by:
//   - Checking each error in the unwrapping chain via CheckIsTimeout
//   - Following both Unwrap() error and Unwrap() []error patterns
//   - Detecting legacy net.Error.Timeout() implementations
//   - Detecting modern IsTimeout() bool implementations
//
// Returns true if any error in the chain indicates a timeout condition.
// Returns false for nil errors or chains with no timeout indicators.
//
// Use this function when you need to distinguish timeout errors from
// other types of temporary errors for specialized retry logic or
// timeout-specific error handling.
func IsTimeout(err error) bool {
	is, _ := IsErrorFn2(CheckIsTimeout, err)
	return is
}
