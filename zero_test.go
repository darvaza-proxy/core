package core

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

// TestCase validations
var _ TestCase = zeroTestCase[int]{}
var _ TestCase = zeroRefTestCase[int]{}
var _ TestCase = isZeroTestCase{}
var _ TestCase = isNilTestCase{}
var _ TestCase = isNilVsIsZeroTestCase{}
var _ TestCase = initializationSemanticsTestCase{}
var _ TestCase = isSameTestCase{}
var _ TestCase = isSameStackOverflowTestCase{}

type zeroTestCase[T comparable] struct {
	expected T
	factory  func() *T
	name     string
}

func (tc zeroTestCase[T]) Name() string {
	return tc.name
}

func (tc zeroTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := Zero(tc.factory())
	AssertEqual(t, tc.expected, result, "Zero[%s]", tc.name)
}

// Generic test case for reference types that should return nil
type zeroRefTestCase[T any] struct {
	factory func() *T
	name    string
}

func (tc zeroRefTestCase[T]) Name() string {
	return tc.name
}

func (tc zeroRefTestCase[T]) Test(t *testing.T) {
	t.Helper()
	result := Zero(tc.factory())
	AssertNil(t, result, "Zero[%T]", *new(T))
}

func newZeroTestCase[T comparable](name string, expected T, factory func() *T) zeroTestCase[T] {
	return zeroTestCase[T]{
		expected: expected,
		factory:  factory,
		name:     name,
	}
}

func newZeroRefTestCase[T any](name string, factory func() *T) zeroRefTestCase[T] {
	return zeroRefTestCase[T]{
		factory: factory,
		name:    name,
	}
}

var zeroIntTestCases = S(
	newZeroTestCase("nil pointer", 0, func() *int { return nil }),
	newZeroTestCase("non-nil pointer", 0, func() *int { v := 42; return &v }),
)

var zeroStringTestCases = S(
	newZeroTestCase("nil pointer", "", func() *string { return nil }),
	newZeroTestCase("non-nil pointer", "", func() *string { v := "hello"; return &v }),
)

var zeroBoolTestCases = S(
	newZeroTestCase("nil pointer", false, func() *bool { return nil }),
	newZeroTestCase("non-nil pointer", false, func() *bool { v := true; return &v }),
)

var zeroTimeTestCases = S(
	newZeroTestCase("nil pointer", time.Time{}, func() *time.Time { return nil }),
	newZeroTestCase("non-nil pointer", time.Time{}, func() *time.Time { v := time.Now(); return &v }),
)

var zeroSliceTestCases = S(
	newZeroRefTestCase("nil slice pointer", func() *[]int { return nil }),
	newZeroRefTestCase("non-nil slice pointer", func() *[]int { v := S(1, 2, 3); return &v }),
)

var zeroMapTestCases = S(
	newZeroRefTestCase("nil map pointer", func() *map[string]int { return nil }),
	newZeroRefTestCase("non-nil map pointer", func() *map[string]int { v := map[string]int{"a": 1}; return &v }),
)

var zeroPointerTestCases = S(
	newZeroRefTestCase("nil double pointer", func() **int { return nil }),
	newZeroRefTestCase("non-nil double pointer", func() **int { v := 42; vp := &v; return &vp }),
)

var zeroInterfaceTestCases = S(
	newZeroRefTestCase("nil interface pointer", func() *any { return nil }),
	newZeroRefTestCase("non-nil interface pointer", func() *any { var v any = 42; return &v }),
)

var zeroChannelTestCases = S(
	newZeroRefTestCase("nil channel pointer", func() *chan int { return nil }),
	newZeroRefTestCase("non-nil channel pointer", func() *chan int { ch := make(chan int); return &ch }),
)

var zeroFuncTestCases = S(
	newZeroRefTestCase("nil func pointer", func() *func() { return nil }),
	newZeroRefTestCase("non-nil func pointer", func() *func() { fn := func() {}; return &fn }),
)

func TestZero(t *testing.T) {
	t.Run("int", func(t *testing.T) { RunTestCases(t, zeroIntTestCases) })
	t.Run("string", func(t *testing.T) { RunTestCases(t, zeroStringTestCases) })
	t.Run("bool", func(t *testing.T) { RunTestCases(t, zeroBoolTestCases) })
	t.Run("slice", func(t *testing.T) { RunTestCases(t, zeroSliceTestCases) })
	t.Run("map", func(t *testing.T) { RunTestCases(t, zeroMapTestCases) })
	t.Run("pointer", func(t *testing.T) { RunTestCases(t, zeroPointerTestCases) })
	t.Run("struct", testZeroStruct)
	t.Run("interface", func(t *testing.T) { RunTestCases(t, zeroInterfaceTestCases) })
	t.Run("channel", func(t *testing.T) { RunTestCases(t, zeroChannelTestCases) })
	t.Run("func", func(t *testing.T) { RunTestCases(t, zeroFuncTestCases) })
	t.Run("time", func(t *testing.T) { RunTestCases(t, zeroTimeTestCases) })
}

func testZeroStruct(t *testing.T) {
	t.Helper()

	type testStruct struct {
		Name string
		Age  int
	}

	newTestStructPtr := func() *testStruct {
		v := testStruct{Name: "John", Age: 30}
		return &v
	}

	structTests := []zeroTestCase[testStruct]{
		newZeroTestCase("nil pointer", testStruct{}, func() *testStruct { return nil }),
		newZeroTestCase("non-nil pointer", testStruct{}, newTestStructPtr),
	}

	RunTestCases(t, structTests)
}

type isZeroTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isZeroTestCase) Name() string {
	return tc.name
}

func (tc isZeroTestCase) Test(t *testing.T) {
	t.Helper()

	result := IsZero(tc.value)
	AssertEqual(t, tc.expected, result, "IsZero result")
}

func newIsZeroTestCase(name string, value any, expected bool) isZeroTestCase {
	return isZeroTestCase{
		value:    value,
		name:     name,
		expected: expected,
	}
}

func TestIsZero(t *testing.T) {
	// Basic types
	zeroTests := []isZeroTestCase{
		newIsZeroTestCase("nil", nil, true),
		newIsZeroTestCase("zero int", 0, true),
		newIsZeroTestCase("non-zero int", 42, false),
		newIsZeroTestCase("zero string", "", true),
		newIsZeroTestCase("non-zero string", "hello", false),
		newIsZeroTestCase("zero bool", false, true),
		newIsZeroTestCase("non-zero bool", true, false),
		newIsZeroTestCase("zero float64", 0.0, true),
		newIsZeroTestCase("non-zero float64", 3.14, false),
		newIsZeroTestCase("zero uint", uint(0), true),
		newIsZeroTestCase("non-zero uint", uint(42), false),
		newIsZeroTestCase("zero int8", int8(0), true),
		newIsZeroTestCase("non-zero int8", int8(42), false),
		newIsZeroTestCase("zero int16", int16(0), true),
		newIsZeroTestCase("non-zero int16", int16(42), false),
		newIsZeroTestCase("zero int32", int32(0), true),
		newIsZeroTestCase("non-zero int32", int32(42), false),
		newIsZeroTestCase("zero int64", int64(0), true),
		newIsZeroTestCase("non-zero int64", int64(42), false),
		newIsZeroTestCase("zero uint8", uint8(0), true),
		newIsZeroTestCase("non-zero uint8", uint8(42), false),
		newIsZeroTestCase("zero uint16", uint16(0), true),
		newIsZeroTestCase("non-zero uint16", uint16(42), false),
		newIsZeroTestCase("zero uint32", uint32(0), true),
		newIsZeroTestCase("non-zero uint32", uint32(42), false),
		newIsZeroTestCase("zero uint64", uint64(0), true),
		newIsZeroTestCase("non-zero uint64", uint64(42), false),
		newIsZeroTestCase("zero float32", float32(0), true),
		newIsZeroTestCase("non-zero float32", float32(3.14), false),
		newIsZeroTestCase("zero complex64", complex64(0), true),
		newIsZeroTestCase("non-zero complex64", complex64(1+2i), false),
		newIsZeroTestCase("zero complex128", complex128(0), true),
		newIsZeroTestCase("non-zero complex128", complex128(1+2i), false),
		newIsZeroTestCase("zero rune", rune(0), true),
		newIsZeroTestCase("non-zero rune", rune('A'), false),
		newIsZeroTestCase("zero byte", byte(0), true),
		newIsZeroTestCase("non-zero byte", byte(42), false),
		newIsZeroTestCase("zero uintptr", uintptr(0), true),
		newIsZeroTestCase("non-zero uintptr", uintptr(0x12345678), false),
	}

	RunTestCases(t, zeroTests)
}

// Test IsZero with interfaces containing values
func TestIsZeroInterface(t *testing.T) {
	var nilIntPtr *int
	zeroIntPtr := new(int)

	interfaceTests := []isZeroTestCase{
		// nil interface
		newIsZeroTestCase("nil interface", nil, true),
		// interface containing nil pointer
		newIsZeroTestCase("interface containing nil pointer", nilIntPtr, true),
		// interface containing non-nil pointer to zero
		newIsZeroTestCase("interface containing pointer to zero", zeroIntPtr, false),
		// interface containing zero value
		newIsZeroTestCase("interface containing zero int", any(0), true),
		// interface containing non-zero value
		newIsZeroTestCase("interface containing non-zero int", any(42), false),
		// interface containing empty string
		newIsZeroTestCase("interface containing empty string", any(""), true),
		// interface containing non-empty string
		newIsZeroTestCase("interface containing non-empty string", any("hello"), false),
		// interface containing empty slice
		newIsZeroTestCase("interface containing empty slice", any(S[int]()), false), // empty slice is not a zero value
		// interface containing nil slice
		newIsZeroTestCase("interface containing nil slice", any([]int(nil)), true),
	}

	RunTestCases(t, interfaceTests)
}

// Test IsZero with edge cases
func TestIsZeroEdgeCases(t *testing.T) {
	edgeTests := []isZeroTestCase{
		// Slices
		newIsZeroTestCase("nil slice", []int(nil), true),
		newIsZeroTestCase("empty slice", S[int](), false),
		newIsZeroTestCase("non-empty slice", S(1, 2, 3), false),
		// Maps
		newIsZeroTestCase("nil map", map[string]int(nil), true),
		newIsZeroTestCase("empty map", map[string]int{}, false),
		newIsZeroTestCase("non-empty map", map[string]int{"a": 1}, false),
		// Channels
		newIsZeroTestCase("nil channel", chan int(nil), true),
		newIsZeroTestCase("non-nil channel", make(chan int), false),
		// Functions
		newIsZeroTestCase("nil function", (func())(nil), true),
		newIsZeroTestCase("non-nil function", func() {}, false),
		// Arrays
		newIsZeroTestCase("zero array", [3]int{}, true),
		newIsZeroTestCase("non-zero array", [3]int{1, 2, 3}, false),
		newIsZeroTestCase("partially zero array", [3]int{0, 0, 1}, false),
	}

	RunTestCases(t, edgeTests)
}

func TestZeroEdgeCases(t *testing.T) {
	// Test Zero function returns expected values when passed nil
	t.Run("nil slice pointer", testZeroEdgeCasesNilSlicePointer)

	t.Run("nil map pointer", testZeroEdgeCasesNilMapPointer)

	t.Run("nil pointer pointer", testZeroEdgeCasesNilPointerPointer)

	t.Run("nil channel pointer", testZeroEdgeCasesNilChannelPointer)

	t.Run("nil func pointer", testZeroEdgeCasesNilFuncPointer)
}

func isZeroReflectValueTestCases() []isZeroTestCase {
	return S(
		// Invalid reflect.Value
		newIsZeroTestCase("invalid reflect.Value", reflect.Value{}, true),

		// Valid reflect.Value containing zero
		newIsZeroTestCase("reflect.Value containing zero int", reflect.ValueOf(0), true),
		newIsZeroTestCase("reflect.Value containing empty string", reflect.ValueOf(""), true),
		newIsZeroTestCase("reflect.Value containing false", reflect.ValueOf(false), true),

		// Valid reflect.Value containing non-zero
		newIsZeroTestCase("reflect.Value containing non-zero int", reflect.ValueOf(42), false),
		newIsZeroTestCase("reflect.Value containing non-empty string", reflect.ValueOf("hello"), false),
		newIsZeroTestCase("reflect.Value containing true", reflect.ValueOf(true), false),

		// reflect.Value containing nil pointer
		newIsZeroTestCase("reflect.Value containing nil pointer", reflect.ValueOf((*int)(nil)), true),

		// reflect.Value containing non-nil pointer
		newIsZeroTestCase("reflect.Value containing non-nil pointer", reflect.ValueOf(new(int)), false),
	)
}

func TestIsZeroReflectValue(t *testing.T) {
	RunTestCases(t, isZeroReflectValueTestCases())
}

type complexStruct struct {
	MapField    map[string]int
	PtrField    *int
	ChanField   chan int
	FuncField   func()
	StringField string
	SliceField  []int
	IntField    int
}

func TestIsZeroComplexStruct(t *testing.T) {
	complexTests := []isZeroTestCase{
		newIsZeroTestCase("zero struct", complexStruct{}, true),
		newIsZeroTestCase("struct with int field", complexStruct{
			IntField: 42,
		}, false),
		newIsZeroTestCase("struct with string field", complexStruct{
			StringField: "hello",
		}, false),
		newIsZeroTestCase("struct with nil slice", complexStruct{
			SliceField: nil,
		}, true),
		newIsZeroTestCase("struct with empty slice", complexStruct{
			SliceField: S[int](),
		}, false),
		newIsZeroTestCase("struct with nil map", complexStruct{
			MapField: nil,
		}, true),
		newIsZeroTestCase("struct with empty map", complexStruct{
			MapField: map[string]int{},
		}, false),
	}

	RunTestCases(t, complexTests)
}

// Test IsZero with pointers
func TestIsZeroPointers(t *testing.T) {
	var nilInt *int
	nonNilInt := new(int)
	zeroInt := 0
	nonZeroInt := 42
	ptrToZero := &zeroInt
	ptrToNonZero := &nonZeroInt

	var nilIntPtr *int
	ptrToNilPtr := &nilIntPtr

	ptrTests := []isZeroTestCase{
		newIsZeroTestCase("nil pointer", nilInt, true),
		newIsZeroTestCase("non-nil pointer to zero", nonNilInt, false),
		newIsZeroTestCase("pointer to zero value", ptrToZero, false),
		newIsZeroTestCase("pointer to non-zero value", ptrToNonZero, false),
		newIsZeroTestCase("pointer to nil pointer", ptrToNilPtr, false),
	}

	RunTestCases(t, ptrTests)
}

type outerStruct struct {
	Inner innerStruct
}

type innerStruct struct {
	Value int
}

// Test IsZero with nested structs
func TestIsZeroNestedStructs(t *testing.T) {
	var nilOuterPtr *outerStruct
	zeroOuterPtr := &outerStruct{}

	nestedTests := []isZeroTestCase{
		newIsZeroTestCase("zero nested struct", outerStruct{}, true),
		newIsZeroTestCase("nested struct with value", outerStruct{
			Inner: innerStruct{
				Value: 42,
			},
		}, false),
		newIsZeroTestCase("pointer to zero nested struct", zeroOuterPtr, false),
		newIsZeroTestCase("nil pointer to nested struct", nilOuterPtr, true),
	}

	RunTestCases(t, nestedTests)
}

// isNilTestCase tests IsNil function
type isNilTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isNilTestCase) Name() string {
	return tc.name
}

func (tc isNilTestCase) Test(t *testing.T) {
	t.Helper()

	result := IsNil(tc.value)
	AssertEqual(t, tc.expected, result, "IsNil result")
}

func newIsNilTestCase(name string, value any, expected bool) isNilTestCase {
	return isNilTestCase{
		value:    value,
		name:     name,
		expected: expected,
	}
}

func TestIsNil(t *testing.T) {
	t.Run("basic", testIsNilBasic)
	t.Run("pointers", testIsNilPointers)
	t.Run("slices", testIsNilSlices)
	t.Run("maps", testIsNilMaps)
	t.Run("channels", testIsNilChannels)
	t.Run("functions", testIsNilFunctions)
	t.Run("interfaces", testIsNilInterfaces)
}

func testIsNilBasic(t *testing.T) {
	t.Helper()

	nilTests := []isNilTestCase{
		newIsNilTestCase("nil", nil, true),
		newIsNilTestCase("non-nil int", 42, false),
		newIsNilTestCase("zero int", 0, false),
		newIsNilTestCase("non-nil string", "hello", false),
		newIsNilTestCase("empty string", "", false),
		newIsNilTestCase("non-nil bool", true, false),
		newIsNilTestCase("false bool", false, false),
	}

	RunTestCases(t, nilTests)
}

func testIsNilPointers(t *testing.T) {
	t.Helper()

	var nilPtr *int
	nonNilPtr := new(int)

	ptrTests := []isNilTestCase{
		newIsNilTestCase("nil pointer", nilPtr, true),
		newIsNilTestCase("non-nil pointer", nonNilPtr, false),
	}

	RunTestCases(t, ptrTests)
}

// isNilVsIsZeroTestCase tests both IsNil and IsZero behaviour on the same value
type isNilVsIsZeroTestCase struct {
	value        any
	name         string
	description  string
	expectedNil  bool
	expectedZero bool
}

func (tc isNilVsIsZeroTestCase) Name() string {
	return tc.name
}

func (tc isNilVsIsZeroTestCase) Test(t *testing.T) {
	t.Helper()

	nilResult := IsNil(tc.value)
	zeroResult := IsZero(tc.value)

	AssertEqual(t, tc.expectedNil, nilResult, tc.description+" - IsNil result")
	AssertEqual(t, tc.expectedZero, zeroResult, tc.description+" - IsZero result")
}

func newIsNilVsIsZeroTestCase(name string, value any, expectedNil, expectedZero bool,
	description string) isNilVsIsZeroTestCase {
	return isNilVsIsZeroTestCase{
		name:         name,
		value:        value,
		expectedNil:  expectedNil,
		expectedZero: expectedZero,
		description:  description,
	}
}

func isNilVsIsZeroTestCases() []isNilVsIsZeroTestCase {
	return S(
		// Basic types: not nil but can be zero
		newIsNilVsIsZeroTestCase("zero int", 0, false, true, "zero int is not nil but is zero"),
		newIsNilVsIsZeroTestCase("non-zero int", 42, false, false, "non-zero int is neither nil nor zero"),
		newIsNilVsIsZeroTestCase("zero string", "", false, true, "zero string is not nil but is zero"),
		newIsNilVsIsZeroTestCase("non-zero string", "hello", false, false, "non-zero string is neither nil nor zero"),

		// Pointer types: can be both nil and zero
		newIsNilVsIsZeroTestCase("nil pointer", (*int)(nil), true, true, "nil pointer is both nil and zero"),
		newIsNilVsIsZeroTestCase("non-nil pointer", new(int), false, false, "non-nil pointer is neither nil nor zero"),

		// Slice types: nil vs empty distinction
		newIsNilVsIsZeroTestCase("nil slice", []int(nil), true, true, "nil slice is both nil and zero"),
		newIsNilVsIsZeroTestCase("empty slice", []int{}, false, false,
			"empty slice is neither nil nor zero"),
		newIsNilVsIsZeroTestCase("non-empty slice", []int{1, 2}, false, false,
			"non-empty slice is neither nil nor zero"),

		// Map types: nil vs empty distinction
		newIsNilVsIsZeroTestCase("nil map", map[string]int(nil), true, true, "nil map is both nil and zero"),
		newIsNilVsIsZeroTestCase("empty map", map[string]int{}, false, false, "empty map is neither nil nor zero"),
		newIsNilVsIsZeroTestCase("non-empty map", map[string]int{"a": 1}, false, false,
			"non-empty map is neither nil nor zero"),

		// Channel types
		newIsNilVsIsZeroTestCase("nil channel", (chan int)(nil), true, true, "nil channel is both nil and zero"),
		newIsNilVsIsZeroTestCase("non-nil channel", make(chan int), false, false,
			"non-nil channel is neither nil nor zero"),

		// Function types
		newIsNilVsIsZeroTestCase("nil function", (func())(nil), true, true, "nil function is both nil and zero"),
		newIsNilVsIsZeroTestCase("non-nil function", func() {}, false, false,
			"non-nil function is neither nil nor zero"),

		// Interface types
		newIsNilVsIsZeroTestCase("nil interface", (any)(nil), true, true, "nil interface is both nil and zero"),
		newIsNilVsIsZeroTestCase("non-nil interface", any(42), false, false,
			"non-nil interface is neither nil nor zero"),

		// Struct types: cannot be nil
		newIsNilVsIsZeroTestCase("zero struct", struct{}{}, false, true, "zero struct is not nil but is zero"),
		newIsNilVsIsZeroTestCase("non-zero struct", struct{ Name string }{Name: "test"}, false, false,
			"non-zero struct is neither nil nor zero"),
	)
}

func TestIsNilVsIsZero(t *testing.T) {
	RunTestCases(t, isNilVsIsZeroTestCases())
}

func isNilReflectValueTestCases() []isNilTestCase {
	return S(
		// Invalid reflect.Value
		newIsNilTestCase("invalid reflect.Value", reflect.Value{}, true),

		// Valid reflect.Value containing nil
		newIsNilTestCase("reflect.Value containing nil pointer", reflect.ValueOf((*int)(nil)), true),
		newIsNilTestCase("reflect.Value containing nil slice", reflect.ValueOf([]int(nil)), true),
		newIsNilTestCase("reflect.Value containing nil map", reflect.ValueOf(map[string]int(nil)), true),

		// Valid reflect.Value containing non-nil
		newIsNilTestCase("reflect.Value containing non-nil pointer", reflect.ValueOf(new(int)), false),
		newIsNilTestCase("reflect.Value containing empty slice", reflect.ValueOf([]int{}), false),
		newIsNilTestCase("reflect.Value containing empty map", reflect.ValueOf(map[string]int{}), false),

		// reflect.Value containing basic types (cannot be nil)
		newIsNilTestCase("reflect.Value containing int", reflect.ValueOf(42), false),
		newIsNilTestCase("reflect.Value containing string", reflect.ValueOf("hello"), false),
		newIsNilTestCase("reflect.Value containing bool", reflect.ValueOf(true), false),
	)
}

func TestIsNilWithReflectValue(t *testing.T) {
	RunTestCases(t, isNilReflectValueTestCases())
}

func testIsNilSlices(t *testing.T) {
	t.Helper()

	var nilSlice []int
	emptySlice := S[int]()
	nonEmptySlice := S(1, 2, 3)

	sliceTests := []isNilTestCase{
		newIsNilTestCase("nil slice", nilSlice, true),
		newIsNilTestCase("empty slice", emptySlice, false),
		newIsNilTestCase("non-empty slice", nonEmptySlice, false),
	}

	RunTestCases(t, sliceTests)

	// Additional edge case tests from HEAD - interface containing typed nil
	var nilPtr *int
	var vi any = nilPtr
	AssertTrue(t, IsNil(vi), "interface containing typed nil")

	// Slice of pointers with nil elements
	var ptrSlice []*int
	ptrSlice = append(ptrSlice, nil)
	AssertFalse(t, IsNil(ptrSlice), "slice containing nil elements")
	AssertTrue(t, IsNil(ptrSlice[0]), "nil element in slice")
}

func testIsNilMaps(t *testing.T) {
	t.Helper()

	var nilMap map[string]int
	emptyMap := map[string]int{}
	nonEmptyMap := map[string]int{"a": 1}

	mapTests := []isNilTestCase{
		newIsNilTestCase("nil map", nilMap, true),
		newIsNilTestCase("empty map", emptyMap, false),
		newIsNilTestCase("non-empty map", nonEmptyMap, false),
	}

	RunTestCases(t, mapTests)
}

func testIsNilChannels(t *testing.T) {
	t.Helper()

	var nilChan chan int
	nonNilChan := make(chan int)

	chanTests := []isNilTestCase{
		newIsNilTestCase("nil channel", nilChan, true),
		newIsNilTestCase("non-nil channel", nonNilChan, false),
	}

	RunTestCases(t, chanTests)
}

func testIsNilFunctions(t *testing.T) {
	t.Helper()

	var nilFunc func()
	nonNilFunc := func() {}

	funcTests := []isNilTestCase{
		newIsNilTestCase("nil function", nilFunc, true),
		newIsNilTestCase("non-nil function", nonNilFunc, false),
	}

	RunTestCases(t, funcTests)
}

func testIsNilInterfaces(t *testing.T) {
	t.Helper()

	var nilInterface fmt.Stringer
	var nonNilInterface fmt.Stringer = time.Second // time.Duration implements fmt.Stringer

	interfaceTests := []isNilTestCase{
		newIsNilTestCase("nil interface", nilInterface, true),
		newIsNilTestCase("non-nil interface", nonNilInterface, false),
	}

	RunTestCases(t, interfaceTests)
}

// Test IsNil with reflect.Value
// TestIsNilReflectValue is now consolidated with TestIsNilWithReflectValue
// This test was duplicating the same functionality

func isNilTypedInterfaceTestCases() []isNilTestCase {
	var nilPtr *int
	var typedNilInterface any = nilPtr

	ptrArray := []*int{nil, nil, nil}
	nilValueMap := map[string]*int{"key": nil}
	var nilChan chan int
	closedChan := make(chan int)
	close(closedChan)

	return S(
		// Typed nil in interface
		newIsNilTestCase("interface containing typed nil pointer", typedNilInterface, true),

		// Array/slice with nil elements
		newIsNilTestCase("array with nil elements", ptrArray, false),
		newIsNilTestCase("nil element in array", ptrArray[0], true),

		// Map with nil values
		newIsNilTestCase("map with nil values", nilValueMap, false),
		newIsNilTestCase("nil value in map", nilValueMap["key"], true),

		// Channel operations
		newIsNilTestCase("nil channel", nilChan, true),
		newIsNilTestCase("closed channel", closedChan, false),
	)
}

// Test IsNil with typed nil in interface
func TestIsNilTypedInterface(t *testing.T) {
	RunTestCases(t, isNilTypedInterfaceTestCases())
}

// initializationSemanticsTestCase tests initialization semantics patterns
type initializationSemanticsTestCase struct {
	value       any
	name        string
	description string
	expected    bool
}

func (tc initializationSemanticsTestCase) Name() string {
	return tc.name
}

func (tc initializationSemanticsTestCase) Test(t *testing.T) {
	t.Helper()

	result := IsZero(tc.value)
	AssertEqual(t, tc.expected, result, tc.description)
}

func newInitializationSemanticsTestCase(
	name string, value any, expected bool, description string,
) initializationSemanticsTestCase {
	return initializationSemanticsTestCase{
		name:        name,
		value:       value,
		expected:    expected,
		description: description,
	}
}

func initializationSemanticsTestCases() []initializationSemanticsTestCase {
	return S(
		// Slices: nil vs initialized empty
		newInitializationSemanticsTestCase("nil slice", []int(nil), true, "nil slice needs initialization"),
		newInitializationSemanticsTestCase("empty slice", []int{}, false, "empty slice is already initialized"),
		newInitializationSemanticsTestCase("non-empty slice", []int{1, 2}, false, "non-empty slice is initialized"),

		// Maps: nil vs initialized empty
		newInitializationSemanticsTestCase("nil map", map[string]int(nil), true,
			"nil map needs initialization"),
		newInitializationSemanticsTestCase("empty map", map[string]int{}, false,
			"empty map is already initialized"),
		newInitializationSemanticsTestCase("non-empty map", map[string]int{"a": 1}, false,
			"non-empty map is initialized"),

		// Pointers: nil vs assigned
		newInitializationSemanticsTestCase("nil pointer", (*int)(nil), true, "nil pointer can be assigned"),
		newInitializationSemanticsTestCase("pointer to zero", new(int), false,
			"pointer is already assigned (even to zero)"),
		newInitializationSemanticsTestCase("pointer to value",
			func() *int { i := 42; return &i }(), false, "pointer to value is assigned"),

		// Channels: nil vs created
		newInitializationSemanticsTestCase("nil channel", (chan int)(nil), true, "nil channel needs initialization"),
		newInitializationSemanticsTestCase("created channel", make(chan int), false, "created channel is initialized"),

		// Functions: nil vs assigned
		newInitializationSemanticsTestCase("nil function", (func())(nil), true, "nil function can be assigned"),
		newInitializationSemanticsTestCase("assigned function", func() {}, false, "assigned function is initialized"),

		// Basic types: zero vs non-zero
		newInitializationSemanticsTestCase("zero int", 0, true, "zero int is uninitialized"),
		newInitializationSemanticsTestCase("non-zero int", 42, false, "non-zero int is initialized"),
		newInitializationSemanticsTestCase("zero string", "", true, "zero string is uninitialized"),
		newInitializationSemanticsTestCase("non-zero string", "hello", false, "non-zero string is initialized"),
		newInitializationSemanticsTestCase("zero bool", false, true, "zero bool is uninitialized"),
		newInitializationSemanticsTestCase("non-zero bool", true, false, "non-zero bool is initialized"),
	)
}

func TestIsZeroInitializationSemantics(t *testing.T) {
	// Test the key insight: initialized vs uninitialized state
	RunTestCases(t, initializationSemanticsTestCases())
}

func TestIsZeroPracticalInitialization(t *testing.T) {
	// Demonstrate practical initialization patterns using IsZero

	// Example 1: Lazy initialization of slice
	var items []string
	if IsZero(items) {
		items = []string{"default", "values"}
	}
	AssertSliceEqual(t, []string{"default", "values"}, items, "slice should be initialized")

	// Items is now initialized, so IsZero returns false
	AssertFalse(t, IsZero(items), "initialized slice zero")

	// Example 2: Conditional map initialization
	var cache map[string]int
	if IsZero(cache) {
		cache = make(map[string]int)
	}
	cache["key"] = 42
	AssertEqual(t, 42, cache["key"], "map should be usable after initialization")

	// Example 3: Pointer initialization
	var ptr *int
	if IsZero(ptr) {
		value := 100
		ptr = &value
	}
	AssertEqual(t, 100, *ptr, "pointer should point to initialized value")

	// Example 4: Channel initialization
	var ch chan int
	if IsZero(ch) {
		ch = make(chan int, 1)
	}
	ch <- 42
	result := <-ch
	AssertEqual(t, 42, result, "channel should be usable after initialization")

	// Example 5: Function initialization
	var fn func() string
	if IsZero(fn) {
		fn = func() string { return "initialized" }
	}
	AssertEqual(t, "initialized", fn(), "function should be callable after initialization")
}

func testZeroEdgeCasesNilSlicePointer(t *testing.T) {
	t.Helper()
	items := Zero[[]string](nil)
	AssertNil(t, items, "Zero(nil)")

	// Initialize the slice
	items = S("default", "values")
	AssertSliceEqual(t, S("default", "values"), items, "initialized slice")

	// Check that it's not zero after initialization
	AssertEqual(t, false, IsZero(items), "IsZero(initialized)")
}

func testZeroEdgeCasesNilMapPointer(t *testing.T) {
	t.Helper()
	cache := Zero[map[string]int](nil)
	AssertNil(t, cache, "Zero(nil)")

	// Initialize and use the map
	cache = make(map[string]int)
	cache["key"] = 42
	AssertEqual(t, 42, cache["key"], "map value")
}

func testZeroEdgeCasesNilPointerPointer(t *testing.T) {
	t.Helper()
	ptr := Zero[*int](nil)
	AssertNil(t, ptr, "Zero(nil)")

	// Initialize the pointer
	value := 100
	ptr = &value
	AssertEqual(t, 100, *ptr, "pointer value")
}

func testZeroEdgeCasesNilChannelPointer(t *testing.T) {
	t.Helper()
	ch := Zero[chan int](nil)
	AssertNil(t, ch, "Zero(nil)")

	// Initialize and use the channel
	ch = make(chan int, 1)
	ch <- 42
	result := <-ch
	AssertEqual(t, 42, result, "channel value")
}

func testZeroEdgeCasesNilFuncPointer(t *testing.T) {
	t.Helper()
	fn := Zero[func() string](nil)
	AssertNil(t, fn, "Zero(nil)")

	// Initialize the function
	fn = func() string { return "initialized" }
	AssertEqual(t, "initialized", fn(), "function result")
}

// isSameTestCase tests IsSame function
type isSameTestCase struct {
	// Large fields first (interfaces, strings) - 8+ bytes
	a           any
	b           any
	description string
	name        string

	// Small fields last (booleans) - 1 byte
	expected bool
}

func (tc isSameTestCase) Name() string {
	return tc.name
}

func (tc isSameTestCase) Test(t *testing.T) {
	t.Helper()

	result := IsSame(tc.a, tc.b)
	AssertEqual(t, tc.expected, result, tc.description)
}

func newIsSameTestCase(name string, a, b any, expected bool,
	description string) isSameTestCase {
	return isSameTestCase{
		name:        name,
		a:           a,
		b:           b,
		expected:    expected,
		description: description,
	}
}

func isSameValueTypeTestCases() []isSameTestCase {
	return S(
		// Value types - compared by equality
		newIsSameTestCase("same integers", 42, 42, true,
			"same integers should be same"),
		newIsSameTestCase("different integers", 42, 43, false,
			"different integers should not be same"),
		newIsSameTestCase("same strings", "hello", "hello", true,
			"same strings should be same"),
		newIsSameTestCase("different strings", "hello", "world", false,
			"different strings should not be same"),
		newIsSameTestCase("same booleans", true, true, true,
			"same booleans should be same"),
		newIsSameTestCase("different booleans", true, false, false,
			"different booleans should not be same"),
		newIsSameTestCase("same floats", 3.14, 3.14, true,
			"same floats should be same"),
		newIsSameTestCase("different floats", 3.14, 2.71, false,
			"different floats should not be same"),
	)
}

func isSameReferenceTypeTestCases() []isSameTestCase {
	slice1 := S(1, 2, 3)
	slice2 := slice1
	slice3 := S(1, 2, 3)

	map1 := map[string]int{"a": 1}
	map2 := map1
	map3 := map[string]int{"a": 1}

	ch1 := make(chan int)
	ch2 := ch1
	ch3 := make(chan int)

	fn1 := func() {}
	fn2 := fn1
	fn3 := func() {}

	x := 42
	ptr1 := &x
	ptr2 := ptr1
	ptr3 := &x

	// Empty slices from separate make() calls
	emptySlice1 := make([]int, 0)
	emptySlice2 := make([]int, 0)

	// Zero-length slices from same backing array
	baseArray := []int{1, 2, 3, 4, 5}
	zeroLen1 := baseArray[2:2] // zero-length slice at position 2
	zeroLen2 := baseArray[2:2] // same zero-length slice
	zeroLen3 := baseArray[3:3] // different zero-length slice from same array

	return S(
		// Slices - compared by backing array pointer
		newIsSameTestCase("same slice reference", slice1, slice2, true,
			"slices with same backing array should be same"),
		newIsSameTestCase("different slice reference", slice1, slice3, false,
			"slices with different backing arrays should not be same"),

		// Empty slice tests
		newIsSameTestCase("two empty slices from separate make()", emptySlice1, emptySlice2, false,
			"empty slices from separate make() calls should not be same"),
		newIsSameTestCase("zero-length slices from same position", zeroLen1, zeroLen2, true,
			"zero-length slices from same array position should be same"),
		newIsSameTestCase("zero-length slices from different positions", zeroLen1, zeroLen3, false,
			"zero-length slices from different positions should not be same"),

		// Maps - compared by map pointer
		newIsSameTestCase("same map reference", map1, map2, true,
			"maps with same reference should be same"),
		newIsSameTestCase("different map reference", map1, map3, false,
			"maps with different references should not be same"),

		// Channels - compared by channel pointer
		newIsSameTestCase("same channel reference", ch1, ch2, true,
			"channels with same reference should be same"),
		newIsSameTestCase("different channel reference", ch1, ch3, false,
			"channels with different references should not be same"),

		// Functions - compared by function pointer
		newIsSameTestCase("same function reference", fn1, fn2, true,
			"functions with same reference should be same"),
		newIsSameTestCase("different function reference", fn1, fn3, false,
			"functions with different references should not be same"),

		// Pointers - compared by address
		newIsSameTestCase("same pointer reference", ptr1, ptr2, true,
			"pointers with same reference should be same"),
		newIsSameTestCase("pointers to same address", ptr1, ptr3, true,
			"pointers to same address should be same"),
	)
}

func isSameNilTestCases() []isSameTestCase {
	var nilSlice1, nilSlice2 []int
	var nilMap1, nilMap2 map[string]int
	var nilPtr1, nilPtr2 *int
	var nilChan1, nilChan2 chan int
	var nilFunc1, nilFunc2 func()

	return S(
		// Untyped nil
		newIsSameTestCase("both untyped nil", nil, nil, true,
			"both untyped nil should be same"),
		newIsSameTestCase("untyped nil vs typed nil", nil, nilPtr1, false,
			"untyped nil vs typed nil should not be same"),

		// Typed nils of same type
		newIsSameTestCase("same type nil slices", nilSlice1, nilSlice2, true,
			"nil slices of same type should be same"),
		newIsSameTestCase("same type nil maps", nilMap1, nilMap2, true,
			"nil maps of same type should be same"),
		newIsSameTestCase("same type nil pointers", nilPtr1, nilPtr2, true,
			"nil pointers of same type should be same"),
		newIsSameTestCase("same type nil channels", nilChan1, nilChan2, true,
			"nil channels of same type should be same"),
		newIsSameTestCase("same type nil functions", nilFunc1, nilFunc2, true,
			"nil functions of same type should be same"),

		// Nil vs non-nil
		newIsSameTestCase("nil vs empty slice", nilSlice1, S[int](), false,
			"nil slice vs empty slice should not be same"),
		newIsSameTestCase("nil vs empty map", nilMap1, map[string]int{}, false,
			"nil map vs empty map should not be same"),
		newIsSameTestCase("nil vs non-nil pointer", nilPtr1, new(int), false,
			"nil pointer vs non-nil pointer should not be same"),
	)
}

func isSameDifferentTypeTestCases() []isSameTestCase {
	return S(
		// Different types should never be same
		newIsSameTestCase("int vs string", 42, "42", false,
			"different types should not be same"),
		newIsSameTestCase("int vs float", 42, 42.0, false,
			"different numeric types should not be same"),
		newIsSameTestCase("slice vs array", S(1, 2, 3), [3]int{1, 2, 3}, false,
			"slice vs array should not be same"),
		newIsSameTestCase("int slice vs string slice", S(1, 2, 3), S("1", "2", "3"), false,
			"slices of different types should not be same"),
	)
}

func isSameInterfaceTestCases() []isSameTestCase {
	x := 42
	ptr1 := &x
	ptr2 := ptr1

	var interface1 any = ptr1
	var interface2 any = ptr2
	var interface3 any = &x

	slice1 := S(1, 2, 3)
	slice2 := slice1
	var interface4 any = slice1
	var interface5 any = slice2

	// Critical nil interface cases for interface comparison coverage
	var nilInterface1 any
	var nilInterface2 any
	var nonNilInterface any = 42

	// Nested interface cases
	var nestedInterface1 = interface1
	var nestedInterface2 = interface2

	return S(
		// Both interfaces nil - handled by isSameTypedNil
		newIsSameTestCase("both interfaces nil", nilInterface1, nilInterface2, true,
			"both nil interfaces should be same"),

		// One nil, one not nil - handled by isSameTypedNil
		newIsSameTestCase("nil vs non-nil interface", nilInterface1, nonNilInterface, false,
			"nil interface vs non-nil interface should not be same"),
		newIsSameTestCase("non-nil vs nil interface", nonNilInterface, nilInterface1, false,
			"non-nil interface vs nil interface should not be same"),

		// Both non-nil interfaces - handled by interface case in isSamePointer
		newIsSameTestCase("interfaces with same pointer", interface1, interface2, true,
			"interfaces containing same pointer should be same"),
		newIsSameTestCase("interfaces with pointers to same address", interface1, interface3, true,
			"interfaces with pointers to same address should be same"),
		newIsSameTestCase("interfaces with same slice", interface4, interface5, true,
			"interfaces containing same slice should be same"),

		// Interfaces with values
		newIsSameTestCase("interfaces with same value", any(42), any(42), true,
			"interfaces with same value should be same"),
		newIsSameTestCase("interfaces with different values", any(42), any(43), false,
			"interfaces with different values should not be same"),

		// Nested interface comparisons - recursive interface handling
		newIsSameTestCase("nested interfaces same", nestedInterface1, nestedInterface2, true,
			"nested interfaces with same content should be same"),
		newIsSameTestCase("interface vs nested interface", interface1, nestedInterface1, true,
			"interface and nested interface with same content should be same"),

		// Interface containing different reference types
		newIsSameTestCase("interfaces with different slices", any(S(1, 2, 3)), any(S(1, 2, 3)), false,
			"interfaces with different slice backing arrays should not be same"),
		newIsSameTestCase("interfaces with different maps",
			any(map[string]int{"a": 1}), any(map[string]int{"a": 1}), false,
			"interfaces with different map references should not be same"),
	)
}

func TestIsSame(t *testing.T) {
	t.Run("value types", func(t *testing.T) {
		RunTestCases(t, isSameValueTypeTestCases())
	})
	t.Run("reference types", func(t *testing.T) {
		RunTestCases(t, isSameReferenceTypeTestCases())
	})
	t.Run("nil handling", func(t *testing.T) {
		RunTestCases(t, isSameNilTestCases())
	})
	t.Run("different types", func(t *testing.T) {
		RunTestCases(t, isSameDifferentTypeTestCases())
	})
	t.Run("interfaces", func(t *testing.T) {
		RunTestCases(t, isSameInterfaceTestCases())
	})
	t.Run("unsafe pointers", func(t *testing.T) {
		RunTestCases(t, isSameUnsafePointerTestCases())
	})
	t.Run("complex numbers", func(t *testing.T) {
		RunTestCases(t, isSameComplexTestCases())
	})
}

func isSameInterfaceReflectValueTestCases() []isSameTestCase {
	// Create interface variables
	var interface1 any = 42
	var interface2 any = 42
	var interface3 any = 43

	// Create reflect.Values with Kind() == reflect.Interface
	v1 := reflect.ValueOf(&interface1).Elem()
	v2 := reflect.ValueOf(&interface2).Elem()
	v3 := reflect.ValueOf(&interface3).Elem()

	return S(
		// Same values in interface reflect.Values
		newIsSameTestCase("interface reflect.Values with same value", v1, v2, true,
			"interface reflect.Values containing same value should be same"),

		// Different values in interface reflect.Values
		newIsSameTestCase("interface reflect.Values with different values", v1, v3, false,
			"interface reflect.Values containing different values should not be same"),
	)
}

// Test reflect.Values with Kind() == reflect.Interface
func TestIsSameInterfaceReflectValues(t *testing.T) {
	RunTestCases(t, isSameInterfaceReflectValueTestCases())
}

func isSameInvalidReflectValueTestCases() []isSameTestCase {
	invalid1 := reflect.Value{}
	invalid2 := reflect.Value{}
	valid := reflect.ValueOf(42)
	var nilValue reflect.Value

	return S(
		// Two invalid reflect.Values - now returns false since we can't compare them
		newIsSameTestCase("two invalid reflect.Values", invalid1, invalid2, false,
			"two invalid reflect.Values cannot be compared"),

		// Invalid vs valid
		newIsSameTestCase("invalid vs valid reflect.Value", invalid1, valid, false,
			"invalid and valid reflect.Value should not be same"),
		newIsSameTestCase("valid vs invalid reflect.Value", valid, invalid1, false,
			"valid and invalid reflect.Value should not be same"),

		// Nil reflect.Value vs invalid
		newIsSameTestCase("nil vs invalid reflect.Value", nilValue, invalid1, false,
			"nil and invalid reflect.Values cannot be compared"),
	)
}

// Also test IsZero and IsNil with invalid reflect.Values
func isInvalidReflectValueZeroNilTestCases() []isZeroTestCase {
	invalid := reflect.Value{}

	return S(
		newIsZeroTestCase("invalid reflect.Value IsZero", invalid, true),
	)
}

func isInvalidReflectValueNilTestCases() []isNilTestCase {
	invalid := reflect.Value{}

	return S(
		newIsNilTestCase("invalid reflect.Value IsNil", invalid, true),
	)
}

// Test invalid reflect.Values in IsSame
func TestIsSameInvalidReflectValues(t *testing.T) {
	t.Run("IsSame", func(t *testing.T) {
		RunTestCases(t, isSameInvalidReflectValueTestCases())
	})

	t.Run("IsZero", func(t *testing.T) {
		RunTestCases(t, isInvalidReflectValueZeroNilTestCases())
	})

	t.Run("IsNil", func(t *testing.T) {
		RunTestCases(t, isInvalidReflectValueNilTestCases())
	})
}

// isSameStackOverflowTestCase tests stack overflow scenarios for IsSame function
type isSameStackOverflowTestCase struct {
	// Large fields first - interfaces and strings (8+ bytes)
	setupFunc   func() (any, any)
	description string
	name        string

	// Small fields last - booleans (1 byte)
	expected bool
}

func (tc isSameStackOverflowTestCase) Name() string {
	return tc.name
}

func (tc isSameStackOverflowTestCase) Test(t *testing.T) {
	t.Helper()

	a, b := tc.setupFunc()
	result := IsSame(a, b)
	AssertEqual(t, tc.expected, result, tc.description)
}

func newIsSameStackOverflowTestCase(name string,
	setupFunc func() (any, any), expected bool,
	description string) isSameStackOverflowTestCase {
	return isSameStackOverflowTestCase{
		name:        name,
		setupFunc:   setupFunc,
		expected:    expected,
		description: description,
	}
}

func isSameStackOverflowTestCases() []isSameStackOverflowTestCase {
	return S(
		// Circular interface references
		newIsSameStackOverflowTestCase("circular interface references", func() (any, any) {
			var a, b any
			a = &b
			b = &a
			return a, b
		}, false, "circular interface references should not be same"),

		newIsSameStackOverflowTestCase("same circular reference", func() (any, any) {
			var a any
			a = &a
			c := &a
			d := &a
			return c, d
		}, true, "pointers to same circular reference should be same"),

		// Deeply nested interfaces
		newIsSameStackOverflowTestCase("different deep nested chains", func() (any, any) {
			var current1 any = 42
			for i := 0; i < 100; i++ {
				next := current1
				current1 = &next
			}

			var current2 any = 42
			for i := 0; i < 100; i++ {
				next := current2
				current2 = &next
			}

			return current1, current2
		}, false, "different deep nested chains should not be same"),

		newIsSameStackOverflowTestCase("same deep nested chain", func() (any, any) {
			var current any = 42
			for i := 0; i < 100; i++ {
				next := current
				current = &next
			}

			return current, current
		}, true, "same deep nested chain should be same"),

		// Self-referential structures
		newIsSameStackOverflowTestCase("different self-referential nodes", func() (any, any) {
			type Node struct {
				Next  *Node
				Value int
			}

			node1 := &Node{Value: 1}
			node1.Next = node1

			node2 := &Node{Value: 1}
			node2.Next = node2

			var interface1 any = node1
			var interface2 any = node2

			return interface1, interface2
		}, false, "different self-referential nodes should not be same"),

		newIsSameStackOverflowTestCase("same self-referential node", func() (any, any) {
			type Node struct {
				Next  *Node
				Value int
			}

			node1 := &Node{Value: 1}
			node1.Next = node1

			var interface1 any = node1
			var interface2 any = node1

			return interface1, interface2
		}, true, "same self-referential node should be same"),

		newIsSameStackOverflowTestCase("nested interface with self-referential content", func() (any, any) {
			type Node struct {
				Next  *Node
				Value int
			}

			node := &Node{Value: 1}
			node.Next = node

			var interface1 any = node
			nested1 := interface1
			nested2 := interface1

			return nested1, nested2
		}, true, "nested interfaces with same self-referential content should be same"),
	)
}

func isSameUnsafePointerTestCases() []isSameTestCase {
	x := 42
	ptr := &x

	// skipcq: GSC-G103 - Testing IsSame's handling of unsafe.Pointer
	unsafePtr1 := unsafe.Pointer(ptr)
	unsafePtr2 := unsafePtr1
	// skipcq: GSC-G103 - Testing IsSame's handling of unsafe.Pointer
	unsafePtr3 := unsafe.Pointer(&x) // Same address, different unsafe.Pointer

	y := 42
	// skipcq: GSC-G103 - Testing IsSame's handling of unsafe.Pointer
	unsafePtr4 := unsafe.Pointer(&y) // Different address

	return S(
		newIsSameTestCase("same unsafe pointer", unsafePtr1, unsafePtr2, true,
			"same unsafe pointer should be same"),
		newIsSameTestCase("unsafe pointers to same address", unsafePtr1, unsafePtr3, true,
			"unsafe pointers to same address should be same"),
		newIsSameTestCase("unsafe pointers to different addresses", unsafePtr1, unsafePtr4, false,
			"unsafe pointers to different addresses should not be same"),

		// Nil unsafe pointers
		newIsSameTestCase("nil unsafe pointers",
			// skipcq: GSC-G103 - Testing IsSame's handling of unsafe.Pointer
			unsafe.Pointer(nil), unsafe.Pointer(nil), true,
			"nil unsafe pointers should be same"),
	)
}

func isSameComplexTestCases() []isSameTestCase {
	return S(
		// Complex64
		newIsSameTestCase("same complex64", complex64(1+2i), complex64(1+2i), true,
			"same complex64 values should be same"),
		newIsSameTestCase("different complex64", complex64(1+2i), complex64(1+3i), false,
			"different complex64 values should not be same"),

		// Complex128
		newIsSameTestCase("same complex128", complex128(1+2i), complex128(1+2i), true,
			"same complex128 values should be same"),
		newIsSameTestCase("different complex128", complex128(1+2i), complex128(2+2i), false,
			"different complex128 values should not be same"),

		// Edge cases with NaN and Inf
		newIsSameTestCase("complex with NaN",
			complex(math.NaN(), 1), complex(math.NaN(), 1), false,
			"complex with NaN should not be same (NaN != NaN)"),
		newIsSameTestCase("complex with Inf",
			complex(math.Inf(1), 1), complex(math.Inf(1), 1), true,
			"complex with same Inf should be same"),
	)
}

func TestIsSameStackOverflow(t *testing.T) {
	RunTestCases(t, isSameStackOverflowTestCases())
}
