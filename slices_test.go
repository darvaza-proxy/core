package core

import (
	"fmt"
	"math"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = sliceReverseTestCase{}
	_ TestCase = sliceContainsTestCase{}
	_ TestCase = sliceMapTestCase[int, string]{}
	_ TestCase = sliceReversedTestCase{}
	_ TestCase = sliceReversedFnTestCase[int]{}
	_ TestCase = sliceSortFnTestCase[int]{}
	_ TestCase = sliceSortOrderedTestCase[int]{}
	_ TestCase = sliceMinusTestCase{}
	_ TestCase = sliceMinusFnTestCase{}
	_ TestCase = sliceRandomTestCase{}
)

type sliceReverseTestCase struct {
	name string
	a, b []int
}

func (tc sliceReverseTestCase) Name() string {
	return tc.name
}

func (tc sliceReverseTestCase) Test(t *testing.T) {
	t.Helper()
	c := SliceCopy(tc.a)
	SliceReverse(c)
	AssertSliceEqual(t, tc.b, c, "SliceReverse", tc.a)
	if SliceEqual(c, tc.b) {
		t.Logf("%s(%q) → %q", "SliceReverse", tc.a, c)
	}
}

// Factory function for sliceReverseTestCase
func newSliceReverseTestCase(name string, a, b []int) sliceReverseTestCase {
	return sliceReverseTestCase{
		name: name,
		a:    a,
		b:    b,
	}
}

var sliceReverseTestCases = []sliceReverseTestCase{
	newSliceReverseTestCase("empty", S[int](), S[int]()),
	newSliceReverseTestCase("single", S(1), S(1)),
	newSliceReverseTestCase("two elements", S(1, 2), S(2, 1)),
	newSliceReverseTestCase("three elements", S(1, 2, 3), S(3, 2, 1)),
	newSliceReverseTestCase("four elements", S(1, 2, 3, 4), S(4, 3, 2, 1)),
	newSliceReverseTestCase("five elements", S(1, 2, 3, 4, 5), S(5, 4, 3, 2, 1)),
	newSliceReverseTestCase("six elements", S(1, 2, 3, 4, 5, 6), S(6, 5, 4, 3, 2, 1)),
}

func TestSliceReverse(t *testing.T) {
	RunTestCases(t, sliceReverseTestCases)
}

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
	AssertSliceEqual(t, after, s, "SliceUnique")

	s = SliceUniqueFn(before, eq[T])
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniqueFn")

	s = SliceCopyFn(before, nil)
	s2 := SliceUniquify(&s)
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniquify")
	AssertSliceEqual(t, s, s2, "return value")

	s = SliceCopy(before)
	s2 = SliceUniquifyFn(&s, eq[T])
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniquifyFn")
	AssertSliceEqual(t, s, s2, "return value")
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

// Test cases for SliceMinus function
type sliceMinusTestCase struct {
	name     string
	a        []int
	b        []int
	expected []int
}

func (tc sliceMinusTestCase) Name() string {
	return tc.name
}

func (tc sliceMinusTestCase) Test(t *testing.T) {
	t.Helper()
	result := SliceMinus(tc.a, tc.b)
	AssertSliceEqual(t, tc.expected, result, "SliceMinus", tc.a, tc.b)
}

func newSliceMinusTestCase(name string, a, b, expected []int) sliceMinusTestCase {
	return sliceMinusTestCase{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
	}
}

func TestSliceMinus(t *testing.T) {
	testCases := []sliceMinusTestCase{
		newSliceMinusTestCase("empty slices", S[int](), S[int](), S[int]()),
		newSliceMinusTestCase("empty a", S[int](), S(1, 2), S[int]()),
		newSliceMinusTestCase("empty b", S(1, 2), S[int](), S(1, 2)),
		newSliceMinusTestCase("no overlap", S(1, 2), S(3, 4), S(1, 2)),
		newSliceMinusTestCase("partial overlap", S(1, 2, 3, 4, 5), S(3, 4, 6, 7), S(1, 2, 5)),
		newSliceMinusTestCase("complete overlap", S(1, 2, 3), S(1, 2, 3), S[int]()),
	}

	RunTestCases(t, testCases)
}

// Test cases for SliceMinusFn function
type sliceMinusFnTestCase struct {
	equal    func(string, string) bool
	name     string
	a        []string
	b        []string
	expected []string
}

func (tc sliceMinusFnTestCase) Name() string {
	return tc.name
}

func (tc sliceMinusFnTestCase) Test(t *testing.T) {
	t.Helper()
	result := SliceMinusFn(tc.a, tc.b, tc.equal)
	AssertSliceEqual(t, tc.expected, result, "SliceMinusFn", tc.a, tc.b)
}

func newSliceMinusFnTestCase(name string, a, b, expected []string,
	equal func(string, string) bool) sliceMinusFnTestCase {
	return sliceMinusFnTestCase{
		name:     name,
		a:        a,
		b:        b,
		expected: expected,
		equal:    equal,
	}
}

func TestSliceMinusFn(t *testing.T) {
	equal := func(va, vb string) bool {
		return va == vb
	}

	testCases := []sliceMinusFnTestCase{
		newSliceMinusFnTestCase("empty slices", S[string](), S[string](), S[string](), equal),
		newSliceMinusFnTestCase("empty a", S[string](), S("x", "y"), S[string](), equal),
		newSliceMinusFnTestCase("empty b", S("x", "y"), S[string](), S("x", "y"), equal),
		newSliceMinusFnTestCase("no overlap", S("apple", "cherry"), S("banana", "date"), S("apple", "cherry"), equal),
		newSliceMinusFnTestCase("partial overlap", S("apple", "banana", "cherry"),
			S("banana", "date"), S("apple", "cherry"), equal),
		newSliceMinusFnTestCase("complete overlap", S("apple", "banana"), S("apple", "banana"), S[string](), equal),
	}

	RunTestCases(t, testCases)
}

// Test cases for SliceRandom function
type sliceRandomTestCase struct {
	name   string
	want   string
	input  []string
	wantOK bool
}

func (tc sliceRandomTestCase) Name() string {
	return tc.name
}

func (tc sliceRandomTestCase) Test(t *testing.T) {
	t.Helper()
	if tc.name != "random" {
		got, ok := SliceRandom(tc.input)
		AssertEqual(t, tc.wantOK, ok, "SliceRandom ok for %s", tc.name)
		AssertEqual(t, tc.want, got, "SliceRandom result mismatch for %s", tc.name)
	} else {
		u, ok := SliceRandom(tc.input)
		AssertEqual(t, tc.wantOK, ok, "SliceRandom ok for %s", tc.name)
		t.Logf("random from one,two,three,four,five,six: %s", u)
	}
}

func newSliceRandomTestCase(name string, input []string, want string, wantOK bool) sliceRandomTestCase {
	return sliceRandomTestCase{
		name:   name,
		input:  input,
		want:   want,
		wantOK: wantOK,
	}
}

func TestSliceRandom(t *testing.T) {
	testCases := []sliceRandomTestCase{
		newSliceRandomTestCase("empty", S[string](), "", false),
		newSliceRandomTestCase("one", S("one"), "one", true),
		newSliceRandomTestCase("random", S("one", "two", "three", "four", "five", "six"), "", true),
	}

	RunTestCases(t, testCases)
}

// Test cases for SliceContains function
type sliceContainsTestCase struct {
	name     string
	slice    []int
	value    int
	expected bool
}

func (tc sliceContainsTestCase) Name() string {
	return tc.name
}

func (tc sliceContainsTestCase) Test(t *testing.T) {
	t.Helper()

	result := SliceContains(tc.slice, tc.value)
	AssertEqual(t, tc.expected, result, "SliceContains result")
}

// Factory function for sliceContainsTestCase
func newSliceContainsTestCase(name string, slice []int, value int, expected bool) sliceContainsTestCase {
	return sliceContainsTestCase{
		name:     name,
		slice:    slice,
		value:    value,
		expected: expected,
	}
}

func TestSliceContains(t *testing.T) {
	testCases := []sliceContainsTestCase{
		newSliceContainsTestCase("empty slice", S[int](), 42, false),
		newSliceContainsTestCase("single element found", S(42), 42, true),
		newSliceContainsTestCase("single element not found", S(42), 99, false),
		newSliceContainsTestCase("multiple elements found", S(1, 2, 3, 42, 5), 42, true),
		newSliceContainsTestCase("multiple elements not found", S(1, 2, 3, 4, 5), 42, false),
		newSliceContainsTestCase("first element", S(42, 1, 2), 42, true),
		newSliceContainsTestCase("last element", S(1, 2, 42), 42, true),
		newSliceContainsTestCase("middle element", S(1, 42, 2), 42, true),
		newSliceContainsTestCase("duplicate elements", S(1, 42, 2, 42, 3), 42, true),
	}

	RunTestCases(t, testCases)
}

// Test cases for SliceMap function
type sliceMapTestCase[T1, T2 any] struct {
	name     string
	input    []T1
	fn       func([]T2, T1) []T2
	expected []T2
}

func (tc sliceMapTestCase[T1, T2]) Name() string {
	return tc.name
}

func (tc sliceMapTestCase[T1, T2]) Test(t *testing.T) {
	t.Helper()

	result := SliceMap(tc.input, tc.fn)
	AssertSliceEqual(t, tc.expected, result, "SliceMap")
}

// Factory function for sliceMapTestCase
func newSliceMapTestCase[T1, T2 any](name string, input []T1, fn func([]T2, T1) []T2,
	expected []T2) sliceMapTestCase[T1, T2] {
	return sliceMapTestCase[T1, T2]{
		name:     name,
		input:    input,
		fn:       fn,
		expected: expected,
	}
}

func TestSliceMap(t *testing.T) {
	// Simple transformation: returns one element per input
	intToString := func(_ []string, i int) []string {
		return S(fmt.Sprintf("num_%d", i))
	}

	// Simple case first
	t.Run("debug", testSliceMapDebug)

	// Test simple mapping
	testCases := []sliceMapTestCase[int, string]{
		newSliceMapTestCase("single element", S(42), intToString, S("num_42")),
		newSliceMapTestCase("multiple elements", S(1, 2, 3), intToString, S("num_1", "num_2", "num_3")),
		newSliceMapTestCase("negative numbers", S(-1, 0, 1), intToString, S("num_-1", "num_0", "num_1")),
	}

	// Test empty slice separately
	t.Run("empty slice", testSliceMapEmpty)

	RunTestCases(t, testCases)
}

func testSliceMapEmpty(t *testing.T) {
	t.Helper()
	intToString := func(_ []string, i int) []string {
		return S(fmt.Sprintf("num_%d", i))
	}
	result := SliceMap(S[int](), intToString)
	AssertEqual(t, 0, len(result), "result slice length")
}

func testSliceMapDebug(t *testing.T) {
	t.Helper()
	debug := func(partial []int, i int) []int {
		t.Logf("partial=%v, i=%d", partial, i)
		return S(i)
	}
	result := SliceMap(S(1, 2, 3), debug)
	t.Logf("result=%v", result)
}

// Test cases for SliceReversed function
type sliceReversedTestCase struct {
	name     string
	input    []int
	expected []int
}

func (tc sliceReversedTestCase) Name() string {
	return tc.name
}

func (tc sliceReversedTestCase) Test(t *testing.T) {
	t.Helper()

	result := SliceReversed(tc.input)
	AssertSliceEqual(t, tc.expected, result, "SliceReversed")

	// Verify original slice is unchanged
	originalCopy := SliceCopy(tc.input)
	AssertSliceEqual(t, originalCopy, tc.input, "original unchanged")
}

// Factory function for sliceReversedTestCase
func newSliceReversedTestCase(name string, input, expected []int) sliceReversedTestCase {
	return sliceReversedTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSliceReversed(t *testing.T) {
	testCases := []sliceReversedTestCase{
		newSliceReversedTestCase("empty slice", S[int](), S[int]()),
		newSliceReversedTestCase("single element", S(42), S(42)),
		newSliceReversedTestCase("two elements", S(1, 2), S(2, 1)),
		newSliceReversedTestCase("three elements", S(1, 2, 3), S(3, 2, 1)),
		newSliceReversedTestCase("four elements", S(1, 2, 3, 4), S(4, 3, 2, 1)),
		newSliceReversedTestCase("five elements", S(1, 2, 3, 4, 5), S(5, 4, 3, 2, 1)),
	}

	RunTestCases(t, testCases)
}

// Test cases for SliceReversedFn function
type sliceReversedFnTestCase[T any] struct {
	name     string
	input    []T
	fn       func([]T, T) (T, bool)
	expected []T
}

func (tc sliceReversedFnTestCase[T]) Name() string {
	return tc.name
}

func (tc sliceReversedFnTestCase[T]) Test(t *testing.T) {
	t.Helper()

	result := SliceReversedFn(tc.input, tc.fn)
	AssertSliceEqual(t, tc.expected, result, "SliceReversedFn")
}

// Factory function for sliceReversedFnTestCase
func newSliceReversedFnTestCase[T any](name string, input []T, fn func([]T, T) (T, bool),
	expected []T) sliceReversedFnTestCase[T] {
	return sliceReversedFnTestCase[T]{
		name:     name,
		input:    input,
		fn:       fn,
		expected: expected,
	}
}

func TestSliceReversedFn(t *testing.T) {
	double := func(_ []int, i int) (int, bool) { return i * 2, true }
	negate := func(_ []int, i int) (int, bool) { return -i, true }
	filter := func(_ []int, i int) (int, bool) { return i, i > 0 }

	testCases := []sliceReversedFnTestCase[int]{
		newSliceReversedFnTestCase("empty slice", S[int](), double, S[int]()),
		newSliceReversedFnTestCase("single element", S(42), double, S(84)),
		newSliceReversedFnTestCase("double and reverse", S(1, 2, 3), double, S(6, 4, 2)),
		newSliceReversedFnTestCase("negate and reverse", S(1, 2, 3), negate, S(-3, -2, -1)),
		newSliceReversedFnTestCase("filter and reverse", S(-1, 2, -3, 4), filter, S(4, 2)),
		newSliceReversedFnTestCase("identity function", S(1, 2, 3, 4),
			func(_ []int, i int) (int, bool) { return i, true }, S(4, 3, 2, 1)),
	}

	RunTestCases(t, testCases)
}

// Test cases for SliceSortFn function
type sliceSortFnTestCase[T any] struct {
	name     string
	input    []T
	less     func(T, T) bool
	expected []T
}

func (tc sliceSortFnTestCase[T]) Name() string {
	return tc.name
}

func (tc sliceSortFnTestCase[T]) Test(t *testing.T) {
	t.Helper()

	// Make a copy to avoid modifying the original
	result := SliceCopy(tc.input)
	SliceSortFn(result, tc.less)
	AssertSliceEqual(t, tc.expected, result, "SliceSortFn")
}

// Factory function for sliceSortFnTestCase
func newSliceSortFnTestCase[T any](name string, input []T, less func(T, T) bool, expected []T) sliceSortFnTestCase[T] {
	return sliceSortFnTestCase[T]{
		name:     name,
		input:    input,
		less:     less,
		expected: expected,
	}
}

func TestSliceSortFn(t *testing.T) {
	intLess := func(a, b int) bool { return a < b }
	intGreater := func(a, b int) bool { return a > b }
	stringLess := func(a, b string) bool { return a < b }

	testCases := []sliceSortFnTestCase[int]{
		newSliceSortFnTestCase("empty slice", S[int](), intLess, S[int]()),
		newSliceSortFnTestCase("single element", S(42), intLess, S(42)),
		newSliceSortFnTestCase("two elements ascending", S(2, 1), intLess, S(1, 2)),
		newSliceSortFnTestCase("two elements descending", S(1, 2), intGreater, S(2, 1)),
		newSliceSortFnTestCase("multiple elements", S(3, 1, 4, 1, 5), intLess, S(1, 1, 3, 4, 5)),
		newSliceSortFnTestCase("already sorted", S(1, 2, 3, 4, 5), intLess, S(1, 2, 3, 4, 5)),
		newSliceSortFnTestCase("reverse sorted", S(5, 4, 3, 2, 1), intLess, S(1, 2, 3, 4, 5)),
		newSliceSortFnTestCase("duplicates", S(3, 1, 3, 1, 3), intLess, S(1, 1, 3, 3, 3)),
		newSliceSortFnTestCase("negative numbers", S(-1, 3, -2, 0, 1), intLess, S(-2, -1, 0, 1, 3)),
	}

	RunTestCases(t, testCases)

	// Test with string slice
	stringTests := []sliceSortFnTestCase[string]{
		newSliceSortFnTestCase("empty string slice", S[string](), stringLess, S[string]()),
		newSliceSortFnTestCase("single string", S("hello"), stringLess, S("hello")),
		newSliceSortFnTestCase("multiple strings", S("cherry", "apple", "banana"), stringLess,
			S("apple", "banana", "cherry")),
		newSliceSortFnTestCase("string duplicates", S("b", "a", "b", "c", "a"), stringLess, S("a", "a", "b", "b", "c")),
	}

	RunTestCases(t, stringTests)

	// Test edge cases
	t.Run("nil less function", testSliceSortFnNilFunction)
}

func testSliceSortFnNilFunction(t *testing.T) {
	t.Helper()
	original := S(3, 1, 2)
	result := SliceCopy(original)
	SliceSortFn(result, nil)
	AssertSliceEqual(t, original, result, "nil function")
}

// Test cases for SliceSortOrdered function
type sliceSortOrderedTestCase[T Ordered] struct {
	name     string
	input    []T
	expected []T
}

func (tc sliceSortOrderedTestCase[T]) Name() string {
	return tc.name
}

func (tc sliceSortOrderedTestCase[T]) Test(t *testing.T) {
	t.Helper()

	// Make a copy to avoid modifying the original
	result := SliceCopy(tc.input)
	SliceSortOrdered(result)
	AssertSliceEqual(t, tc.expected, result, "SliceSortOrdered")
}

// Factory function for sliceSortOrderedTestCase
func newSliceSortOrderedTestCase[T Ordered](name string, input, expected []T) sliceSortOrderedTestCase[T] {
	return sliceSortOrderedTestCase[T]{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSliceSortOrdered(t *testing.T) {
	// Test with int
	intTestCases := []sliceSortOrderedTestCase[int]{
		newSliceSortOrderedTestCase("empty int slice", S[int](), S[int]()),
		newSliceSortOrderedTestCase("single int", S(42), S(42)),
		newSliceSortOrderedTestCase("two ints", S(2, 1), S(1, 2)),
		newSliceSortOrderedTestCase("multiple ints", S(3, 1, 4, 1, 5), S(1, 1, 3, 4, 5)),
		newSliceSortOrderedTestCase("already sorted ints", S(1, 2, 3, 4, 5), S(1, 2, 3, 4, 5)),
		newSliceSortOrderedTestCase("reverse sorted ints", S(5, 4, 3, 2, 1), S(1, 2, 3, 4, 5)),
		newSliceSortOrderedTestCase("negative ints", S(-1, 3, -2, 0, 1), S(-2, -1, 0, 1, 3)),
	}

	RunTestCases(t, intTestCases)

	// Test with string
	stringTestCases := []sliceSortOrderedTestCase[string]{
		newSliceSortOrderedTestCase("empty string slice", S[string](), S[string]()),
		newSliceSortOrderedTestCase("single string", S("hello"), S("hello")),
		newSliceSortOrderedTestCase("multiple strings", S("cherry", "apple", "banana"), S("apple", "banana", "cherry")),
		newSliceSortOrderedTestCase("string duplicates", S("b", "a", "b", "c", "a"), S("a", "a", "b", "b", "c")),
		newSliceSortOrderedTestCase("empty strings", S("", "b", "", "a"), S("", "", "a", "b")),
	}

	RunTestCases(t, stringTestCases)

	// Test with float64
	floatTestCases := []sliceSortOrderedTestCase[float64]{
		newSliceSortOrderedTestCase("empty float slice", S[float64](), S[float64]()),
		newSliceSortOrderedTestCase("single float", S(3.14), S(3.14)),
		newSliceSortOrderedTestCase("multiple floats", S(3.14, 1.41, 2.71), S(1.41, 2.71, 3.14)),
		newSliceSortOrderedTestCase("float duplicates", S(1.0, 2.0, 1.0, 3.0), S(1.0, 1.0, 2.0, 3.0)),
		newSliceSortOrderedTestCase("negative floats", S(-1.5, 2.5, -0.5, 0.0), S(-1.5, -0.5, 0.0, 2.5)),
	}

	RunTestCases(t, floatTestCases)
}
