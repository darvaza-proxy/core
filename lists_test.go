package core

import (
	"container/list"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = listContainsTestCase{}
	_ TestCase = listCopyTestCase{}
)

func testListIteration(t *testing.T, name string, iterFn func(*list.List, func(int) bool), values, expected []int) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		runListIterationTest(t, iterFn, values, expected)
	})
}

func runListIterationTest(t *testing.T, iterFn func(*list.List, func(int) bool), values, expected []int) {
	t.Helper()
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
}

func testListElementIteration(
	t *testing.T,
	name string,
	iterFn func(*list.List, func(*list.Element) bool),
	values, expected []int,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		runListElementIterationTest(t, iterFn, values, expected)
	})
}

func runListElementIterationTest(
	t *testing.T,
	iterFn func(*list.List, func(*list.Element) bool),
	values, expected []int,
) {
	t.Helper()
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
}

func TestListForEach(t *testing.T) {
	testListIteration(t, "empty list", ListForEach[int], S[int](), S[int]())
	testListIteration(t, "single element", ListForEach[int], S(1), S(1))
	testListIteration(t, "multiple elements", ListForEach[int], S(1, 2, 3), S(1, 2, 3))

	testListForEachNilAndEarlyReturn(t, "ListForEach", func(l *list.List, fn func(int) bool) {
		ListForEach(l, fn)
	})
}

func TestListForEachElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachElement, S[int](), S[int]())
	testListElementIteration(t, "single element", ListForEachElement, S(1), S(1))
	testListElementIteration(t, "multiple elements", ListForEachElement, S(1, 2, 3), S(1, 2, 3))

	testListForEachElementNilAndEarlyReturn(t, "ListForEachElement", func(l *list.List, fn func(*list.Element) bool) {
		ListForEachElement(l, fn)
	})
}

func TestListForEachBackward(t *testing.T) {
	testListIteration(t, "empty list", ListForEachBackward[int], S[int](), S[int]())
	testListIteration(t, "single element", ListForEachBackward[int], S(1), S(1))
	testListIteration(t, "multiple elements", ListForEachBackward[int], S(1, 2, 3), S(3, 2, 1))

	testListForEachBackwardNilAndEarlyReturn(t, "ListForEachBackward", func(l *list.List, fn func(int) bool) {
		ListForEachBackward(l, fn)
	})
}

func TestListForEachBackwardElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachBackwardElement, S[int](), S[int]())
	testListElementIteration(t, "single element", ListForEachBackwardElement, S(1), S(1))
	testListElementIteration(t, "multiple elements", ListForEachBackwardElement, S(1, 2, 3), S(3, 2, 1))

	testListForEachBackwardElementNilAndEarlyReturn(t, "ListForEachBackwardElement",
		func(l *list.List, fn func(*list.Element) bool) {
			ListForEachBackwardElement(l, fn)
		})
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

// Test cases for ListContains function
type listContainsTestCase struct {
	name     string
	values   []int
	target   int
	expected bool
}

// newListContainsTestCase creates a new listContainsTestCase
func newListContainsTestCase(name string, values []int, target int, expected bool) listContainsTestCase {
	return listContainsTestCase{
		name:     name,
		values:   values,
		target:   target,
		expected: expected,
	}
}

func (tc listContainsTestCase) Name() string {
	return tc.name
}

func (tc listContainsTestCase) Test(t *testing.T) {
	t.Helper()

	l := list.New()
	for _, v := range tc.values {
		l.PushBack(v)
	}

	result := ListContains(l, tc.target)
	AssertEqual(t, tc.expected, result, "ListContains result")
}

func TestListContains(t *testing.T) {
	testCases := []listContainsTestCase{
		newListContainsTestCase("empty list", S[int](), 42, false),
		newListContainsTestCase("single element found", S(42), 42, true),
		newListContainsTestCase("single element not found", S(42), 99, false),
		newListContainsTestCase("multiple elements found", S(1, 2, 3, 42, 5), 42, true),
		newListContainsTestCase("multiple elements not found", S(1, 2, 3, 4, 5), 42, false),
		newListContainsTestCase("first element", S(42, 1, 2), 42, true),
		newListContainsTestCase("last element", S(1, 2, 42), 42, true),
		newListContainsTestCase("middle element", S(1, 42, 2), 42, true),
		newListContainsTestCase("duplicate elements", S(1, 42, 2, 42, 3), 42, true),
	}

	RunTestCases(t, testCases)

	// Test nil list
	t.Run("nil list", func(t *testing.T) {
		result := ListContains((*list.List)(nil), 42)
		AssertFalse(t, result, "ListContains nil list")
	})
}

// Test cases for ListCopy function
type listCopyTestCase struct {
	name   string
	values []int
}

// newListCopyTestCase creates a new listCopyTestCase
func newListCopyTestCase(name string, values []int) listCopyTestCase {
	return listCopyTestCase{
		name:   name,
		values: values,
	}
}

func (tc listCopyTestCase) Name() string {
	return tc.name
}

func (tc listCopyTestCase) Test(t *testing.T) {
	t.Helper()

	// Create original list
	orig := list.New()
	for _, v := range tc.values {
		orig.PushBack(v)
	}

	// Copy the list
	copied := ListCopy[int](orig)

	// Verify same length
	AssertEqual(t, orig.Len(), copied.Len(), "length")

	// Verify same elements
	origElem := orig.Front()
	copiedElem := copied.Front()
	for origElem != nil && copiedElem != nil {
		AssertEqual(t, origElem.Value.(int), copiedElem.Value.(int), "element value")
		origElem = origElem.Next()
		copiedElem = copiedElem.Next()
	}

	// Verify they are different lists (not the same pointer)
	if orig == copied {
		t.Error("ListCopy should return a different list instance")
	}

	// Verify independence - modifying one doesn't affect the other
	if orig.Len() > 0 {
		orig.PushBack(999)
		AssertEqual(t, orig.Len()-1, copied.Len(), "independence")
	}
}

func TestListCopy(t *testing.T) {
	testCases := []listCopyTestCase{
		newListCopyTestCase("empty list", S[int]()),
		newListCopyTestCase("single element", S(42)),
		newListCopyTestCase("multiple elements", S(1, 2, 3, 4, 5)),
		newListCopyTestCase("negative numbers", S(-1, 0, 1)),
		newListCopyTestCase("duplicates", S(1, 1, 2, 2, 3)),
	}

	RunTestCases(t, testCases)
}

// Test ListCopy with nil input
func TestListCopyNil(t *testing.T) {
	result := ListCopy[int](nil)
	if result == nil {
		t.Error("ListCopy(nil) should return a new empty list, not nil")
	}
	if result.Len() != 0 {
		t.Errorf("ListCopy(nil) should return empty list, got length %d", result.Len())
	}
}

// Test ListContainsFn function
func TestListContainsFn(t *testing.T) {
	t.Run("custom function", testListContainsFnCustomFunction)
	t.Run("nil list", testListContainsFnNilList)
	t.Run("nil function", testListContainsFnNilFunction)
}

func testListContainsFnCustomFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// Find element greater than 2
	result := ListContainsFn(l, 0, func(_, val int) bool {
		return val > 2
	})
	AssertTrue(t, result, "ListContainsFn custom function")
}

func testListContainsFnNilList(t *testing.T) {
	t.Helper()
	result := ListContainsFn((*list.List)(nil), 42, func(a, b int) bool {
		return a == b
	})
	AssertFalse(t, result, "ListContainsFn nil list")
}

func testListContainsFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := ListContainsFn(l, 42, nil)
	AssertFalse(t, result, "ListContainsFn nil function")
}

// Test ListCopyFn function
func TestListCopyFn(t *testing.T) {
	t.Run("transformation function", testListCopyFnTransformation)
	t.Run("filtering function", testListCopyFnFiltering)
	t.Run("nil list", testListCopyFnNilList)
	t.Run("nil function", testListCopyFnNilFunction)
}

func testListCopyFnTransformation(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// Double each value
	result := ListCopyFn(l, func(v int) (int, bool) {
		return v * 2, true
	})

	AssertEqual(t, 3, result.Len(), "length")
	expected := S(2, 4, 6)
	i := 0
	for e := result.Front(); e != nil; e = e.Next() {
		AssertEqual(t, expected[i], e.Value.(int), "value[%d]", i)
		i++
	}
}

func testListCopyFnFiltering(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.PushBack(4)

	// Only keep even numbers
	result := ListCopyFn(l, func(v int) (int, bool) {
		return v, v%2 == 0
	})

	AssertEqual(t, 2, result.Len(), "filtered length")
	expected := S(2, 4)
	i := 0
	for e := result.Front(); e != nil; e = e.Next() {
		AssertEqual(t, expected[i], e.Value.(int), "value[%d]", i)
		i++
	}
}

func testListCopyFnNilList(t *testing.T) {
	t.Helper()
	result := ListCopyFn((*list.List)(nil), func(v int) (int, bool) {
		return v, true
	})
	AssertEqual(t, 0, result.Len(), "nil list")
}

func testListCopyFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := ListCopyFn[int](l, nil)
	AssertEqual(t, 1, result.Len(), "nil function")
	AssertEqual(t, 42, result.Front().Value.(int), "value")
}

func testListForEachNilAndEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()

	t.Run("nil function", func(t *testing.T) {
		testListForEachNilFunction(t, name, iterFn)
	})
	t.Run("early return", func(t *testing.T) {
		testListForEachEarlyReturn(t, name, iterFn)
	})
	t.Run("nil list", func(t *testing.T) {
		testListForEachNilList(t, name, iterFn)
	})
}

func testListForEachNilFunction(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)

	var result []int
	iterFn(l, func(int) bool { return false })
	AssertEqual(t, 0, len(result), name+" nil function")
}

func testListForEachEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	var result []int
	iterFn(l, func(v int) bool {
		result = append(result, v)
		return v == 2 // Stop when we hit 2
	})
	AssertEqual(t, 2, len(result), name+" early return")
	if name == "ListForEach" {
		AssertEqual(t, 1, result[0], "First element")
		AssertEqual(t, 2, result[1], "Second element")
	} else {
		AssertEqual(t, 3, result[0], "first (backward)")
		AssertEqual(t, 2, result[1], "second (backward)")
	}
}

func testListForEachNilList(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	var result []int
	iterFn((*list.List)(nil), func(v int) bool {
		result = append(result, v)
		return false
	})
	AssertEqual(t, 0, len(result), name+" nil list")
}

func testListForEachElementNilAndEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()

	t.Run("nil function", func(t *testing.T) {
		testListForEachElementNilFunction(t, name, iterFn)
	})
	t.Run("early return", func(t *testing.T) {
		testListForEachElementEarlyReturn(t, name, iterFn)
	})
	t.Run("nil list", func(t *testing.T) {
		testListForEachElementNilList(t, name, iterFn)
	})
}

func testListForEachElementNilFunction(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)

	var called bool
	iterFn(l, nil)
	AssertFalse(t, called, name+" nil function")
}

func testListForEachElementEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	var result []int
	iterFn(l, func(e *list.Element) bool {
		result = append(result, e.Value.(int))
		return e.Value.(int) == 2 // Stop when we hit 2
	})
	AssertEqual(t, 2, len(result), name+" early return")
	if name == "ListForEachElement" {
		AssertEqual(t, 1, result[0], "First element")
		AssertEqual(t, 2, result[1], "Second element")
	} else {
		AssertEqual(t, 3, result[0], "first (backward)")
		AssertEqual(t, 2, result[1], "second (backward)")
	}
}

func testListForEachElementNilList(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	var result []int
	iterFn((*list.List)(nil), func(e *list.Element) bool {
		result = append(result, e.Value.(int))
		return false
	})
	AssertEqual(t, 0, len(result), name+" nil list")
}

func testListForEachBackwardNilAndEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	testListForEachNilAndEarlyReturn(t, name, iterFn)
}

func testListForEachBackwardElementNilAndEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	testListForEachElementNilAndEarlyReturn(t, name, iterFn)
}
