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
