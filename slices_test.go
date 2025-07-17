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
		if SliceEqual(c, tc.b) {
			t.Logf("%s(%q) → %q", "SliceReverse", tc.a, c)
		} else {
			t.Fatalf("ERROR: %s(%q) → %q (expected %q)", "SliceReverse", tc.a, c, tc.b)
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
	if !SliceEqual(s, after) {
		t.Errorf("%v != %v", s, after)
	}

	s = SliceUniqueFn(before, eq[T])
	SliceSort(s, cmp[T])
	if !SliceEqual(s, after) {
		t.Errorf("%v != %v", s, after)
	}

	s = SliceCopyFn(before, nil)
	s2 := SliceUniquify(&s)
	SliceSort(s, cmp[T])
	if !SliceEqual(s, after) {
		t.Errorf("%v != %v", s, after)
	}
	if !SliceEqual(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}

	s = SliceCopy(before)
	s2 = SliceUniquifyFn(&s, eq[T])
	SliceSort(s, cmp[T])
	if !SliceEqual(s, after) {
		t.Errorf("%v != %v", s, after)
	}
	if !SliceEqual(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}
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
			if !SliceEqual(result, tc.expected) {
				t.Errorf("SliceMinus(%v, %v) = %v, want %v", tc.a, tc.b, result, tc.expected)
			}
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
			if !SliceEqual(result, tc.expected) {
				t.Errorf("SliceMinusFn(%v, %v) = %v, want %v", tc.a, tc.b, result, tc.expected)
			}
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
			if ok != tc.wantok {
				t.Fatalf("an error occurred in %s", tc.name)
			}
			if tc.want != got {
				t.Fatalf("%s: expected: %v, got: %v", tc.name, tc.want, got)
			}
		} else {
			u, ok := SliceRandom(tc.input)
			if ok != tc.wantok {
				t.Fatalf("error occurred in %s", tc.name)
			}
			t.Logf("random from one,two,three,four,five,six: %s", u)
		}
	}
}
