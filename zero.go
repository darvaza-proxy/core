package core

import "reflect"

// Zero returns the zero value of a type
// for which we got a pointer.
func Zero[T any](_ *T) T {
	var zero T
	return zero
}

// IsZero checks if a non-zero value has been set
// either by using the `IsZero() bool“ interface
// or reflection.
// nil and (*T)(nil) are considered to be zero.
func IsZero(vi any) bool {
	switch p := vi.(type) {
	case nil:
		// nil
		return true
	case interface {
		IsZero() bool
	}:
		// interface
		return p.IsZero()
	default:
		// reflection
		v := reflect.ValueOf(vi)
		if v.IsValid() {
			return v.IsZero()
		}

		// nil
		return true
	}
}
