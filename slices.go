package core

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
func SliceMinusFn[T any](a []T, b []T, eq func(T, T) bool) []T {
	out := make([]T, 0, len(a))

	for _, v := range a {
		if !SliceContainsFn(b, v, eq) {
			out = append(out, v)
		}
	}

	return out
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

// SliceUnique returns a new slice containing only
// unique elements
func SliceUnique[T comparable](a []T) []T {
	keys := make(map[T]bool, len(a))
	list := make([]T, 0, len(a))
	for _, entry := range a {
		if _, known := keys[entry]; !known {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// SliceUniquify returns the same slice, reduced to
// only contain unique elements
func SliceUniquify[T comparable](ptr *[]T) []T {
	if ptr == nil {
		return []T{}
	}

	keys := make(map[T]bool, len(*ptr))
	j := 0
	for i, entry := range *ptr {
		if _, known := keys[entry]; !known {
			keys[entry] = true
			if i != j {
				(*ptr)[j] = entry
			}
			j++
		}
	}

	*ptr = (*ptr)[:j]
	return *ptr
}
