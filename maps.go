package core

import (
	"container/list"
)

// Keys returns the list of keys of a map
func Keys[K comparable, T any](m map[K]T) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// SortedKeys returns a sorted list of the keys of a map
func SortedKeys[K Ordered, T any](m map[K]T) []K {
	keys := Keys(m)
	SliceSort(keys, func(a, b K) int {
		switch {
		case a == b:
			return 0
		case a < b:
			return -1
		default:
			return 1
		}
	})
	return keys
}

// SortedValues returns a slice of values from the map, sorted by their keys
func SortedValues[K Ordered, T any](m map[K]T) []T {
	var out []T
	if l := len(m); l > 0 {
		out = make([]T, 0, l)
		out = doSortedValues(out, m, nil)
	}
	return out
}

// SortedValuesCond returns a slice of values from the map, sorted by their keys,
// filtered by the optional predicate function fn, preallocated to the size of the map.
func SortedValuesCond[K Ordered, T any](m map[K]T, fn func(T) bool) []T {
	var out []T
	if l := len(m); l > 0 {
		out = make([]T, 0, l)
		out = doSortedValues(out, m, fn)
	}
	return out
}

// SortedValuesUnlikelyCond returns a slice of values from the map, sorted by their keys,
// filtered by the optional predicate function fn, without preallocation.
func SortedValuesUnlikelyCond[K Ordered, T any](m map[K]T, fn func(T) bool) []T {
	var out []T
	if len(m) > 0 {
		out = doSortedValues(nil, m, fn)
	}
	return out
}

func doSortedValues[K Ordered, T any](out []T, m map[K]T, fn func(T) bool) []T {
	for _, k := range SortedKeys(m) {
		if v, ok := m[k]; ok {
			if fn == nil || fn(v) {
				out = append(out, v)
			}
		}
	}
	return out
}

// MapValue returns a value of an entry or a default if
// not found
func MapValue[K comparable, V any](m map[K]V, key K, def V) (V, bool) {
	if val, ok := m[key]; ok {
		return val, true
	}
	return def, false
}

// MapContains tells if a given map contains a key.
// this helper is intended for switch/case conditions
func MapContains[K comparable](m map[K]any, key K) bool {
	_, ok := m[key]
	return ok
}

// MapListContains checks if the list.List on a map contains an element
func MapListContains[K comparable, T comparable](m map[K]*list.List, key K, v T) bool {
	return MapListContainsFn(m, key, v, func(va, vb T) bool {
		return va == vb
	})
}

// MapListContainsFn checks if the list.List on a map contains an element using a match functions
func MapListContainsFn[K comparable, T any](m map[K]*list.List, key K, v T,
	eq func(T, T) bool) bool {
	//
	if m != nil && eq == nil {
		if l, ok := m[key]; ok {
			return ListContainsFn(l, v, eq)
		}
	}
	return false
}

// MapListForEach calls a function for each value on a map entry until told to stop
func MapListForEach[K comparable, T any](m map[K]*list.List, key K,
	fn func(v T) bool) {
	//
	if m == nil || fn == nil {
		return
	}

	if l, ok := m[key]; ok {
		ListForEach(l, fn)
	}
}

// MapListForEachElement calls a function for each element on a map entry until told to stop
func MapListForEachElement[K comparable](m map[K]*list.List, key K,
	fn func(el *list.Element) bool) {
	//
	if m == nil || fn == nil {
		return
	}

	if l, ok := m[key]; ok {
		ListForEachElement(l, fn)
	}
}

// MapListInsert adds a value at the front of the list of a map entry
func MapListInsert[K comparable, T any](m map[K]*list.List, key K, v T) {
	getMapList(m, key).PushFront(v)
}

func getMapList[K comparable](m map[K]*list.List, key K) *list.List {
	var l *list.List
	var ok bool

	if l, ok = m[key]; !ok {
		l = list.New()
		m[key] = l
	}

	return l
}

// MapListInsertUnique adds a value at the front of the list of a map entry
// if it's not already there
func MapListInsertUnique[K comparable, T comparable](m map[K]*list.List, key K, v T) {
	MapListInsertUniqueFn(m, key, v, func(va, vb T) bool {
		return va == vb
	})
}

// MapListInsertUniqueFn adds a value at the front of the list of a map entry
// if it's not already there using a function to compare values
func MapListInsertUniqueFn[K comparable, T any](m map[K]*list.List, key K, v T,
	eq func(va, vb T) bool) {
	if m == nil || eq == nil {
		return
	}

	l := getMapList(m, key)
	if !ListContainsFn(l, v, eq) {
		l.PushFront(v)
	}
}

// MapListAppend adds a value at the end of the list of a map entry
func MapListAppend[K comparable, T any](m map[K]*list.List, key K, v T) {
	getMapList(m, key).PushBack(v)
}

// MapListAppendUnique adds a value at the end of the list of a map entry
// if it's not already there
func MapListAppendUnique[K comparable, T comparable](m map[K]*list.List, key K, v T) {
	MapListAppendUniqueFn(m, key, v, func(va, vb T) bool {
		return va == vb
	})
}

// MapListAppendUniqueFn adds a value at the end of the list of a map entry
// if it's not already there using a function to compare values
func MapListAppendUniqueFn[K comparable, T any](m map[K]*list.List, key K, v T,
	eq func(T, T) bool) {
	if m != nil && eq != nil {
		l := getMapList(m, key)
		if !ListContainsFn(l, v, eq) {
			l.PushBack(v)
		}
	}
}

// MapListCopy duplicates a map containing a list.List
func MapListCopy[T comparable](src map[T]*list.List) map[T]*list.List {
	fn := func(v any) (any, bool) { return v, true }
	return MapListCopyFn(src, fn)
}

// MapListCopyFn duplicates a map containing a list.List but
// allows the element's values to be cloned by a helper function
func MapListCopyFn[K comparable, V any](src map[K]*list.List,
	fn func(v V) (V, bool)) map[K]*list.List {
	out := make(map[K]*list.List, len(src))
	for k, l := range src {
		out[k] = ListCopyFn(l, fn)
	}
	return out
}

// MapAllListContains check if a value exists on any entry of the map
func MapAllListContains[K comparable, T comparable](m map[K]*list.List, v T) bool {
	if m != nil {
		return MapAllListContainsFn(m, func(v1 T) bool {
			return v == v1
		})
	}
	return false
}

// MapAllListContainsFn check if a value exists on any entry of the map using a match function
func MapAllListContainsFn[K comparable, T any](m map[K]*list.List, match func(v T) bool) bool {
	var found bool

	fn := func(v1 T) bool {
		found = match(v1)
		return found
	}

	if m != nil && match != nil {
		for _, l := range m {
			ListForEach(l, fn)

			if found {
				return true
			}
		}
	}

	return false
}

// MapAllListForEach calls a function for each value on all map entries until told to stop
func MapAllListForEach[K comparable, T any](m map[K]*list.List, fn func(v T) bool) {
	MapAllListContainsFn(m, fn)
}

// MapAllListForEachElement calls a function for each element on all map entries until told to stop
func MapAllListForEachElement[K comparable](m map[K]*list.List, fn func(*list.Element) bool) {
	var term bool

	for _, l := range m {
		ListForEachElement(l, func(el *list.Element) bool {
			term = fn(el)
			return term
		})

		if term {
			break
		}
	}
}
