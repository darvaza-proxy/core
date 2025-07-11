package core

import (
	"container/list"
	"testing"
)

func testListIteration(t *testing.T, name string, iterFn func(*list.List, func(int) bool), values, expected []int) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		l := list.New()
		for _, v := range values {
			l.PushBack(v)
		}

		var result []int
		iterFn(l, func(v int) bool {
			result = append(result, v)
			return false
		})

		assertIntSlicesEqual(t, expected, result)
	})
}

func testListElementIteration(
	t *testing.T,
	name string,
	iterFn func(*list.List, func(*list.Element) bool),
	values, expected []int,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		l := list.New()
		for _, v := range values {
			l.PushBack(v)
		}

		var result []int
		iterFn(l, func(e *list.Element) bool {
			result = append(result, e.Value.(int))
			return false
		})

		assertIntSlicesEqual(t, expected, result)
	})
}

func TestListForEach(t *testing.T) {
	testListIteration(t, "empty list", ListForEach[int], []int{}, []int{})
	testListIteration(t, "single element", ListForEach[int], []int{1}, []int{1})
	testListIteration(t, "multiple elements", ListForEach[int], []int{1, 2, 3}, []int{1, 2, 3})
}

func TestListForEachElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachElement, []int{}, []int{})
	testListElementIteration(t, "single element", ListForEachElement, []int{1}, []int{1})
	testListElementIteration(t, "multiple elements", ListForEachElement, []int{1, 2, 3}, []int{1, 2, 3})
}

func TestListForEachBackward(t *testing.T) {
	testListIteration(t, "empty list", ListForEachBackward[int], []int{}, []int{})
	testListIteration(t, "single element", ListForEachBackward[int], []int{1}, []int{1})
	testListIteration(t, "multiple elements", ListForEachBackward[int], []int{1, 2, 3}, []int{3, 2, 1})
}

func TestListForEachBackwardElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachBackwardElement, []int{}, []int{})
	testListElementIteration(t, "single element", ListForEachBackwardElement, []int{1}, []int{1})
	testListElementIteration(t, "multiple elements", ListForEachBackwardElement, []int{1, 2, 3}, []int{3, 2, 1})
}

func assertIntSlicesEqual(t *testing.T, expected, actual []int) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(actual))
	}
	for i, v := range actual {
		if i >= len(expected) || v != expected[i] {
			t.Errorf("Expected %v, got %v", expected, actual)
			break
		}
	}
}
