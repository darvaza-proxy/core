package core

import (
	"math"
	"testing"
)

// coalesceTestCase tests Coalesce function with generic type support
type coalesceTestCase[T comparable] struct {
	expected T
	name     string
	inputs   []T
}

func (tc coalesceTestCase[T]) test(t *testing.T) {
	t.Helper()

	got := Coalesce(tc.inputs...)
	if got != tc.expected {
		t.Errorf("Coalesce(%v) = %v, want %v", tc.inputs, got, tc.expected)
	}
}

var coalesceIntTestCases = []coalesceTestCase[int]{
	{
		name:     "all zeros",
		inputs:   S(0, 0, 0),
		expected: 0,
	},
	{
		name:     "first non-zero",
		inputs:   S(0, 42, 0, 100),
		expected: 42,
	},
	{
		name:     "last non-zero",
		inputs:   S(0, 0, 0, 100),
		expected: 100,
	},
	{
		name:     "single value",
		inputs:   S(42),
		expected: 42,
	},
	{
		name:     "single zero",
		inputs:   S(0),
		expected: 0,
	},
	{
		name:     "empty inputs",
		inputs:   S[int](),
		expected: 0,
	},
	{
		name:     "negative values",
		inputs:   S(0, -42, 0, -100),
		expected: -42,
	},
	{
		name:     "mixed positive and negative",
		inputs:   S(0, -1, 0, 1),
		expected: -1,
	},
	{
		name:     "large numbers",
		inputs:   S(0, 0, 2147483647, 0),
		expected: 2147483647,
	},
	{
		name:     "no zeros",
		inputs:   S(1, 2, 3, 4),
		expected: 1,
	},
}

func TestCoalesceInt(t *testing.T) {
	for _, tc := range coalesceIntTestCases {
		t.Run(tc.name, tc.test)
	}
}

var coalesceStringTestCases = []coalesceTestCase[string]{
	{
		name:     "all empty",
		inputs:   S("", "", ""),
		expected: "",
	},
	{
		name:     "first non-empty",
		inputs:   S("", "hello", "", "world"),
		expected: "hello",
	},
	{
		name:     "last non-empty",
		inputs:   S("", "", "", "world"),
		expected: "world",
	},
	{
		name:     "single value",
		inputs:   S("hello"),
		expected: "hello",
	},
	{
		name:     "single empty",
		inputs:   S(""),
		expected: "",
	},
	{
		name:     "empty inputs",
		inputs:   S[string](),
		expected: "",
	},
	{
		name:     "spaces are not empty",
		inputs:   S("", " ", "", "world"),
		expected: " ",
	},
	{
		name:     "unicode strings",
		inputs:   S("", "🚀", "", "world"),
		expected: "🚀",
	},
	{
		name:     "long strings",
		inputs:   S("", "a very long string that should not be truncated", "", "short"),
		expected: "a very long string that should not be truncated",
	},
	{
		name:     "newlines and tabs",
		inputs:   S("", "\n\t", "", "world"),
		expected: "\n\t",
	},
	{
		name:     "no empty strings",
		inputs:   S("first", "second", "third"),
		expected: "first",
	},
}

func TestCoalesceString(t *testing.T) {
	for _, tc := range coalesceStringTestCases {
		t.Run(tc.name, tc.test)
	}
}

// coalescePointerTestCase tests Coalesce with pointers (requires special comparison logic)
type coalescePointerTestCase[T comparable] struct {
	expected *T
	name     string
	inputs   []*T
}

func (tc coalescePointerTestCase[T]) test(t *testing.T) {
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

func intPtr(v int) *int {
	return &v
}

var coalescePointerTestCases = []coalescePointerTestCase[int]{
	{
		name:     "all nil",
		inputs:   S[*int](nil, nil, nil),
		expected: nil,
	},
	{
		name:     "first non-nil",
		inputs:   S(nil, intPtr(42), nil, intPtr(100)),
		expected: intPtr(42),
	},
	{
		name:     "last non-nil",
		inputs:   S(nil, nil, nil, intPtr(100)),
		expected: intPtr(100),
	},
	{
		name:     "single value",
		inputs:   S(intPtr(42)),
		expected: intPtr(42),
	},
	{
		name:     "single nil",
		inputs:   S[*int](nil),
		expected: nil,
	},
	{
		name:     "empty inputs",
		inputs:   S[*int](),
		expected: nil,
	},
	{
		name:     "zero value pointer is not nil",
		inputs:   S(nil, intPtr(0), nil, intPtr(100)),
		expected: intPtr(0),
	},
}

func TestCoalescePointer(t *testing.T) {
	for _, tc := range coalescePointerTestCases {
		t.Run(tc.name, tc.test)
	}
}

// testStruct for testing struct coalescing
type testStruct struct {
	Value string
	Count int
}

var coalesceStructTestCases = []coalesceTestCase[testStruct]{
	{
		name:     "all zero",
		inputs:   S(testStruct{}, testStruct{}, testStruct{}),
		expected: testStruct{},
	},
	{
		name: "first non-zero",
		inputs: S(
			testStruct{},
			testStruct{Value: "hello", Count: 42},
			testStruct{},
			testStruct{Value: "world", Count: 100},
		),
		expected: testStruct{Value: "hello", Count: 42},
	},
	{
		name: "partial zero struct",
		inputs: S(
			testStruct{},
			testStruct{Value: "hello"},
			testStruct{Count: 42},
			testStruct{Value: "world", Count: 100},
		),
		expected: testStruct{Value: "hello"},
	},
	{
		name:     "empty inputs",
		inputs:   S[testStruct](),
		expected: testStruct{},
	},
}

func TestCoalesceStruct(t *testing.T) {
	for _, tc := range coalesceStructTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Additional test cases for different numeric types
var coalesceFloat64TestCases = []coalesceTestCase[float64]{
	{
		name:     "all zeros",
		inputs:   S(0.0, 0.0, 0.0),
		expected: 0.0,
	},
	{
		name:     "first non-zero",
		inputs:   S(0.0, 3.14, 0.0, 2.71),
		expected: 3.14,
	},
	{
		name:     "negative float",
		inputs:   S(0.0, -1.5, 0.0, 1.5),
		expected: -1.5,
	},
	{
		name:     "very small numbers",
		inputs:   S(0.0, 1e-10, 0.0, 1e-5),
		expected: 1e-10,
	},
	{
		name:     "infinity",
		inputs:   S(0.0, math.Inf(1), 2.0),
		expected: math.Inf(1),
	},
	{
		name:     "no zeros",
		inputs:   S(1.1, 2.2, 3.3),
		expected: 1.1,
	},
}

func TestCoalesceFloat64(t *testing.T) {
	for _, tc := range coalesceFloat64TestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for bool type
var coalesceBoolTestCases = []coalesceTestCase[bool]{
	{
		name:     "all false",
		inputs:   S(false, false, false),
		expected: false,
	},
	{
		name:     "first true",
		inputs:   S(false, true, false),
		expected: true,
	},
	{
		name:     "single true",
		inputs:   S(true),
		expected: true,
	},
	{
		name:     "empty inputs",
		inputs:   S[bool](),
		expected: false,
	},
	{
		name:     "no false values",
		inputs:   S(true, true, true),
		expected: true,
	},
}

func TestCoalesceBool(t *testing.T) {
	for _, tc := range coalesceBoolTestCases {
		t.Run(tc.name, tc.test)
	}
}

// iifTestCase tests IIf function
type iifTestCase struct {
	name     string
	cond     bool
	yes      int
	no       int
	expected int
}

var iifTestCases = []iifTestCase{
	{
		name:     "true condition",
		cond:     true,
		yes:      42,
		no:       100,
		expected: 42,
	},
	{
		name:     "false condition",
		cond:     false,
		yes:      42,
		no:       100,
		expected: 100,
	},
	{
		name:     "true with zeros",
		cond:     true,
		yes:      0,
		no:       100,
		expected: 0,
	},
	{
		name:     "false with zeros",
		cond:     false,
		yes:      42,
		no:       0,
		expected: 0,
	},
	{
		name:     "same values",
		cond:     true,
		yes:      42,
		no:       42,
		expected: 42,
	},
	{
		name:     "negative values true",
		cond:     true,
		yes:      -42,
		no:       -100,
		expected: -42,
	},
	{
		name:     "negative values false",
		cond:     false,
		yes:      -42,
		no:       -100,
		expected: -100,
	},
}

func (tc iifTestCase) test(t *testing.T) {
	t.Helper()

	got := IIf(tc.cond, tc.yes, tc.no)
	if got != tc.expected {
		t.Errorf("IIf(%v, %v, %v) = %v, want %v", tc.cond, tc.yes, tc.no, got, tc.expected)
	}
}

func TestIIfInt(t *testing.T) {
	for _, tc := range iifTestCases {
		t.Run(tc.name, tc.test)
	}
}

// iifStringTestCase tests IIf with strings
type iifStringTestCase struct {
	name     string
	yes      string
	no       string
	expected string
	cond     bool
}

var iifStringTestCases = []iifStringTestCase{
	{
		name:     "true condition",
		cond:     true,
		yes:      "hello",
		no:       "world",
		expected: "hello",
	},
	{
		name:     "false condition",
		cond:     false,
		yes:      "hello",
		no:       "world",
		expected: "world",
	},
	{
		name:     "true with empty",
		cond:     true,
		yes:      "",
		no:       "world",
		expected: "",
	},
	{
		name:     "false with empty",
		cond:     false,
		yes:      "hello",
		no:       "",
		expected: "",
	},
	{
		name:     "same values",
		cond:     true,
		yes:      "same",
		no:       "same",
		expected: "same",
	},
}

func (tc iifStringTestCase) test(t *testing.T) {
	t.Helper()

	got := IIf(tc.cond, tc.yes, tc.no)
	if got != tc.expected {
		t.Errorf("IIf(%v, %v, %v) = %v, want %v", tc.cond, tc.yes, tc.no, got, tc.expected)
	}
}

func TestIIfString(t *testing.T) {
	for _, tc := range iifStringTestCases {
		t.Run(tc.name, tc.test)
	}
}

// iifPointerTestCase tests IIf with pointers
type iifPointerTestCase struct {
	yes      *int
	no       *int
	expected *int
	name     string
	cond     bool
}

var iifPointerTestCases = []iifPointerTestCase{
	{
		name:     "true condition",
		cond:     true,
		yes:      intPtr(42),
		no:       intPtr(100),
		expected: intPtr(42),
	},
	{
		name:     "false condition",
		cond:     false,
		yes:      intPtr(42),
		no:       intPtr(100),
		expected: intPtr(100),
	},
	{
		name:     "true with nil",
		cond:     true,
		yes:      nil,
		no:       intPtr(100),
		expected: nil,
	},
	{
		name:     "false with nil",
		cond:     false,
		yes:      intPtr(42),
		no:       nil,
		expected: nil,
	},
	{
		name:     "both nil true",
		cond:     true,
		yes:      nil,
		no:       nil,
		expected: nil,
	},
	{
		name:     "both nil false",
		cond:     false,
		yes:      nil,
		no:       nil,
		expected: nil,
	},
}

func (tc iifPointerTestCase) test(t *testing.T) {
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

func TestIIfPointer(t *testing.T) {
	for _, tc := range iifPointerTestCases {
		t.Run(tc.name, tc.test)
	}
}

// iifStructTestCase tests IIf with structs
type iifStructTestCase struct {
	name     string
	yes      testStruct
	no       testStruct
	expected testStruct
	cond     bool
}

var iifStructTestCases = []iifStructTestCase{
	{
		name:     "true condition",
		cond:     true,
		yes:      testStruct{Value: "hello", Count: 42},
		no:       testStruct{Value: "world", Count: 100},
		expected: testStruct{Value: "hello", Count: 42},
	},
	{
		name:     "false condition",
		cond:     false,
		yes:      testStruct{Value: "hello", Count: 42},
		no:       testStruct{Value: "world", Count: 100},
		expected: testStruct{Value: "world", Count: 100},
	},
	{
		name:     "true with zero",
		cond:     true,
		yes:      testStruct{},
		no:       testStruct{Value: "world", Count: 100},
		expected: testStruct{},
	},
	{
		name:     "false with zero",
		cond:     false,
		yes:      testStruct{Value: "hello", Count: 42},
		no:       testStruct{},
		expected: testStruct{},
	},
}

func (tc iifStructTestCase) test(t *testing.T) {
	t.Helper()

	got := IIf(tc.cond, tc.yes, tc.no)
	if got != tc.expected {
		t.Errorf("IIf(%v, %+v, %+v) = %+v, want %+v", tc.cond, tc.yes, tc.no, got, tc.expected)
	}
}

func TestIIfStruct(t *testing.T) {
	for _, tc := range iifStructTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test IIf with function evaluation (ensure both branches are evaluated before selection)
func TestIIfEvaluation(t *testing.T) {
	// Track which functions were called
	var calledYes, calledNo bool

	yes := func() int {
		calledYes = true
		return 42
	}()

	no := func() int {
		calledNo = true
		return 100
	}()

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
