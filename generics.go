package core

// Coalesce returns the first non-zero argument
func Coalesce[T comparable](opts ...T) T {
	var zero T

	for _, v := range opts {
		if v != zero {
			return v
		}
	}

	return zero
}

// revive:disable:flag-parameter

// IIf returns one value or the other depending
// on a condition.
func IIf[T any](cond bool, yes, no T) T {
	// revive:enable:flag-parameter
	if cond {
		return yes
	}
	return no
}
