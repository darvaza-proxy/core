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
