package core

import (
	"math"
	"testing"
)

var _ TestCase = deeperTestCase{}

// deeperTestCase pins the skip normalisation every panic constructor
// applies. A negative skip is clamped to 1 rather than passed down to
// getCallers, which rejects it and yields an empty stack.
type deeperTestCase struct {
	name string
	skip int
	want int
}

func newDeeperTestCase(name string, skip, want int) deeperTestCase {
	return deeperTestCase{
		name: name,
		skip: skip,
		want: want,
	}
}

func (tc deeperTestCase) Name() string {
	return tc.name
}

func (tc deeperTestCase) Test(t *testing.T) {
	AssertEqual(t, tc.want, deeper(tc.skip), "deeper(%d)", tc.skip)
}

// The clamped rows all state 1, the same answer the zero row states, so
// a negative skip is indistinguishable from no skip at all. The far
// positive row declares that the positive side counts on, and the max
// skip row that it stops at math.MaxInt instead of wrapping.
func deeperTestCases() []deeperTestCase {
	return []deeperTestCase{
		newDeeperTestCase("no skip", 0, 1),
		newDeeperTestCase("one frame", 1, 2),
		newDeeperTestCase("far positive", 9999, 10000),
		newDeeperTestCase("max skip", math.MaxInt, math.MaxInt),
		newDeeperTestCase("minus one", -1, 1),
		newDeeperTestCase("far negative", -9999, 1),
	}
}

func TestDeeper(t *testing.T) {
	RunTestCases(t, deeperTestCases())
}
