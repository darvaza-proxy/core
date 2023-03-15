package core

import (
	"sync/atomic"
)

// Recovered is an error caught from a panic call
type Recovered interface {
	Error() string
	Recovered() any
}

// AsRecovered receives the value from recover()
// and wraps it as a Recovered error
func AsRecovered(rvr any) Recovered {
	if rvr == nil {
		// no panic
		return nil
	} else if p, ok := rvr.(Recovered); ok {
		// pass-through
		return p
	} else {
		// wrap it
		return NewPanicError(2, rvr)
	}
}

// Catcher is a runner that catches panics
type Catcher struct {
	recovered atomic.Value
}

// Do calls a function, returning its organic error,
// or the caught panic
func (p *Catcher) Do(fn func() error) error {
	if err := p.Try(fn); err != nil {
		// natural death
		return err
	}

	if err := p.Recovered(); err != nil {
		// recovered panic
		return err
	}

	// all good
	return nil
}

// Try calls a function, returning its organic error,
// or storing the recovered error for later consumption
func (p *Catcher) Try(fn func() error) error {
	if fn != nil {
		defer func() {
			if err := AsRecovered(recover()); err != nil {
				p.recovered.CompareAndSwap(nil, err)
			}
		}()

		return fn()
	}
	return nil
}

// Recovered returns the error corresponding to a
// panic when the Catcher was running a function
func (p *Catcher) Recovered() Recovered {
	if err, ok := p.recovered.Load().(Recovered); ok {
		return err
	}
	return nil
}
