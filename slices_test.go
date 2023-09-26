package core

import (
	"math"
	"testing"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

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

func testSliceUnique[T constraints.Ordered](t *testing.T, before, after []T) {
	slices.Sort(after)

	eq := func(a, b T) bool {
		return a == b
	}

	s := SliceUnique(before)
	slices.Sort(s)
	if !slices.Equal(s, after) {
		t.Errorf("%v != %v", s, after)
	}

	s = SliceUniqueFn(before, eq)
	slices.Sort(s)
	if !slices.Equal(s, after) {
		t.Errorf("%v != %v", s, after)
	}

	s = slices.Clone(before)
	s2 := SliceUniquify(&s)
	slices.Sort(s)
	if !slices.Equal(s, after) {
		t.Errorf("%v != %v", s, after)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}

	s = slices.Clone(before)
	s2 = SliceUniquifyFn(&s, eq)
	slices.Sort(s)
	if !slices.Equal(s, after) {
		t.Errorf("%v != %v", s, after)
	}
	if !slices.Equal(s2, s) {
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
