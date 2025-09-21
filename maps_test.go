package core

import (
	"container/list"
	"fmt"
	"testing"
)

// TestCase interface validations
var _ TestCase = keysTestCase{}
var _ TestCase = sortedKeysTestCase{}
var _ TestCase = sortedValuesTestCase{}
var _ TestCase = sortedValuesCondTestCase{}
var _ TestCase = mapValueTestCase{}
var _ TestCase = mapContainsTestCase{}
var _ TestCase = mapListContainsTestCase{}
var _ TestCase = mapListForEachElementTestCase{}
var _ TestCase = mapAllListContainsTestCase{}

// keysTestCase tests Keys function
type keysTestCase struct {
	input    map[string]int
	name     string
	expected int
}

func makeKeysTestCases() []TestCase {
	return S(
		newKeysTestCase("empty map", map[string]int{}, 0),
		newKeysTestCase("single entry", map[string]int{"a": 1}, 1),
		newKeysTestCase("multiple entries", map[string]int{"a": 1, "b": 2, "c": 3}, 3),
		newKeysTestCase("nil map", nil, 0),
	)
}

func (tc keysTestCase) Name() string {
	return tc.name
}

func (tc keysTestCase) Test(t *testing.T) {
	t.Helper()

	got := Keys(tc.input)
	AssertEqual(t, tc.expected, len(got), "Keys", tc.input)

	tc.verifyAllKeysPresent(t, got)
}

func (tc keysTestCase) verifyAllKeysPresent(t *testing.T, got []string) {
	t.Helper()
	for k := range tc.input {
		AssertTrue(t, SliceContains(got, k), "contains %v", k)
	}
}

// Factory function for keysTestCase
func newKeysTestCase(name string, input map[string]int, expected int) TestCase {
	return keysTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestKeys(t *testing.T) {
	RunTestCases(t, makeKeysTestCases())
}

// sortedKeysTestCase tests SortedKeys function
type sortedKeysTestCase struct {
	name     string
	input    map[string]int
	expected []string
}

func makeSortedKeysTestCases() []TestCase {
	return S(
		newSortedKeysTestCase("empty map", map[string]int{}, S[string]()),
		newSortedKeysTestCase("single entry", map[string]int{"a": 1}, S("a")),
		newSortedKeysTestCase("multiple entries", map[string]int{"c": 3, "a": 1, "b": 2}, S("a", "b", "c")),
		newSortedKeysTestCase("numeric string keys",
			map[string]int{"10": 10, "2": 2, "1": 1, "20": 20},
			S("1", "10", "2", "20")),
		newSortedKeysTestCase("nil map", nil, S[string]()),
	)
}

func (tc sortedKeysTestCase) Name() string {
	return tc.name
}

func (tc sortedKeysTestCase) Test(t *testing.T) {
	t.Helper()

	got := SortedKeys(tc.input)
	AssertSliceEqual(t, tc.expected, got, "SortedKeys", tc.input)
}

// Factory function for sortedKeysTestCase
func newSortedKeysTestCase(name string, input map[string]int, expected []string) TestCase {
	return sortedKeysTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSortedKeys(t *testing.T) {
	t.Run("string keys", func(t *testing.T) {
		RunTestCases(t, makeSortedKeysTestCases())
	})
	t.Run("int keys", func(t *testing.T) {
		input := map[int]string{10: "ten", 2: "two", 1: "one", 20: "twenty"}
		expected := S(1, 2, 10, 20)

		got := SortedKeys(input)
		AssertSliceEqual(t, expected, got, "SortedKeys[int]")
	})
}

// sortedValuesTestCase tests SortedValues functions
type sortedValuesTestCase struct {
	name     string
	input    map[string]int
	expected []int
}

func makeSortedValuesTestCases() []TestCase {
	return S(
		newSortedValuesTestCase("empty map", map[string]int{}, nil),
		newSortedValuesTestCase("single entry", map[string]int{"a": 1}, S(1)),
		newSortedValuesTestCase("multiple entries", map[string]int{"c": 3, "a": 1, "b": 2}, S(1, 2, 3)),
		newSortedValuesTestCase("nil map", nil, nil),
	)
}

func (tc sortedValuesTestCase) Name() string {
	return tc.name
}

func (tc sortedValuesTestCase) Test(t *testing.T) {
	t.Helper()

	got := SortedValues(tc.input)
	AssertSliceEqual(t, tc.expected, got, "SortedValues", tc.input)
}

// Factory function for sortedValuesTestCase
func newSortedValuesTestCase(name string, input map[string]int, expected []int) TestCase {
	return sortedValuesTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSortedValues(t *testing.T) {
	RunTestCases(t, makeSortedValuesTestCases())
}

// sortedValuesCondTestCase tests SortedValuesCond function
type sortedValuesCondTestCase struct {
	name      string
	input     map[string]int
	predicate func(int) bool
	expected  []int
}

func makeSortedValuesCondTestCases() []TestCase {
	return S(
		newSortedValuesCondTestCase("empty map", map[string]int{}, func(v int) bool { return v > 0 }, nil),
		newSortedValuesCondTestCase("filter even values",
			map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
			func(v int) bool { return v%2 == 0 }, S(2, 4)),
		newSortedValuesCondTestCase("filter greater than 2",
			map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
			func(v int) bool { return v > 2 }, S(3, 4)),
		newSortedValuesCondTestCase("filter none",
			map[string]int{"a": 1, "b": 2, "c": 3},
			func(v int) bool { return v > 10 }, S[int]()),
		newSortedValuesCondTestCase("nil predicate", map[string]int{"a": 1, "b": 2, "c": 3}, nil, S(1, 2, 3)),
		newSortedValuesCondTestCase("nil map", nil, func(v int) bool { return v > 0 }, nil),
	)
}

func (tc sortedValuesCondTestCase) Name() string {
	return tc.name
}

func (tc sortedValuesCondTestCase) Test(t *testing.T) {
	t.Helper()

	got := SortedValuesCond(tc.input, tc.predicate)
	AssertSliceEqual(t, tc.expected, got, "SortedValuesCond", tc.input)
}

// Factory function for sortedValuesCondTestCase
func newSortedValuesCondTestCase(name string, input map[string]int,
	predicate func(int) bool, expected []int) TestCase {
	return sortedValuesCondTestCase{
		name:      name,
		input:     input,
		predicate: predicate,
		expected:  expected,
	}
}

func TestSortedValuesCond(t *testing.T) {
	RunTestCases(t, makeSortedValuesCondTestCases())
}

// Test SortedValuesUnlikelyCond
func TestSortedValuesUnlikelyCond(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	predicate := func(v int) bool { return v == 3 }
	expected := S(3)

	got := SortedValuesUnlikelyCond(input, predicate)
	AssertSliceEqual(t, expected, got, "SortedValuesUnlikelyCond")

	// Test empty result
	predicate2 := func(v int) bool { return v > 10 }
	got2 := SortedValuesUnlikelyCond(input, predicate2)
	AssertEqual(t, 0, len(got2), "SortedValuesUnlikelyCond empty")

	// Test nil map
	got3 := SortedValuesUnlikelyCond[string, int](nil, predicate)
	AssertNil(t, got3, "SortedValuesUnlikelyCond(nil)")

	// Test nil predicate
	got4 := SortedValuesUnlikelyCond(input, nil)
	expected4 := S(1, 2, 3, 4)
	AssertSliceEqual(t, expected4, got4, "SortedValuesUnlikelyCond nil predicate")
}

// mapValueTestCase tests MapValue function
type mapValueTestCase struct {
	name     string
	m        map[string]int
	key      string
	def      int
	expected int
	found    bool
}

func makeMapValueTestCases() []TestCase {
	return S(
		newMapValueTestCase("existing key", map[string]int{"a": 1, "b": 2}, "a", 99, 1),
		newMapValueTestCase("missing key", map[string]int{"a": 1, "b": 2}, "c", 88, 88),
		newMapValueTestCase("nil map", nil, "a", 77, 77),
		newMapValueTestCase("zero value exists", map[string]int{"a": 0}, "a", 66, 0),
	)
}

func (tc mapValueTestCase) Name() string {
	return tc.name
}

func (tc mapValueTestCase) Test(t *testing.T) {
	t.Helper()

	got, found := MapValue(tc.m, tc.key, tc.def)
	AssertEqual(t, tc.expected, got, "MapValue value", tc.m, tc.key, tc.def)
	AssertEqual(t, tc.found, found, "MapValue found", tc.m, tc.key, tc.def)
}

// Factory function for mapValueTestCase
func newMapValueTestCase(name string, m map[string]int, key string,
	def int, expected int) TestCase {
	return mapValueTestCase{
		name:     name,
		m:        m,
		key:      key,
		def:      def,
		expected: expected,
		found:    expected != def, // derive found from whether expected == def
	}
}

func TestMapValue(t *testing.T) {
	RunTestCases(t, makeMapValueTestCases())
}

// mapContainsTestCase tests MapContains function
type mapContainsTestCase struct {
	name     string
	m        map[string]any
	key      string
	expected bool
}

func makeMapContainsTestCases() []TestCase {
	return S(
		newMapContainsTestCase("existing key", map[string]any{"a": 1, "b": "two"}, "a", true),
		newMapContainsTestCase("missing key", map[string]any{"a": 1, "b": "two"}, "c", false),
		newMapContainsTestCase("nil map", nil, "a", false),
		newMapContainsTestCase("nil value exists", map[string]any{"a": nil}, "a", true),
	)
}

func newMapContainsTestCase(name string, m map[string]any, key string, expected bool) TestCase {
	return mapContainsTestCase{
		name:     name,
		m:        m,
		key:      key,
		expected: expected,
	}
}

func (tc mapContainsTestCase) Name() string {
	return tc.name
}

func (tc mapContainsTestCase) Test(t *testing.T) {
	t.Helper()

	got := MapContains(tc.m, tc.key)
	AssertEqual(t, tc.expected, got, "MapContains", tc.m, tc.key)
}

func TestMapContains(t *testing.T) {
	RunTestCases(t, makeMapContainsTestCases())
}

func TestMapListInsert(t *testing.T) {
	m := make(map[string]*list.List)

	// Insert into empty map
	MapListInsert(m, "key1", "value1")
	l, ok := m["key1"]
	AssertTrue(t, ok, "MapListInsert")
	AssertEqual(t, 1, l.Len(), "list length")

	// Insert another value
	MapListInsert(m, "key1", "value2")
	AssertEqual(t, 2, m["key1"].Len(), "list length after insert")

	// Verify order (insert at front)
	AssertEqual(t, "value2", m["key1"].Front().Value, "front value")
}

func TestMapListAppend(t *testing.T) {
	m := make(map[string]*list.List)

	// Append to empty map
	MapListAppend(m, "key1", "value1")
	l, ok := m["key1"]
	AssertTrue(t, ok, "MapListAppend")
	AssertEqual(t, 1, l.Len(), "list length")

	// Append another value
	MapListAppend(m, "key1", "value2")
	AssertEqual(t, 2, m["key1"].Len(), "list length after append")

	// Verify order (append at back)
	AssertEqual(t, "value2", m["key1"].Back().Value, "back value")
}

type mapListContainsTestCase struct {
	name     string
	m        map[string]*list.List
	key      string
	value    string
	expected bool
}

func newMapListContainsTestCase(name string, m map[string]*list.List, key, value string,
	expected bool) TestCase {
	return mapListContainsTestCase{
		name:     name,
		m:        m,
		key:      key,
		value:    value,
		expected: expected,
	}
}

func (tc mapListContainsTestCase) Name() string {
	return tc.name
}

func (tc mapListContainsTestCase) Test(t *testing.T) {
	t.Helper()
	got := MapListContains(tc.m, tc.key, tc.value)
	AssertEqual(t, tc.expected, got, "MapListContains", tc.key, tc.value)
}

func TestMapListContains(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "value1")
	MapListAppend(m, "key1", "value2")
	MapListAppend(m, "key2", "value3")

	tests := S(
		newMapListContainsTestCase("existing value", m, "key1", "value1", true),
		newMapListContainsTestCase("existing value 2", m, "key1", "value2", true),
		newMapListContainsTestCase("wrong key", m, "key2", "value1", false),
		newMapListContainsTestCase("missing key", m, "key3", "value1", false),
		newMapListContainsTestCase("missing value", m, "key1", "value3", false),
		newMapListContainsTestCase("nil map", nil, "key", "value", false),
	)

	RunTestCases(t, tests)
}

func TestMapListContainsFn(t *testing.T) {
	type customType struct {
		name string
		id   int
	}

	m := make(map[string]*list.List)
	MapListAppend(m, "key1", customType{id: 1, name: "one"})
	MapListAppend(m, "key1", customType{id: 2, name: "two"})

	eq := func(a, b customType) bool { return a.id == b.id }

	// Test existing value
	AssertTrue(t, MapListContainsFn(m, "key1", customType{id: 1, name: "different"}, eq), "MapListContainsFn")

	// Test missing value
	AssertFalse(t, MapListContainsFn(m, "key1", customType{id: 3, name: "three"}, eq), "MapListContainsFn missing")

	// Test nil eq function
	AssertFalse(t, MapListContainsFn(m, "key1", customType{id: 1, name: "one"}, nil), "MapListContainsFn nil eq")
}

func TestMapListInsertUnique(t *testing.T) {
	m := make(map[string]*list.List)

	// Insert first value
	MapListInsertUnique(m, "key1", "value1")
	AssertEqual(t, 1, m["key1"].Len(), "length after first insert")

	// Try to insert duplicate
	MapListInsertUnique(m, "key1", "value1")
	AssertEqual(t, 1, m["key1"].Len(), "length after duplicate")

	// Insert different value
	MapListInsertUnique(m, "key1", "value2")
	AssertEqual(t, 2, m["key1"].Len(), "length after unique")
}

func TestMapListAppendUnique(t *testing.T) {
	m := make(map[string]*list.List)

	// Append first value
	MapListAppendUnique(m, "key1", "value1")
	AssertEqual(t, 1, m["key1"].Len(), "length after first append")

	// Try to append duplicate
	MapListAppendUnique(m, "key1", "value1")
	AssertEqual(t, 1, m["key1"].Len(), "length after duplicate")

	// Append different value
	MapListAppendUnique(m, "key1", "value2")
	AssertEqual(t, 2, m["key1"].Len(), "length after unique")
}

func TestMapListForEach(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "a")
	MapListAppend(m, "key1", "b")
	MapListAppend(m, "key1", "c")

	var result []string
	MapListForEach(m, "key1", func(v string) bool {
		result = append(result, v)
		return false // continue
	})

	expected := S("a", "b", "c")
	AssertSliceEqual(t, expected, result, "MapListForEach")

	// Test early termination
	result = nil
	MapListForEach(m, "key1", func(v string) bool {
		result = append(result, v)
		return v == "b" // stop at "b"
	})

	AssertEqual(t, 2, len(result), "early termination length")
	AssertEqual(t, "b", result[1], "stopped value")

	// Test missing key
	result = nil
	MapListForEach(m, "missing", func(v string) bool {
		result = append(result, v)
		return false
	})
	AssertEqual(t, 0, len(result), "missing key")

	// Test nil map
	MapListForEach(nil, "key", func(_ string) bool { return false })
	// Should not panic

	// Test nil function
	MapListForEach[string, string](m, "key1", nil)
	// Should not panic
}

type mapListForEachElementTestCase struct {
	name     string
	m        map[string]*list.List
	key      string
	fn       func(*list.Element) bool
	expected []string
}

func makeMapListForEachElementTestCases() []TestCase {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "a")
	MapListAppend(m, "key1", "b")
	MapListAppend(m, "key2", "c")

	return S(
		newMapListForEachElementTestCase("iterate all elements", m, "key1", func(_ *list.Element) bool {
			return false // continue
		}, S("a", "b")),
		newMapListForEachElementTestCase("early termination", m, "key1", func(_ *list.Element) bool {
			return true // stop after first
		}, S("a")),
		newMapListForEachElementTestCase("missing key", m, "missing",
			func(_ *list.Element) bool { return false }, S[string]()),
		newMapListForEachElementTestCase("nil map", nil, "key",
			func(_ *list.Element) bool { return false }, S[string]()),
		newMapListForEachElementTestCase("nil function", m, "key1", nil, S[string]()),
	)
}

func (tc mapListForEachElementTestCase) Name() string {
	return tc.name
}

func (tc mapListForEachElementTestCase) Test(t *testing.T) {
	t.Helper()
	var values []string
	MapListForEachElement(tc.m, tc.key, func(el *list.Element) bool {
		if tc.fn != nil {
			values = append(values, el.Value.(string))
			return tc.fn(el)
		}
		return false
	})

	AssertEqual(t, len(tc.expected), len(values), "value count")
	for i, v := range values {
		if i < len(tc.expected) {
			AssertEqual(t, tc.expected[i], v, "values[%d] mismatch", i)
		}
	}
}

// Factory function for mapListForEachElementTestCase
func newMapListForEachElementTestCase(name string, m map[string]*list.List,
	key string, fn func(*list.Element) bool, expected []string) TestCase {
	return mapListForEachElementTestCase{
		name:     name,
		m:        m,
		key:      key,
		fn:       fn,
		expected: expected,
	}
}

func TestMapListForEachElement(t *testing.T) {
	testCases := makeMapListForEachElementTestCases()

	RunTestCases(t, testCases)
}

func TestMapListCopy(t *testing.T) {
	src := make(map[string]*list.List)
	MapListAppend(src, "key1", "a")
	MapListAppend(src, "key1", "b")
	MapListAppend(src, "key2", "c")

	dst := MapListCopy(src)

	// Verify structure
	AssertEqual(t, len(src), len(dst), "key count")

	// Verify contents
	for key, srcList := range src {
		dstList, ok := dst[key]
		if !ok {
			t.Errorf("MapListCopy missing key %q", key)
			continue
		}
		AssertEqual(t, srcList.Len(), dstList.Len(), "list length[%q]", key)
	}

	// Verify deep copy (modify dst shouldn't affect src)
	MapListAppend(dst, "key1", "new")
	AssertEqual(t, 2, src["key1"].Len(), "source length")
	AssertEqual(t, 3, dst["key1"].Len(), "destination length")
}

func TestMapListCopyFn(t *testing.T) {
	type data struct {
		value string
	}

	src := make(map[string]*list.List)
	MapListAppend(src, "key1", data{"a"})
	MapListAppend(src, "key1", data{"b"})

	// Copy with transformation
	dst := MapListCopyFn(src, func(v data) (data, bool) {
		return data{value: v.value + "-copy"}, true
	})

	// Verify transformation
	el := dst["key1"].Front()
	v, ok := el.Value.(data)
	AssertTrue(t, ok, "MapListCopyFn type check")
	AssertEqual(t, "a-copy", v.value, "transformed value")

	// Test filtering
	dst2 := MapListCopyFn(src, func(v data) (data, bool) {
		return v, v.value != "b" // exclude "b"
	})

	AssertEqual(t, 1, dst2["key1"].Len(), "filtered length")
}

type mapAllListContainsTestCase struct {
	name     string
	m        map[string]*list.List
	value    string
	expected bool
}

func (tc mapAllListContainsTestCase) Name() string {
	return tc.name
}

func (tc mapAllListContainsTestCase) Test(t *testing.T) {
	t.Helper()
	got := MapAllListContains(tc.m, tc.value)
	AssertEqual(t, tc.expected, got, "MapAllListContains", tc.value)
}

// Factory function for mapAllListContainsTestCase
func newMapAllListContainsTestCase(name string, m map[string]*list.List,
	value string, expected bool) TestCase {
	return mapAllListContainsTestCase{
		name:     name,
		m:        m,
		value:    value,
		expected: expected,
	}
}

func TestMapAllListContains(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "value1")
	MapListAppend(m, "key2", "value2")
	MapListAppend(m, "key3", "value1") // duplicate value in different key

	tests := S(
		newMapAllListContainsTestCase("existing value in key1", m, "value1", true),
		newMapAllListContainsTestCase("existing value in key2", m, "value2", true),
		newMapAllListContainsTestCase("missing value", m, "value3", false),
		newMapAllListContainsTestCase("nil map", nil, "value", false),
	)

	RunTestCases(t, tests)
}

func TestMapAllListContainsFn(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", 1)
	MapListAppend(m, "key1", 2)
	MapListAppend(m, "key2", 3)
	MapListAppend(m, "key2", 4)

	// Test finding even number
	found := MapAllListContainsFn(m, func(v int) bool {
		return v%2 == 0
	})
	AssertTrue(t, found, "MapAllListContainsFn even")

	// Test not finding large number
	found = MapAllListContainsFn(m, func(v int) bool {
		return v > 10
	})
	AssertFalse(t, found, "MapAllListContainsFn >10")

	// Test nil function
	found = MapAllListContainsFn[string, int](m, nil)
	AssertFalse(t, found, "MapAllListContainsFn(nil)")
}

func TestMapAllListForEach(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", 1)
	MapListAppend(m, "key1", 2)
	MapListAppend(m, "key2", 3)
	MapListAppend(m, "key2", 4)

	var sum int
	MapAllListForEach(m, func(v int) bool {
		sum += v
		return false // continue
	})

	AssertEqual(t, 10, sum, "sum")

	// Test early termination
	sum = 0
	MapAllListForEach(m, func(v int) bool {
		sum += v
		return v == 3 // stop at 3
	})

	// Sum should be less than 10 due to early termination
	AssertTrue(t, sum < 10, "MapAllListForEach early stop")
}

func TestMapAllListForEachElement(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "a")
	MapListAppend(m, "key1", "b")
	MapListAppend(m, "key2", "c")

	var count int
	MapAllListForEachElement(m, func(_ *list.Element) bool {
		count++
		return false // continue
	})

	AssertEqual(t, 3, count, "count")

	// Test early termination
	count = 0
	MapAllListForEachElement(m, func(_ *list.Element) bool {
		count++
		return count == 2 // stop after 2
	})

	AssertEqual(t, 2, count, "early stop count")
}

// Test edge cases
func TestMapListEdgeCases(t *testing.T) {
	// Test MapListInsertUniqueFn with nil map
	MapListInsertUniqueFn(nil, "key", "value", func(a, b string) bool { return a == b })
	// Should not panic

	// Test MapListInsertUniqueFn with nil eq
	m := make(map[string]*list.List)
	MapListInsertUniqueFn(m, "key", "value", nil)
	AssertEqual(t, 0, len(m), "nil eq")

	// Test MapListAppendUniqueFn with nil map
	MapListAppendUniqueFn(nil, "key", "value", func(a, b string) bool { return a == b })
	// Should not panic

	// Test MapListAppendUniqueFn with nil eq
	MapListAppendUniqueFn(m, "key", "value", nil)
	AssertEqual(t, 0, len(m), "nil eq")
}

// Benchmark tests
func BenchmarkKeys(b *testing.B) {
	m := make(map[string]int)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		m[key] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Keys(m)
	}
}

func BenchmarkSortedKeys(b *testing.B) {
	m := make(map[string]int)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		m[key] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SortedKeys(m)
	}
}

func BenchmarkMapListAppend(b *testing.B) {
	m := make(map[string]*list.List)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapListAppend(m, "key", i)
	}
}

func BenchmarkMapListContains(b *testing.B) {
	m := make(map[string]*list.List)
	for i := 0; i < 100; i++ {
		MapListAppend(m, "key", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MapListContains(m, "key", 50)
	}
}
