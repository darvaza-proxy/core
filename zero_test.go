package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type zeroTestCase[T comparable] struct {
	expected T
	factory  func() *T
	name     string
}

func (tc zeroTestCase[T]) test(t *testing.T) {
	t.Helper()
	result := Zero(tc.factory())
	AssertEqual(t, tc.expected, result, "Zero[%s]", tc.name)
}

// Generic test case for reference types that should return nil
type zeroRefTestCase[T any] struct {
	factory func() *T
	name    string
}

func (tc zeroRefTestCase[T]) test(t *testing.T) {
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

// testZeroT tests Zero function for reference types that should return nil
func testZeroT[T any](t *testing.T, testCases ...zeroRefTestCase[T]) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
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
	t.Run("int", func(t *testing.T) { runZeroTestCases(t, zeroIntTestCases) })
	t.Run("string", func(t *testing.T) { runZeroTestCases(t, zeroStringTestCases) })
	t.Run("bool", func(t *testing.T) { runZeroTestCases(t, zeroBoolTestCases) })
	t.Run("slice", func(t *testing.T) { testZeroT(t, zeroSliceTestCases...) })
	t.Run("map", func(t *testing.T) { testZeroT(t, zeroMapTestCases...) })
	t.Run("pointer", func(t *testing.T) { testZeroT(t, zeroPointerTestCases...) })
	t.Run("struct", testZeroStruct)
	t.Run("interface", func(t *testing.T) { testZeroT(t, zeroInterfaceTestCases...) })
	t.Run("channel", func(t *testing.T) { testZeroT(t, zeroChannelTestCases...) })
	t.Run("func", func(t *testing.T) { testZeroT(t, zeroFuncTestCases...) })
	t.Run("time", func(t *testing.T) { runZeroTestCases(t, zeroTimeTestCases) })
}

func runZeroTestCases[T comparable](t *testing.T, testCases []zeroTestCase[T]) {
	t.Helper()
	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func testZeroStruct(t *testing.T) {
	t.Helper()

	type testStruct struct {
		Name string
		Age  int
	}

	structTests := []zeroTestCase[testStruct]{
		{
			expected: testStruct{},
			factory:  func() *testStruct { return nil },
			name:     "nil pointer",
		},
		{
			expected: testStruct{},
			factory:  func() *testStruct { v := testStruct{Name: "John", Age: 30}; return &v },
			name:     "non-nil pointer",
		},
	}

	for _, tc := range structTests {
		t.Run(tc.name, tc.test)
	}
}

type isZeroTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isZeroTestCase) test(t *testing.T) {
	t.Helper()

	result := IsZero(tc.value)
	AssertEqual(t, tc.expected, result, "IsZero")
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

	for _, tc := range zeroTests {
		t.Run(tc.name, tc.test)
	}
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

	for _, tc := range interfaceTests {
		t.Run(tc.name, tc.test)
	}
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

	for _, tc := range edgeTests {
		t.Run(tc.name, tc.test)
	}
}

func TestZeroEdgeCases(t *testing.T) {
	// Test Zero function returns expected values when passed nil
	t.Run("nil slice pointer", func(t *testing.T) {
		items := Zero[[]string](nil)
		AssertNil(t, items, "Zero(nil)")

		// Initialize the slice
		items = S("default", "values")
		AssertSliceEqual(t, S("default", "values"), items, "initialized slice")

		// Check that it's not zero after initialization
		AssertEqual(t, false, IsZero(items), "IsZero(initialized)")
	})

	t.Run("nil map pointer", func(t *testing.T) {
		cache := Zero[map[string]int](nil)
		AssertNil(t, cache, "Zero(nil)")

		// Initialize and use the map
		cache = make(map[string]int)
		cache["key"] = 42
		AssertEqual(t, 42, cache["key"], "map value")
	})

	t.Run("nil pointer pointer", func(t *testing.T) {
		ptr := Zero[*int](nil)
		AssertNil(t, ptr, "Zero(nil)")

		// Initialize the pointer
		value := 100
		ptr = &value
		AssertEqual(t, 100, *ptr, "pointer value")
	})

	t.Run("nil channel pointer", func(t *testing.T) {
		ch := Zero[chan int](nil)
		AssertNil(t, ch, "Zero(nil)")

		// Initialize and use the channel
		ch = make(chan int, 1)
		ch <- 42
		result := <-ch
		AssertEqual(t, 42, result, "channel value")
	})

	t.Run("nil func pointer", func(t *testing.T) {
		fn := Zero[func() string](nil)
		AssertNil(t, fn, "Zero(nil)")

		// Initialize the function
		fn = func() string { return "initialized" }
		AssertEqual(t, "initialized", fn(), "function result")
	})
}

func TestIsZeroReflectValue(t *testing.T) {
	// Test with invalid reflect.Value
	var v reflect.Value
	result := IsZero(v)
	AssertEqual(t, true, result, "IsZero(zero reflect.Value)")

	// Test with valid reflect.Value containing zero
	v2 := reflect.ValueOf(0)
	result2 := IsZero(v2)
	AssertEqual(t, true, result2, "IsZero(valid reflect.Value with zero)")

	// Test with valid reflect.Value containing non-zero
	v3 := reflect.ValueOf(42)
	result3 := IsZero(v3)
	AssertEqual(t, false, result3, "IsZero(valid reflect.Value with non-zero)")
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

	for _, tc := range complexTests {
		t.Run(tc.name, tc.test)
	}
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

	for _, tc := range ptrTests {
		t.Run(tc.name, tc.test)
	}
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

	for _, tc := range nestedTests {
		t.Run(tc.name, tc.test)
	}
}

// isNilTestCase tests IsNil function
type isNilTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isNilTestCase) test(t *testing.T) {
	t.Helper()

	result := IsNil(tc.value)
	AssertEqual(t, tc.expected, result, "IsNil")
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

	for _, tc := range nilTests {
		t.Run(tc.name, tc.test)
	}
}

func testIsNilPointers(t *testing.T) {
	t.Helper()

	var nilPtr *int
	nonNilPtr := new(int)

	ptrTests := []isNilTestCase{
		newIsNilTestCase("nil pointer", nilPtr, true),
		newIsNilTestCase("non-nil pointer", nonNilPtr, false),
	}

	for _, tc := range ptrTests {
		t.Run(tc.name, tc.test)
	}
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

	for _, tc := range sliceTests {
		t.Run(tc.name, tc.test)
	}
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

	for _, tc := range mapTests {
		t.Run(tc.name, tc.test)
	}
}

func testIsNilChannels(t *testing.T) {
	t.Helper()

	var nilChan chan int
	nonNilChan := make(chan int)

	chanTests := []isNilTestCase{
		newIsNilTestCase("nil channel", nilChan, true),
		newIsNilTestCase("non-nil channel", nonNilChan, false),
	}

	for _, tc := range chanTests {
		t.Run(tc.name, tc.test)
	}
}

func testIsNilFunctions(t *testing.T) {
	t.Helper()

	var nilFunc func()
	nonNilFunc := func() {}

	funcTests := []isNilTestCase{
		newIsNilTestCase("nil function", nilFunc, true),
		newIsNilTestCase("non-nil function", nonNilFunc, false),
	}

	for _, tc := range funcTests {
		t.Run(tc.name, tc.test)
	}
}

func testIsNilInterfaces(t *testing.T) {
	t.Helper()

	var nilInterface fmt.Stringer
	var nonNilInterface fmt.Stringer = time.Second // time.Duration implements fmt.Stringer

	interfaceTests := []isNilTestCase{
		newIsNilTestCase("nil interface", nilInterface, true),
		newIsNilTestCase("non-nil interface", nonNilInterface, false),
	}

	for _, tc := range interfaceTests {
		t.Run(tc.name, tc.test)
	}
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
	AssertEqual(t, false, IsNil(nilMap), "map with nils")
	AssertEqual(t, true, IsNil(nilMap["key"]), "nil value")

	// Test closed channel
	ch := make(chan int)
	close(ch)
	AssertEqual(t, false, IsNil(ch), "closed channel")
}
