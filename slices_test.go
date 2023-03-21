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
	uniqInt := SliceUnique(ints)
	slices.Sort(uniqInt)
	slices.Sort(expectInts)
	if !slices.Equal(uniqInt, expectInts) {
		t.Fail()
	}
}

func TestSliceUniqueFloat(t *testing.T) {
	uniqFloat := SliceUnique(float64s)
	slices.Sort(uniqFloat)
	slices.Sort(expectFloat64s)
	if !slices.Equal(uniqFloat, expectFloat64s) {
		t.Fail()
	}
}

func TestSliceUniqueString(t *testing.T) {
	uniqStr := SliceUnique(strs)
	slices.Sort(uniqStr)
	slices.Sort(expectStrs)
	if !slices.Equal(uniqStr, expectStrs) {
		t.Fail()
	}
}
