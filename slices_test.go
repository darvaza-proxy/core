package core_test

import (
	"cmp"
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = sliceReverseTestCase{}
	_ core.TestCase = sliceContainsTestCase{}
	_ core.TestCase = sliceMapTestCase[int, string]{}
	_ core.TestCase = sliceReversedTestCase{}
	_ core.TestCase = sliceReversedFnTestCase[int]{}
	_ core.TestCase = sliceSortFnTestCase[int]{}
	_ core.TestCase = sliceSortTestCase{}
	_ core.TestCase = sliceSortOrderedTestCase[int]{}
	_ core.TestCase = sliceMinusTestCase{}
	_ core.TestCase = sliceMinusFnTestCase{}
	_ core.TestCase = sliceRandomTestCase{}
	_ core.TestCase = slicePIntTestCase{}
	_ core.TestCase = slicePFloatTestCase{}
	_ core.TestCase = slicePStringTestCase{}
	_ core.TestCase = sliceReplaceFnNilFnTestCase{}
)

type sliceReverseTestCase struct {
	name string
	a, b []int
}

func (tc sliceReverseTestCase) Name() string {
	return tc.name
}

func (tc sliceReverseTestCase) Test(t *testing.T) {
	c := core.SliceCopy(tc.a)
	core.SliceReverse(c)
	core.AssertSliceEqual(t, tc.b, c, "SliceReverse(%v)", tc.a)
	if core.SliceEqual(c, tc.b) {
		t.Logf("%s(%v) → %v", "SliceReverse", tc.a, c)
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
	newSliceReverseTestCase("empty", core.S[int](), core.S[int]()),
	newSliceReverseTestCase("single", core.S(1), core.S(1)),
	newSliceReverseTestCase("two elements", core.S(1, 2), core.S(2, 1)),
	newSliceReverseTestCase("three elements", core.S(1, 2, 3), core.S(3, 2, 1)),
	newSliceReverseTestCase("four elements", core.S(1, 2, 3, 4), core.S(4, 3, 2, 1)),
	newSliceReverseTestCase("five elements", core.S(1, 2, 3, 4, 5), core.S(5, 4, 3, 2, 1)),
	newSliceReverseTestCase("six elements", core.S(1, 2, 3, 4, 5, 6), core.S(6, 5, 4, 3, 2, 1)),
}

func TestSliceReverse(t *testing.T) {
	core.RunTestCases(t, sliceReverseTestCases)
}

var (
	ints       = core.S(74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586)
	expectInts = core.S(74, 59, 238, -784, 9845, 959, 905, 0, 42, 7586, -5467984)

	float64s = core.S(
		74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, 7.8, 7.8, 74.3,
		59.0, math.Inf(1), 238.2, -784.0, 2.3,
	)
	expectFloat64s = core.S(74.3, 59.0, math.Inf(1), 238.2, -784, 2.3, 7.8)

	strs       = core.S("", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&")
	expectStrs = core.S("", "Hello", "foo", "bar", "f00", "%*&^*&^&")
)

func eq[T core.Ordered](a, b T) bool {
	return a == b
}

// The unique helpers keep the first occurrence of each element in its
// place, so the expectations are in input order and compared as they are.
func testSliceUnique[T core.Ordered](t *testing.T, before, after []T) {
	t.Helper()

	s := core.SliceUnique(before)
	core.AssertSliceEqual(t, after, s, "SliceUnique")

	s = core.SliceUniqueFn(before, eq[T])
	core.AssertSliceEqual(t, after, s, "SliceUniqueFn")

	s = core.SliceCopyFn(before, nil)
	s2 := core.SliceUniquify(&s)
	core.AssertSliceEqual(t, after, s, "SliceUniquify")
	core.AssertSliceEqual(t, after, s2, "return value")
	core.AssertSame(t, s, s2, "same slice")

	s = core.SliceCopy(before)
	s2 = core.SliceUniquifyFn(&s, eq[T])
	core.AssertSliceEqual(t, after, s, "SliceUniquifyFn")
	core.AssertSliceEqual(t, after, s2, "return value")
	core.AssertSame(t, s, s2, "same slice")
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
	result := core.SliceMinus(tc.a, tc.b)
	core.AssertSliceEqual(t, tc.expected, result, "SliceMinus(%v, %v)", tc.a, tc.b)
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
		newSliceMinusTestCase("empty slices", core.S[int](), core.S[int](), core.S[int]()),
		newSliceMinusTestCase("empty a", core.S[int](), core.S(1, 2), core.S[int]()),
		newSliceMinusTestCase("empty b", core.S(1, 2), core.S[int](), core.S(1, 2)),
		newSliceMinusTestCase("no overlap", core.S(1, 2), core.S(3, 4), core.S(1, 2)),
		newSliceMinusTestCase("partial overlap", core.S(1, 2, 3, 4, 5), core.S(3, 4, 6, 7), core.S(1, 2, 5)),
		newSliceMinusTestCase("complete overlap", core.S(1, 2, 3), core.S(1, 2, 3), core.S[int]()),
	}

	core.RunTestCases(t, testCases)
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
	result := core.SliceMinusFn(tc.a, tc.b, tc.equal)
	core.AssertSliceEqual(t, tc.expected, result, "SliceMinusFn(%q, %q, _)", tc.a, tc.b)
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
		newSliceMinusFnTestCase("empty slices", core.S[string](), core.S[string](), core.S[string](), equal),
		newSliceMinusFnTestCase("empty a", core.S[string](), core.S("x", "y"), core.S[string](), equal),
		newSliceMinusFnTestCase("empty b", core.S("x", "y"), core.S[string](), core.S("x", "y"), equal),
		newSliceMinusFnTestCase("no overlap", core.S("apple", "cherry"),
			core.S("banana", "date"), core.S("apple", "cherry"), equal),
		newSliceMinusFnTestCase("partial overlap", core.S("apple", "banana", "cherry"),
			core.S("banana", "date"), core.S("apple", "cherry"), equal),
		newSliceMinusFnTestCase("complete overlap", core.S("apple", "banana"),
			core.S("apple", "banana"), core.S[string](), equal),
	}

	core.RunTestCases(t, testCases)
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
	if tc.name != "random" {
		got, ok := core.SliceRandom(tc.input)
		core.AssertEqual(t, tc.wantOK, ok, "SliceRandom ok for %s", tc.name)
		core.AssertEqual(t, tc.want, got, "SliceRandom result mismatch for %s", tc.name)
	} else {
		u, ok := core.SliceRandom(tc.input)
		core.AssertEqual(t, tc.wantOK, ok, "SliceRandom ok for %s", tc.name)
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
		newSliceRandomTestCase("empty", core.S[string](), "", false),
		newSliceRandomTestCase("one", core.S("one"), "one", true),
		newSliceRandomTestCase("random", core.S("one", "two", "three", "four", "five", "six"), "", true),
	}

	core.RunTestCases(t, testCases)
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
	result := core.SliceContains(tc.slice, tc.value)
	core.AssertEqual(t, tc.expected, result, "SliceContains result")
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
		newSliceContainsTestCase("empty slice", core.S[int](), 42, false),
		newSliceContainsTestCase("single element found", core.S(42), 42, true),
		newSliceContainsTestCase("single element not found", core.S(42), 99, false),
		newSliceContainsTestCase("multiple elements found", core.S(1, 2, 3, 42, 5), 42, true),
		newSliceContainsTestCase("multiple elements not found", core.S(1, 2, 3, 4, 5), 42, false),
		newSliceContainsTestCase("first element", core.S(42, 1, 2), 42, true),
		newSliceContainsTestCase("last element", core.S(1, 2, 42), 42, true),
		newSliceContainsTestCase("middle element", core.S(1, 42, 2), 42, true),
		newSliceContainsTestCase("duplicate elements", core.S(1, 42, 2, 42, 3), 42, true),
	}

	core.RunTestCases(t, testCases)
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
	result := core.SliceMap(tc.input, tc.fn)
	core.AssertSliceEqual(t, tc.expected, result, "SliceMap")
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
	t.Run("int to string", runTestSliceMapIntToString)
	t.Run("string to int", runTestSliceMapStringToInt)
}

func runTestSliceMapIntToString(t *testing.T) {
	// Simple transformation: returns one element per input
	intToString := func(_ []string, i int) []string {
		return core.S(fmt.Sprintf("num_%d", i))
	}

	testCases := []sliceMapTestCase[int, string]{
		newSliceMapTestCase("single element", core.S(42), intToString, core.S("num_42")),
		newSliceMapTestCase("multiple elements", core.S(1, 2, 3), intToString, core.S("num_1", "num_2", "num_3")),
		newSliceMapTestCase("negative numbers", core.S(-1, 0, 1), intToString, core.S("num_-1", "num_0", "num_1")),
		newSliceMapTestCase("empty slice", core.S[int](), intToString, nil),
		newSliceMapTestCase("nil slice", nil, intToString, nil),
		newSliceMapTestCase[int, string]("nil function", core.S(1, 2), nil, nil),
	}

	core.RunTestCases(t, testCases)
}

func runTestSliceMapStringToInt(t *testing.T) {
	// Maps each string to its length and drops the empty ones, so a row
	// can state a result shorter than its input and one that is nil
	// though both the input and the function are not.
	lengths := func(_ []int, s string) []int {
		if s == "" {
			return nil
		}
		return core.S(len(s))
	}

	// Adds each length to the sum of the result so far, so a row can
	// state that partial holds what was mapped before the element.
	running := func(partial []int, s string) []int {
		var sum int
		for _, n := range partial {
			sum += n
		}
		return core.S(sum + len(s))
	}

	testCases := []sliceMapTestCase[string, int]{
		newSliceMapTestCase("single element", core.S("hello"), lengths, core.S(5)),
		newSliceMapTestCase("multiple elements", core.S("a", "bb", "ccc"), lengths, core.S(1, 2, 3)),
		newSliceMapTestCase("running total", core.S("a", "bb", "ccc"), running, core.S(1, 3, 7)),
		newSliceMapTestCase("element dropped", core.S("a", "", "ccc"), lengths, core.S(1, 3)),
		newSliceMapTestCase("every element dropped", core.S("", ""), lengths, nil),
		newSliceMapTestCase("empty slice", core.S[string](), lengths, nil),
		newSliceMapTestCase("nil slice", nil, lengths, nil),
		newSliceMapTestCase[string, int]("nil function", core.S("a"), nil, nil),
	}

	core.RunTestCases(t, testCases)
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
	original := core.SliceCopy(tc.input)

	result := core.SliceReversed(tc.input)
	core.AssertSliceEqual(t, tc.expected, result, "SliceReversed")
	core.AssertSliceEqual(t, original, tc.input, "original unchanged")
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
		newSliceReversedTestCase("empty slice", core.S[int](), core.S[int]()),
		newSliceReversedTestCase("single element", core.S(42), core.S(42)),
		newSliceReversedTestCase("two elements", core.S(1, 2), core.S(2, 1)),
		newSliceReversedTestCase("three elements", core.S(1, 2, 3), core.S(3, 2, 1)),
		newSliceReversedTestCase("four elements", core.S(1, 2, 3, 4), core.S(4, 3, 2, 1)),
		newSliceReversedTestCase("five elements", core.S(1, 2, 3, 4, 5), core.S(5, 4, 3, 2, 1)),
	}

	core.RunTestCases(t, testCases)
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
	result := core.SliceReversedFn(tc.input, tc.fn)
	core.AssertSliceEqual(t, tc.expected, result, "SliceReversedFn")
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
		newSliceReversedFnTestCase("empty slice", core.S[int](), double, core.S[int]()),
		newSliceReversedFnTestCase("single element", core.S(42), double, core.S(84)),
		newSliceReversedFnTestCase("double and reverse", core.S(1, 2, 3), double, core.S(6, 4, 2)),
		newSliceReversedFnTestCase("negate and reverse", core.S(1, 2, 3), negate, core.S(-3, -2, -1)),
		newSliceReversedFnTestCase("filter and reverse", core.S(-1, 2, -3, 4), filter, core.S(4, 2)),
		newSliceReversedFnTestCase("identity function", core.S(1, 2, 3, 4),
			func(_ []int, i int) (int, bool) { return i, true }, core.S(4, 3, 2, 1)),
	}

	core.RunTestCases(t, testCases)
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
	// Make a copy to avoid modifying the original
	result := core.SliceCopy(tc.input)
	core.SliceSortFn(result, tc.less)
	core.AssertSliceEqual(t, tc.expected, result, "SliceSortFn")
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
		newSliceSortFnTestCase("empty slice", core.S[int](), intLess, core.S[int]()),
		newSliceSortFnTestCase("single element", core.S(42), intLess, core.S(42)),
		newSliceSortFnTestCase("two elements ascending", core.S(2, 1), intLess, core.S(1, 2)),
		newSliceSortFnTestCase("two elements descending", core.S(1, 2), intGreater, core.S(2, 1)),
		newSliceSortFnTestCase("multiple elements", core.S(3, 1, 4, 1, 5), intLess, core.S(1, 1, 3, 4, 5)),
		newSliceSortFnTestCase("already sorted", core.S(1, 2, 3, 4, 5), intLess, core.S(1, 2, 3, 4, 5)),
		newSliceSortFnTestCase("reverse sorted", core.S(5, 4, 3, 2, 1), intLess, core.S(1, 2, 3, 4, 5)),
		newSliceSortFnTestCase("duplicates", core.S(3, 1, 3, 1, 3), intLess, core.S(1, 1, 3, 3, 3)),
		newSliceSortFnTestCase("negative numbers", core.S(-1, 3, -2, 0, 1), intLess, core.S(-2, -1, 0, 1, 3)),
	}

	core.RunTestCases(t, testCases)

	// Test with string slice
	stringTests := []sliceSortFnTestCase[string]{
		newSliceSortFnTestCase("empty string slice", core.S[string](), stringLess, core.S[string]()),
		newSliceSortFnTestCase("single string", core.S("hello"), stringLess, core.S("hello")),
		newSliceSortFnTestCase("multiple strings", core.S("cherry", "apple", "banana"), stringLess,
			core.S("apple", "banana", "cherry")),
		newSliceSortFnTestCase("string duplicates", core.S("b", "a", "b", "c", "a"), stringLess,
			core.S("a", "a", "b", "b", "c")),
	}

	core.RunTestCases(t, stringTests)

	// Test edge cases
	t.Run("nil less function", testSliceSortFnNilFunction)
}

func testSliceSortFnNilFunction(t *testing.T) {
	original := core.S(3, 1, 2)
	result := core.SliceCopy(original)
	core.SliceSortFn(result, nil)
	core.AssertSliceEqual(t, original, result, "nil function")
}

// sliceSortTestCase states what SliceSort leaves in a copy of input
// after sorting it with compare.
type sliceSortTestCase struct {
	compare  func(a, b int) int
	name     string
	input    []int
	expected []int
}

func (tc sliceSortTestCase) Name() string {
	return tc.name
}

func (tc sliceSortTestCase) Test(t *testing.T) {
	result := core.SliceCopy(tc.input)
	core.SliceSort(result, tc.compare)
	core.AssertSliceEqual(t, tc.expected, result, "SliceSort")
}

func newSliceSortTestCase(name string, input []int,
	compare func(a, b int) int, expected []int) sliceSortTestCase {
	return sliceSortTestCase{
		compare:  compare,
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSliceSort(t *testing.T) {
	unsorted := core.S(3, 1, 2)
	ascending := cmp.Compare[int]

	core.RunTestCases(t, core.S(
		newSliceSortTestCase("ascending", unsorted, ascending, core.S(1, 2, 3)),
		// a nil compare leaves the slice as it is
		newSliceSortTestCase("nil compare", unsorted, nil, core.S(3, 1, 2)),
	))
}

// Test cases for SliceSortOrdered function
type sliceSortOrderedTestCase[T core.Ordered] struct {
	name     string
	input    []T
	expected []T
}

func (tc sliceSortOrderedTestCase[T]) Name() string {
	return tc.name
}

func (tc sliceSortOrderedTestCase[T]) Test(t *testing.T) {
	// Make a copy to avoid modifying the original
	result := core.SliceCopy(tc.input)
	core.SliceSortOrdered(result)
	core.AssertSliceEqual(t, tc.expected, result, "SliceSortOrdered")
}

// Factory function for sliceSortOrderedTestCase
func newSliceSortOrderedTestCase[T core.Ordered](name string, input, expected []T) sliceSortOrderedTestCase[T] {
	return sliceSortOrderedTestCase[T]{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSliceSortOrdered(t *testing.T) {
	// Test with int
	intTestCases := []sliceSortOrderedTestCase[int]{
		newSliceSortOrderedTestCase("empty int slice", core.S[int](), core.S[int]()),
		newSliceSortOrderedTestCase("single int", core.S(42), core.S(42)),
		newSliceSortOrderedTestCase("two ints", core.S(2, 1), core.S(1, 2)),
		newSliceSortOrderedTestCase("multiple ints", core.S(3, 1, 4, 1, 5), core.S(1, 1, 3, 4, 5)),
		newSliceSortOrderedTestCase("already sorted ints", core.S(1, 2, 3, 4, 5), core.S(1, 2, 3, 4, 5)),
		newSliceSortOrderedTestCase("reverse sorted ints", core.S(5, 4, 3, 2, 1), core.S(1, 2, 3, 4, 5)),
		newSliceSortOrderedTestCase("negative ints", core.S(-1, 3, -2, 0, 1), core.S(-2, -1, 0, 1, 3)),
	}

	core.RunTestCases(t, intTestCases)

	// Test with string
	stringTestCases := []sliceSortOrderedTestCase[string]{
		newSliceSortOrderedTestCase("empty string slice", core.S[string](), core.S[string]()),
		newSliceSortOrderedTestCase("single string", core.S("hello"), core.S("hello")),
		newSliceSortOrderedTestCase("multiple strings", core.S("cherry", "apple", "banana"),
			core.S("apple", "banana", "cherry")),
		newSliceSortOrderedTestCase("string duplicates", core.S("b", "a", "b", "c", "a"),
			core.S("a", "a", "b", "b", "c")),
		newSliceSortOrderedTestCase("empty strings", core.S("", "b", "", "a"), core.S("", "", "a", "b")),
	}

	core.RunTestCases(t, stringTestCases)

	// Test with float64
	floatTestCases := []sliceSortOrderedTestCase[float64]{
		newSliceSortOrderedTestCase("empty float slice", core.S[float64](), core.S[float64]()),
		newSliceSortOrderedTestCase("single float", core.S(3.14), core.S(3.14)),
		newSliceSortOrderedTestCase("multiple floats", core.S(3.14, 1.41, 2.71), core.S(1.41, 2.71, 3.14)),
		newSliceSortOrderedTestCase("float duplicates", core.S(1.0, 2.0, 1.0, 3.0), core.S(1.0, 1.0, 2.0, 3.0)),
		newSliceSortOrderedTestCase("negative floats", core.S(-1.5, 2.5, -0.5, 0.0), core.S(-1.5, -0.5, 0.0, 2.5)),
	}

	core.RunTestCases(t, floatTestCases)
}

// Test cases for SliceP function with int
type slicePIntTestCase struct {
	name     string
	input    []int
	p        float64
	expected int
}

func (tc slicePIntTestCase) Name() string {
	return tc.name
}

func (tc slicePIntTestCase) Test(t *testing.T) {
	// Create a copy to test with unsorted data
	inputCopy := core.SliceCopy(tc.input)

	result := core.SliceP(inputCopy, tc.p)
	core.AssertEqual(t, tc.expected, result, "SliceP(%v, %.2f)", tc.input, tc.p)
}

// Factory function for slicePIntTestCase
func newSlicePIntTestCase(name string, input []int, p float64, expected int) slicePIntTestCase {
	return slicePIntTestCase{
		name:     name,
		input:    input,
		p:        p,
		expected: expected,
	}
}

// Test cases for SliceP function with float64
type slicePFloatTestCase struct {
	name     string
	input    []float64
	p        float64
	expected float64
}

func (tc slicePFloatTestCase) Name() string {
	return tc.name
}

func (tc slicePFloatTestCase) Test(t *testing.T) {
	// Create a copy to test with unsorted data
	inputCopy := core.SliceCopy(tc.input)

	result := core.SliceP(inputCopy, tc.p)
	core.AssertEqual(t, tc.expected, result, "SliceP(%v, %.2f)", tc.input, tc.p)
}

// Factory function for slicePFloatTestCase
func newSlicePFloatTestCase(name string, input []float64, p, expected float64) slicePFloatTestCase {
	return slicePFloatTestCase{
		name:     name,
		input:    input,
		p:        p,
		expected: expected,
	}
}

// Test cases for SliceP function with string
type slicePStringTestCase struct {
	name     string
	expected string
	input    []string
	p        float64
}

func (tc slicePStringTestCase) Name() string {
	return tc.name
}

func (tc slicePStringTestCase) Test(t *testing.T) {
	// Create a copy to test with unsorted data
	inputCopy := core.SliceCopy(tc.input)

	result := core.SliceP(inputCopy, tc.p)
	core.AssertEqual(t, tc.expected, result, "SliceP(%v, %.2f)", tc.input, tc.p)
}

// Factory function for slicePStringTestCase
func newSlicePStringTestCase(name string, input []string, p float64, expected string) slicePStringTestCase {
	return slicePStringTestCase{
		name:     name,
		input:    input,
		p:        p,
		expected: expected,
	}
}

func slicePIntTestCases() []core.TestCase {
	return []core.TestCase{
		// Empty slice
		newSlicePIntTestCase("empty slice", core.S[int](), 0.5, 0),

		// Single element
		newSlicePIntTestCase("single element p=0.0", core.S(42), 0.0, 42),
		newSlicePIntTestCase("single element p=0.5", core.S(42), 0.5, 42),
		newSlicePIntTestCase("single element p=1.0", core.S(42), 1.0, 42),

		// Common percentiles
		newSlicePIntTestCase("P50 (median)", core.S(5, 2, 8, 1, 9), 0.50, 5),
		newSlicePIntTestCase("P95", core.S(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 0.95, 10),
		newSlicePIntTestCase("P99", core.S(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 0.99, 10),

		// Min and max
		newSlicePIntTestCase("P0 (min)", core.S(9, 3, 7, 1, 5), 0.0, 1),
		newSlicePIntTestCase("P100 (max)", core.S(9, 3, 7, 1, 5), 1.0, 9),

		// Unsorted data
		newSlicePIntTestCase("unsorted P25", core.S(10, 5, 8, 3, 1, 9, 4, 7, 2, 6), 0.25, 3),
		newSlicePIntTestCase("unsorted P75", core.S(10, 5, 8, 3, 1, 9, 4, 7, 2, 6), 0.75, 8),

		// Negative numbers
		newSlicePIntTestCase("negative P50", core.S(-5, -2, -8, -1, -9), 0.5, -5),
		newSlicePIntTestCase("mixed signs P50", core.S(-2, 5, -1, 3, 0), 0.5, 0),

		// Duplicates
		newSlicePIntTestCase("duplicates P50", core.S(3, 1, 3, 1, 3, 1, 3), 0.5, 3),

		// Invalid percentiles
		newSlicePIntTestCase("invalid p=-0.1", core.S(1, 2, 3, 4, 5), -0.1, 0),
		newSlicePIntTestCase("invalid p=1.5", core.S(1, 2, 3, 4, 5), 1.5, 0),
	}
}

func slicePFloatTestCases() []core.TestCase {
	return []core.TestCase{
		newSlicePFloatTestCase("float P50", core.S(3.14, 1.41, 2.71, 0.5, 4.0), 0.5, 2.71),
		newSlicePFloatTestCase("float P90", core.S(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.0), 0.9, 9.9),
		newSlicePFloatTestCase("float empty", core.S[float64](), 0.5, 0.0),
	}
}

func slicePStringTestCases() []core.TestCase {
	return []core.TestCase{
		newSlicePStringTestCase("string P50", core.S("cherry", "apple", "banana", "date"), 0.5, "banana"),
		newSlicePStringTestCase("string P0", core.S("zebra", "alpha", "beta"), 0.0, "alpha"),
		newSlicePStringTestCase("string P100", core.S("zebra", "alpha", "beta"), 1.0, "zebra"),
	}
}

func TestSliceP(t *testing.T) {
	core.RunTestCases(t, slicePIntTestCases())
	core.RunTestCases(t, slicePFloatTestCases())
	core.RunTestCases(t, slicePStringTestCases())
}

// Cover the short-circuit branches of SliceEqual/SliceEqualFn.
func TestSliceEqualShortCircuits(t *testing.T) {
	a := core.S(1, 2, 3)
	b := core.S(1, 2)
	core.AssertFalse(t, core.SliceEqual(a, b), "different lengths")
	core.AssertFalse(t, core.SliceEqualFn(a, b, func(x, y int) bool { return x == y }),
		"SliceEqualFn different lengths")
	core.AssertFalse(t, core.SliceEqualFn(a, a, nil), "SliceEqualFn nil comparator")
	core.AssertFalse(t, core.SliceEqual(core.S(1, 2, 3), core.S(1, 2, 4)), "values differ")
}

// Cover nil-pointer early-returns and nil-fn NO-OP in slice helpers.
func TestSliceUniquifyNilPtr(t *testing.T) {
	result := core.SliceUniquify[int](nil)
	core.AssertEqual(t, 0, len(result), "SliceUniquify nil ptr returns empty slice")

	result = core.SliceUniquifyFn[int](nil, func(x, y int) bool { return x == y })
	core.AssertEqual(t, 0, len(result), "SliceUniquifyFn nil ptr returns empty slice")
}

type sliceReplaceFnNilFnTestCase struct {
	name string
	in   []int
}

func (tc sliceReplaceFnNilFnTestCase) Name() string { return tc.name }

// A nil function returns the slice it was given, untouched.
func (tc sliceReplaceFnNilFnTestCase) Test(t *testing.T) {
	original := core.SliceCopy(tc.in)

	result := core.SliceReplaceFn(tc.in, nil)
	core.AssertSame(t, tc.in, result, "same slice")
	core.AssertSliceEqual(t, original, result, "unchanged")
}

func newSliceReplaceFnNilFnTestCase(name string, in []int) sliceReplaceFnNilFnTestCase {
	return sliceReplaceFnNilFnTestCase{name: name, in: in}
}

func TestSliceReplaceFnNilFn(t *testing.T) {
	core.RunTestCases(t, []sliceReplaceFnNilFnTestCase{
		newSliceReplaceFnNilFnTestCase("non-empty", core.S(1, 2, 3)),
		// IsSame compares a slice by its backing array, which an empty
		// slice only has when it was given capacity.
		newSliceReplaceFnNilFnTestCase("empty", make([]int, 0, 1)),
	})
}
