package core

// As returns a value cast to a different type
func As[T, V any](v T) (V, bool) {
	x, ok := any(v).(V)
	return x, ok
}

// AsFn returns a value cast to a different type using a helper function
func AsFn[T, V any](fn func(T) (V, bool), v T) (V, bool) {
	if fn == nil {
		var zero V
		return zero, false
	}

	return fn(v)
}

// SliceAs returns a subset of a slice that could be cast into a different type
func SliceAs[T, V any](vv []T) []V {
	return SliceAsFn(As[T, V], vv)
}

// SliceAsFn returns a subset of a slice that could be cast into a different type,
// using a helper function.
func SliceAsFn[T, V any](fn func(T) (V, bool), vv []T) []V {
	if fn == nil || len(vv) == 0 {
		return nil
	}

	out := make([]V, 0, len(vv))
	for _, v := range vv {
		if x, ok := fn(v); ok {
			out = append(out, x)
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}
