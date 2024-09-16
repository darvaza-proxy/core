package core

import (
	"crypto/rand"
	"math/big"
	"sort"
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
	if fn == nil {
		// NO-OP
		return s
	}

	j := 0
	for _, v := range s {
		if w, ok := fn(s[:j], v); ok {
			s[j] = w
			j++
		}
	}
	return s[:j]
}

// SliceCopyFn makes a copy of a slice, optionally modifying in-flight
// the items using a function. If no function is provided,
// the destination will be a shallow copy of the source slice.
func SliceCopyFn[T any](s []T,
	fn func(partial []T, before T) (after T, include bool),
) []T {
	//
	if fn == nil {
		return SliceCopy(s)
	}

	result := make([]T, 0, len(s))
	for _, v := range s {
		if w, ok := fn(result, v); ok {
			result = append(result, w)
		}
	}

	return result
}

// SliceCopy makes a shallow copy of a given slice
func SliceCopy[T any](s []T) []T {
	l := len(s)
	result := make([]T, l)
	if l > 0 {
		copy(result, s)
	}
	return result
}

// SliceMap takes a []T1 and uses a function to produce a []T2
// by processing each item on the source slice.
func SliceMap[T1 any, T2 any](a []T1,
	fn func(partial []T2, v T1) (newEntries []T2)) []T2 {
	//
	var result []T2
	if fn != nil {
		for _, v := range a {
			result = append(result, fn(result, v)...)
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

// SliceSortFn sorts the slice x in ascending order as a less function.
// This sort is not guaranteed to be stable.
// less(a, b) should true when a < b
func SliceSortFn[T any](x []T, less func(a, b T) bool) {
	if less != nil && len(x) > 0 {
		doSliceSort(x, less)
	}
}

// SliceSort sorts the slice x in ascending order as determined by the cmp
// function. This sort is not guaranteed to be stable.
// cmp(a, b) should return a negative number when a < b, a positive number when
// a > b and zero when a == b.
func SliceSort[T any](x []T, cmp func(a, b T) int) {
	if cmp != nil && len(x) > 0 {
		doSliceSort(x, func(a, b T) bool {
			return cmp(a, b) < 0
		})
	}
}

// SliceSortOrdered sorts the slice x of an [Ordered] type in ascending order.
func SliceSortOrdered[T Ordered](x []T) {
	if len(x) > 0 {
		doSliceSort(x, func(a, b T) bool {
			return a < b
		})
	}
}

func doSliceSort[T any](x []T, less func(a, b T) bool) {
	s := sortable[T]{
		x:    x,
		less: less,
	}

	sort.Sort(s)
}

var _ sort.Interface = sortable[any]{}

type sortable[T any] struct {
	x    []T
	less func(a, b T) bool
}

func (s sortable[T]) Len() int {
	return len(s.x)
}

func (s sortable[T]) Less(i, j int) bool {
	// this is only accessible from sort.Sort() so
	// we can trust the indexes
	return s.less(s.x[i], s.x[j])
}

func (s sortable[T]) Swap(i, j int) {
	// this is only accessible from sort.Sort() so
	// we can trust the indexes
	s.x[j], s.x[i] = s.x[i], s.x[j]
}
