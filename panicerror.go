package core

import (
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
	return &PanicError{
		payload: payload,
		stack:   StackTrace(skip + 1),
	}
}

// NewPanicErrorf creates a new PanicError with a formated string as payload
func NewPanicErrorf(skip int, format string, args ...any) *PanicError {
	return &PanicError{
		payload: fmt.Errorf(format, args...),
		stack:   StackTrace(skip + 1),
	}
}

// NewPanicWrap creates a new PanicError wrapping a given error
// annotated with a string message
func NewPanicWrap(skip int, err error, msg string) *PanicError {
	return &PanicError{
		payload: Wrapf(err, msg),
		stack:   StackTrace(skip + 1),
	}
}

// NewPanicWrapf creates a new PanicError wrapping a given error
// annotated with a formated string
func NewPanicWrapf(skip int, err error, format string, args ...any) *PanicError {
	return &PanicError{
		payload: Wrapf(err, format, args...),
		stack:   StackTrace(skip + 1),
	}
}

// Panic emits a PanicError with the given payload
func Panic(payload any) {
	panic(NewPanicError(1, payload))
}

// Panicf emits a PanicError with a formated string as payload
func Panicf(format string, args ...any) {
	panic(NewPanicErrorf(1, format, args...))
}

// PanicWrap emits a PanicError wrapping an annotated error
func PanicWrap(err error, msg string) {
	panic(NewPanicWrap(1, err, msg))
}

// PanicWrapf emits a PanicError wrapping an error annotated with a formated string
func PanicWrapf(err error, format string, args ...any) {
	panic(NewPanicWrapf(1, err, format, args...))
}
