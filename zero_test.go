package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type zeroTestCase struct {
	testFunc func(t *testing.T)
	name     string
}

func newZeroTestCase(name string, testFunc func(t *testing.T)) zeroTestCase {
	return zeroTestCase{
		name:     name,
		testFunc: testFunc,
	}
}

func zeroTestCases() []zeroTestCase {
	return S(
		newZeroTestCase("int", testZeroInt),
		newZeroTestCase("string", testZeroString),
		newZeroTestCase("bool", testZeroBool),
		newZeroTestCase("slice", testZeroSlice),
		newZeroTestCase("map", testZeroMap),
		newZeroTestCase("pointer", testZeroPointer),
		newZeroTestCase("struct", testZeroStruct),
		newZeroTestCase("interface", testZeroInterface),
		newZeroTestCase("channel", testZeroChannel),
		newZeroTestCase("func", testZeroFunc),
		newZeroTestCase("time", testZeroTime),
	)
}

func TestZero(t *testing.T) {
	for _, tc := range zeroTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t)
		})
	}
}

func testZeroInt(t *testing.T) {
	var p *int
	result := Zero(p)
	AssertEqual(t, 0, result, "Zero() should return 0 for int")
}

func testZeroString(t *testing.T) {
	var p *string
	result := Zero(p)
	AssertEqual(t, "", result, "Zero() should return empty string for string")
}

func testZeroBool(t *testing.T) {
	var p *bool
	result := Zero(p)
	AssertEqual(t, false, result, "Zero() should return false for bool")
}

func testZeroSlice(t *testing.T) {
	var p *[]int
	result := Zero(p)
	if result != nil {
		t.Errorf("Zero() should return nil slice for []int, got %v", result)
	}
}

func testZeroMap(t *testing.T) {
	var p *map[string]int
	result := Zero(p)
	if result != nil {
		t.Errorf("Zero() should return nil map for map[string]int, got %v", result)
	}
}

func testZeroPointer(t *testing.T) {
	var p **int
	result := Zero(p)
	AssertEqual(t, (*int)(nil), result, "Zero() should return nil pointer for *int")
}

func testZeroStruct(t *testing.T) {
	type testStruct struct {
		Name string
		Age  int
	}
	var p *testStruct
	result := Zero(p)
	expected := testStruct{}
	AssertEqual(t, expected, result, "Zero() should return zero struct for testStruct")
}

func testZeroInterface(t *testing.T) {
	var p *any
	result := Zero(p)
	AssertEqual(t, any(nil), result, "Zero() should return nil interface for any")
}

func testZeroChannel(t *testing.T) {
	var p *chan int
	result := Zero(p)
	AssertEqual(t, (chan int)(nil), result, "Zero() should return nil channel for chan int")
}

func testZeroFunc(t *testing.T) {
	var p *func()
	result := Zero(p)
	if result != nil {
		t.Error("Zero() should return nil func for func()")
	}
}

func testZeroTime(t *testing.T) {
	var p *time.Time
	result := Zero(p)
	AssertEqual(t, time.Time{}, result, "Zero() should return zero time for time.Time")
}

type isZeroTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isZeroTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, "IsZero() result")
}

func newIsZeroTestCase(name string, value any, expected bool) isZeroTestCase {
	return isZeroTestCase{
		name:     name,
		value:    value,
		expected: expected,
	}
}

func isZeroTestCases() []isZeroTestCase {
	return S(
		newIsZeroTestCase("nil", nil, true),
		newIsZeroTestCase("zero int", 0, true),
		newIsZeroTestCase("non-zero int", 42, false),
		newIsZeroTestCase("zero string", "", true),
		newIsZeroTestCase("non-zero string", "hello", false),
		newIsZeroTestCase("zero bool", false, true),
		newIsZeroTestCase("non-zero bool", true, false),
		newIsZeroTestCase("nil slice", []int(nil), true),
		newIsZeroTestCase("empty slice", []int{}, false),
		newIsZeroTestCase("non-empty slice", []int{1, 2, 3}, false),
		newIsZeroTestCase("nil map", map[string]int(nil), true),
		newIsZeroTestCase("empty map", map[string]int{}, false),
		newIsZeroTestCase("non-empty map", map[string]int{"a": 1}, false),
		newIsZeroTestCase("nil pointer", (*int)(nil), true),
		newIsZeroTestCase("non-nil pointer", new(int), false),
		newIsZeroTestCase("zero struct", struct{ Name string }{}, true),
		newIsZeroTestCase("non-zero struct", struct{ Name string }{Name: "test"}, false),
		newIsZeroTestCase("nil interface", any(nil), true),
		newIsZeroTestCase("non-nil interface", any(42), false),
		newIsZeroTestCase("nil channel", (chan int)(nil), true),
		newIsZeroTestCase("non-nil channel", make(chan int), false),
		newIsZeroTestCase("nil func", (func())(nil), true),
		newIsZeroTestCase("non-nil func", func() {}, false),
		newIsZeroTestCase("zero time", time.Time{}, true),
		newIsZeroTestCase("non-zero time", time.Now(), false),
	)
}

func TestIsZero(t *testing.T) {
	for _, tc := range isZeroTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type zeroChecker struct {
	isZero bool
}

func (z zeroChecker) IsZero() bool {
	return z.isZero
}

type isZeroInterfaceTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isZeroInterfaceTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, "IsZero() with interface")
}

func newIsZeroInterfaceTestCase(name string, value any, expected bool) isZeroInterfaceTestCase {
	return isZeroInterfaceTestCase{
		name:     name,
		value:    value,
		expected: expected,
	}
}

func isZeroInterfaceTestCases() []isZeroInterfaceTestCase {
	return S(
		newIsZeroInterfaceTestCase("zero checker - true", zeroChecker{isZero: true}, true),
		newIsZeroInterfaceTestCase("zero checker - false", zeroChecker{isZero: false}, false),
		newIsZeroInterfaceTestCase("pointer to zero checker - true", &zeroChecker{isZero: true}, true),
		newIsZeroInterfaceTestCase("pointer to zero checker - false", &zeroChecker{isZero: false}, false),
	)
}

func TestIsZeroWithInterface(t *testing.T) {
	for _, tc := range isZeroInterfaceTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type isZeroEdgeCaseTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc isZeroEdgeCaseTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, "IsZero() edge case")
}

func newIsZeroEdgeCaseTestCase(name string, value any, expected bool) isZeroEdgeCaseTestCase {
	return isZeroEdgeCaseTestCase{
		name:     name,
		value:    value,
		expected: expected,
	}
}

func isZeroEdgeCaseTestCases() []isZeroEdgeCaseTestCase {
	return S(
		newIsZeroEdgeCaseTestCase("typed nil pointer", (*int)(nil), true),
		newIsZeroEdgeCaseTestCase("typed nil slice", []int(nil), true),
		newIsZeroEdgeCaseTestCase("typed nil map", map[string]int(nil), true),
		newIsZeroEdgeCaseTestCase("typed nil channel", (chan int)(nil), true),
		newIsZeroEdgeCaseTestCase("typed nil func", (func())(nil), true),
		newIsZeroEdgeCaseTestCase("typed nil interface", (*any)(nil), true),
		newIsZeroEdgeCaseTestCase("empty array", [0]int{}, true),
		newIsZeroEdgeCaseTestCase("non-empty array", [1]int{0}, true),
		newIsZeroEdgeCaseTestCase("zero-value array", [3]int{0, 0, 0}, true),
		newIsZeroEdgeCaseTestCase("non-zero-value array", [3]int{1, 0, 0}, false),
	)
}

func TestIsZeroEdgeCases(t *testing.T) {
	for _, tc := range isZeroEdgeCaseTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type initializationSemanticsTestCase struct {
	value       any
	name        string
	description string
	expected    bool
}

func (tc initializationSemanticsTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, tc.description)
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
	for _, tc := range initializationSemanticsTestCases() {
		t.Run(tc.name, tc.test)
	}
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
	AssertBool(t, IsZero(items), false, "initialized slice should not be zero")

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

func TestIsZeroWithReflectValue(t *testing.T) {
	// Test that a zero reflect.Value is considered zero
	var zeroValue reflect.Value
	result := IsZero(zeroValue)
	AssertBool(t, result, true, "IsZero() with zero reflect.Value should return true")

	// Test that a valid reflect.Value is not considered zero
	validValue := reflect.ValueOf(42)
	result2 := IsZero(validValue)
	AssertBool(t, result2, false, "IsZero() with valid reflect.Value should return false")
}

type complexStruct struct {
	Config  map[string]any
	Handler func()
	Name    string
	Tags    []string
	Age     int
	Active  bool
}

type complexStructTestCase struct {
	name     string
	value    complexStruct
	expected bool
}

func (tc complexStructTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, "IsZero() with complex struct")
}

func newComplexStructTestCase(name string, value complexStruct, expected bool) complexStructTestCase {
	return complexStructTestCase{
		name:     name,
		value:    value,
		expected: expected,
	}
}

func complexStructTestCases() []complexStructTestCase {
	return S(
		newComplexStructTestCase("zero complex struct", complexStruct{}, true),
		newComplexStructTestCase("non-zero name", complexStruct{Name: "test"}, false),
		newComplexStructTestCase("non-zero age", complexStruct{Age: 25}, false),
		newComplexStructTestCase("non-zero active", complexStruct{Active: true}, false),
		newComplexStructTestCase("non-zero tags", complexStruct{Tags: []string{"tag1"}}, false),
		newComplexStructTestCase("non-zero config", complexStruct{Config: map[string]any{"key": "value"}}, false),
		newComplexStructTestCase("non-zero handler", complexStruct{Handler: func() {}}, false),
	)
}

func TestIsZeroWithComplexTypes(t *testing.T) {
	for _, tc := range complexStructTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type pointerTestCase struct {
	value    any
	name     string
	expected bool
}

func (tc pointerTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, "IsZero() with pointers")
}

func newPointerTestCase(name string, value any) pointerTestCase {
	return pointerTestCase{
		name:     name,
		value:    value,
		expected: false, // All pointer tests expect false
	}
}

func pointerTestCases() []pointerTestCase {
	intVal := 42
	stringVal := "hello"
	boolVal := true

	return S(
		newPointerTestCase("pointer to int", &intVal),
		newPointerTestCase("pointer to string", &stringVal),
		newPointerTestCase("pointer to bool", &boolVal),
		newPointerTestCase("pointer to zero int", new(int)),
		newPointerTestCase("pointer to zero string", new(string)),
		newPointerTestCase("pointer to zero bool", new(bool)),
	)
}

func TestIsZeroWithPointers(t *testing.T) {
	for _, tc := range pointerTestCases() {
		t.Run(tc.name, tc.test)
	}
}

type inner struct {
	Value int
}

type outer struct {
	Name  string
	Inner inner
}

type nestedStructTestCase struct {
	name     string
	value    outer
	expected bool
}

func (tc nestedStructTestCase) test(t *testing.T) {
	result := IsZero(tc.value)
	AssertBool(t, result, tc.expected, "IsZero() with nested structs")
}

func newNestedStructTestCase(name string, value outer, expected bool) nestedStructTestCase {
	return nestedStructTestCase{
		name:     name,
		value:    value,
		expected: expected,
	}
}

func nestedStructTestCases() []nestedStructTestCase {
	return S(
		newNestedStructTestCase("zero nested struct", outer{}, true),
		newNestedStructTestCase("non-zero inner", outer{Inner: inner{Value: 1}}, false),
		newNestedStructTestCase("non-zero name", outer{Name: "test"}, false),
		newNestedStructTestCase("both non-zero", outer{Inner: inner{Value: 1}, Name: "test"}, false),
	)
}

func TestIsZeroWithNestedStructs(t *testing.T) {
	for _, tc := range nestedStructTestCases() {
		t.Run(tc.name, tc.test)
	}
}

func BenchmarkZero(b *testing.B) {
	var p *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Zero(p)
	}
}

type benchmarkTestCase struct {
	value any
	name  string
}

func newBenchmarkTestCase(name string, value any) benchmarkTestCase {
	return benchmarkTestCase{
		name:  name,
		value: value,
	}
}

func benchmarkTestCases() []benchmarkTestCase {
	return S(
		newBenchmarkTestCase("nil", nil),
		newBenchmarkTestCase("int", 0),
		newBenchmarkTestCase("string", ""),
		newBenchmarkTestCase("slice", []int{}),
		newBenchmarkTestCase("map", map[string]int{}),
		newBenchmarkTestCase("struct", struct{ Name string }{}),
		newBenchmarkTestCase("interface", zeroChecker{isZero: true}),
	)
}

func BenchmarkIsZero(b *testing.B) {
	for _, tc := range benchmarkTestCases() {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = IsZero(tc.value)
			}
		})
	}
}

func ExampleZero() {
	var p *int
	result := Zero(p)
	_, _ = fmt.Printf("Zero value of int: %d\n", result)

	var s *string
	strResult := Zero(s)
	_, _ = fmt.Printf("Zero value of string: %q\n", strResult)

	type Person struct {
		Name string
		Age  int
	}
	var person *Person
	personResult := Zero(person)
	_, _ = fmt.Printf("Zero value of Person: %+v\n", personResult)

	// Output:
	// Zero value of int: 0
	// Zero value of string: ""
	// Zero value of Person: {Name: Age:0}
}

func ExampleIsZero() {
	// Basic zero value detection
	_, _ = fmt.Printf("IsZero(nil): %t\n", IsZero(nil))
	_, _ = fmt.Printf("IsZero(0): %t\n", IsZero(0))
	_, _ = fmt.Printf("IsZero(42): %t\n", IsZero(42))
	_, _ = fmt.Printf("IsZero(\"\"): %t\n", IsZero(""))
	_, _ = fmt.Printf("IsZero(\"hello\"): %t\n", IsZero("hello"))

	// The key distinction: initialized vs uninitialized
	_, _ = fmt.Printf("IsZero([]int(nil)): %t\n", IsZero([]int(nil))) // nil slice - needs initialization
	_, _ = fmt.Printf("IsZero([]int{}): %t\n", IsZero([]int{}))       // empty slice - already initialized
	_, _ = fmt.Printf("IsZero([]int{1, 2, 3}): %t\n", IsZero([]int{1, 2, 3}))

	// With IsZero() interface
	checker := zeroChecker{isZero: true}
	_, _ = fmt.Printf("IsZero(zeroChecker{isZero: true}): %t\n", IsZero(checker))

	checker2 := zeroChecker{isZero: false}
	_, _ = fmt.Printf("IsZero(zeroChecker{isZero: false}): %t\n", IsZero(checker2))

	// Output:
	// IsZero(nil): true
	// IsZero(0): true
	// IsZero(42): false
	// IsZero(""): true
	// IsZero("hello"): false
	// IsZero([]int(nil)): true
	// IsZero([]int{}): false
	// IsZero([]int{1, 2, 3}): false
	// IsZero(zeroChecker{isZero: true}): true
	// IsZero(zeroChecker{isZero: false}): false
}

func ExampleIsZero_initialization() {
	// Practical initialization patterns

	// Lazy initialization of slice
	var items []string
	if IsZero(items) {
		items = []string{"default", "item"}
		_, _ = fmt.Printf("Initialized slice: %v\n", items)
	}

	// Conditional map initialization
	var cache map[string]int
	if IsZero(cache) {
		cache = make(map[string]int)
		_, _ = fmt.Printf("Initialized map: %v\n", cache)
	}

	// Pointer initialization
	var ptr *int
	if IsZero(ptr) {
		value := 42
		ptr = &value
		_, _ = fmt.Printf("Initialized pointer: %d\n", *ptr)
	}

	// Already initialized values won't be re-initialized
	items2 := []string{"existing"}
	if IsZero(items2) {
		_, _ = fmt.Println("This won't print - slice is already initialized")
	} else {
		_, _ = fmt.Printf("Slice already initialized: %v\n", items2)
	}

	// Output:
	// Initialized slice: [default item]
	// Initialized map: map[]
	// Initialized pointer: 42
	// Slice already initialized: [existing]
}
