package core

import (
	"math"
	"testing"
)

func TestSliceReverse(t *testing.T) {
	for _, tc := range []struct{ a, b []int }{
		{S[int](), S[int]()},
		{S(1), S(1)},
		{S(1, 2), S(2, 1)},
		{S(1, 2, 3), S(3, 2, 1)},
		{S(1, 2, 3, 4), S(4, 3, 2, 1)},
		{S(1, 2, 3, 4, 5), S(5, 4, 3, 2, 1)},
		{S(1, 2, 3, 4, 5, 6), S(6, 5, 4, 3, 2, 1)},
	} {
		c := SliceCopy(tc.a)
		SliceReverse(c)
		AssertSliceEqual(t, tc.b, c, "SliceReverse(%q) failed", tc.a)
		if SliceEqual(c, tc.b) {
			t.Logf("%s(%q) → %q", "SliceReverse", tc.a, c)
		}
	}
}

// revive:disable
var (
	ints       = S(74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586)
	expectInts = S(74, 59, 238, -784, 9845, 959, 905, 0, 42, 7586, -5467984)

	float64s = S(
		74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, 7.8, 7.8, 74.3,
		59.0, math.Inf(1), 238.2, -784.0, 2.3,
	)
	expectFloat64s = S(74.3, 59.0, math.Inf(1), 238.2, -784, 2.3, 7.8)

	strs       = S("", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&")
	expectStrs = S("", "Hello", "foo", "bar", "f00", "%*&^*&^&")
)

func eq[T Ordered](a, b T) bool {
	return a == b
}

func cmp[T Ordered](a, b T) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}

func testSliceUnique[T Ordered](t *testing.T, before, after []T) {
	SliceSort(after, cmp[T])

	s := SliceUnique(before)
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUnique failed")

	s = SliceUniqueFn(before, eq[T])
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniqueFn failed")

	s = SliceCopyFn(before, nil)
	s2 := SliceUniquify(&s)
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniquify failed")
	AssertSliceEqual(t, s, s2, "SliceUniquify return value mismatch")

	s = SliceCopy(before)
	s2 = SliceUniquifyFn(&s, eq[T])
	SliceSort(s, cmp[T])
	AssertSliceEqual(t, after, s, "SliceUniquifyFn failed")
	AssertSliceEqual(t, s, s2, "SliceUniquifyFn return value mismatch")
}

func TestSliceUniqueInt(t *testing.T) {
	testSliceUnique(t, ints, expectInts)
}

func TestSliceUniqueFloat(t *testing.T) {
	testSliceUnique(t, float64s, expectFloat64s)
}

func TestSliceUniqueString(t *testing.T) {
	testSliceUnique(t, strs, expectStrs)
}

func TestSliceMinus(t *testing.T) {
	for _, tc := range []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{"empty slices", S[int](), S[int](), S[int]()},
		{"empty a", S[int](), S(1, 2), S[int]()},
		{"empty b", S(1, 2), S[int](), S(1, 2)},
		{"no overlap", S(1, 2), S(3, 4), S(1, 2)},
		{"partial overlap", S(1, 2, 3, 4, 5), S(3, 4, 6, 7), S(1, 2, 5)},
		{"complete overlap", S(1, 2, 3), S(1, 2, 3), S[int]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := SliceMinus(tc.a, tc.b)
			AssertSliceEqual(t, tc.expected, result, "SliceMinus(%v, %v) failed", tc.a, tc.b)
		})
	}
}

func TestSliceMinusFn(t *testing.T) {
	equal := func(va, vb string) bool {
		return va == vb
	}

	for _, tc := range []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{"empty slices", S[string](), S[string](), S[string]()},
		{"empty a", S[string](), S("x", "y"), S[string]()},
		{"empty b", S("x", "y"), S[string](), S("x", "y")},
		{"no overlap", S("apple", "cherry"), S("banana", "date"), S("apple", "cherry")},
		{"partial overlap", S("apple", "banana", "cherry"), S("banana", "date"), S("apple", "cherry")},
		{"complete overlap", S("apple", "banana"), S("apple", "banana"), S[string]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := SliceMinusFn(tc.a, tc.b, equal)
			AssertSliceEqual(t, tc.expected, result, "SliceMinusFn(%v, %v) failed", tc.a, tc.b)
		})
	}
}

func TestSliceRandom(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		input  []string
		wantok bool
	}{
		{name: "empty", input: S[string](), want: string(""), wantok: false},
		{name: "one", input: S("one"), want: string("one"), wantok: true},
		{name: "random", input: S("one", "two", "three", "four", "five", "six"), want: string(""), wantok: true},
	}
	for _, tc := range tests {
		if tc.name != "random" {
			got, ok := SliceRandom(tc.input)
			AssertBool(t, ok, tc.wantok, "SliceRandom ok status mismatch for %s", tc.name)
			AssertEqual(t, tc.want, got, "SliceRandom result mismatch for %s", tc.name)
		} else {
			u, ok := SliceRandom(tc.input)
			AssertBool(t, ok, tc.wantok, "SliceRandom ok status mismatch for %s", tc.name)
			t.Logf("random from one,two,three,four,five,six: %s", u)
		}
	}
}
