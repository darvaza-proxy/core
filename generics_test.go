package core

import (
	"math"
	"testing"
)

// TestCase interface validations
var _ TestCase = coalesceTestCase[int]{}
var _ TestCase = coalescePointerTestCase[int]{}
var _ TestCase = iifTestCase{}
var _ TestCase = iifStringTestCase{}
var _ TestCase = iifPointerTestCase{}
var _ TestCase = iifStructTestCase{}

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
	if got != tc.expected {
		t.Errorf("Coalesce(%v) = %v, want %v", tc.inputs, got, tc.expected)
	}
}

// Factory function for coalesceTestCase
func newCoalesceTestCase[T comparable](name string, inputs []T, expected T) coalesceTestCase[T] {
	return coalesceTestCase[T]{
		name:     name,
		inputs:   inputs,
		expected: expected,
	}
}

func makeCoalesceIntTestCases() []TestCase {
	return []TestCase{
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
}

func makeCoalesceStringTestCases() []TestCase {
	return []TestCase{
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

	// Compare pointer values
	if tc.expected == nil && got == nil {
		return
	}
	if tc.expected == nil || got == nil {
		t.Errorf("Coalesce(%v) = %v, want %v", tc.inputs, got, tc.expected)
		return
	}
	if *got != *tc.expected {
		t.Errorf("Coalesce(%v) = %v, want %v", tc.inputs, *got, *tc.expected)
	}
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

func makeCoalescePointerTestCases() []TestCase {
	return []TestCase{
		newCoalescePointerTestCase("all nil", S[*int](nil, nil, nil), nil),
		newCoalescePointerTestCase("first non-nil", S(nil, intPtr(42), nil, intPtr(100)), intPtr(42)),
		newCoalescePointerTestCase("last non-nil", S(nil, nil, nil, intPtr(100)), intPtr(100)),
		newCoalescePointerTestCase("single value", S(intPtr(42)), intPtr(42)),
		newCoalescePointerTestCase("single nil", S[*int](nil), nil),
		newCoalescePointerTestCase("empty inputs", S[*int](), nil),
		newCoalescePointerTestCase("zero value pointer is not nil", S(nil, intPtr(0), nil, intPtr(100)), intPtr(0)),
	}
}

// testStruct for testing struct coalescing
type testStruct struct {
	Value string
	Count int
}

func makeCoalesceStructTestCases() []TestCase {
	return []TestCase{
		newCoalesceTestCase("all zero", S(testStruct{}, testStruct{}, testStruct{}), testStruct{}),
		newCoalesceTestCase("first non-zero", S(
			testStruct{},
			testStruct{Value: "hello", Count: 42},
			testStruct{},
			testStruct{Value: "world", Count: 100},
		), testStruct{Value: "hello", Count: 42}),
		newCoalesceTestCase("partial zero struct", S(
			testStruct{},
			testStruct{Value: "hello"},
			testStruct{Count: 42},
			testStruct{Value: "world", Count: 100},
		), testStruct{Value: "hello"}),
		newCoalesceTestCase("empty inputs", S[testStruct](), testStruct{}),
	}
}

// Additional test cases for different numeric types
func makeCoalesceFloat64TestCases() []TestCase {
	return []TestCase{
		newCoalesceTestCase("all zeros", S(0.0, 0.0, 0.0), 0.0),
		newCoalesceTestCase("first non-zero", S(0.0, 3.14, 0.0, 2.71), 3.14),
		newCoalesceTestCase("negative float", S(0.0, -1.5, 0.0, 1.5), -1.5),
		newCoalesceTestCase("very small numbers", S(0.0, 1e-10, 0.0, 1e-5), 1e-10),
		newCoalesceTestCase("infinity", S(0.0, math.Inf(1), 2.0), math.Inf(1)),
		newCoalesceTestCase("no zeros", S(1.1, 2.2, 3.3), 1.1),
	}
}

// Test cases for bool type
func makeCoalesceBoolTestCases() []TestCase {
	return []TestCase{
		newCoalesceTestCase("all false", S(false, false, false), false),
		newCoalesceTestCase("first true", S(false, true, false), true),
		newCoalesceTestCase("single true", S(true), true),
		newCoalesceTestCase("empty inputs", S[bool](), false),
		newCoalesceTestCase("no false values", S(true, true, true), true),
	}
}

func TestCoalesce(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		RunTestCases(t, makeCoalesceIntTestCases())
	})
	t.Run("string", func(t *testing.T) {
		RunTestCases(t, makeCoalesceStringTestCases())
	})
	t.Run("pointer", func(t *testing.T) {
		RunTestCases(t, makeCoalescePointerTestCases())
	})
	t.Run("struct", func(t *testing.T) {
		RunTestCases(t, makeCoalesceStructTestCases())
	})
	t.Run("float64", func(t *testing.T) {
		RunTestCases(t, makeCoalesceFloat64TestCases())
	})
	t.Run("bool", func(t *testing.T) {
		RunTestCases(t, makeCoalesceBoolTestCases())
	})
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
	if got != tc.expected {
		t.Errorf("IIf(%v, %v, %v) = %v, want %v", tc.cond, tc.yes, tc.no, got, tc.expected)
	}
}

// Factory function for iifTestCase
func newIifTestCase(name string, cond bool, yes, no, expected int) TestCase {
	return iifTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

func makeIIfTestCases() []TestCase {
	return S(
		newIifTestCase("true condition", true, 42, 100, 42),
		newIifTestCase("false condition", false, 42, 100, 100),
		newIifTestCase("true with zeros", true, 0, 100, 0),
		newIifTestCase("false with zeros", false, 42, 0, 0),
		newIifTestCase("same values", true, 42, 42, 42),
		newIifTestCase("negative values true", true, -42, -100, -42),
		newIifTestCase("negative values false", false, -42, -100, -100),
	)
}

func TestIIf(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		RunTestCases(t, makeIIfTestCases())
	})
	t.Run("string", func(t *testing.T) {
		RunTestCases(t, makeIIfStringTestCases())
	})
	t.Run("pointer", func(t *testing.T) {
		RunTestCases(t, makeIIfPointerTestCases())
	})
	t.Run("struct", func(t *testing.T) {
		RunTestCases(t, makeIIfStructTestCases())
	})
	t.Run("evaluation", runTestIIfEvaluation)
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
	if got != tc.expected {
		t.Errorf("IIf(%v, %v, %v) = %v, want %v", tc.cond, tc.yes, tc.no, got, tc.expected)
	}
}

// Factory function for iifStringTestCase
func newIifStringTestCase(name string, cond bool, yes, no, expected string) TestCase {
	return iifStringTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

func makeIIfStringTestCases() []TestCase {
	return S(
		newIifStringTestCase("true condition", true, "hello", "world", "hello"),
		newIifStringTestCase("false condition", false, "hello", "world", "world"),
		newIifStringTestCase("true with empty", true, "", "world", ""),
		newIifStringTestCase("false with empty", false, "hello", "", ""),
		newIifStringTestCase("same values", true, "same", "same", "same"),
	)
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

	// Compare pointer values
	if tc.expected == nil && got == nil {
		return
	}
	if tc.expected == nil || got == nil {
		t.Errorf("IIf(%v, %v, %v) = %v, want %v", tc.cond, tc.yes, tc.no, got, tc.expected)
		return
	}
	if *got != *tc.expected {
		t.Errorf("IIf(%v, %v, %v) = %v, want %v", tc.cond, tc.yes, tc.no, *got, *tc.expected)
	}
}

// Factory function for iifPointerTestCase
func newIifPointerTestCase(name string, cond bool, yes, no, expected *int) TestCase {
	return iifPointerTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

func makeIIfPointerTestCases() []TestCase {
	return S(
		newIifPointerTestCase("true condition", true, intPtr(42), intPtr(100), intPtr(42)),
		newIifPointerTestCase("false condition", false, intPtr(42), intPtr(100), intPtr(100)),
		newIifPointerTestCase("true with nil", true, nil, intPtr(100), nil),
		newIifPointerTestCase("false with nil", false, intPtr(42), nil, nil),
		newIifPointerTestCase("both nil true", true, nil, nil, nil),
		newIifPointerTestCase("both nil false", false, nil, nil, nil),
	)
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
	if got != tc.expected {
		t.Errorf("IIf(%v, %+v, %+v) = %+v, want %+v", tc.cond, tc.yes, tc.no, got, tc.expected)
	}
}

// Factory function for iifStructTestCase
func newIifStructTestCase(name string, cond bool, yes, no, expected testStruct) TestCase {
	return iifStructTestCase{
		name:     name,
		cond:     cond,
		yes:      yes,
		no:       no,
		expected: expected,
	}
}

func makeIIfStructTestCases() []TestCase {
	return S(
		newIifStructTestCase("true condition", true,
			testStruct{Value: "hello", Count: 42},
			testStruct{Value: "world", Count: 100},
			testStruct{Value: "hello", Count: 42}),
		newIifStructTestCase("false condition", false,
			testStruct{Value: "hello", Count: 42},
			testStruct{Value: "world", Count: 100},
			testStruct{Value: "world", Count: 100}),
		newIifStructTestCase("true with zero", true,
			testStruct{},
			testStruct{Value: "world", Count: 100},
			testStruct{}),
		newIifStructTestCase("false with zero", false,
			testStruct{Value: "hello", Count: 42},
			testStruct{},
			testStruct{}),
	)
}

func runTestIIfEvaluation(t *testing.T) {
	t.Helper()
	// Test IIf with function evaluation (ensure both branches are evaluated before selection)
	// Track which functions were called
	var calledYes, calledNo bool

	yes := evaluateYesFunction(&calledYes)
	no := evaluateNoFunction(&calledNo)

	// Both should be evaluated before IIf is called
	if !calledYes || !calledNo {
		t.Errorf("Functions not evaluated before IIf call: yes=%v, no=%v", calledYes, calledNo)
	}

	result := IIf(true, yes, no)
	if result != 42 {
		t.Errorf("IIf(true, 42, 100) = %v, want 42", result)
	}

	result = IIf(false, yes, no)
	if result != 100 {
		t.Errorf("IIf(false, 42, 100) = %v, want 100", result)
	}
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Coalesce(inputs...)
	}
}

func BenchmarkCoalesceString(b *testing.B) {
	inputs := S("", "", "", "hello", "", "world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Coalesce(inputs...)
	}
}

func BenchmarkIIfInt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IIf(i%2 == 0, 42, 100)
	}
}

func BenchmarkIIfString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IIf(i%2 == 0, "hello", "world")
	}
}
