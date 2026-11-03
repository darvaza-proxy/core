package core_test

import (
	"reflect"
	"testing"
	"time"

	"darvaza.org/core"
)

// TestCase validations
var (
	_ core.TestCase = zeroTestCase[int]{}
	_ core.TestCase = zeroRefTestCase[int]{}
	_ core.TestCase = isZeroTestCase{}
	_ core.TestCase = initializationSemanticsTestCase{}
	_ core.TestCase = isZeroCustomInterfaceTestCase{}
)

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
	result := core.Zero(tc.factory())
	core.AssertEqual(t, tc.expected, result, "Zero[%T]", tc.expected)
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
	result := core.Zero(tc.factory())
	core.AssertNil(t, result, "Zero[%T]", *new(T))
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

var zeroIntTestCases = core.S(
	newZeroTestCase("nil pointer", 0, func() *int { return nil }),
	newZeroTestCase("non-nil pointer", 0, func() *int { v := 42; return &v }),
)

var zeroStringTestCases = core.S(
	newZeroTestCase("nil pointer", "", func() *string { return nil }),
	newZeroTestCase("non-nil pointer", "", func() *string { v := "hello"; return &v }),
)

var zeroBoolTestCases = core.S(
	newZeroTestCase("nil pointer", false, func() *bool { return nil }),
	newZeroTestCase("non-nil pointer", false, func() *bool { v := true; return &v }),
)

var zeroTimeTestCases = core.S(
	newZeroTestCase("nil pointer", time.Time{}, func() *time.Time { return nil }),
	newZeroTestCase("non-nil pointer", time.Time{}, func() *time.Time { v := time.Now(); return &v }),
)

var zeroSliceTestCases = core.S(
	newZeroRefTestCase("nil slice pointer", func() *[]int { return nil }),
	newZeroRefTestCase("non-nil slice pointer", func() *[]int { v := core.S(1, 2, 3); return &v }),
)

var zeroMapTestCases = core.S(
	newZeroRefTestCase("nil map pointer", func() *map[string]int { return nil }),
	newZeroRefTestCase("non-nil map pointer", func() *map[string]int { v := map[string]int{"a": 1}; return &v }),
)

var zeroPointerTestCases = core.S(
	newZeroRefTestCase("nil double pointer", func() **int { return nil }),
	newZeroRefTestCase("non-nil double pointer", func() **int { v := 42; vp := &v; return &vp }),
)

var zeroInterfaceTestCases = core.S(
	newZeroRefTestCase("nil interface pointer", func() *any { return nil }),
	newZeroRefTestCase("non-nil interface pointer", func() *any { var v any = 42; return &v }),
)

var zeroChannelTestCases = core.S(
	newZeroRefTestCase("nil channel pointer", func() *chan int { return nil }),
	newZeroRefTestCase("non-nil channel pointer", func() *chan int { ch := make(chan int); return &ch }),
)

var zeroFuncTestCases = core.S(
	newZeroRefTestCase("nil func pointer", func() *func() { return nil }),
	newZeroRefTestCase("non-nil func pointer", func() *func() { fn := func() {}; return &fn }),
)

func TestZero(t *testing.T) {
	t.Run("int", func(t *testing.T) { core.RunTestCases(t, zeroIntTestCases) })
	t.Run("string", func(t *testing.T) { core.RunTestCases(t, zeroStringTestCases) })
	t.Run("bool", func(t *testing.T) { core.RunTestCases(t, zeroBoolTestCases) })
	t.Run("slice", func(t *testing.T) { core.RunTestCases(t, zeroSliceTestCases) })
	t.Run("map", func(t *testing.T) { core.RunTestCases(t, zeroMapTestCases) })
	t.Run("pointer", func(t *testing.T) { core.RunTestCases(t, zeroPointerTestCases) })
	t.Run("struct", testZeroStruct)
	t.Run("interface", func(t *testing.T) { core.RunTestCases(t, zeroInterfaceTestCases) })
	t.Run("channel", func(t *testing.T) { core.RunTestCases(t, zeroChannelTestCases) })
	t.Run("func", func(t *testing.T) { core.RunTestCases(t, zeroFuncTestCases) })
	t.Run("time", func(t *testing.T) { core.RunTestCases(t, zeroTimeTestCases) })
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

	core.RunTestCases(t, structTests)
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

	result := core.IsZero(tc.value)
	core.AssertEqual(t, tc.expected, result, "IsZero result")
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

	core.RunTestCases(t, zeroTests)
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
		newIsZeroTestCase("interface containing empty slice", any(core.S[int]()),
			false), // empty slice is not a zero value
		// interface containing nil slice
		newIsZeroTestCase("interface containing nil slice", any([]int(nil)), true),
	}

	core.RunTestCases(t, interfaceTests)
}

// Test IsZero with edge cases
func TestIsZeroEdgeCases(t *testing.T) {
	edgeTests := []isZeroTestCase{
		// Slices
		newIsZeroTestCase("nil slice", []int(nil), true),
		newIsZeroTestCase("empty slice", core.S[int](), false),
		newIsZeroTestCase("non-empty slice", core.S(1, 2, 3), false),
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

	core.RunTestCases(t, edgeTests)
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
	return core.S(
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
	core.RunTestCases(t, isZeroReflectValueTestCases())
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
			SliceField: core.S[int](),
		}, false),
		newIsZeroTestCase("struct with nil map", complexStruct{
			MapField: nil,
		}, true),
		newIsZeroTestCase("struct with empty map", complexStruct{
			MapField: map[string]int{},
		}, false),
	}

	core.RunTestCases(t, complexTests)
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

	core.RunTestCases(t, ptrTests)
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

	core.RunTestCases(t, nestedTests)
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

	result := core.IsZero(tc.value)
	core.AssertEqual(t, tc.expected, result, tc.description)
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
	return core.S(
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
	core.RunTestCases(t, initializationSemanticsTestCases())
}

func TestIsZeroPracticalInitialization(t *testing.T) {
	// Demonstrate practical initialization patterns using IsZero

	// Example 1: Lazy initialization of slice
	var items []string
	if core.IsZero(items) {
		items = []string{"default", "values"}
	}
	core.AssertSliceEqual(t, []string{"default", "values"}, items, "slice should be initialized")

	// Items is now initialized, so IsZero returns false
	core.AssertFalse(t, core.IsZero(items), "initialized slice zero")

	// Example 2: Conditional map initialization
	var cache map[string]int
	if core.IsZero(cache) {
		cache = make(map[string]int)
	}
	cache["key"] = 42
	core.AssertEqual(t, 42, cache["key"], "map should be usable after initialization")

	// Example 3: Pointer initialization
	var ptr *int
	if core.IsZero(ptr) {
		value := 100
		ptr = &value
	}
	core.AssertEqual(t, 100, *ptr, "pointer should point to initialized value")

	// Example 4: Channel initialization
	var ch chan int
	if core.IsZero(ch) {
		ch = make(chan int, 1)
	}
	ch <- 42
	result := <-ch
	core.AssertEqual(t, 42, result, "channel should be usable after initialization")

	// Example 5: Function initialization
	var fn func() string
	if core.IsZero(fn) {
		fn = func() string { return "initialized" }
	}
	core.AssertEqual(t, "initialized", fn(), "function should be callable after initialization")
}

func testZeroEdgeCasesNilSlicePointer(t *testing.T) {
	t.Helper()
	items := core.Zero[[]string](nil)
	core.AssertNil(t, items, "Zero(nil)")

	// Initialize the slice
	items = core.S("default", "values")
	core.AssertSliceEqual(t, core.S("default", "values"), items, "initialized slice")

	// Check that it's not zero after initialization
	core.AssertEqual(t, false, core.IsZero(items), "IsZero(initialized)")
}

func testZeroEdgeCasesNilMapPointer(t *testing.T) {
	t.Helper()
	cache := core.Zero[map[string]int](nil)
	core.AssertNil(t, cache, "Zero(nil)")

	// Initialize and use the map
	cache = make(map[string]int)
	cache["key"] = 42
	core.AssertEqual(t, 42, cache["key"], "map value")
}

func testZeroEdgeCasesNilPointerPointer(t *testing.T) {
	t.Helper()
	ptr := core.Zero[*int](nil)
	core.AssertNil(t, ptr, "Zero(nil)")

	// Initialize the pointer
	value := 100
	ptr = &value
	core.AssertEqual(t, 100, *ptr, "pointer value")
}

func testZeroEdgeCasesNilChannelPointer(t *testing.T) {
	t.Helper()
	ch := core.Zero[chan int](nil)
	core.AssertNil(t, ch, "Zero(nil)")

	// Initialize and use the channel
	ch = make(chan int, 1)
	ch <- 42
	result := <-ch
	core.AssertEqual(t, 42, result, "channel value")
}

func testZeroEdgeCasesNilFuncPointer(t *testing.T) {
	t.Helper()
	fn := core.Zero[func() string](nil)
	core.AssertNil(t, fn, "Zero(nil)")

	// Initialize the function
	fn = func() string { return "initialized" }
	core.AssertEqual(t, "initialized", fn(), "function result")
}

// Custom type that implements IsZero() bool for the interface branch.
type hasIsZeroType struct {
	zero bool
}

func (h hasIsZeroType) IsZero() bool { return h.zero }

type isZeroCustomInterfaceTestCase struct {
	name string
	in   hasIsZeroType
	want bool
}

func (tc isZeroCustomInterfaceTestCase) Name() string { return tc.name }

func (tc isZeroCustomInterfaceTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, core.IsZero(tc.in), tc.name)
}

func newIsZeroCustomInterfaceTestCase(name string, in hasIsZeroType,
	want bool) isZeroCustomInterfaceTestCase {
	return isZeroCustomInterfaceTestCase{name: name, in: in, want: want}
}

// Cover the `interface{ IsZero() bool }` branch of IsZero.
func TestIsZeroCustomInterface(t *testing.T) {
	core.RunTestCases(t, []isZeroCustomInterfaceTestCase{
		newIsZeroCustomInterfaceTestCase("custom IsZero returns true",
			hasIsZeroType{zero: true}, true),
		newIsZeroCustomInterfaceTestCase("custom IsZero returns false",
			hasIsZeroType{zero: false}, false),
	})
}

// Benchmarks
//
// AreEqual buys its panic-free breadth with reflection, and these measure
// what that costs against a plain ==. They exist to document the gap, not
// to be closed: the Assert family that calls AreEqual runs in tests, where
// a decisive answer and a useful message beat nanoseconds.

// benchEqualSink keeps the compiler from eliding the comparisons below.
var benchEqualSink bool

func BenchmarkEqualOperator(b *testing.B) {
	x, y := 42, 42
	for b.Loop() {
		benchEqualSink = x == y
	}
}

func BenchmarkAreEqualComparable(b *testing.B) {
	x, y := 42, 42
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}

func BenchmarkAreEqualMethod(b *testing.B) {
	x, y := core.EqSliceForTest{1, 2, 3}, core.EqSliceForTest{1, 2, 3}
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}

func BenchmarkAreEqualSlice(b *testing.B) {
	x, y := core.S(1, 2, 3), core.S(1, 2, 3)
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}

// BenchmarkAreEqualUndecided measures the walk that ends in no answer,
// the path Assert functions turn into a failure.
func BenchmarkAreEqualUndecided(b *testing.B) {
	x, y := map[string]int{"a": 1}, map[string]int{"a": 1}
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}
