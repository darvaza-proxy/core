package core_test

import (
	"container/list"
	"fmt"
	"testing"

	"darvaza.org/core"
)

// TestCase interface validations
var (
	_ core.TestCase = keysTestCase{}
	_ core.TestCase = sortedKeysTestCase{}
	_ core.TestCase = sortedValuesTestCase{}
	_ core.TestCase = sortedValuesCondTestCase{}
	_ core.TestCase = mapValueTestCase{}
	_ core.TestCase = mapContainsTestCase{}
	_ core.TestCase = mapListContainsTestCase{}
	_ core.TestCase = mapListForEachElementTestCase{}
	_ core.TestCase = mapAllListContainsTestCase{}
)

// keysTestCase tests Keys function
type keysTestCase struct {
	input    map[string]int
	name     string
	expected int
}

var keysTestCases = []keysTestCase{
	newKeysTestCase("empty map", map[string]int{}, 0),
	newKeysTestCase("single entry", map[string]int{"a": 1}, 1),
	newKeysTestCase("multiple entries", map[string]int{"a": 1, "b": 2, "c": 3}, 3),
	newKeysTestCase("nil map", nil, 0),
}

func (tc keysTestCase) Name() string {
	return tc.name
}

func (tc keysTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.Keys(tc.input)
	core.AssertEqual(t, tc.expected, len(got), "Keys(%v)", tc.input)

	for k := range tc.input {
		core.AssertTrue(t, core.SliceContains(got, k), "contains %v", k)
	}
}

// Factory function for keysTestCase
func newKeysTestCase(name string, input map[string]int, expected int) keysTestCase {
	return keysTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestKeys(t *testing.T) {
	core.RunTestCases(t, keysTestCases)
}

// sortedKeysTestCase tests SortedKeys function
type sortedKeysTestCase struct {
	name     string
	input    map[string]int
	expected []string
}

var sortedKeysTestCases = []sortedKeysTestCase{
	newSortedKeysTestCase("empty map", map[string]int{}, core.S[string]()),
	newSortedKeysTestCase("single entry", map[string]int{"a": 1}, core.S("a")),
	newSortedKeysTestCase("multiple entries", map[string]int{"c": 3, "a": 1, "b": 2}, core.S("a", "b", "c")),
	newSortedKeysTestCase("numeric string keys",
		map[string]int{"10": 10, "2": 2, "1": 1, "20": 20},
		core.S("1", "10", "2", "20")),
	newSortedKeysTestCase("nil map", nil, core.S[string]()),
}

func (tc sortedKeysTestCase) Name() string {
	return tc.name
}

func (tc sortedKeysTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.SortedKeys(tc.input)
	core.AssertSliceEqual(t, tc.expected, got, "SortedKeys(%v)", tc.input)
}

// Factory function for sortedKeysTestCase
func newSortedKeysTestCase(name string, input map[string]int, expected []string) sortedKeysTestCase {
	return sortedKeysTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSortedKeys(t *testing.T) {
	core.RunTestCases(t, sortedKeysTestCases)
}

// Test SortedKeys with int keys
func TestSortedKeysInt(t *testing.T) {
	input := map[int]string{10: "ten", 2: "two", 1: "one", 20: "twenty"}
	expected := core.S(1, 2, 10, 20)

	got := core.SortedKeys(input)
	core.AssertSliceEqual(t, expected, got, "SortedKeys[int]")
}

// sortedValuesTestCase tests SortedValues functions
type sortedValuesTestCase struct {
	name     string
	input    map[string]int
	expected []int
}

var sortedValuesTestCases = []sortedValuesTestCase{
	newSortedValuesTestCase("empty map", map[string]int{}, nil),
	newSortedValuesTestCase("single entry", map[string]int{"a": 1}, core.S(1)),
	newSortedValuesTestCase("multiple entries", map[string]int{"c": 3, "a": 1, "b": 2}, core.S(1, 2, 3)),
	newSortedValuesTestCase("nil map", nil, nil),
}

func (tc sortedValuesTestCase) Name() string {
	return tc.name
}

func (tc sortedValuesTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.SortedValues(tc.input)
	core.AssertSliceEqual(t, tc.expected, got, "SortedValues(%v)", tc.input)
}

// Factory function for sortedValuesTestCase
func newSortedValuesTestCase(name string, input map[string]int, expected []int) sortedValuesTestCase {
	return sortedValuesTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func TestSortedValues(t *testing.T) {
	core.RunTestCases(t, sortedValuesTestCases)
}

// sortedValuesCondTestCase tests SortedValuesCond function
type sortedValuesCondTestCase struct {
	name      string
	input     map[string]int
	predicate func(int) bool
	expected  []int
}

var sortedValuesCondTestCases = []sortedValuesCondTestCase{
	newSortedValuesCondTestCase("empty map", map[string]int{}, func(v int) bool { return v > 0 }, nil),
	newSortedValuesCondTestCase("filter even values",
		map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		func(v int) bool { return v%2 == 0 }, core.S(2, 4)),
	newSortedValuesCondTestCase("filter greater than 2",
		map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		func(v int) bool { return v > 2 }, core.S(3, 4)),
	newSortedValuesCondTestCase("filter none",
		map[string]int{"a": 1, "b": 2, "c": 3},
		func(v int) bool { return v > 10 }, core.S[int]()),
	newSortedValuesCondTestCase("nil predicate", map[string]int{"a": 1, "b": 2, "c": 3}, nil, core.S(1, 2, 3)),
	newSortedValuesCondTestCase("nil map", nil, func(v int) bool { return v > 0 }, nil),
}

func (tc sortedValuesCondTestCase) Name() string {
	return tc.name
}

func (tc sortedValuesCondTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.SortedValuesCond(tc.input, tc.predicate)
	core.AssertSliceEqual(t, tc.expected, got, "SortedValuesCond(%v, _)", tc.input)
}

// Factory function for sortedValuesCondTestCase
func newSortedValuesCondTestCase(name string, input map[string]int,
	predicate func(int) bool, expected []int) sortedValuesCondTestCase {
	return sortedValuesCondTestCase{
		name:      name,
		input:     input,
		predicate: predicate,
		expected:  expected,
	}
}

func TestSortedValuesCond(t *testing.T) {
	core.RunTestCases(t, sortedValuesCondTestCases)
}

// Test SortedValuesUnlikelyCond
func TestSortedValuesUnlikelyCond(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	predicate := func(v int) bool { return v == 3 }
	expected := core.S(3)

	got := core.SortedValuesUnlikelyCond(input, predicate)
	core.AssertSliceEqual(t, expected, got, "SortedValuesUnlikelyCond")

	// Test empty result
	predicate2 := func(v int) bool { return v > 10 }
	got2 := core.SortedValuesUnlikelyCond(input, predicate2)
	core.AssertEqual(t, 0, len(got2), "SortedValuesUnlikelyCond empty")

	// Test nil map
	got3 := core.SortedValuesUnlikelyCond[string](nil, predicate)
	core.AssertNil(t, got3, "SortedValuesUnlikelyCond(nil)")

	// Test nil predicate
	got4 := core.SortedValuesUnlikelyCond(input, nil)
	expected4 := core.S(1, 2, 3, 4)
	core.AssertSliceEqual(t, expected4, got4, "SortedValuesUnlikelyCond nil predicate")
}

// mapValueTestCase tests MapValue by the value it returns and whether it
// found the key.
type mapValueTestCase struct {
	name     string
	m        map[string]int
	key      string
	def      int
	expected int
	found    bool
}

func (tc mapValueTestCase) Name() string {
	return tc.name
}

func (tc mapValueTestCase) Test(t *testing.T) {
	t.Helper()

	got, found := core.MapValue(tc.m, tc.key, tc.def)
	core.AssertEqual(t, tc.expected, got, "value")
	core.AssertEqual(t, tc.found, found, "found")
}

func newMapValueTestCase(name string, m map[string]int, key string,
	def, expected int) mapValueTestCase {
	return mapValueTestCase{
		name:     name,
		m:        m,
		key:      key,
		def:      def,
		expected: expected,
		found:    true,
	}
}

// newMapValueTestCaseMissing declares a key the map does not hold, where
// the default is the value MapValue must return.
func newMapValueTestCaseMissing(name string, m map[string]int, key string,
	def int) mapValueTestCase {
	return mapValueTestCase{
		name:     name,
		m:        m,
		key:      key,
		def:      def,
		expected: def,
		found:    false,
	}
}

func mapValueTestCases() []mapValueTestCase {
	return core.S(
		newMapValueTestCase("existing key", map[string]int{"a": 1, "b": 2}, "a", 99, 1),
		newMapValueTestCase("zero value exists", map[string]int{"a": 0}, "a", 66, 0),
		newMapValueTestCase("value equals default", map[string]int{"a": 66}, "a", 66, 66),
		newMapValueTestCaseMissing("missing key", map[string]int{"a": 1, "b": 2}, "c", 88),
		newMapValueTestCaseMissing("nil map", nil, "a", 77),
	)
}

func TestMapValue(t *testing.T) {
	core.RunTestCases(t, mapValueTestCases())
}

// mapContainsTestCase tests MapContains function
type mapContainsTestCase struct {
	name     string
	m        map[string]any
	key      string
	expected bool
}

func mapContainsTestCases() []mapContainsTestCase {
	m := map[string]any{"a": 1, "b": "two"}

	return []mapContainsTestCase{
		newMapContainsTestCase("existing key", m, "a", true),
		newMapContainsTestCase("missing key", m, "c", false),
		newMapContainsTestCase("nil map", nil, "a", false),
		newMapContainsTestCase("nil value exists", map[string]any{"a": nil}, "a", true),
	}
}

func newMapContainsTestCase(name string, m map[string]any, key string, expected bool) mapContainsTestCase {
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

	got := core.MapContains(tc.m, tc.key)
	core.AssertEqual(t, tc.expected, got, "MapContains(%v, %q)", tc.m, tc.key)
}

func TestMapContains(t *testing.T) {
	core.RunTestCases(t, mapContainsTestCases())
}

func TestMapListInsert(t *testing.T) {
	m := make(map[string]*list.List)

	// Insert into empty map
	core.MapListInsert(m, "key1", "value1")
	l, ok := m["key1"]
	core.AssertTrue(t, ok, "MapListInsert")
	core.AssertEqual(t, 1, l.Len(), "list length")

	// Insert another value
	core.MapListInsert(m, "key1", "value2")
	core.AssertEqual(t, 2, m["key1"].Len(), "list length after insert")

	// Verify order (insert at front)
	core.AssertEqual(t, "value2", m["key1"].Front().Value, "front value")
}

func TestMapListAppend(t *testing.T) {
	m := make(map[string]*list.List)

	// Append to empty map
	core.MapListAppend(m, "key1", "value1")
	l, ok := m["key1"]
	core.AssertTrue(t, ok, "MapListAppend")
	core.AssertEqual(t, 1, l.Len(), "list length")

	// Append another value
	core.MapListAppend(m, "key1", "value2")
	core.AssertEqual(t, 2, m["key1"].Len(), "list length after append")

	// Verify order (append at back)
	core.AssertEqual(t, "value2", m["key1"].Back().Value, "back value")
}

type mapListContainsTestCase struct {
	name     string
	m        map[string]*list.List
	key      string
	value    string
	expected bool
}

func newMapListContainsTestCase(name string, m map[string]*list.List, key, value string,
	expected bool) mapListContainsTestCase {
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
	got := core.MapListContains(tc.m, tc.key, tc.value)
	core.AssertEqual(t, tc.expected, got, "MapListContains(_, %q, %q)", tc.key, tc.value)
}

func TestMapListContains(t *testing.T) {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", "value1")
	core.MapListAppend(m, "key1", "value2")
	core.MapListAppend(m, "key2", "value3")

	tests := []mapListContainsTestCase{
		newMapListContainsTestCase("existing value", m, "key1", "value1", true),
		newMapListContainsTestCase("existing value 2", m, "key1", "value2", true),
		newMapListContainsTestCase("wrong key", m, "key2", "value1", false),
		newMapListContainsTestCase("missing key", m, "key3", "value1", false),
		newMapListContainsTestCase("missing value", m, "key1", "value3", false),
		newMapListContainsTestCase("nil map", nil, "key", "value", false),
	}

	core.RunTestCases(t, tests)
}

func TestMapListContainsFn(t *testing.T) {
	type customType struct {
		name string
		id   int
	}

	newCustomType := func(id int, name string) customType {
		return customType{id: id, name: name}
	}

	first := newCustomType(1, "one")

	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", first)
	core.MapListAppend(m, "key1", newCustomType(2, "two"))

	eq := func(a, b customType) bool { return a.id == b.id }

	// Test existing value
	core.AssertTrue(t, core.MapListContainsFn(m, "key1", newCustomType(1, "different"), eq), "MapListContainsFn")

	// Test missing value
	core.AssertFalse(t, core.MapListContainsFn(m, "key1", newCustomType(3, "three"), eq), "MapListContainsFn missing")

	// Test nil eq function
	core.AssertFalse(t, core.MapListContainsFn(m, "key1", first, nil), "MapListContainsFn nil eq")
}

func TestMapListInsertUnique(t *testing.T) {
	m := make(map[string]*list.List)

	// Insert first value
	core.MapListInsertUnique(m, "key1", "value1")
	core.AssertEqual(t, 1, m["key1"].Len(), "length after first insert")

	// Try to insert duplicate
	core.MapListInsertUnique(m, "key1", "value1")
	core.AssertEqual(t, 1, m["key1"].Len(), "length after duplicate")

	// Insert different value
	core.MapListInsertUnique(m, "key1", "value2")
	core.AssertEqual(t, 2, m["key1"].Len(), "length after unique")
}

func TestMapListAppendUnique(t *testing.T) {
	m := make(map[string]*list.List)

	// Append first value
	core.MapListAppendUnique(m, "key1", "value1")
	core.AssertEqual(t, 1, m["key1"].Len(), "length after first append")

	// Try to append duplicate
	core.MapListAppendUnique(m, "key1", "value1")
	core.AssertEqual(t, 1, m["key1"].Len(), "length after duplicate")

	// Append different value
	core.MapListAppendUnique(m, "key1", "value2")
	core.AssertEqual(t, 2, m["key1"].Len(), "length after unique")
}

func TestMapListForEach(t *testing.T) {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", "a")
	core.MapListAppend(m, "key1", "b")
	core.MapListAppend(m, "key1", "c")

	var result []string
	core.MapListForEach(m, "key1", func(v string) bool {
		result = append(result, v)
		return false // continue
	})

	expected := core.S("a", "b", "c")
	core.AssertSliceEqual(t, expected, result, "MapListForEach")

	// Test early termination
	result = nil
	core.MapListForEach(m, "key1", func(v string) bool {
		result = append(result, v)
		return v == "b" // stop at "b"
	})

	core.AssertEqual(t, 2, len(result), "early termination length")
	core.AssertEqual(t, "b", result[1], "stopped value")

	// Test missing key
	result = nil
	core.MapListForEach(m, "missing", func(v string) bool {
		result = append(result, v)
		return false
	})
	core.AssertEqual(t, 0, len(result), "missing key")

	// Test nil map
	core.MapListForEach(nil, "key", func(_ string) bool { return false })
	// Should not panic

	// Test nil function
	core.MapListForEach[string, string](m, "key1", nil)
	// Should not panic
}

type mapListForEachElementTestCase struct {
	name     string
	m        map[string]*list.List
	key      string
	fn       func(*list.Element) bool
	expected []string
}

func createMapListForEachElementTestCases() []mapListForEachElementTestCase {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", "a")
	core.MapListAppend(m, "key1", "b")
	core.MapListAppend(m, "key2", "c")

	return []mapListForEachElementTestCase{
		newMapListForEachElementTestCase("iterate all elements", m, "key1", func(_ *list.Element) bool {
			return false // continue
		}, core.S("a", "b")),
		newMapListForEachElementTestCase("early termination", m, "key1", func(_ *list.Element) bool {
			return true // stop after first
		}, core.S("a")),
		newMapListForEachElementTestCase("missing key", m, "missing",
			func(_ *list.Element) bool { return false }, core.S[string]()),
		newMapListForEachElementTestCase("nil map", nil, "key",
			func(_ *list.Element) bool { return false }, core.S[string]()),
		newMapListForEachElementTestCase("nil function", m, "key1", nil, core.S[string]()),
	}
}

func (tc mapListForEachElementTestCase) Name() string {
	return tc.name
}

func (tc mapListForEachElementTestCase) Test(t *testing.T) {
	t.Helper()
	var values []string
	core.MapListForEachElement(tc.m, tc.key, func(el *list.Element) bool {
		if tc.fn != nil {
			values = append(values, el.Value.(string))
			return tc.fn(el)
		}
		return false
	})

	core.AssertEqual(t, len(tc.expected), len(values), "value count")
	for i, v := range values {
		if i < len(tc.expected) {
			core.AssertEqual(t, tc.expected[i], v, "values[%d] mismatch", i)
		}
	}
}

// Factory function for mapListForEachElementTestCase
func newMapListForEachElementTestCase(name string, m map[string]*list.List,
	key string, fn func(*list.Element) bool, expected []string) mapListForEachElementTestCase {
	return mapListForEachElementTestCase{
		name:     name,
		m:        m,
		key:      key,
		fn:       fn,
		expected: expected,
	}
}

func TestMapListForEachElement(t *testing.T) {
	testCases := createMapListForEachElementTestCases()

	core.RunTestCases(t, testCases)
}

func TestMapListCopy(t *testing.T) {
	src := make(map[string]*list.List)
	core.MapListAppend(src, "key1", "a")
	core.MapListAppend(src, "key1", "b")
	core.MapListAppend(src, "key2", "c")

	dst := core.MapListCopy(src)

	// Verify structure
	core.AssertEqual(t, len(src), len(dst), "key count")

	// Verify contents
	for key, srcList := range src {
		dstList, ok := dst[key]
		core.AssertMustTrue(t, ok, "key %q copied", key)
		core.AssertMustNotNil(t, dstList, "key %q list", key)
		core.AssertEqual(t, srcList.Len(), dstList.Len(), "list length[%q]", key)
	}

	// Verify deep copy (modify dst shouldn't affect src)
	core.MapListAppend(dst, "key1", "new")
	core.AssertEqual(t, 2, src["key1"].Len(), "source length")
	core.AssertEqual(t, 3, dst["key1"].Len(), "destination length")
}

func TestMapListCopyFn(t *testing.T) {
	type data struct {
		value string
	}

	src := make(map[string]*list.List)
	core.MapListAppend(src, "key1", data{"a"})
	core.MapListAppend(src, "key1", data{"b"})

	// Copy with transformation
	dst := core.MapListCopyFn(src, func(v data) (data, bool) {
		return data{value: v.value + "-copy"}, true
	})

	// Verify transformation
	copied := dst["key1"]
	core.AssertMustNotNil(t, copied, "key1 copied")
	el := copied.Front()
	core.AssertMustNotNil(t, el, "first element")
	v := core.AssertMustTypeIs[data](t, el.Value, "first element value")
	core.AssertEqual(t, "a-copy", v.value, "transformed value")

	// Test filtering
	dst2 := core.MapListCopyFn(src, func(v data) (data, bool) {
		return v, v.value != "b" // exclude "b"
	})

	filtered := dst2["key1"]
	core.AssertMustNotNil(t, filtered, "key1 filtered")
	core.AssertEqual(t, 1, filtered.Len(), "filtered length")
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
	got := core.MapAllListContains(tc.m, tc.value)
	core.AssertEqual(t, tc.expected, got, "MapAllListContains(_, %q)", tc.value)
}

// Factory function for mapAllListContainsTestCase
func newMapAllListContainsTestCase(name string, m map[string]*list.List,
	value string, expected bool) mapAllListContainsTestCase {
	return mapAllListContainsTestCase{
		name:     name,
		m:        m,
		value:    value,
		expected: expected,
	}
}

func TestMapAllListContains(t *testing.T) {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", "value1")
	core.MapListAppend(m, "key2", "value2")
	core.MapListAppend(m, "key3", "value1") // duplicate value in different key

	tests := []mapAllListContainsTestCase{
		newMapAllListContainsTestCase("existing value in key1", m, "value1", true),
		newMapAllListContainsTestCase("existing value in key2", m, "value2", true),
		newMapAllListContainsTestCase("missing value", m, "value3", false),
		newMapAllListContainsTestCase("nil map", nil, "value", false),
	}

	core.RunTestCases(t, tests)
}

func TestMapAllListContainsFn(t *testing.T) {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", 1)
	core.MapListAppend(m, "key1", 2)
	core.MapListAppend(m, "key2", 3)
	core.MapListAppend(m, "key2", 4)

	// Test finding even number
	found := core.MapAllListContainsFn(m, func(v int) bool {
		return v%2 == 0
	})
	core.AssertTrue(t, found, "MapAllListContainsFn even")

	// Test not finding large number
	found = core.MapAllListContainsFn(m, func(v int) bool {
		return v > 10
	})
	core.AssertFalse(t, found, "MapAllListContainsFn >10")

	// Test nil function
	found = core.MapAllListContainsFn[string, int](m, nil)
	core.AssertFalse(t, found, "MapAllListContainsFn(nil)")
}

func TestMapAllListForEach(t *testing.T) {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", 1)
	core.MapListAppend(m, "key1", 2)
	core.MapListAppend(m, "key2", 3)
	core.MapListAppend(m, "key2", 4)

	var sum int
	core.MapAllListForEach(m, func(v int) bool {
		sum += v
		return false // continue
	})

	core.AssertEqual(t, 10, sum, "sum")

	// Test early termination
	var count int
	core.MapAllListForEach(m, func(_ int) bool {
		count++
		return count == 2 // stop after 2
	})

	core.AssertEqual(t, 2, count, "early stop count")
}

func TestMapAllListForEachElement(t *testing.T) {
	m := make(map[string]*list.List)
	core.MapListAppend(m, "key1", "a")
	core.MapListAppend(m, "key1", "b")
	core.MapListAppend(m, "key2", "c")

	var count int
	core.MapAllListForEachElement(m, func(_ *list.Element) bool {
		count++
		return false // continue
	})

	core.AssertEqual(t, 3, count, "count")

	// Test early termination
	count = 0
	core.MapAllListForEachElement(m, func(_ *list.Element) bool {
		count++
		return count == 2 // stop after 2
	})

	core.AssertEqual(t, 2, count, "early stop count")
}

// Test edge cases
func TestMapListEdgeCases(t *testing.T) {
	// Test MapListInsertUniqueFn with nil map
	core.MapListInsertUniqueFn(nil, "key", "value", func(a, b string) bool { return a == b })
	// Should not panic

	// Test MapListInsertUniqueFn with nil eq
	m := make(map[string]*list.List)
	core.MapListInsertUniqueFn(m, "key", "value", nil)
	core.AssertEqual(t, 0, len(m), "nil eq")

	// Test MapListAppendUniqueFn with nil map
	core.MapListAppendUniqueFn(nil, "key", "value", func(a, b string) bool { return a == b })
	// Should not panic

	// Test MapListAppendUniqueFn with nil eq
	core.MapListAppendUniqueFn(m, "key", "value", nil)
	core.AssertEqual(t, 0, len(m), "nil eq")
}

// Benchmark tests
func BenchmarkKeys(b *testing.B) {
	m := make(map[string]int)
	for i := range 100 {
		key := fmt.Sprintf("key%d", i)
		m[key] = i
	}

	for b.Loop() {
		_ = core.Keys(m)
	}
}

func BenchmarkSortedKeys(b *testing.B) {
	m := make(map[string]int)
	for i := range 100 {
		key := fmt.Sprintf("key%d", i)
		m[key] = i
	}

	for b.Loop() {
		_ = core.SortedKeys(m)
	}
}

func BenchmarkMapListAppend(b *testing.B) {
	m := make(map[string]*list.List)

	i := 0
	for b.Loop() {
		core.MapListAppend(m, "key", i)
		i++
	}
}

func BenchmarkMapListContains(b *testing.B) {
	m := make(map[string]*list.List)
	for i := range 100 {
		core.MapListAppend(m, "key", i)
	}

	for b.Loop() {
		_ = core.MapListContains(m, "key", 50)
	}
}
