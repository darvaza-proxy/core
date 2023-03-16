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
