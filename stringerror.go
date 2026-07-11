package core

import "fmt"

// StringError is an error whose message is the string value itself.
//
// Unlike an error from [errors.New], a StringError can be declared as a
// constant, so sentinels become immutable typed constants instead of
// package-level variables:
//
//	const ErrClosed StringError = "already closed"
//
// Values with equal text compare equal, so [errors.Is] matches
// StringError sentinels by value without a dedicated Is method.
type StringError string

// Compile-time checks that StringError is usable as a constant and
// satisfies the error interface.
const _ StringError = "boom"

var _ error = StringError("boom")

// Error returns the message, which is the string value itself.
func (e StringError) Error() string {
	return string(e)
}

// AsError returns the receiver as an error, or nil when its text is
// empty, so a zero StringError stands for "no error". It follows the
// AsError() convention recognised by [AsError].
func (e StringError) AsError() error {
	if e == "" {
		return nil
	}
	return e
}

// OK reports whether the StringError is empty, meaning it represents no
// error. It is the inverse of a non-nil [StringError.AsError] result.
func (e StringError) OK() bool {
	return e == ""
}

// IsZero reports whether the StringError is empty. It lets [IsZero]
// recognise a zero StringError through the IsZero() convention.
func (e StringError) IsZero() bool {
	return e == ""
}

// NewStringError formats its arguments into a [StringError]. Without
// arguments the format string becomes the message verbatim, so a literal
// per-cent sign is left untouched. It returns nil when the message is
// empty, matching [StringError.AsError].
func NewStringError(format string, args ...any) error {
	var s StringError
	if len(args) > 0 {
		s = StringError(fmt.Sprintf(format, args...))
	} else {
		s = StringError(format)
	}
	return s.AsError()
}
