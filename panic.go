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
	}

	if p, ok := rvr.(Recovered); ok {
		// pass-through
		return p
	}

	// wrap it
	return NewPanicError(2, rvr)
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

// Catch uses a [Catcher] to safely call a function and
// return the organic error or the [Recovered] [PanicError].
func Catch(fn func() error) error {
	var p Catcher
	return p.Do(fn)
}

// Must panics if err is not nil, otherwise returns value.
// This is useful for situations where errors should never occur, such as
// test setup or configuration loading. It follows the common Go pattern
// of Must* functions that panic on error. The panic includes proper stack
// traces pointing to the caller.
//
// Example usage:
//
//	config := Must(loadConfig("config.json"))  // panics if loadConfig returns error
//	conn := Must(net.Dial("tcp", "localhost:8080"))  // panics if dial fails
//	data := Must(json.Marshal(obj))  // panics if marshal fails
func Must[V any](value V, err error) V {
	if err != nil {
		panic(NewPanicWrap(1, err, "core.Must"))
	}
	return value
}

// Maybe returns the value, ignoring any error.
// This is useful when you want to proceed with a default or zero value
// regardless of whether an error occurred. Unlike Must, it never panics.
//
// Example usage:
//
//	// Use empty string if ReadFile fails
//	content := Maybe(os.ReadFile("optional.txt"))
//
//	// Use zero value if parsing fails
//	count := Maybe(strconv.Atoi(userInput))
//
//	// Chain operations where errors are non-critical
//	data := Maybe(json.Marshal(obj))
func Maybe[V any](value V, _ error) V {
	return value
}

// MustOK panics if ok is false, otherwise returns value.
// This is useful for situations where operations should always succeed,
// such as accessing map values that are known to exist or type assertions
// that are guaranteed to be valid. It follows the common Go pattern
// of Must* functions that panic on failure. The panic includes proper stack
// traces pointing to the caller.
//
// Example usage:
//
//	value := MustOK(MapValue(m, "key", 0))  // panics if key doesn't exist in map
//	str := MustOK(As[any, string](v))  // panics if v is not a string
//	result := MustOK(someFunc())  // panics if someFunc returns false for ok
//
// revive:disable-next-line:flag-parameter
func MustOK[V any](value V, ok bool) V {
	if !ok {
		panic(NewPanicError(1, "core.MustOK: operation failed"))
	}
	return value
}

// MaybeOK returns the value, ignoring the ok flag.
// This is useful when you want to proceed with a default or zero value
// regardless of whether the operation succeeded. Unlike MustOK, it never panics.
//
// Example usage:
//
//	// Use zero value if key doesn't exist in map
//	value := MaybeOK(MapValue(m, "key", 0))
//
//	// Use zero value if type assertion fails
//	str := MaybeOK(As[any, string](v))
//
//	// Chain operations where success is non-critical
//	result := MaybeOK(someFunc())
func MaybeOK[V any](value V, _ bool) V {
	return value
}
