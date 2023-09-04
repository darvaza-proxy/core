package core

import "reflect"

// Zero returns the zero value of a type
// for which we got a pointer.
func Zero[T any](_ *T) T {
	var zero T
	return zero
}

// IsZero checks if a given value is zero, either using
// the IsZero() bool interface or reflection
func IsZero(vi any) bool {
	if p, ok := vi.(interface {
		IsZero() bool
	}); ok {
		return p.IsZero()
	}

	return reflect.ValueOf(vi).IsZero()
}
