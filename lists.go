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

// ListForEachBackward calls a function for each value until told to stop
func ListForEachBackward[T any](l *list.List, fn func(v T) bool) {
	if l == nil || fn == nil {
		return
	}

	ListForEachBackwardElement(l, func(e *list.Element) bool {
		if v, ok := e.Value.(T); ok {
			return fn(v)
		}
		return false
	})
}

// ListForEachBackwardElement calls a function for each element until told to stop
func ListForEachBackwardElement(l *list.List, fn func(*list.Element) bool) {
	if l == nil || fn == nil {
		return
	}

	e, prev := listIterBackwardStep(l.Back())
	for e != nil {
		if fn(e) {
			break
		}
		e, prev = listIterBackwardStep(prev)
	}
}

func listIterBackwardStep(ref *list.Element) (e *list.Element, prev *list.Element) {
	if ref != nil {
		prev = ref.Prev()
	}
	return ref, prev
}

// ListCopy makes a shallow copy of a list
func ListCopy[T any](src *list.List) *list.List {
	return ListCopyFn[T](src, nil)
}

// ListCopyFn makes a copy of a list using the given helper
func ListCopyFn[T any](src *list.List, fn func(v T) (T, bool)) *list.List {
	if fn == nil {
		fn = func(v T) (T, bool) {
			return v, true
		}
	}

	out := list.New()
	ListForEach(src, func(v0 T) bool {
		if v1, ok := fn(v0); ok {
			out.PushBack(v1)
		}
		return false
	})
	return out
}
