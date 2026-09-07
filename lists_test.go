package core

import (
	"container/list"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = listContainsTestCase{}
	_ TestCase = listCopyTestCase{}
	_ TestCase = listForEachTypeMismatchTestCase{}
)

// listValues collects the values of a list of int, walking it through
// container/list rather than through any subject under test.
func listValues(l *list.List) []int {
	out := make([]int, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		out = append(out, e.Value.(int))
	}
	return out
}

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

	AssertSliceEqual(t, expected, result, "visited values")
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

	AssertSliceEqual(t, expected, result, "visited values")
}

func TestListForEach(t *testing.T) {
	testListIteration(t, "empty list", ListForEach[int], nil, nil)
	testListIteration(t, "single element", ListForEach[int], S(1), S(1))
	testListIteration(t, "multiple elements", ListForEach[int], S(1, 2, 3), S(1, 2, 3))

	testListForEachNilAndEarlyReturn(t, "ListForEach", func(l *list.List, fn func(int) bool) {
		ListForEach(l, fn)
	}, S(1, 2))
}

func TestListForEachElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachElement, nil, nil)
	testListElementIteration(t, "single element", ListForEachElement, S(1), S(1))
	testListElementIteration(t, "multiple elements", ListForEachElement, S(1, 2, 3), S(1, 2, 3))

	testListForEachElementNilAndEarlyReturn(t, "ListForEachElement",
		func(l *list.List, fn func(*list.Element) bool) {
			ListForEachElement(l, fn)
		}, S(1, 2))
}

func TestListForEachBackward(t *testing.T) {
	testListIteration(t, "empty list", ListForEachBackward[int], nil, nil)
	testListIteration(t, "single element", ListForEachBackward[int], S(1), S(1))
	testListIteration(t, "multiple elements", ListForEachBackward[int], S(1, 2, 3), S(3, 2, 1))

	testListForEachNilAndEarlyReturn(t, "ListForEachBackward", func(l *list.List, fn func(int) bool) {
		ListForEachBackward(l, fn)
	}, S(3, 2))
}

func TestListForEachBackwardElement(t *testing.T) {
	testListElementIteration(t, "empty list", ListForEachBackwardElement, nil, nil)
	testListElementIteration(t, "single element", ListForEachBackwardElement, S(1), S(1))
	testListElementIteration(t, "multiple elements", ListForEachBackwardElement, S(1, 2, 3), S(3, 2, 1))

	testListForEachElementNilAndEarlyReturn(t, "ListForEachBackwardElement",
		func(l *list.List, fn func(*list.Element) bool) {
			ListForEachBackwardElement(l, fn)
		}, S(3, 2))
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

	// Everything below reads the copy, so a copy that is the original
	// or holds the wrong values makes the rest meaningless.
	AssertMustNotSame(t, orig, copied, "list instance")
	AssertMustSliceEqual(t, tc.values, listValues(copied), "copied values")

	// Appending to the original must not reach the copy.
	orig.PushBack(999)
	AssertSliceEqual(t, tc.values, listValues(copied), "copy after appending to the original")
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
	AssertMustNotNil(t, result, "ListCopy(nil)")
	AssertEqual(t, 0, result.Len(), "length")
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

	AssertMustNotNil(t, result, "ListCopyFn(_, double)")
	AssertSliceEqual(t, S(2, 4, 6), listValues(result), "doubled values")
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

	AssertMustNotNil(t, result, "ListCopyFn(_, keepEven)")
	AssertSliceEqual(t, S(2, 4), listValues(result), "kept values")
}

func testListCopyFnNilList(t *testing.T) {
	t.Helper()
	result := ListCopyFn((*list.List)(nil), func(v int) (int, bool) {
		return v, true
	})
	AssertMustNotNil(t, result, "ListCopyFn(nil, _)")
	AssertSliceEqual(t, S[int](), listValues(result), "nil list")
}

func testListCopyFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := ListCopyFn[int](l, nil)
	AssertMustNotNil(t, result, "ListCopyFn(_, nil)")
	AssertSliceEqual(t, S(42), listValues(result), "nil function copies everything")
}

func testListForEachNilAndEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(int) bool), wantEarly []int) {
	t.Helper()

	t.Run("nil function", func(t *testing.T) {
		testListForEachNilFunction(t, name, iterFn)
	})
	t.Run("early return", func(t *testing.T) {
		testListForEachEarlyReturn(t, name, iterFn, wantEarly)
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

	AssertNoPanic(t, func() { iterFn(l, nil) }, name+" nil function")
}

func testListForEachEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(int) bool), want []int) {
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
	AssertSliceEqual(t, want, result, name+" early return")
}

func testListForEachNilList(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	visits := 0
	AssertNoPanic(t, func() {
		iterFn((*list.List)(nil), func(int) bool {
			visits++
			return false
		})
	}, name+" nil list")
	AssertEqual(t, 0, visits, name+" nil list visits")
}

func testListForEachElementNilAndEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool), wantEarly []int) {
	t.Helper()

	t.Run("nil function", func(t *testing.T) {
		testListForEachElementNilFunction(t, name, iterFn)
	})
	t.Run("early return", func(t *testing.T) {
		testListForEachElementEarlyReturn(t, name, iterFn, wantEarly)
	})
	t.Run("nil list", func(t *testing.T) {
		testListForEachElementNilList(t, name, iterFn)
	})
}

func testListForEachElementNilFunction(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	l := list.New()
	l.PushBack(1)

	AssertNoPanic(t, func() { iterFn(l, nil) }, name+" nil function")
}

func testListForEachElementEarlyReturn(t *testing.T, name string,
	iterFn func(*list.List, func(*list.Element) bool), want []int) {
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
	AssertSliceEqual(t, want, result, name+" early return")
}

func testListForEachElementNilList(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	visits := 0
	AssertNoPanic(t, func() {
		iterFn((*list.List)(nil), func(*list.Element) bool {
			visits++
			return false
		})
	}, name+" nil list")
	AssertEqual(t, 0, visits, name+" nil list visits")
}

type listForEachTypeMismatchTestCase struct {
	list *list.List
	iter func(*list.List, func(int) bool)
	name string
	want []int
}

func (tc listForEachTypeMismatchTestCase) Name() string { return tc.name }

func (tc listForEachTypeMismatchTestCase) Test(t *testing.T) {
	t.Helper()
	collected := make([]int, 0, len(tc.want))
	tc.iter(tc.list, func(v int) bool {
		collected = append(collected, v)
		return false
	})
	AssertSliceEqual(t, tc.want, collected, "visited values")
}

func newListForEachTypeMismatchTestCase(name string, l *list.List,
	iter func(*list.List, func(int) bool), want []int) listForEachTypeMismatchTestCase {
	return listForEachTypeMismatchTestCase{
		name: name,
		list: l,
		iter: iter,
		want: want,
	}
}

// Cover the type-mismatch branch in ListForEach/Backward: elements whose
// Value is not of type T are silently skipped (inner closure returns false
// to continue iteration).
func TestListForEachTypeMismatch(t *testing.T) {
	l := list.New()
	l.PushBack(1)
	l.PushBack("not-an-int")
	l.PushBack(3)

	RunTestCases(t, []listForEachTypeMismatchTestCase{
		newListForEachTypeMismatchTestCase("forward", l, ListForEach[int], S(1, 3)),
		newListForEachTypeMismatchTestCase("backward", l, ListForEachBackward[int], S(3, 1)),
	})
}
