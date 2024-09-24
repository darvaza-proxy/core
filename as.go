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

// AsError attempts to convert a value to an error,
// checking different interfaces.
//
// * AsError() error
// * Error() string
// * OK() bool
// * IsZero() bool
func AsError[T any](v T) error {
	switch x := any(v).(type) {
	case interface {
		AsError() error
	}:
		return x.AsError()
	case interface {
		Error() string
		OK() bool
	}:
		if !x.OK() {
			return x
		}
		return nil
	case error:
		if IsZero(x) {
			return nil
		}
		return x
	default:
		return nil
	}
}

// AsErrors uses AsError to return the subset of
// the elements that are errors.
func AsErrors[T any](vv []T) []error {
	return SliceAsFn(func(v T) (error, bool) {
		err := AsError[T](v)
		return err, err != nil
	}, vv)
}
