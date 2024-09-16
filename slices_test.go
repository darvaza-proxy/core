package core

import (
	"math"
	"testing"
)

func S[T comparable](v ...T) []T { return v }

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
	ints       = []int{74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586}
	expectInts = []int{74, 59, 238, -784, 9845, 959, 905, 0, 42, 7586, -5467984}

	float64s = []float64{
		74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, 7.8, 7.8, 74.3,
		59.0, math.Inf(1), 238.2, -784.0, 2.3,
	}
	expectFloat64s = []float64{74.3, 59.0, math.Inf(1), 238.2, -784, 2.3, 7.8}

	strs       = []string{"", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&"}
	expectStrs = []string{"", "Hello", "foo", "bar", "f00", "%*&^*&^&"}
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

func TestSliceRandom(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		want   string
		wantok bool
	}{
		{name: "empty", input: []string{}, want: string(""), wantok: false},
		{name: "one", input: []string{"one"}, want: string("one"), wantok: true},
		{name: "random", input: []string{"one", "two", "three", "four", "five", "six"}, want: string(""), wantok: true},
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
