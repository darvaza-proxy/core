package core

import (
	"errors"
	"fmt"
)

var (
	_ Recovered   = (*PanicError)(nil)
	_ Unwrappable = (*PanicError)(nil)
	_ CallStacker = (*PanicError)(nil)
)

// PanicError is an error to be sent via panic, ideally
// to be caught using slog.Recover()
type PanicError struct {
	payload any
	stack   Stack
}

// Error returns the payload as a string
func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", p.payload)
}

// Unwrap returns the payload if it's and error
func (p *PanicError) Unwrap() error {
	if err, ok := p.payload.(error); ok {
		return err
	}
	return nil
}

// Recovered returns the payload of the panic
func (p *PanicError) Recovered() any {
	return p.payload
}

// CallStack returns the call stack associated to this panic() event
func (p *PanicError) CallStack() Stack {
	return p.stack
}

// NewPanicError creates a new PanicError with arbitrary payload
func NewPanicError(skip int, payload any) *PanicError {
	if s, ok := payload.(string); ok {
		payload = errors.New(s)
	}
	return &PanicError{
		payload: payload,
		stack:   StackTrace(deeper(skip)),
	}
}

// NewPanicErrorf creates a new PanicError annotated with
// a string, optionally formatted. %w is expanded.
func NewPanicErrorf(skip int, format string, args ...any) *PanicError {
	var payload error
	if len(args) > 0 {
		payload = fmt.Errorf(format, args...)
	} else {
		payload = errors.New(format)
	}

	return &PanicError{
		payload: payload,
		stack:   StackTrace(deeper(skip)),
	}
}

// NewPanicWrap creates a new PanicError wrapping a given error
// annotated with a single string.
func NewPanicWrap(skip int, err error, note string) *PanicError {
	return &PanicError{
		payload: Wrap(err, note),
		stack:   StackTrace(deeper(skip)),
	}
}

// NewPanicWrapf creates a new PanicError wrapping a given error
// annotated with a formatted string.
func NewPanicWrapf(skip int, err error, format string, args ...any) *PanicError {
	return &PanicError{
		payload: Wrapf(err, format, args...),
		stack:   StackTrace(deeper(skip)),
	}
}

// Panic emits a PanicError with the given payload. A *PanicError payload
// is raised as it is, keeping the stack it already carries, and a nil one
// counts as no payload.
func Panic(payload any) {
	PanicFrom(1, payload)
}

// PanicFrom emits a PanicError with the given payload, as [Panic] does,
// its stack starting skip frames above the caller: 0 is PanicFrom's own
// caller. A *PanicError payload keeps the stack it already carries,
// whatever the skip.
func PanicFrom(skip int, payload any) {
	pe, ok := payload.(*PanicError)
	if pe == nil {
		if ok {
			// a nil *PanicError carries nothing
			payload = nil
		}
		pe = NewPanicError(deeper(skip), payload)
	}
	panic(pe)
}

// Panicf emits a PanicError with a formatted string as payload
func Panicf(format string, args ...any) {
	PanicfFrom(1, format, args...)
}

// PanicfFrom emits a PanicError with a formatted string as payload, its
// stack starting skip frames above the caller: 0 is PanicfFrom's own
// caller.
func PanicfFrom(skip int, format string, args ...any) {
	panic(NewPanicErrorf(deeper(skip), format, args...))
}

// PanicWrap emits a PanicError wrapping an annotated error.
func PanicWrap(err error, note string) {
	panic(NewPanicWrap(1, err, note))
}

// PanicWrapf emits a PanicError wrapping an annotated error using
// a formatting string.
func PanicWrapf(err error, format string, args ...any) {
	panic(NewPanicWrapf(1, err, format, args...))
}

// NewUnreachableErrorf creates a new annotated ErrUnreachable with callstack.
func NewUnreachableErrorf(skip int, err error, format string, args ...any) error {
	return NewUnreachableError(deeper(skip), err, fmt.Sprintf(format, args...))
}

// NewUnreachableError creates a new annotated ErrUnreachable with callstack.
func NewUnreachableError(skip int, err error, note string) error {
	switch err {
	case nil, ErrUnreachable:
		err = ErrUnreachable
	default:
		err = QuietWrap(NewCompoundError(ErrUnreachable, err),
			"%s: %s", ErrUnreachable, err)
	}

	if note == "" {
		return NewPanicError(deeper(skip), err)
	}
	return NewPanicWrap(deeper(skip), err, note)
}

// deeper accounts for the frame of the function calling it, so a skip of
// 0 attributes to that function's own caller. A negative skip clamps to
// 1, that same caller, instead of reaching [StackTrace], which rejects it
// and captures nothing.
func deeper(skip int) int {
	if skip < 0 {
		return 1
	}
	return skip + 1
}
