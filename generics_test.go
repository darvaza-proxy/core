package core_test

import (
	"fmt"
	"math"
	"testing"

	"darvaza.org/core"
)

// TestCase interface validations
var _ core.TestCase = typeNameTestCase[int]{}
var _ core.TestCase = coalesceTestCase[int]{}
var _ core.TestCase = coalescePointerTestCase[int]{}
var _ core.TestCase = iifTestCase{}
var _ core.TestCase = iifStringTestCase{}
var _ core.TestCase = iifPointerTestCase{}
var _ core.TestCase = iifStructTestCase{}

// typeNameTestCase states what TypeName renders for one type.
type typeNameTestCase[T any] struct {
	want string
	name string
}

func newTypeNameTestCase[T any](name, want string) core.TestCase {
	return typeNameTestCase[T]{
		want: want,
		name: name,
	}
}

func (tc typeNameTestCase[T]) Name() string {
	return tc.name
}

func (tc typeNameTestCase[T]) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, core.TypeName[T](), "name")
}

func typeNameTestCases() []core.TestCase {
	return core.S(
		newTypeNameTestCase[int]("builtin", "int"),
		newTypeNameTestCase[core.PanicError]("struct", "core.PanicError"),
		newTypeNameTestCase[*core.PanicError]("pointer", "*core.PanicError"),
		newTypeNameTestCase[[]string]("slice", "[]string"),
		newTypeNameTestCase[error]("builtin interface", "error"),
		newTypeNameTestCase[fmt.Stringer]("interface", "fmt.Stringer"),
		newTypeNameTestCase[core.Recovered]("package interface", "core.Recovered"),
		newTypeNameTestCase[interface{ Len() int }]("anonymous interface",
			"interface { Len() int }"),
		newTypeNameTestCase[any]("any", "any"),
		newTypeNameTestCase[[]any]("any slice", "[]any"),
	)
}

func TestTypeName(t *testing.T) {
	core.RunTestCases(t, typeNameTestCases())
}

// coalesceTestCase tests Coalesce function with generic type support
type coalesceTestCase[T comparable] struct {
	expected T
	name     string
	inputs   []T
}

func (tc coalesceTestCase[T]) Name() string {
	return tc.name
}

func (tc coalesceTestCase[T]) Test(t *testing.T) {
	t.Helper()

	got := core.Coalesce(tc.inputs...)
	core.AssertEqual(t, tc.expected, got, "Coalesce")
}

// Factory function for coalesceTestCase
func newCoalesceTestCase[T comparable](name string, inputs []T, expected T) coalesceTestCase[T] {
	return coalesceTestCase[T]{
		name:     name,
		inputs:   inputs,
		expected: expected,
	}
}

var coalesceIntTestCases = []coalesceTestCase[int]{
	newCoalesceTestCase("all zeros", core.S(0, 0, 0), 0),
	newCoalesceTestCase("first non-zero", core.S(0, 42, 0, 100), 42),
	newCoalesceTestCase("last non-zero", core.S(0, 0, 0, 100), 100),
	newCoalesceTestCase("single value", core.S(42), 42),
	newCoalesceTestCase("single zero", core.S(0), 0),
	newCoalesceTestCase("empty inputs", core.S[int](), 0),
	newCoalesceTestCase("negative values", core.S(0, -42, 0, -100), -42),
	newCoalesceTestCase("mixed positive and negative", core.S(0, -1, 0, 1), -1),
	newCoalesceTestCase("large numbers", core.S(0, 0, 2147483647, 0), 2147483647),
	newCoalesceTestCase("no zeros", core.S(1, 2, 3, 4), 1),
}

func TestCoalesceInt(t *testing.T) {
	core.RunTestCases(t, coalesceIntTestCases)
}

var coalesceStringTestCases = []coalesceTestCase[string]{
	newCoalesceTestCase("all empty", core.S("", "", ""), ""),
	newCoalesceTestCase("first non-empty", core.S("", "hello", "", "world"), "hello"),
	newCoalesceTestCase("last non-empty", core.S("", "", "", "world"), "world"),
	newCoalesceTestCase("single value", core.S("hello"), "hello"),
	newCoalesceTestCase("single empty", core.S(""), ""),
	newCoalesceTestCase("empty inputs", core.S[string](), ""),
	newCoalesceTestCase("spaces are not empty", core.S("", " ", "", "world"), " "),
	newCoalesceTestCase("unicode strings", core.S("", "🚀", "", "world"), "🚀"),
	newCoalesceTestCase("long strings",
		core.S("", "a very long string that should not be truncated", "", "short"),
		"a very long string that should not be truncated"),
	newCoalesceTestCase("newlines and tabs", core.S("", "\n\t", "", "world"), "\n\t"),
	newCoalesceTestCase("no empty strings", core.S("first", "second", "third"), "first"),
}

func TestCoalesceString(t *testing.T) {
	core.RunTestCases(t, coalesceStringTestCases)
}

// coalescePointerTestCase tests Coalesce with pointers (requires special comparison logic)
type coalescePointerTestCase[T comparable] struct {
	expected *T
	name     string
	inputs   []*T
}

func (tc coalescePointerTestCase[T]) Name() string {
	return tc.name
}

func (tc coalescePointerTestCase[T]) Test(t *testing.T) {
	t.Helper()

	got := core.Coalesce(tc.inputs...)
	core.AssertSame(t, tc.expected, got, "Coalesce")
}

// Factory function for coalescePointerTestCase
func newCoalescePointerTestCase[T comparable](name string, inputs []*T, expected *T) coalescePointerTestCase[T] {
	return coalescePointerTestCase[T]{
		name:     name,
		inputs:   inputs,
		expected: expected,
	}
}

func intPtr(v int) *int {
	return &v
}

// coalescePointerTestCases pairs each row with the input pointer it
// expects back, rather than with a separately allocated one of equal
// value, so that the identity Coalesce promises is what the rows
// declare.
func coalescePointerTestCases() []coalescePointerTestCase[int] {
	first, last, zero := intPtr(42), intPtr(100), intPtr(0)

	return []coalescePointerTestCase[int]{
		newCoalescePointerTestCase("all nil", core.S[*int](nil, nil, nil), nil),
		newCoalescePointerTestCase("first non-nil", core.S(nil, first, nil, last), first),
		newCoalescePointerTestCase("last non-nil", core.S(nil, nil, nil, last), last),
		newCoalescePointerTestCase("single value", core.S(first), first),
		newCoalescePointerTestCase("single nil", core.S[*int](nil), nil),
		newCoalescePointerTestCase("empty inputs", core.S[*int](), nil),
		newCoalescePointerTestCase("zero value pointer is not nil", core.S(nil, zero, nil, last), zero),
	}
}

func TestCoalescePointer(t *testing.T) {
	core.RunTestCases(t, coalescePointerTestCases())
}

// testStruct for testing struct coalescing
type testStruct struct {
	Value string
	Count int
}

func newTestStruct(value string, count int) testStruct {
	return testStruct{
		Value: value,
		Count: count,
	}
}

func coalesceStructTestCases() []coalesceTestCase[testStruct] {
	var zero testStruct
	first, last := newTestStruct("hello", 42), newTestStruct("world", 100)
	noCount, noValue := newTestStruct("hello", 0), newTestStruct("", 42)

	return []coalesceTestCase[testStruct]{
		newCoalesceTestCase("all zero", core.S(zero, zero, zero), zero),
		newCoalesceTestCase("first non-zero", core.S(zero, first, zero, last), first),
		newCoalesceTestCase("partial zero struct", core.S(zero, noCount, noValue, last), noCount),
		newCoalesceTestCase("empty inputs", core.S[testStruct](), zero),
	}
}

func TestCoalesceStruct(t *testing.T) {
	core.RunTestCases(t, coalesceStructTestCases())
}

// Additional test cases for different numeric types
var coalesceFloat64TestCases = []coalesceTestCase[float64]{
	newCoalesceTestCase("all zeros", core.S(0.0, 0.0, 0.0), 0.0),
	newCoalesceTestCase("first non-zero", core.S(0.0, 3.14, 0.0, 2.71), 3.14),
	newCoalesceTestCase("negative float", core.S(0.0, -1.5, 0.0, 1.5), -1.5),
	newCoalesceTestCase("very small numbers", core.S(0.0, 1e-10, 0.0, 1e-5), 1e-10),
	newCoalesceTestCase("infinity", core.S(0.0, math.Inf(1), 2.0), math.Inf(1)),
	newCoalesceTestCase("no zeros", core.S(1.1, 2.2, 3.3), 1.1),
}

func TestCoalesceFloat64(t *testing.T) {
	core.RunTestCases(t, coalesceFloat64TestCases)
}

// Test cases for bool type
var coalesceBoolTestCases = []coalesceTestCase[bool]{
	newCoalesceTestCase("all false", core.S(false, false, false), false),
	newCoalesceTestCase("first true", core.S(false, true, false), true),
	newCoalesceTestCase("single true", core.S(true), true),
	newCoalesceTestCase("empty inputs", core.S[bool](), false),
	newCoalesceTestCase("no false values", core.S(true, true, true), true),
}

func TestCoalesceBool(t *testing.T) {
	core.RunTestCases(t, coalesceBoolTestCases)
}

// iifTestCase tests IIf function
type iifTestCase struct {
	name     string
	cond     bool
	yes      int
	no       int
	expected int
}

func (tc iifTestCase) Name() string {
	return tc.name
}

func (tc iifTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.IIf(tc.cond, tc.yes, tc.no)
	core.AssertEqual(t, tc.expected, got, "IIf")
}

// Factory function for iifTestCase
func newIIfTestCase(name string, cond bool, yes, no, expected int) iifTestCase {
	return iifTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

var iifTestCases = []iifTestCase{
	newIIfTestCase("true condition", true, 42, 100, 42),
	newIIfTestCase("false condition", false, 42, 100, 100),
	newIIfTestCase("true with zeros", true, 0, 100, 0),
	newIIfTestCase("false with zeros", false, 42, 0, 0),
	newIIfTestCase("same values", true, 42, 42, 42),
	newIIfTestCase("negative values true", true, -42, -100, -42),
	newIIfTestCase("negative values false", false, -42, -100, -100),
}

func TestIIfInt(t *testing.T) {
	core.RunTestCases(t, iifTestCases)
}

// iifStringTestCase tests IIf with strings
type iifStringTestCase struct {
	name     string
	yes      string
	no       string
	expected string
	cond     bool
}

func (tc iifStringTestCase) Name() string {
	return tc.name
}

func (tc iifStringTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.IIf(tc.cond, tc.yes, tc.no)
	core.AssertEqual(t, tc.expected, got, "IIf")
}

// Factory function for iifStringTestCase
func newIIfStringTestCase(name string, cond bool, yes, no, expected string) iifStringTestCase {
	return iifStringTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

var iifStringTestCases = []iifStringTestCase{
	newIIfStringTestCase("true condition", true, "hello", "world", "hello"),
	newIIfStringTestCase("false condition", false, "hello", "world", "world"),
	newIIfStringTestCase("true with empty", true, "", "world", ""),
	newIIfStringTestCase("false with empty", false, "hello", "", ""),
	newIIfStringTestCase("same values", true, "same", "same", "same"),
}

func TestIIfString(t *testing.T) {
	core.RunTestCases(t, iifStringTestCases)
}

// iifPointerTestCase tests IIf with pointers
type iifPointerTestCase struct {
	yes      *int
	no       *int
	expected *int
	name     string
	cond     bool
}

func (tc iifPointerTestCase) Name() string {
	return tc.name
}

func (tc iifPointerTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.IIf(tc.cond, tc.yes, tc.no)
	core.AssertSame(t, tc.expected, got, "IIf")
}

// Factory function for iifPointerTestCase
func newIIfPointerTestCase(name string, cond bool, yes, no, expected *int) iifPointerTestCase {
	return iifPointerTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

// iifPointerTestCases declares as the expectation the very pointer IIf
// should hand back, since IIf returns one of its arguments and identity
// is what says which.
func iifPointerTestCases() []iifPointerTestCase {
	yes, no := intPtr(42), intPtr(100)

	return []iifPointerTestCase{
		newIIfPointerTestCase("true condition", true, yes, no, yes),
		newIIfPointerTestCase("false condition", false, yes, no, no),
		newIIfPointerTestCase("true with nil", true, nil, no, nil),
		newIIfPointerTestCase("false with nil", false, yes, nil, nil),
		newIIfPointerTestCase("both nil true", true, nil, nil, nil),
		newIIfPointerTestCase("both nil false", false, nil, nil, nil),
	}
}

func TestIIfPointer(t *testing.T) {
	core.RunTestCases(t, iifPointerTestCases())
}

// iifStructTestCase tests IIf with structs
type iifStructTestCase struct {
	name     string
	yes      testStruct
	no       testStruct
	expected testStruct
	cond     bool
}

func (tc iifStructTestCase) Name() string {
	return tc.name
}

func (tc iifStructTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.IIf(tc.cond, tc.yes, tc.no)
	core.AssertEqual(t, tc.expected, got, "IIf")
}

// Factory function for iifStructTestCase
func newIIfStructTestCase(name string, cond bool, yes, no, expected testStruct) iifStructTestCase {
	return iifStructTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

func iifStructTestCases() []iifStructTestCase {
	var zero testStruct
	yes, no := newTestStruct("hello", 42), newTestStruct("world", 100)

	return []iifStructTestCase{
		newIIfStructTestCase("true condition", true, yes, no, yes),
		newIIfStructTestCase("false condition", false, yes, no, no),
		newIIfStructTestCase("true with zero", true, zero, no, zero),
		newIIfStructTestCase("false with zero", false, yes, zero, zero),
	}
}

func TestIIfStruct(t *testing.T) {
	core.RunTestCases(t, iifStructTestCases())
}

// Test IIf with function evaluation (ensure both branches are evaluated before selection)
func TestIIfEvaluation(t *testing.T) {
	// Track which functions were called
	var calledYes, calledNo bool

	yes := evaluateYesFunction(&calledYes)
	no := evaluateNoFunction(&calledNo)

	// Both should be evaluated before IIf is called
	core.AssertTrue(t, calledYes, "yes evaluated")
	core.AssertTrue(t, calledNo, "no evaluated")

	core.AssertEqual(t, 42, core.IIf(true, yes, no), "IIf true")
	core.AssertEqual(t, 100, core.IIf(false, yes, no), "IIf false")
}

func evaluateYesFunction(called *bool) int {
	*called = true
	return 42
}

func evaluateNoFunction(called *bool) int {
	*called = true
	return 100
}

// Benchmark tests
func BenchmarkCoalesceInt(b *testing.B) {
	inputs := core.S(0, 0, 0, 42, 0, 100)
	for b.Loop() {
		_ = core.Coalesce(inputs...)
	}
}

func BenchmarkCoalesceString(b *testing.B) {
	inputs := core.S("", "", "", "hello", "", "world")
	for b.Loop() {
		_ = core.Coalesce(inputs...)
	}
}

func BenchmarkIIfInt(b *testing.B) {
	i := 0
	for b.Loop() {
		_ = core.IIf(i%2 == 0, 42, 100)
		i++
	}
}

func BenchmarkIIfString(b *testing.B) {
	i := 0
	for b.Loop() {
		_ = core.IIf(i%2 == 0, "hello", "world")
		i++
	}
}
