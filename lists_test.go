package core_test

import (
	"container/list"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = listForEachTestCase{}
	_ core.TestCase = listForEachElementTestCase{}
	_ core.TestCase = listContainsTestCase{}
	_ core.TestCase = listCopyTestCase{}
	_ core.TestCase = listForEachTypeMismatchTestCase{}
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

// listOf builds a list of int, pushing the values in the order given.
func listOf(values ...int) *list.List {
	l := list.New()
	for _, v := range values {
		l.PushBack(v)
	}
	return l
}

// listForEachTestCase visits a list through one of the iteration
// functions taking a value callback. That function is the suite's
// subject rather than row data: the list factory hands the same one to
// every row of a table.
type listForEachTestCase struct {
	iterFn   func(*list.List, func(int) bool)
	name     string
	values   []int
	expected []int
}

func (tc listForEachTestCase) Name() string {
	return tc.name
}

func (tc listForEachTestCase) Test(t *testing.T) {
	t.Helper()

	var result []int
	tc.iterFn(listOf(tc.values...), func(v int) bool {
		result = append(result, v)
		return false
	})

	core.AssertSliceEqual(t, tc.expected, result, "visited values")
}

func newListForEachTestCase(name string, iterFn func(*list.List, func(int) bool),
	values, expected []int) listForEachTestCase {
	return listForEachTestCase{
		name:     name,
		iterFn:   iterFn,
		values:   values,
		expected: expected,
	}
}

// listForEachTestCases takes the order in which the subject visits
// several elements, which is what separates a forward function from a
// backward one.
func listForEachTestCases(iterFn func(*list.List, func(int) bool),
	wantMultiple []int) []listForEachTestCase {
	return []listForEachTestCase{
		newListForEachTestCase("empty list", iterFn, nil, nil),
		newListForEachTestCase("single element", iterFn, core.S(1), core.S(1)),
		newListForEachTestCase("multiple elements", iterFn, core.S(1, 2, 3), wantMultiple),
	}
}

// listForEachElementTestCase is the same suite for the iteration
// functions taking an element callback.
type listForEachElementTestCase struct {
	iterFn   func(*list.List, func(*list.Element) bool)
	name     string
	values   []int
	expected []int
}

func (tc listForEachElementTestCase) Name() string {
	return tc.name
}

func (tc listForEachElementTestCase) Test(t *testing.T) {
	t.Helper()

	var result []int
	tc.iterFn(listOf(tc.values...), func(e *list.Element) bool {
		result = append(result, e.Value.(int))
		return false
	})

	core.AssertSliceEqual(t, tc.expected, result, "visited values")
}

func newListForEachElementTestCase(name string, iterFn func(*list.List, func(*list.Element) bool),
	values, expected []int) listForEachElementTestCase {
	return listForEachElementTestCase{
		name:     name,
		iterFn:   iterFn,
		values:   values,
		expected: expected,
	}
}

func listForEachElementTestCases(iterFn func(*list.List, func(*list.Element) bool),
	wantMultiple []int) []listForEachElementTestCase {
	return []listForEachElementTestCase{
		newListForEachElementTestCase("empty list", iterFn, nil, nil),
		newListForEachElementTestCase("single element", iterFn, core.S(1), core.S(1)),
		newListForEachElementTestCase("multiple elements", iterFn, core.S(1, 2, 3), wantMultiple),
	}
}

func TestListForEach(t *testing.T) {
	core.RunTestCases(t, listForEachTestCases(core.ListForEach[int], core.S(1, 2, 3)))

	testListForEachNilAndEarlyReturn(t, "ListForEach", func(l *list.List, fn func(int) bool) {
		core.ListForEach(l, fn)
	}, core.S(1, 2))
}

func TestListForEachElement(t *testing.T) {
	core.RunTestCases(t, listForEachElementTestCases(core.ListForEachElement, core.S(1, 2, 3)))

	testListForEachElementNilAndEarlyReturn(t, "ListForEachElement",
		func(l *list.List, fn func(*list.Element) bool) {
			core.ListForEachElement(l, fn)
		}, core.S(1, 2))
}

func TestListForEachBackward(t *testing.T) {
	core.RunTestCases(t, listForEachTestCases(core.ListForEachBackward[int], core.S(3, 2, 1)))

	testListForEachNilAndEarlyReturn(t, "ListForEachBackward", func(l *list.List, fn func(int) bool) {
		core.ListForEachBackward(l, fn)
	}, core.S(3, 2))
}

func TestListForEachBackwardElement(t *testing.T) {
	core.RunTestCases(t, listForEachElementTestCases(core.ListForEachBackwardElement, core.S(3, 2, 1)))

	testListForEachElementNilAndEarlyReturn(t, "ListForEachBackwardElement",
		func(l *list.List, fn func(*list.Element) bool) {
			core.ListForEachBackwardElement(l, fn)
		}, core.S(3, 2))
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

	result := core.ListContains(l, tc.target)
	core.AssertEqual(t, tc.expected, result, "ListContains result")
}

func TestListContains(t *testing.T) {
	testCases := []listContainsTestCase{
		newListContainsTestCase("empty list", core.S[int](), 42, false),
		newListContainsTestCase("single element found", core.S(42), 42, true),
		newListContainsTestCase("single element not found", core.S(42), 99, false),
		newListContainsTestCase("multiple elements found", core.S(1, 2, 3, 42, 5), 42, true),
		newListContainsTestCase("multiple elements not found", core.S(1, 2, 3, 4, 5), 42, false),
		newListContainsTestCase("first element", core.S(42, 1, 2), 42, true),
		newListContainsTestCase("last element", core.S(1, 2, 42), 42, true),
		newListContainsTestCase("middle element", core.S(1, 42, 2), 42, true),
		newListContainsTestCase("duplicate elements", core.S(1, 42, 2, 42, 3), 42, true),
	}

	core.RunTestCases(t, testCases)

	// Test nil list
	t.Run("nil list", func(t *testing.T) {
		result := core.ListContains((*list.List)(nil), 42)
		core.AssertFalse(t, result, "ListContains nil list")
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
	copied := core.ListCopy[int](orig)

	// Everything below reads the copy, so a copy that is the original
	// or holds the wrong values makes the rest meaningless.
	core.AssertMustNotSame(t, orig, copied, "list instance")
	core.AssertMustSliceEqual(t, tc.values, listValues(copied), "copied values")

	// Appending to the original must not reach the copy.
	orig.PushBack(999)
	core.AssertSliceEqual(t, tc.values, listValues(copied), "copy after appending to the original")
}

func TestListCopy(t *testing.T) {
	testCases := []listCopyTestCase{
		newListCopyTestCase("empty list", core.S[int]()),
		newListCopyTestCase("single element", core.S(42)),
		newListCopyTestCase("multiple elements", core.S(1, 2, 3, 4, 5)),
		newListCopyTestCase("negative numbers", core.S(-1, 0, 1)),
		newListCopyTestCase("duplicates", core.S(1, 1, 2, 2, 3)),
	}

	core.RunTestCases(t, testCases)
}

// Test ListCopy with nil input
func TestListCopyNil(t *testing.T) {
	result := core.ListCopy[int](nil)
	core.AssertMustNotNil(t, result, "ListCopy(nil)")
	core.AssertEqual(t, 0, result.Len(), "length")
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
	result := core.ListContainsFn(l, 0, func(_, val int) bool {
		return val > 2
	})
	core.AssertTrue(t, result, "ListContainsFn custom function")
}

func testListContainsFnNilList(t *testing.T) {
	t.Helper()
	result := core.ListContainsFn((*list.List)(nil), 42, func(a, b int) bool {
		return a == b
	})
	core.AssertFalse(t, result, "ListContainsFn nil list")
}

func testListContainsFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := core.ListContainsFn(l, 42, nil)
	core.AssertFalse(t, result, "ListContainsFn nil function")
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
	result := core.ListCopyFn(l, func(v int) (int, bool) {
		return v * 2, true
	})

	core.AssertMustNotNil(t, result, "ListCopyFn(_, double)")
	core.AssertSliceEqual(t, core.S(2, 4, 6), listValues(result), "doubled values")
}

func testListCopyFnFiltering(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.PushBack(4)

	// Only keep even numbers
	result := core.ListCopyFn(l, func(v int) (int, bool) {
		return v, v%2 == 0
	})

	core.AssertMustNotNil(t, result, "ListCopyFn(_, keepEven)")
	core.AssertSliceEqual(t, core.S(2, 4), listValues(result), "kept values")
}

func testListCopyFnNilList(t *testing.T) {
	t.Helper()
	result := core.ListCopyFn((*list.List)(nil), func(v int) (int, bool) {
		return v, true
	})
	core.AssertMustNotNil(t, result, "ListCopyFn(nil, _)")
	core.AssertSliceEqual(t, core.S[int](), listValues(result), "nil list")
}

func testListCopyFnNilFunction(t *testing.T) {
	t.Helper()
	l := list.New()
	l.PushBack(42)

	result := core.ListCopyFn[int](l, nil)
	core.AssertMustNotNil(t, result, "ListCopyFn(_, nil)")
	core.AssertSliceEqual(t, core.S(42), listValues(result), "copied values")
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

	core.AssertNoPanic(t, func() { iterFn(l, nil) }, name+" nil function")
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
	core.AssertSliceEqual(t, want, result, name+" early return")
}

func testListForEachNilList(t *testing.T, name string, iterFn func(*list.List, func(int) bool)) {
	t.Helper()
	visits := 0
	core.AssertNoPanic(t, func() {
		iterFn((*list.List)(nil), func(int) bool {
			visits++
			return false
		})
	}, name+" nil list")
	core.AssertEqual(t, 0, visits, name+" nil list visits")
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

	core.AssertNoPanic(t, func() { iterFn(l, nil) }, name+" nil function")
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
	core.AssertSliceEqual(t, want, result, name+" early return")
}

func testListForEachElementNilList(t *testing.T, name string, iterFn func(*list.List, func(*list.Element) bool)) {
	t.Helper()
	visits := 0
	core.AssertNoPanic(t, func() {
		iterFn((*list.List)(nil), func(*list.Element) bool {
			visits++
			return false
		})
	}, name+" nil list")
	core.AssertEqual(t, 0, visits, name+" nil list visits")
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
	core.AssertSliceEqual(t, tc.want, collected, "visited values")
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

	core.RunTestCases(t, []listForEachTypeMismatchTestCase{
		newListForEachTypeMismatchTestCase("forward", l, core.ListForEach[int], core.S(1, 3)),
		newListForEachTypeMismatchTestCase("backward", l, core.ListForEachBackward[int], core.S(3, 1)),
	})
}

// ListContains and ListContainsFn walk the list through ListForEach, so
// they inherit that skip: the type parameter selects which of a mixed
// list's elements they can match at all.
func TestListContainsTypeMismatch(t *testing.T) {
	l := list.New()
	l.PushBack(1)
	l.PushBack("not-an-int")
	l.PushBack(3)

	core.AssertTrue(t, core.ListContains(l, 3), "ListContains(_, 3)")
	core.AssertFalse(t, core.ListContains(l, 2), "ListContains(_, 2)")
	core.AssertTrue(t, core.ListContains(l, "not-an-int"), `ListContains(_, "not-an-int")`)
}

// ListCopy inherits the skip as well, which drops the elements of
// another type rather than carrying them over.
func TestListCopyTypeMismatch(t *testing.T) {
	l := list.New()
	l.PushBack(1)
	l.PushBack("not-an-int")
	l.PushBack(3)

	copied := core.ListCopy[int](l)
	core.AssertMustNotNil(t, copied, "ListCopy[int]")
	core.AssertSliceEqual(t, core.S(1, 3), listValues(copied), "copied values")
	core.AssertEqual(t, 3, l.Len(), "source length")
}

// ListCopy is documented as shallow, so the copy holds the values the
// original holds rather than duplicates of them.
func TestListCopyIsShallow(t *testing.T) {
	v := 42
	orig := list.New()
	orig.PushBack(&v)

	copied := core.ListCopy[*int](orig)
	core.AssertMustNotNil(t, copied, "ListCopy[*int]")
	core.AssertMustEqual(t, 1, copied.Len(), "length")

	got := core.AssertMustTypeIs[*int](t, copied.Front().Value, "copied element")
	core.AssertSame(t, &v, got, "element pointer")
}
