package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// TestCase validations
var _ TestCase = zeroTestCase[int]{}
var _ TestCase = zeroRefTestCase[int]{}
var _ TestCase = isZeroTestCase{}
var _ TestCase = isNilTestCase{}
var _ TestCase = isNilVsIsZeroTestCase{}
var _ TestCase = initializationSemanticsTestCase{}

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

func TestIsZeroReflectValue(t *testing.T) {
	// Test with invalid reflect.Value
	var v reflect.Value
	result := IsZero(v)
	AssertTrue(t, result, "IsZero with zero reflect.Value")

	// Test with valid reflect.Value containing zero
	v2 := reflect.ValueOf(0)
	result2 := IsZero(v2)
	AssertTrue(t, result2, "IsZero with valid reflect.Value containing zero")

	// Test with valid reflect.Value containing non-zero
	v3 := reflect.ValueOf(42)
	result3 := IsZero(v3)
	AssertFalse(t, result3, "IsZero with valid reflect.Value containing non-zero")
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

func TestIsNilWithReflectValue(t *testing.T) {
	// Test that an invalid reflect.Value is considered nil
	var invalidValue reflect.Value
	result := IsNil(invalidValue)
	AssertTrue(t, result, "IsNil with invalid reflect.Value")

	// Test that a valid reflect.Value with nil content is considered nil
	var nilPtr *int
	nilPtrValue := reflect.ValueOf(nilPtr)
	result2 := IsNil(nilPtrValue)
	AssertTrue(t, result2, "IsNil with reflect.Value containing nil")

	// Test that a valid reflect.Value with non-nil content is not considered nil
	nonNilPtr := new(int)
	nonNilPtrValue := reflect.ValueOf(nonNilPtr)
	result3 := IsNil(nonNilPtrValue)
	AssertFalse(t, result3, "IsNil with reflect.Value containing non-nil")

	// Test that a valid reflect.Value with basic type is not considered nil
	intValue := reflect.ValueOf(42)
	result4 := IsNil(intValue)
	AssertFalse(t, result4, "IsNil with reflect.Value containing basic type")
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
func TestIsNilReflectValue(t *testing.T) {
	// Test with invalid reflect.Value
	var v reflect.Value
	result := IsNil(v)
	AssertEqual(t, true, result, "IsNil(invalid reflect.Value)")

	// Test with valid reflect.Value containing nil
	var p *int
	v2 := reflect.ValueOf(p)
	result2 := IsNil(v2)
	AssertEqual(t, true, result2, "IsNil(reflect.Value nil)")

	// Test with valid reflect.Value containing non-nil
	i := 42
	v3 := reflect.ValueOf(&i)
	result3 := IsNil(v3)
	AssertEqual(t, false, result3, "IsNil(reflect.Value non-nil)")

	// Test with reflect.Value containing basic type
	v4 := reflect.ValueOf(42)
	result4 := IsNil(v4)
	AssertEqual(t, false, result4, "IsNil(reflect.Value basic)")
}

// Test IsNil with typed nil in interface
func TestIsNilTypedInterface(t *testing.T) {
	var p *int
	var vi any = p
	AssertEqual(t, true, IsNil(vi), "typed nil")

	// Test slice with nil elements
	ptrSlice := []*int{nil, nil, nil} // array literal, not slice
	AssertEqual(t, false, IsNil(ptrSlice), "slice with nils")
	AssertEqual(t, true, IsNil(ptrSlice[0]), "nil element")

	// Test map with nil values
	nilMap := map[string]*int{"key": nil}
	AssertFalse(t, IsNil(nilMap), "map with nil values")
	AssertTrue(t, IsNil(nilMap["key"]), "nil value in map")

	// Channel operations
	var ch chan int
	AssertTrue(t, IsNil(ch), "nil channel")

	ch = make(chan int)
	close(ch)
	AssertFalse(t, IsNil(ch), "closed channel")
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
