package core

import "fmt"

// Recovered is an error caught from a panic call
type Recovered interface {
	Error() string
	Recovered() any
}

// Recover attempts to recover from a panic
func Recover() Recovered {
	// revive:disable:defer
	if rvr := recover(); rvr == nil {
		// no panic
		return nil
	} else if p, ok := rvr.(Recovered); ok {
		// pass-through
		return p
	} else {
		// wrap it
		return &panicError{rvr}
	}
	// revive:enable:defer
}

type panicError struct {
	payload any
}

func (p *panicError) Error() string {
	return fmt.Sprintf("panic: %s", p.payload)
}

func (p *panicError) Unwrap() error {
	if err, ok := p.payload.(error); ok {
		return err
	}
	return nil
}

func (p *panicError) Recovered() any {
	return p.payload
}
