package core

import (
	"container/list"
	"fmt"
	"reflect"
	"testing"
)

// keysTestCase tests Keys function
type keysTestCase struct {
	name     string
	input    map[string]int
	expected int // expected length
}

var keysTestCases = []keysTestCase{
	{
		name:     "empty map",
		input:    map[string]int{},
		expected: 0,
	},
	{
		name:     "single entry",
		input:    map[string]int{"a": 1},
		expected: 1,
	},
	{
		name:     "multiple entries",
		input:    map[string]int{"a": 1, "b": 2, "c": 3},
		expected: 3,
	},
	{
		name:     "nil map",
		input:    nil,
		expected: 0,
	},
}

// revive:disable-next-line:cognitive-complexity
func (tc keysTestCase) test(t *testing.T) {
	t.Helper()

	got := Keys(tc.input)
	if len(got) != tc.expected {
		t.Errorf("Keys(%v) returned %d keys, want %d", tc.input, len(got), tc.expected)
	}

	// Verify all keys are present
	for k := range tc.input {
		found := false
		for _, gotK := range got {
			if gotK == k {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Key %q not found in result", k)
		}
	}
}

func TestKeys(t *testing.T) {
	for _, tc := range keysTestCases {
		t.Run(tc.name, tc.test)
	}
}

// sortedKeysTestCase tests SortedKeys function
type sortedKeysTestCase struct {
	name     string
	input    map[string]int
	expected []string
}

var sortedKeysTestCases = []sortedKeysTestCase{
	{
		name:     "empty map",
		input:    map[string]int{},
		expected: []string{},
	},
	{
		name:     "single entry",
		input:    map[string]int{"a": 1},
		expected: []string{"a"},
	},
	{
		name:     "multiple entries",
		input:    map[string]int{"c": 3, "a": 1, "b": 2},
		expected: []string{"a", "b", "c"},
	},
	{
		name:     "numeric string keys",
		input:    map[string]int{"10": 10, "2": 2, "1": 1, "20": 20},
		expected: []string{"1", "10", "2", "20"}, // lexicographic order
	},
	{
		name:     "nil map",
		input:    nil,
		expected: []string{},
	},
}

func (tc sortedKeysTestCase) test(t *testing.T) {
	t.Helper()

	got := SortedKeys(tc.input)
	if !reflect.DeepEqual(got, tc.expected) {
		t.Errorf("SortedKeys(%v) = %v, want %v", tc.input, got, tc.expected)
	}
}

func TestSortedKeys(t *testing.T) {
	for _, tc := range sortedKeysTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test SortedKeys with int keys
func TestSortedKeysInt(t *testing.T) {
	input := map[int]string{10: "ten", 2: "two", 1: "one", 20: "twenty"}
	expected := []int{1, 2, 10, 20}

	got := SortedKeys(input)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("SortedKeys(%v) = %v, want %v", input, got, expected)
	}
}

// sortedValuesTestCase tests SortedValues functions
type sortedValuesTestCase struct {
	name     string
	input    map[string]int
	expected []int
}

var sortedValuesTestCases = []sortedValuesTestCase{
	{
		name:     "empty map",
		input:    map[string]int{},
		expected: nil,
	},
	{
		name:     "single entry",
		input:    map[string]int{"a": 1},
		expected: []int{1},
	},
	{
		name:     "multiple entries",
		input:    map[string]int{"c": 3, "a": 1, "b": 2},
		expected: []int{1, 2, 3}, // sorted by keys: a=1, b=2, c=3
	},
	{
		name:     "nil map",
		input:    nil,
		expected: nil,
	},
}

func (tc sortedValuesTestCase) test(t *testing.T) {
	t.Helper()

	got := SortedValues(tc.input)
	if !reflect.DeepEqual(got, tc.expected) {
		t.Errorf("SortedValues(%v) = %v, want %v", tc.input, got, tc.expected)
	}
}

func TestSortedValues(t *testing.T) {
	for _, tc := range sortedValuesTestCases {
		t.Run(tc.name, tc.test)
	}
}

// sortedValuesCondTestCase tests SortedValuesCond function
type sortedValuesCondTestCase struct {
	name      string
	input     map[string]int
	predicate func(int) bool
	expected  []int
}

var sortedValuesCondTestCases = []sortedValuesCondTestCase{
	{
		name:      "empty map",
		input:     map[string]int{},
		predicate: func(v int) bool { return v > 0 },
		expected:  nil,
	},
	{
		name:      "filter even values",
		input:     map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		predicate: func(v int) bool { return v%2 == 0 },
		expected:  []int{2, 4}, // b=2, d=4
	},
	{
		name:      "filter greater than 2",
		input:     map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		predicate: func(v int) bool { return v > 2 },
		expected:  []int{3, 4}, // c=3, d=4
	},
	{
		name:      "filter none",
		input:     map[string]int{"a": 1, "b": 2, "c": 3},
		predicate: func(v int) bool { return v > 10 },
		expected:  []int{},
	},
	{
		name:      "nil predicate",
		input:     map[string]int{"a": 1, "b": 2, "c": 3},
		predicate: nil,
		expected:  []int{1, 2, 3},
	},
	{
		name:      "nil map",
		input:     nil,
		predicate: func(v int) bool { return v > 0 },
		expected:  nil,
	},
}

func (tc sortedValuesCondTestCase) test(t *testing.T) {
	t.Helper()

	got := SortedValuesCond(tc.input, tc.predicate)
	if !reflect.DeepEqual(got, tc.expected) {
		t.Errorf("SortedValuesCond(%v, predicate) = %v, want %v", tc.input, got, tc.expected)
	}
}

func TestSortedValuesCond(t *testing.T) {
	for _, tc := range sortedValuesCondTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test SortedValuesUnlikelyCond
func TestSortedValuesUnlikelyCond(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	predicate := func(v int) bool { return v == 3 }
	expected := []int{3}

	got := SortedValuesUnlikelyCond(input, predicate)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("SortedValuesUnlikelyCond(%v, predicate) = %v, want %v", input, got, expected)
	}

	// Test empty result
	predicate2 := func(v int) bool { return v > 10 }
	got2 := SortedValuesUnlikelyCond(input, predicate2)
	if len(got2) != 0 {
		t.Errorf("SortedValuesUnlikelyCond with no matches returned %v, want empty", got2)
	}

	// Test nil map
	got3 := SortedValuesUnlikelyCond[string, int](nil, predicate)
	if got3 != nil {
		t.Errorf("SortedValuesUnlikelyCond with nil map returned %v, want nil", got3)
	}

	// Test nil predicate
	got4 := SortedValuesUnlikelyCond(input, nil)
	expected4 := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got4, expected4) {
		t.Errorf("SortedValuesUnlikelyCond with nil predicate returned %v, want %v", got4, expected4)
	}
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

var mapValueTestCases = []mapValueTestCase{
	{
		name:     "existing key",
		m:        map[string]int{"a": 1, "b": 2},
		key:      "a",
		def:      99,
		expected: 1,
		found:    true,
	},
	{
		name:     "missing key",
		m:        map[string]int{"a": 1, "b": 2},
		key:      "c",
		def:      99,
		expected: 99,
		found:    false,
	},
	{
		name:     "nil map",
		m:        nil,
		key:      "a",
		def:      99,
		expected: 99,
		found:    false,
	},
	{
		name:     "zero value exists",
		m:        map[string]int{"a": 0},
		key:      "a",
		def:      99,
		expected: 0,
		found:    true,
	},
}

func (tc mapValueTestCase) test(t *testing.T) {
	t.Helper()

	got, found := MapValue(tc.m, tc.key, tc.def)
	if got != tc.expected || found != tc.found {
		t.Errorf("MapValue(%v, %q, %d) = (%d, %v), want (%d, %v)",
			tc.m, tc.key, tc.def, got, found, tc.expected, tc.found)
	}
}

func TestMapValue(t *testing.T) {
	for _, tc := range mapValueTestCases {
		t.Run(tc.name, tc.test)
	}
}

// mapContainsTestCase tests MapContains function
type mapContainsTestCase struct {
	name     string
	m        map[string]any
	key      string
	expected bool
}

var mapContainsTestCases = []mapContainsTestCase{
	{
		name:     "existing key",
		m:        map[string]any{"a": 1, "b": "two"},
		key:      "a",
		expected: true,
	},
	{
		name:     "missing key",
		m:        map[string]any{"a": 1, "b": "two"},
		key:      "c",
		expected: false,
	},
	{
		name:     "nil map",
		m:        nil,
		key:      "a",
		expected: false,
	},
	{
		name:     "nil value exists",
		m:        map[string]any{"a": nil},
		key:      "a",
		expected: true,
	},
}

func (tc mapContainsTestCase) test(t *testing.T) {
	t.Helper()

	got := MapContains(tc.m, tc.key)
	if got != tc.expected {
		t.Errorf("MapContains(%v, %q) = %v, want %v", tc.m, tc.key, got, tc.expected)
	}
}

func TestMapContains(t *testing.T) {
	for _, tc := range mapContainsTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestMapListInsert(t *testing.T) {
	m := make(map[string]*list.List)

	// Insert into empty map
	MapListInsert(m, "key1", "value1")
	if l, ok := m["key1"]; !ok || l.Len() != 1 {
		t.Errorf("MapListInsert failed to create list")
	}

	// Insert another value
	MapListInsert(m, "key1", "value2")
	if m["key1"].Len() != 2 {
		t.Errorf("MapListInsert failed to add second value")
	}

	// Verify order (insert at front)
	if m["key1"].Front().Value != "value2" {
		t.Errorf("MapListInsert should insert at front")
	}
}

func TestMapListAppend(t *testing.T) {
	m := make(map[string]*list.List)

	// Append to empty map
	MapListAppend(m, "key1", "value1")
	if l, ok := m["key1"]; !ok || l.Len() != 1 {
		t.Errorf("MapListAppend failed to create list")
	}

	// Append another value
	MapListAppend(m, "key1", "value2")
	if m["key1"].Len() != 2 {
		t.Errorf("MapListAppend failed to add second value")
	}

	// Verify order (append at back)
	if m["key1"].Back().Value != "value2" {
		t.Errorf("MapListAppend should append at back")
	}
}

type mapListContainsTestCase struct {
	name     string
	m        map[string]*list.List
	key      string
	value    string
	expected bool
}

func (tc mapListContainsTestCase) test(t *testing.T) {
	t.Helper()
	got := MapListContains(tc.m, tc.key, tc.value)
	if got != tc.expected {
		t.Errorf("MapListContains(m, %q, %q) = %v, want %v",
			tc.key, tc.value, got, tc.expected)
	}
}

func TestMapListContains(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "value1")
	MapListAppend(m, "key1", "value2")
	MapListAppend(m, "key2", "value3")

	tests := []mapListContainsTestCase{
		{"existing value", m, "key1", "value1", true},
		{"existing value 2", m, "key1", "value2", true},
		{"wrong key", m, "key2", "value1", false},
		{"missing key", m, "key3", "value1", false},
		{"missing value", m, "key1", "value3", false},
		{"nil map", nil, "key", "value", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func TestMapListContainsFn(t *testing.T) {
	type customType struct {
		id   int
		name string
	}

	m := make(map[string]*list.List)
	MapListAppend(m, "key1", customType{1, "one"})
	MapListAppend(m, "key1", customType{2, "two"})

	eq := func(a, b customType) bool { return a.id == b.id }

	// Test existing value
	if !MapListContainsFn(m, "key1", customType{1, "different"}, eq) {
		t.Errorf("MapListContainsFn should find value by id")
	}

	// Test missing value
	if MapListContainsFn(m, "key1", customType{3, "three"}, eq) {
		t.Errorf("MapListContainsFn should not find missing value")
	}

	// Test nil eq function
	if MapListContainsFn(m, "key1", customType{1, "one"}, nil) {
		t.Errorf("MapListContainsFn with nil eq should return false")
	}
}

func TestMapListInsertUnique(t *testing.T) {
	m := make(map[string]*list.List)

	// Insert first value
	MapListInsertUnique(m, "key1", "value1")
	if m["key1"].Len() != 1 {
		t.Errorf("MapListInsertUnique failed to insert first value")
	}

	// Try to insert duplicate
	MapListInsertUnique(m, "key1", "value1")
	if m["key1"].Len() != 1 {
		t.Errorf("MapListInsertUnique should not insert duplicate")
	}

	// Insert different value
	MapListInsertUnique(m, "key1", "value2")
	if m["key1"].Len() != 2 {
		t.Errorf("MapListInsertUnique should insert different value")
	}
}

func TestMapListAppendUnique(t *testing.T) {
	m := make(map[string]*list.List)

	// Append first value
	MapListAppendUnique(m, "key1", "value1")
	if m["key1"].Len() != 1 {
		t.Errorf("MapListAppendUnique failed to append first value")
	}

	// Try to append duplicate
	MapListAppendUnique(m, "key1", "value1")
	if m["key1"].Len() != 1 {
		t.Errorf("MapListAppendUnique should not append duplicate")
	}

	// Append different value
	MapListAppendUnique(m, "key1", "value2")
	if m["key1"].Len() != 2 {
		t.Errorf("MapListAppendUnique should append different value")
	}
}

// revive:disable:cognitive-complexity
func TestMapListForEach(t *testing.T) {
	// revive:enable:cognitive-complexity
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "a")
	MapListAppend(m, "key1", "b")
	MapListAppend(m, "key1", "c")

	var result []string
	MapListForEach(m, "key1", func(v string) bool {
		result = append(result, v)
		return false // continue
	})

	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("MapListForEach collected %v, want %v", result, expected)
	}

	// Test early termination
	result = nil
	MapListForEach(m, "key1", func(v string) bool {
		result = append(result, v)
		return v == "b" // stop at "b"
	})

	if len(result) != 2 || result[1] != "b" {
		t.Errorf("MapListForEach early termination failed: %v", result)
	}

	// Test missing key
	result = nil
	MapListForEach(m, "missing", func(v string) bool {
		result = append(result, v)
		return false
	})
	if len(result) != 0 {
		t.Errorf("MapListForEach with missing key should not call function")
	}

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

func createMapListForEachElementTestCases() []mapListForEachElementTestCase {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "a")
	MapListAppend(m, "key1", "b")
	MapListAppend(m, "key2", "c")

	return []mapListForEachElementTestCase{
		{
			name: "iterate all elements",
			m:    m,
			key:  "key1",
			fn: func(_ *list.Element) bool {
				return false // continue
			},
			expected: []string{"a", "b"},
		},
		{
			name: "early termination",
			m:    m,
			key:  "key1",
			fn: func(_ *list.Element) bool {
				return true // stop after first
			},
			expected: []string{"a"},
		},
		{
			name:     "missing key",
			m:        m,
			key:      "missing",
			fn:       func(_ *list.Element) bool { return false },
			expected: []string{},
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "key",
			fn:       func(_ *list.Element) bool { return false },
			expected: []string{},
		},
		{
			name:     "nil function",
			m:        m,
			key:      "key1",
			fn:       nil,
			expected: []string{},
		},
	}
}

func (tc mapListForEachElementTestCase) test(t *testing.T) {
	t.Helper()
	var values []string
	MapListForEachElement(tc.m, tc.key, func(el *list.Element) bool {
		if tc.fn != nil {
			values = append(values, el.Value.(string))
			return tc.fn(el)
		}
		return false
	})

	if len(values) != len(tc.expected) {
		t.Errorf("got %d values, want %d", len(values), len(tc.expected))
	}
	for i, v := range values {
		if i < len(tc.expected) && v != tc.expected[i] {
			t.Errorf("values[%d] = %s, want %s", i, v, tc.expected[i])
		}
	}
}

func TestMapListForEachElement(t *testing.T) {
	testCases := createMapListForEachElementTestCases()

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// revive:disable:cognitive-complexity
func TestMapListCopy(t *testing.T) {
	// revive:enable:cognitive-complexity
	src := make(map[string]*list.List)
	MapListAppend(src, "key1", "a")
	MapListAppend(src, "key1", "b")
	MapListAppend(src, "key2", "c")

	dst := MapListCopy(src)

	// Verify structure
	if len(dst) != len(src) {
		t.Errorf("MapListCopy created map with %d keys, want %d", len(dst), len(src))
	}

	// Verify contents
	for key, srcList := range src {
		dstList, ok := dst[key]
		if !ok {
			t.Errorf("MapListCopy missing key %q", key)
			continue
		}
		if srcList.Len() != dstList.Len() {
			t.Errorf("MapListCopy list length mismatch for key %q", key)
		}
	}

	// Verify deep copy (modify dst shouldn't affect src)
	MapListAppend(dst, "key1", "new")
	if src["key1"].Len() != 2 || dst["key1"].Len() != 3 {
		t.Errorf("MapListCopy should create independent lists")
	}
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
	if v, ok := el.Value.(data); !ok || v.value != "a-copy" {
		t.Errorf("MapListCopyFn transformation failed: got %v", v)
	}

	// Test filtering
	dst2 := MapListCopyFn(src, func(v data) (data, bool) {
		return v, v.value != "b" // exclude "b"
	})

	if dst2["key1"].Len() != 1 {
		t.Errorf("MapListCopyFn filtering failed: got %d elements", dst2["key1"].Len())
	}
}

type mapAllListContainsTestCase struct {
	name     string
	m        map[string]*list.List
	value    string
	expected bool
}

func (tc mapAllListContainsTestCase) test(t *testing.T) {
	t.Helper()
	got := MapAllListContains(tc.m, tc.value)
	if got != tc.expected {
		t.Errorf("MapAllListContains(m, %q) = %v, want %v",
			tc.value, got, tc.expected)
	}
}

func TestMapAllListContains(t *testing.T) {
	m := make(map[string]*list.List)
	MapListAppend(m, "key1", "value1")
	MapListAppend(m, "key2", "value2")
	MapListAppend(m, "key3", "value1") // duplicate value in different key

	tests := []mapAllListContainsTestCase{
		{"existing value in key1", m, "value1", true},
		{"existing value in key2", m, "value2", true},
		{"missing value", m, "value3", false},
		{"nil map", nil, "value", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
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
	if !found {
		t.Errorf("MapAllListContainsFn should find even number")
	}

	// Test not finding large number
	found = MapAllListContainsFn(m, func(v int) bool {
		return v > 10
	})
	if found {
		t.Errorf("MapAllListContainsFn should not find number > 10")
	}

	// Test nil function
	found = MapAllListContainsFn[string, int](m, nil)
	if found {
		t.Errorf("MapAllListContainsFn with nil function should return false")
	}
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

	if sum != 10 {
		t.Errorf("MapAllListForEach sum = %d, want 10", sum)
	}

	// Test early termination
	sum = 0
	MapAllListForEach(m, func(v int) bool {
		sum += v
		return v == 3 // stop at 3
	})

	// Sum should be less than 10 due to early termination
	if sum >= 10 {
		t.Errorf("MapAllListForEach should stop early")
	}
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

	if count != 3 {
		t.Errorf("MapAllListForEachElement count = %d, want 3", count)
	}

	// Test early termination
	count = 0
	MapAllListForEachElement(m, func(_ *list.Element) bool {
		count++
		return count == 2 // stop after 2
	})

	if count != 2 {
		t.Errorf("MapAllListForEachElement should stop at 2")
	}
}

// Test edge cases
func TestMapListEdgeCases(t *testing.T) {
	// Test MapListInsertUniqueFn with nil map
	MapListInsertUniqueFn(nil, "key", "value", func(a, b string) bool { return a == b })
	// Should not panic

	// Test MapListInsertUniqueFn with nil eq
	m := make(map[string]*list.List)
	MapListInsertUniqueFn(m, "key", "value", nil)
	if len(m) != 0 {
		t.Errorf("MapListInsertUniqueFn with nil eq should not insert")
	}

	// Test MapListAppendUniqueFn with nil map
	MapListAppendUniqueFn(nil, "key", "value", func(a, b string) bool { return a == b })
	// Should not panic

	// Test MapListAppendUniqueFn with nil eq
	MapListAppendUniqueFn(m, "key", "value", nil)
	if len(m) != 0 {
		t.Errorf("MapListAppendUniqueFn with nil eq should not insert")
	}
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
