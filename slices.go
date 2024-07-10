package core

import (
	"crypto/rand"
	"math/big"
)

// SliceMinus returns a new slice containing only the
// elements of one slice not present on the second
func SliceMinus[T comparable](a []T, b []T) []T {
	return SliceMinusFn(a, b, func(va, vb T) bool {
		return va == vb
	})
}

// SliceMinusFn returns a new slice containing only elements
// of slice A that aren't on slice B according to the callback
// eq
func SliceMinusFn[T any](a, b []T, eq func(T, T) bool) []T {
	fn := func(_ []T, v T) (T, bool) {
		if SliceContainsFn(b, v, eq) {
			return v, false // skip
		}

		return v, true // keep
	}

	return SliceCopyFn(a, fn)
}

// SliceContains tells if a slice contains a given element
func SliceContains[T comparable](a []T, v T) bool {
	return SliceContainsFn(a, v, func(va, vb T) bool {
		return va == vb
	})
}

// SliceContainsFn tells if a slice contains a given element
// according to the callback eq
func SliceContainsFn[T any](a []T, v T, eq func(T, T) bool) bool {
	for _, va := range a {
		if eq(va, v) {
			return true
		}
	}
	return false
}

// SliceEqual tells if two slices are equal.
func SliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	return SliceEqualFn(a, b, func(va, vb T) bool {
		return va == vb
	})
}

// SliceEqualFn tells if two slices are equal using a comparing helper.
func SliceEqualFn[T any](a, b []T, eq func(va, vb T) bool) bool {
	if len(a) != len(b) || eq == nil {
		return false
	}

	for i := range a {
		if !eq(a[i], b[i]) {
			return false
		}
	}

	return true
}

// SliceUnique returns a new slice containing only
// unique elements
func SliceUnique[T comparable](a []T) []T {
	keys := make(map[T]bool, len(a))

	// keep only new elements
	fn := func(_ []T, entry T) (T, bool) {
		var keep bool
		if _, known := keys[entry]; !known {
			keys[entry] = true
			keep = true
		}
		return entry, keep
	}

	return SliceCopyFn(a, fn)
}

// SliceUniqueFn returns a new slice containing only
// unique elements according to the callback eq
func SliceUniqueFn[T any](a []T, eq func(T, T) bool) []T {
	// keep only elements not present on the partial
	// result already
	fn := func(partial []T, entry T) (T, bool) {
		var keep bool

		if !SliceContainsFn(partial, entry, eq) {
			keep = true
		}

		return entry, keep
	}

	return SliceCopyFn(a, fn)
}

// SliceUniquify returns the same slice, reduced to
// only contain unique elements
func SliceUniquify[T comparable](ptr *[]T) []T {
	if ptr == nil {
		return []T{}
	}

	keys := make(map[T]bool, len(*ptr))

	// keep only new elements
	fn := func(_ []T, entry T) (T, bool) {
		var keep bool
		if _, known := keys[entry]; !known {
			keys[entry] = true
			keep = true
		}
		return entry, keep
	}

	*ptr = SliceReplaceFn(*ptr, fn)
	return *ptr
}

// SliceUniquifyFn returns the same slice, reduced to
// only contain unique elements according to the callback eq
func SliceUniquifyFn[T any](ptr *[]T, eq func(T, T) bool) []T {
	if ptr == nil {
		return []T{}
	}

	// keep only elements not present on the partial
	// result already
	fn := func(partial []T, entry T) (T, bool) {
		var keep bool

		if !SliceContainsFn(partial, entry, eq) {
			keep = true
		}
		return entry, keep
	}

	*ptr = SliceReplaceFn(*ptr, fn)
	return *ptr
}

// SliceReplaceFn replaces or skips entries in a slice
func SliceReplaceFn[T any](s []T,
	fn func(partial []T, before T) (after T, replace bool),
) []T {
	//
	j := 0
	for _, v := range s {
		if w, ok := fn(s[:j], v); ok {
			s[j] = w
			j++
		}
	}
	return s[:j]
}

// SliceCopyFn conditionally copies a slice allowing
// modifications of the items
func SliceCopyFn[T any](s []T,
	fn func(partial []T, before T) (after T, replace bool),
) []T {
	//
	result := make([]T, 0, len(s))
	for _, v := range s {
		if w, ok := fn(result, v); ok {
			result = append(result, w)
		}
	}

	return result
}

// SliceRandom returns a random element from a slice
// if the slice is empty it will return the zero value
// of the slice type and false
func SliceRandom[T any](a []T) (T, bool) {
	var result T

	switch len(a) {
	case 0:
		return result, false
	case 1:
		result = a[0]
	default:
		id, _ := rand.Int(rand.Reader, big.NewInt(int64(len(a))))
		result = a[id.Uint64()]
	}
	return result, true
}
