package core

import (
	"fmt"
	"math"
	"testing"
)

// TestCase interface validations
var _ TestCase = typeNameTestCase[int]{}
var _ TestCase = coalesceTestCase[int]{}
var _ TestCase = coalescePointerTestCase[int]{}
var _ TestCase = iifTestCase{}
var _ TestCase = iifStringTestCase{}
var _ TestCase = iifPointerTestCase{}
var _ TestCase = iifStructTestCase{}

// typeNameTestCase states what TypeName renders for one type.
type typeNameTestCase[T any] struct {
	want string
	name string
}

func newTypeNameTestCase[T any](name, want string) TestCase {
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
	AssertEqual(t, tc.want, TypeName[T](), "name")
}

func typeNameTestCases() []TestCase {
	return S(
		newTypeNameTestCase[int]("builtin", "int"),
		newTypeNameTestCase[PanicError]("struct", "core.PanicError"),
		newTypeNameTestCase[*PanicError]("pointer", "*core.PanicError"),
		newTypeNameTestCase[[]string]("slice", "[]string"),
		newTypeNameTestCase[error]("builtin interface", "error"),
		newTypeNameTestCase[fmt.Stringer]("interface", "fmt.Stringer"),
		newTypeNameTestCase[Recovered]("package interface", "core.Recovered"),
		newTypeNameTestCase[interface{ Len() int }]("anonymous interface",
			"interface { Len() int }"),
		newTypeNameTestCase[any]("any", "any"),
		newTypeNameTestCase[[]any]("any slice", "[]any"),
	)
}

func TestTypeName(t *testing.T) {
	RunTestCases(t, typeNameTestCases())
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

	got := Coalesce(tc.inputs...)
	AssertEqual(t, tc.expected, got, "Coalesce")
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
	newCoalesceTestCase("all zeros", S(0, 0, 0), 0),
	newCoalesceTestCase("first non-zero", S(0, 42, 0, 100), 42),
	newCoalesceTestCase("last non-zero", S(0, 0, 0, 100), 100),
	newCoalesceTestCase("single value", S(42), 42),
	newCoalesceTestCase("single zero", S(0), 0),
	newCoalesceTestCase("empty inputs", S[int](), 0),
	newCoalesceTestCase("negative values", S(0, -42, 0, -100), -42),
	newCoalesceTestCase("mixed positive and negative", S(0, -1, 0, 1), -1),
	newCoalesceTestCase("large numbers", S(0, 0, 2147483647, 0), 2147483647),
	newCoalesceTestCase("no zeros", S(1, 2, 3, 4), 1),
}

func TestCoalesceInt(t *testing.T) {
	RunTestCases(t, coalesceIntTestCases)
}

var coalesceStringTestCases = []coalesceTestCase[string]{
	newCoalesceTestCase("all empty", S("", "", ""), ""),
	newCoalesceTestCase("first non-empty", S("", "hello", "", "world"), "hello"),
	newCoalesceTestCase("last non-empty", S("", "", "", "world"), "world"),
	newCoalesceTestCase("single value", S("hello"), "hello"),
	newCoalesceTestCase("single empty", S(""), ""),
	newCoalesceTestCase("empty inputs", S[string](), ""),
	newCoalesceTestCase("spaces are not empty", S("", " ", "", "world"), " "),
	newCoalesceTestCase("unicode strings", S("", "🚀", "", "world"), "🚀"),
	newCoalesceTestCase("long strings",
		S("", "a very long string that should not be truncated", "", "short"),
		"a very long string that should not be truncated"),
	newCoalesceTestCase("newlines and tabs", S("", "\n\t", "", "world"), "\n\t"),
	newCoalesceTestCase("no empty strings", S("first", "second", "third"), "first"),
}

func TestCoalesceString(t *testing.T) {
	RunTestCases(t, coalesceStringTestCases)
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

	got := Coalesce(tc.inputs...)

	// Coalesce returns one of its arguments, so the rows pair it with a
	// separately allocated pointer of equal value. Compare what they
	// point at: AreEqual settles pointers by identity, which would call
	// every such row unequal.
	if tc.expected == nil {
		AssertNil(t, got, "Coalesce")
		return
	}

	AssertMustNotNil(t, got, "Coalesce result")
	AssertEqual(t, *tc.expected, *got, "Coalesce value")
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

var coalescePointerTestCases = []coalescePointerTestCase[int]{
	newCoalescePointerTestCase("all nil", S[*int](nil, nil, nil), nil),
	newCoalescePointerTestCase("first non-nil", S(nil, intPtr(42), nil, intPtr(100)), intPtr(42)),
	newCoalescePointerTestCase("last non-nil", S(nil, nil, nil, intPtr(100)), intPtr(100)),
	newCoalescePointerTestCase("single value", S(intPtr(42)), intPtr(42)),
	newCoalescePointerTestCase("single nil", S[*int](nil), nil),
	newCoalescePointerTestCase("empty inputs", S[*int](), nil),
	newCoalescePointerTestCase("zero value pointer is not nil", S(nil, intPtr(0), nil, intPtr(100)), intPtr(0)),
}

func TestCoalescePointer(t *testing.T) {
	RunTestCases(t, coalescePointerTestCases)
}

// testStruct for testing struct coalescing
type testStruct struct {
	Value string
	Count int
}

var coalesceStructTestCases = []coalesceTestCase[testStruct]{
	newCoalesceTestCase("all zero", S(testStruct{}, testStruct{}, testStruct{}), testStruct{}),
	newCoalesceTestCase("first non-zero", S(
		testStruct{},
		testStruct{Value: testHello, Count: 42},
		testStruct{},
		testStruct{Value: testWorld, Count: 100},
	), testStruct{Value: testHello, Count: 42}),
	newCoalesceTestCase("partial zero struct", S(
		testStruct{},
		testStruct{Value: testHello},
		testStruct{Count: 42},
		testStruct{Value: testWorld, Count: 100},
	), testStruct{Value: testHello}),
	newCoalesceTestCase("empty inputs", S[testStruct](), testStruct{}),
}

func TestCoalesceStruct(t *testing.T) {
	RunTestCases(t, coalesceStructTestCases)
}

// Additional test cases for different numeric types
var coalesceFloat64TestCases = []coalesceTestCase[float64]{
	newCoalesceTestCase("all zeros", S(0.0, 0.0, 0.0), 0.0),
	newCoalesceTestCase("first non-zero", S(0.0, 3.14, 0.0, 2.71), 3.14),
	newCoalesceTestCase("negative float", S(0.0, -1.5, 0.0, 1.5), -1.5),
	newCoalesceTestCase("very small numbers", S(0.0, 1e-10, 0.0, 1e-5), 1e-10),
	newCoalesceTestCase("infinity", S(0.0, math.Inf(1), 2.0), math.Inf(1)),
	newCoalesceTestCase("no zeros", S(1.1, 2.2, 3.3), 1.1),
}

func TestCoalesceFloat64(t *testing.T) {
	RunTestCases(t, coalesceFloat64TestCases)
}

// Test cases for bool type
var coalesceBoolTestCases = []coalesceTestCase[bool]{
	newCoalesceTestCase("all false", S(false, false, false), false),
	newCoalesceTestCase("first true", S(false, true, false), true),
	newCoalesceTestCase("single true", S(true), true),
	newCoalesceTestCase("empty inputs", S[bool](), false),
	newCoalesceTestCase("no false values", S(true, true, true), true),
}

func TestCoalesceBool(t *testing.T) {
	RunTestCases(t, coalesceBoolTestCases)
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

	got := IIf(tc.cond, tc.yes, tc.no)
	AssertEqual(t, tc.expected, got, "IIf")
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
	RunTestCases(t, iifTestCases)
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

	got := IIf(tc.cond, tc.yes, tc.no)
	AssertEqual(t, tc.expected, got, "IIf")
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
	RunTestCases(t, iifStringTestCases)
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

	got := IIf(tc.cond, tc.yes, tc.no)

	// As with Coalesce above, the expected pointer is allocated apart
	// from the one IIf selects, so compare what they point at rather
	// than the pointers themselves.
	if tc.expected == nil {
		AssertNil(t, got, "IIf")
		return
	}

	AssertMustNotNil(t, got, "IIf result")
	AssertEqual(t, *tc.expected, *got, "IIf value")
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

var iifPointerTestCases = []iifPointerTestCase{
	newIIfPointerTestCase("true condition", true, intPtr(42), intPtr(100), intPtr(42)),
	newIIfPointerTestCase("false condition", false, intPtr(42), intPtr(100), intPtr(100)),
	newIIfPointerTestCase("true with nil", true, nil, intPtr(100), nil),
	newIIfPointerTestCase("false with nil", false, intPtr(42), nil, nil),
	newIIfPointerTestCase("both nil true", true, nil, nil, nil),
	newIIfPointerTestCase("both nil false", false, nil, nil, nil),
}

func TestIIfPointer(t *testing.T) {
	RunTestCases(t, iifPointerTestCases)
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

	got := IIf(tc.cond, tc.yes, tc.no)
	AssertEqual(t, tc.expected, got, "IIf")
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

var iifStructTestCases = []iifStructTestCase{
	newIIfStructTestCase("true condition", true,
		testStruct{Value: testHello, Count: 42},
		testStruct{Value: testWorld, Count: 100},
		testStruct{Value: testHello, Count: 42}),
	newIIfStructTestCase("false condition", false,
		testStruct{Value: testHello, Count: 42},
		testStruct{Value: testWorld, Count: 100},
		testStruct{Value: testWorld, Count: 100}),
	newIIfStructTestCase("true with zero", true,
		testStruct{},
		testStruct{Value: testWorld, Count: 100},
		testStruct{}),
	newIIfStructTestCase("false with zero", false,
		testStruct{Value: testHello, Count: 42},
		testStruct{},
		testStruct{}),
}

func TestIIfStruct(t *testing.T) {
	RunTestCases(t, iifStructTestCases)
}

// Test IIf with function evaluation (ensure both branches are evaluated before selection)
func TestIIfEvaluation(t *testing.T) {
	// Track which functions were called
	var calledYes, calledNo bool

	yes := evaluateYesFunction(&calledYes)
	no := evaluateNoFunction(&calledNo)

	// Both should be evaluated before IIf is called
	AssertTrue(t, calledYes, "yes evaluated")
	AssertTrue(t, calledNo, "no evaluated")

	AssertEqual(t, 42, IIf(true, yes, no), "IIf true")
	AssertEqual(t, 100, IIf(false, yes, no), "IIf false")
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
	inputs := S(0, 0, 0, 42, 0, 100)
	for b.Loop() {
		_ = Coalesce(inputs...)
	}
}

func BenchmarkCoalesceString(b *testing.B) {
	inputs := S("", "", "", "hello", "", "world")
	for b.Loop() {
		_ = Coalesce(inputs...)
	}
}

func BenchmarkIIfInt(b *testing.B) {
	i := 0
	for b.Loop() {
		_ = IIf(i%2 == 0, 42, 100)
		i++
	}
}

func BenchmarkIIfString(b *testing.B) {
	i := 0
	for b.Loop() {
		_ = IIf(i%2 == 0, "hello", "world")
		i++
	}
}
