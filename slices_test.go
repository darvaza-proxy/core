package core

import (
	"math"
	"testing"

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

func TestSliceUniqueInt(t *testing.T) {
	slices.Sort(expectInts)

	eq := func(a, b int) bool {
		return a == b
	}

	s := SliceUnique(ints)
	slices.Sort(s)
	if !slices.Equal(s, expectInts) {
		t.Errorf("%v != %v", s, expectInts)
	}

	s = SliceUniqueFn(ints, eq)
	slices.Sort(s)
	if !slices.Equal(s, expectInts) {
		t.Errorf("%v != %v", s, expectInts)
	}

	s = slices.Clone(ints)
	s2 := SliceUniquify(&s)
	slices.Sort(s)
	if !slices.Equal(s, expectInts) {
		t.Errorf("%v != %v", s, expectInts)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}

	s = slices.Clone(ints)
	s2 = SliceUniquifyFn(&s, eq)
	slices.Sort(s)
	if !slices.Equal(s, expectInts) {
		t.Errorf("%v != %v", s, expectInts)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}
}

func TestSliceUniqueFloat(t *testing.T) {
	slices.Sort(expectFloat64s)

	eq := func(a, b float64) bool {
		return a == b
	}

	s := SliceUnique(float64s)
	slices.Sort(s)
	if !slices.Equal(s, expectFloat64s) {
		t.Errorf("%v != %v", s, expectFloat64s)
	}

	s = SliceUniqueFn(float64s, eq)
	slices.Sort(s)
	if !slices.Equal(s, expectFloat64s) {
		t.Errorf("%v != %v", s, expectFloat64s)
	}

	s = slices.Clone(float64s)
	s2 := SliceUniquify(&s)
	slices.Sort(s)
	if !slices.Equal(s, expectFloat64s) {
		t.Errorf("%v != %v", s, expectFloat64s)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}

	s = slices.Clone(float64s)
	s2 = SliceUniquifyFn(&s, eq)
	slices.Sort(s)
	if !slices.Equal(s, expectFloat64s) {
		t.Errorf("%v != %v", s, expectFloat64s)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}
}

func TestSliceUniqueString(t *testing.T) {
	slices.Sort(expectStrs)

	eq := func(a, b string) bool {
		return a == b
	}

	s := SliceUnique(strs)
	slices.Sort(s)
	if !slices.Equal(s, expectStrs) {
		t.Errorf("%v != %v", s, expectStrs)
	}

	s = SliceUniqueFn(strs, eq)
	slices.Sort(s)
	if !slices.Equal(s, expectStrs) {
		t.Errorf("%v != %v", s, expectStrs)
	}

	s = slices.Clone(strs)
	s2 := SliceUniquify(&s)
	slices.Sort(s)
	if !slices.Equal(s, expectStrs) {
		t.Errorf("%v != %v", s, expectStrs)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}

	s = slices.Clone(strs)
	s2 = SliceUniquifyFn(&s, eq)
	slices.Sort(s)
	if !slices.Equal(s, expectStrs) {
		t.Errorf("%v != %v", s, expectStrs)
	}
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}
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
