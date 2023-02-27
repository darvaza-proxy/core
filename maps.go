package core

import "container/list"

// MapContains tells if a given map contains a key.
// this helper is inteded for switch/case conditions
func MapContains[K comparable](m map[K]any, key K) bool {
	_, ok := m[key]
	return ok
}

// MapListContains checks if the list.List on a map contains an element
func MapListContains[K comparable, T comparable](m map[K]*list.List, key K, v T) bool {
	return MapListContainsFn(m, key, func(v1 T) bool {
		return v == v1
	})
}

// MapListContainsFn checks if the list.List on a map contains an element using a match functions
func MapListContainsFn[K comparable, T any](m map[K]*list.List, key K, match func(T) bool) bool {
	if l, ok := m[key]; ok {
		return ListContainsFn(l, match)
	}
	return false
}

// MapListForEach calls a function for each value on a map entry until told to stop
func MapListForEach[K comparable, T any](m map[K]*list.List, key K, fn func(v T) bool) {
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
	if m == nil || fn == nil {
		return
	}

	if l, ok := m[key]; ok {
		ListForEachElement(l, fn)
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
