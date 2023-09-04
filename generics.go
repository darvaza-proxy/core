package core

// Coalesce returns the first non-zero argument
func Coalesce[T any](opts ...T) T {
	for _, v := range opts {
		if !IsZero(v) {
			return v
		}
	}

	return Zero[T](nil)
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
