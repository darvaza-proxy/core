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
	testListIteration(t, "empty list", ListForEach[int], S[int](), S[int]())
	testListIteration(t, "single element", ListForEach[int], S(1), S(1))
	testListIteration(t, "multiple elements", ListForEach[int], S(1, 2, 3), S(1, 2, 3))
}

func TestListForEachElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachElement, S[int](), S[int]())
	testListElementIteration(t, "single element", ListForEachElement, S(1), S(1))
	testListElementIteration(t, "multiple elements", ListForEachElement, S(1, 2, 3), S(1, 2, 3))
}

func TestListForEachBackward(t *testing.T) {
	testListIteration(t, "empty list", ListForEachBackward[int], S[int](), S[int]())
	testListIteration(t, "single element", ListForEachBackward[int], S(1), S(1))
	testListIteration(t, "multiple elements", ListForEachBackward[int], S(1, 2, 3), S(3, 2, 1))
}

func TestListForEachBackwardElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachBackwardElement, S[int](), S[int]())
	testListElementIteration(t, "single element", ListForEachBackwardElement, S(1), S(1))
	testListElementIteration(t, "multiple elements", ListForEachBackwardElement, S(1, 2, 3), S(3, 2, 1))
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
