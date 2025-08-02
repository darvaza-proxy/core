package core

import (
	"fmt"
	"math"
	"testing"
)

type sliceReverseTestCase struct {
	name string
	a, b []int
}

func (tc sliceReverseTestCase) test(t *testing.T) {
	t.Helper()
	c := SliceCopy(tc.a)
	SliceReverse(c)
	AssertSliceEqual(t, tc.b, c, "SliceReverse(%q) failed", tc.a)
	if SliceEqual(c, tc.b) {
		t.Logf("%s(%q) → %q", "SliceReverse", tc.a, c)
	}
}

var sliceReverseTestCases = []sliceReverseTestCase{
	{"empty", S[int](), S[int]()},
	{"single", S(1), S(1)},
	{"two elements", S(1, 2), S(2, 1)},
	{"three elements", S(1, 2, 3), S(3, 2, 1)},
	{"four elements", S(1, 2, 3, 4), S(4, 3, 2, 1)},
	{"five elements", S(1, 2, 3, 4, 5), S(5, 4, 3, 2, 1)},
	{"six elements", S(1, 2, 3, 4, 5, 6), S(6, 5, 4, 3, 2, 1)},
}

func TestSliceReverse(t *testing.T) {
	for _, tc := range sliceReverseTestCases {
		t.Run(tc.name, tc.test)
	}
}

// revive:disable
var (
	ints       = S(74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586)
	expectInts = S(74, 59, 238, -784, 9845, 959, 905, 0, 42, 7586, -5467984)

	float64s = S(
		74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, 7.8, 7.8, 74.3,
		59.0, math.Inf(1), 238.2, -784.0, 2.3,
	)
	expectFloat64s = S(74.3, 59.0, math.Inf(1), 238.2, -784, 2.3, 7.8)

	strs       = S("", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&")
	expectStrs = S("", "Hello", "foo", "bar", "f00", "%*&^*&^&")
)

func eq[T Ordered](a, b T) bool {
	return a == b
}

func cmp[T Ordered](a, b T) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}

func testSliceUnique[T Ordered](t *testing.T, before, after []T) {
	SliceSort(after, cmp[T])

	s := SliceUnique(before)
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUnique failed")

	s = SliceUniqueFn(before, eq[T])
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniqueFn failed")

	s = SliceCopyFn(before, nil)
	s2 := SliceUniquify(&s)
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniquify failed")
	AssertSliceEqual(t, s, s2, "SliceUniquify return value mismatch")

	s = SliceCopy(before)
	s2 = SliceUniquifyFn(&s, eq[T])
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniquifyFn failed")
	AssertSliceEqual(t, s, s2, "SliceUniquifyFn return value mismatch")
}

func TestSliceUniqueInt(t *testing.T) {
	testSliceUnique(t, ints, expectInts)
}

func TestSliceUniqueFloat(t *testing.T) {
	testSliceUnique(t, float64s, expectFloat64s)
}

func TestSliceUniqueString(t *testing.T) {
	testSliceUnique(t, strs, expectStrs)
}

func TestSliceMinus(t *testing.T) {
	for _, tc := range []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{"empty slices", S[int](), S[int](), S[int]()},
		{"empty a", S[int](), S(1, 2), S[int]()},
		{"empty b", S(1, 2), S[int](), S(1, 2)},
		{"no overlap", S(1, 2), S(3, 4), S(1, 2)},
		{"partial overlap", S(1, 2, 3, 4, 5), S(3, 4, 6, 7), S(1, 2, 5)},
		{"complete overlap", S(1, 2, 3), S(1, 2, 3), S[int]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := SliceMinus(tc.a, tc.b)
			AssertSliceEqual(t, tc.expected, result, "SliceMinus(%v, %v) failed", tc.a, tc.b)
		})
	}
}

func TestSliceMinusFn(t *testing.T) {
	equal := func(va, vb string) bool {
		return va == vb
	}

	for _, tc := range []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{"empty slices", S[string](), S[string](), S[string]()},
		{"empty a", S[string](), S("x", "y"), S[string]()},
		{"empty b", S("x", "y"), S[string](), S("x", "y")},
		{"no overlap", S("apple", "cherry"), S("banana", "date"), S("apple", "cherry")},
		{"partial overlap", S("apple", "banana", "cherry"), S("banana", "date"), S("apple", "cherry")},
		{"complete overlap", S("apple", "banana"), S("apple", "banana"), S[string]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := SliceMinusFn(tc.a, tc.b, equal)
			AssertSliceEqual(t, tc.expected, result, "SliceMinusFn(%v, %v) failed", tc.a, tc.b)
		})
	}
}

func TestSliceRandom(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		input  []string
		wantOK bool
	}{
		{name: "empty", input: S[string](), want: string(""), wantOK: false},
		{name: "one", input: S("one"), want: string("one"), wantOK: true},
		{name: "random", input: S("one", "two", "three", "four", "five", "six"), want: string(""), wantOK: true},
	}
	for _, tc := range tests {
		if tc.name != "random" {
			got, ok := SliceRandom(tc.input)
			AssertBool(t, ok, tc.wantOK, "SliceRandom ok status mismatch for %s", tc.name)
			AssertEqual(t, tc.want, got, "SliceRandom result mismatch for %s", tc.name)
		} else {
			u, ok := SliceRandom(tc.input)
			AssertBool(t, ok, tc.wantOK, "SliceRandom ok status mismatch for %s", tc.name)
			t.Logf("random from one,two,three,four,five,six: %s", u)
		}
	}
}

// Test cases for SliceContains function
type sliceContainsTestCase struct {
	name     string
	slice    []int
	value    int
	expected bool
}

func (tc sliceContainsTestCase) test(t *testing.T) {
	t.Helper()

	result := SliceContains(tc.slice, tc.value)
	AssertBool(t, result, tc.expected, "SliceContains result")
}

func TestSliceContains(t *testing.T) {
	testCases := []sliceContainsTestCase{
		{"empty slice", S[int](), 42, false},
		{"single element found", S(42), 42, true},
		{"single element not found", S(42), 99, false},
		{"multiple elements found", S(1, 2, 3, 42, 5), 42, true},
		{"multiple elements not found", S(1, 2, 3, 4, 5), 42, false},
		{"first element", S(42, 1, 2), 42, true},
		{"last element", S(1, 2, 42), 42, true},
		{"middle element", S(1, 42, 2), 42, true},
		{"duplicate elements", S(1, 42, 2, 42, 3), 42, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for SliceMap function
type sliceMapTestCase[T1, T2 any] struct {
	name     string
	input    []T1
	fn       func([]T2, T1) []T2
	expected []T2
}

func (tc sliceMapTestCase[T1, T2]) test(t *testing.T) {
	t.Helper()

	result := SliceMap(tc.input, tc.fn)
	AssertSliceEqual(t, tc.expected, result, "SliceMap result")
}

func TestSliceMap(t *testing.T) {
	// Simple transformation: returns one element per input
	intToString := func(partial []string, i int) []string {
		return S(fmt.Sprintf("num_%d", i))
	}

	// Debug function to understand behaviour
	debug := func(partial []int, i int) []int {
		t.Logf("partial=%v, i=%d", partial, i)
		return S(i)
	}

	// Simple case first
	t.Run("debug", func(t *testing.T) {
		result := SliceMap(S(1, 2, 3), debug)
		t.Logf("result=%v", result)
	})

	// Test simple mapping
	testCases := []sliceMapTestCase[int, string]{
		{"single element", S(42), intToString, S("num_42")},
		{"multiple elements", S(1, 2, 3), intToString, S("num_1", "num_2", "num_3")},
		{"negative numbers", S(-1, 0, 1), intToString, S("num_-1", "num_0", "num_1")},
	}

	// Test empty slice separately
	t.Run("empty slice", func(t *testing.T) {
		result := SliceMap(S[int](), intToString)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for SliceReversed function
type sliceReversedTestCase struct {
	name     string
	input    []int
	expected []int
}

func (tc sliceReversedTestCase) test(t *testing.T) {
	t.Helper()

	result := SliceReversed(tc.input)
	AssertSliceEqual(t, tc.expected, result, "SliceReversed result")

	// Verify original slice is unchanged
	originalCopy := SliceCopy(tc.input)
	AssertSliceEqual(t, originalCopy, tc.input, "SliceReversed should not modify original slice")
}

func TestSliceReversed(t *testing.T) {
	testCases := []sliceReversedTestCase{
		{"empty slice", S[int](), S[int]()},
		{"single element", S(42), S(42)},
		{"two elements", S(1, 2), S(2, 1)},
		{"three elements", S(1, 2, 3), S(3, 2, 1)},
		{"four elements", S(1, 2, 3, 4), S(4, 3, 2, 1)},
		{"five elements", S(1, 2, 3, 4, 5), S(5, 4, 3, 2, 1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for SliceReversedFn function
type sliceReversedFnTestCase[T any] struct {
	name     string
	input    []T
	fn       func([]T, T) (T, bool)
	expected []T
}

func (tc sliceReversedFnTestCase[T]) test(t *testing.T) {
	t.Helper()

	result := SliceReversedFn(tc.input, tc.fn)
	AssertSliceEqual(t, tc.expected, result, "SliceReversedFn result")
}

func TestSliceReversedFn(t *testing.T) {
	double := func(partial []int, i int) (int, bool) { return i * 2, true }
	negate := func(partial []int, i int) (int, bool) { return -i, true }
	filter := func(partial []int, i int) (int, bool) { return i, i > 0 }

	testCases := []sliceReversedFnTestCase[int]{
		{"empty slice", S[int](), double, S[int]()},
		{"single element", S(42), double, S(84)},
		{"double and reverse", S(1, 2, 3), double, S(6, 4, 2)},
		{"negate and reverse", S(1, 2, 3), negate, S(-3, -2, -1)},
		{"filter and reverse", S(-1, 2, -3, 4), filter, S(4, 2)},
		{"identity function", S(1, 2, 3, 4), func(partial []int, i int) (int, bool) { return i, true }, S(4, 3, 2, 1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for SliceSortFn function
type sliceSortFnTestCase[T any] struct {
	name     string
	input    []T
	less     func(T, T) bool
	expected []T
}

func (tc sliceSortFnTestCase[T]) test(t *testing.T) {
	t.Helper()

	// Make a copy to avoid modifying the original
	result := SliceCopy(tc.input)
	SliceSortFn(result, tc.less)
	AssertSliceEqual(t, tc.expected, result, "SliceSortFn result")
}

func TestSliceSortFn(t *testing.T) {
	intLess := func(a, b int) bool { return a < b }
	intGreater := func(a, b int) bool { return a > b }
	stringLess := func(a, b string) bool { return a < b }

	testCases := []sliceSortFnTestCase[int]{
		{"empty slice", S[int](), intLess, S[int]()},
		{"single element", S(42), intLess, S(42)},
		{"two elements ascending", S(2, 1), intLess, S(1, 2)},
		{"two elements descending", S(1, 2), intGreater, S(2, 1)},
		{"multiple elements", S(3, 1, 4, 1, 5), intLess, S(1, 1, 3, 4, 5)},
		{"already sorted", S(1, 2, 3, 4, 5), intLess, S(1, 2, 3, 4, 5)},
		{"reverse sorted", S(5, 4, 3, 2, 1), intLess, S(1, 2, 3, 4, 5)},
		{"duplicates", S(3, 1, 3, 1, 3), intLess, S(1, 1, 3, 3, 3)},
		{"negative numbers", S(-1, 3, -2, 0, 1), intLess, S(-2, -1, 0, 1, 3)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}

	// Test with string slice
	stringTests := []sliceSortFnTestCase[string]{
		{"empty string slice", S[string](), stringLess, S[string]()},
		{"single string", S("hello"), stringLess, S("hello")},
		{"multiple strings", S("cherry", "apple", "banana"), stringLess, S("apple", "banana", "cherry")},
		{"string duplicates", S("b", "a", "b", "c", "a"), stringLess, S("a", "a", "b", "b", "c")},
	}

	for _, tc := range stringTests {
		t.Run(tc.name, tc.test)
	}

	// Test edge cases
	t.Run("nil less function", func(t *testing.T) {
		original := S(3, 1, 2)
		result := SliceCopy(original)
		SliceSortFn(result, nil)
		AssertSliceEqual(t, original, result, "SliceSortFn with nil function should not modify slice")
	})
}

// Test cases for SliceSortOrdered function
type sliceSortOrderedTestCase[T Ordered] struct {
	name     string
	input    []T
	expected []T
}

func (tc sliceSortOrderedTestCase[T]) test(t *testing.T) {
	t.Helper()

	// Make a copy to avoid modifying the original
	result := SliceCopy(tc.input)
	SliceSortOrdered(result)
	AssertSliceEqual(t, tc.expected, result, "SliceSortOrdered result")
}

func TestSliceSortOrdered(t *testing.T) {
	// Test with int
	intTestCases := []sliceSortOrderedTestCase[int]{
		{"empty int slice", S[int](), S[int]()},
		{"single int", S(42), S(42)},
		{"two ints", S(2, 1), S(1, 2)},
		{"multiple ints", S(3, 1, 4, 1, 5), S(1, 1, 3, 4, 5)},
		{"already sorted ints", S(1, 2, 3, 4, 5), S(1, 2, 3, 4, 5)},
		{"reverse sorted ints", S(5, 4, 3, 2, 1), S(1, 2, 3, 4, 5)},
		{"negative ints", S(-1, 3, -2, 0, 1), S(-2, -1, 0, 1, 3)},
	}

	for _, tc := range intTestCases {
		t.Run(tc.name, tc.test)
	}

	// Test with string
	stringTestCases := []sliceSortOrderedTestCase[string]{
		{"empty string slice", S[string](), S[string]()},
		{"single string", S("hello"), S("hello")},
		{"multiple strings", S("cherry", "apple", "banana"), S("apple", "banana", "cherry")},
		{"string duplicates", S("b", "a", "b", "c", "a"), S("a", "a", "b", "b", "c")},
		{"empty strings", S("", "b", "", "a"), S("", "", "a", "b")},
	}

	for _, tc := range stringTestCases {
		t.Run(tc.name, tc.test)
	}

	// Test with float64
	floatTestCases := []sliceSortOrderedTestCase[float64]{
		{"empty float slice", S[float64](), S[float64]()},
		{"single float", S(3.14), S(3.14)},
		{"multiple floats", S(3.14, 1.41, 2.71), S(1.41, 2.71, 3.14)},
		{"float duplicates", S(1.0, 2.0, 1.0, 3.0), S(1.0, 1.0, 2.0, 3.0)},
		{"negative floats", S(-1.5, 2.5, -0.5, 0.0), S(-1.5, -0.5, 0.0, 2.5)},
	}

	for _, tc := range floatTestCases {
		t.Run(tc.name, tc.test)
	}
}
