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
	return fmt.Sprintf("panic: %s", p.payload)
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
		stack:   StackTrace(skip + 1),
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
		stack:   StackTrace(skip + 1),
	}
}

// NewPanicWrap creates a new PanicError wrapping a given error
// annotated, optionally formatted. %w is expanded.
func NewPanicWrap(skip int, err error, format string, args ...any) *PanicError {
	return &PanicError{
		payload: Wrapf(err, format, args...),
		stack:   StackTrace(skip + 1),
	}
}

// Panic emits a PanicError with the given payload
func Panic(payload any) {
	panic(NewPanicError(1, payload))
}

// Panicf emits a PanicError with a formatted string as payload
func Panicf(format string, args ...any) {
	panic(NewPanicErrorf(1, format, args...))
}

// PanicWrap emits a PanicError wrapping an annotated error, optionally
// formatted. %w is expanded.
func PanicWrap(err error, format string, args ...any) {
	panic(NewPanicWrap(1, err, format, args...))
}
