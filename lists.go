package core

import (
	"container/list"
)

// ListContains checks if a container/list contains an element
func ListContains[T comparable](l *list.List, val T) bool {
	return ListContainsFn(l, val, func(va, vb T) bool {
		return va == vb
	})
}

// ListContainsFn checks if a container/list contains an element
// that satisfies a given function
func ListContainsFn[T any](l *list.List, val T, eq func(T, T) bool) bool {
	var found bool

	if l != nil && eq != nil {
		ListForEach(l, func(v T) bool {
			found = eq(val, v)
			return found
		})
	}
	return found
}

// ListForEach calls a function for each value until told to stop
func ListForEach[T any](l *list.List, fn func(v T) bool) {
	if l == nil || fn == nil {
		return
	}

	ListForEachElement(l, func(e *list.Element) bool {
		if v, ok := e.Value.(T); ok {
			return fn(v)
		}
		return false
	})
}

// ListForEachElement calls a function for each element until told to stop
func ListForEachElement(l *list.List, fn func(*list.Element) bool) {
	if l == nil || fn == nil {
		return
	}

	e, next := listIterStep(l.Front())
	for e != nil {
		if fn(e) {
			break
		}
		e, next = listIterStep(next)
	}
}

func listIterStep(ref *list.Element) (e *list.Element, next *list.Element) {
	if ref != nil {
		next = ref.Next()
	}
	return ref, next
}
