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

func runTestListIteration(t *testing.T, name string, iterFn func(*list.List, func(int) bool), values, expected []int) {
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

func runTestListElementIteration(
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
	runTestListIteration(t, "empty list", ListForEach[int], S[int](), S[int]())
	runTestListIteration(t, "single element", ListForEach[int], S(1), S(1))
	runTestListIteration(t, "multiple elements", ListForEach[int], S(1, 2, 3), S(1, 2, 3))

	runTestListForEachNilAndEarlyReturn(t, "ListForEach", func(l *list.List, fn func(int) bool) {
		ListForEach(l, fn)
	})
}

func TestListForEachElement(t *testing.T) {
	runTestListElementIteration(t, "empty list", ListForEachElement, S[int](), S[int]())
	runTestListElementIteration(t, "single element", ListForEachElement, S(1), S(1))
	runTestListElementIteration(t, "multiple elements", ListForEachElement, S(1, 2, 3), S(1, 2, 3))

	runTestListForEachElementNilAndEarlyReturn(t, "ListForEachElement", func(l *list.List, fn func(*list.Element) bool) {
		ListForEachElement(l, fn)
	})
}

func TestListForEachBackward(t *testing.T) {
	runTestListIteration(t, "empty list", ListForEachBackward[int], S[int](), S[int]())
	runTestListIteration(t, "single element", ListForEachBackward[int], S(1), S(1))
	runTestListIteration(t, "multiple elements", ListForEachBackward[int], S(1, 2, 3), S(3, 2, 1))

	runTestListForEachBackwardNilAndEarlyReturn(t, "ListForEachBackward", func(l *list.List, fn func(int) bool) {
		ListForEachBackward(l, fn)
	})
}

func TestListForEachBackwardElement(t *testing.T) {
	runTestListElementIteration(t, "empty list", ListForEachBackwardElement, S[int](), S[int]())
	runTestListElementIteration(t, "single element", ListForEachBackwardElement, S(1), S(1))
	runTestListElementIteration(t, "multiple elements", ListForEachBackwardElement, S(1, 2, 3), S(3, 2, 1))

	runTestListForEachBackwardElementNilAndEarlyReturn(t, "ListForEachBackwardElement",
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
func newListContainsTestCase(name string, values []int, target int, expected bool) TestCase {
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

func makeListContainsBasicTestCases() []TestCase {
	return []TestCase{
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
}

func TestListContains(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		RunTestCases(t, makeListContainsBasicTestCases())
	})

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
func newListCopyTestCase(name string, values []int) TestCase {
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
	AssertNotSame(t, orig, copied, "list instance")

	// Verify independence - modifying one doesn't affect the other
	if orig.Len() > 0 {
		orig.PushBack(999)
		AssertEqual(t, orig.Len()-1, copied.Len(), "independence")
	}
}

func makeListCopyBasicTestCases() []TestCase {
	return []TestCase{
		newListCopyTestCase("empty list", S[int]()),
		newListCopyTestCase("single element", S(42)),
		newListCopyTestCase("multiple elements", S(1, 2, 3, 4, 5)),
		newListCopyTestCase("negative numbers", S(-1, 0, 1)),
		newListCopyTestCase("duplicates", S(1, 1, 2, 2, 3)),
	}
}

func TestListCopy(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		RunTestCases(t, makeListCopyBasicTestCases())
	})
	t.Run("nil", runTestListCopyNil)
}

func runTestListCopyNil(t *testing.T) {
	t.Helper()
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
	t.Run("custom function", runTestListContainsFnCustomFunction)
	t.Run("nil list", runTestListContainsFnNilList)
	t.Run("nil function", runTestListContainsFnNilFunction)
}

func runTestListContainsFnCustomFunction(t *testing.T) {
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

func runTestListContainsFnNilList(t *testing.T) {
	t.Helper()
	result := ListContainsFn((*list.List)(nil), 42, func(a, b int) bool {
		return a == b
	})
	AssertFalse(t, result, "ListContainsFn nil list")
}

func runTestListContainsFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := ListContainsFn(l, 42, nil)
	AssertFalse(t, result, "ListContainsFn nil function")
}

// Test ListCopyFn function
func TestListCopyFn(t *testing.T) {
	t.Run("transformation function", runTestListCopyFnTransformation)
	t.Run("filtering function", runTestListCopyFnFiltering)
	t.Run("nil list", runTestListCopyFnNilList)
	t.Run("nil function", runTestListCopyFnNilFunction)
}

func runTestListCopyFnTransformation(t *testing.T) {
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

func runTestListCopyFnFiltering(t *testing.T) {
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

func runTestListCopyFnNilList(t *testing.T) {
	t.Helper()
	result := ListCopyFn((*list.List)(nil), func(v int) (int, bool) {
		return v, true
	})
	AssertEqual(t, 0, result.Len(), "nil list")
}

func runTestListCopyFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := ListCopyFn[int](l, nil)
	AssertEqual(t, 1, result.Len(), "nil function")
	AssertEqual(t, 42, result.Front().Value.(int), "value")
}

func runTestListForEachNilAndEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()

	t.Run("nil function", func(t *testing.T) {
		runTestListForEachNilFunction(t, name, iterFn)
	})
	t.Run("early return", func(t *testing.T) {
		runTestListForEachEarlyReturn(t, name, iterFn)
	})
	t.Run("nil list", func(t *testing.T) {
		runTestListForEachNilList(t, name, iterFn)
	})
}

func runTestListForEachNilFunction(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)

	var result []int
	iterFn(l, func(int) bool { return false })
	AssertEqual(t, 0, len(result), name+" nil function")
}

func runTestListForEachEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
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

func runTestListForEachNilList(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	var result []int
	iterFn((*list.List)(nil), func(v int) bool {
		result = append(result, v)
		return false
	})
	AssertEqual(t, 0, len(result), name+" nil list")
}

func runTestListForEachElementNilAndEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()

	t.Run("nil function", func(t *testing.T) {
		runTestListForEachElementNilFunction(t, name, iterFn)
	})
	t.Run("early return", func(t *testing.T) {
		runTestListForEachElementEarlyReturn(t, name, iterFn)
	})
	t.Run("nil list", func(t *testing.T) {
		runTestListForEachElementNilList(t, name, iterFn)
	})
}

func runTestListForEachElementNilFunction(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)

	var called bool
	iterFn(l, nil)
	AssertFalse(t, called, name+" nil function")
}

func runTestListForEachElementEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool)) {
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

func runTestListForEachElementNilList(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	var result []int
	iterFn((*list.List)(nil), func(e *list.Element) bool {
		result = append(result, e.Value.(int))
		return false
	})
	AssertEqual(t, 0, len(result), name+" nil list")
}

func runTestListForEachBackwardNilAndEarlyReturn(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	runTestListForEachNilAndEarlyReturn(t, name, iterFn)
}

func runTestListForEachBackwardElementNilAndEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	runTestListForEachElementNilAndEarlyReturn(t, name, iterFn)
}
