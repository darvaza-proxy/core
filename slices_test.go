package core

import (
	"math"
	"testing"

	"golang.org/x/exp/slices"
)

var (
	ints       = []int{74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586}
	expectInts = []int{74, 59, 238, -784, 9845, 959, 905, 0, 42, 7586, -5467984}

	float64s = []float64{74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, 7.8, 7.8, 74.3,
		59.0, math.Inf(1), 238.2, -784.0, 2.3}
	expectFloat64s = []float64{74.3, 59.0, math.Inf(1), 238.2, -784, 2.3, 7.8}

	strs       = []string{"", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&"}
	expectStrs = []string{"", "Hello", "foo", "bar", "f00", "%*&^*&^&"}
)

func TestSliceUniqueInt(t *testing.T) {
	slices.Sort(expectInts)

	s := SliceUnique(ints)
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
}

func TestSliceUniqueFloat(t *testing.T) {
	slices.Sort(expectFloat64s)

	s := SliceUnique(float64s)
	slices.Sort(s)
	if !slices.Equal(s, expectFloat64s) {
		t.Errorf("%v != %v", s, expectFloat64s)
	}

	s = slices.Clone(float64s)
	s2 := SliceUniquify(&s)
	slices.Sort(s)
	if !slices.Equal(s2, s) {
		t.Errorf("%v != %v", s2, s)
	}
}

func TestSliceUniqueString(t *testing.T) {
	slices.Sort(expectStrs)

	s := SliceUnique(strs)
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
}
