package core

import (
	"math"
	"testing"
)

// TestCase validations
var _ TestCase = comparableEqualTestCase{}

// comparableEqualTestCase reads its results with plain checks:
// comparableEqual sits under the assertions that would otherwise read
// them.
type comparableEqualTestCase struct {
	a, b      any
	name      string
	wantEqual bool
	wantKnown bool
}

func (tc comparableEqualTestCase) Name() string { return tc.name }

func (tc comparableEqualTestCase) Test(t *testing.T) {
	equal, known := comparableEqual(tc.a, tc.b)
	plainEqual(t, tc.wantEqual, equal, "equal")
	plainEqual(t, tc.wantKnown, known, "known")
}

func newComparableEqualTestCase(name string, a, b any,
	wantEqual, wantKnown bool) comparableEqualTestCase {
	return comparableEqualTestCase{
		name:      name,
		a:         a,
		b:         b,
		wantEqual: wantEqual,
		wantKnown: wantKnown,
	}
}

func comparableEqualTestCases() []comparableEqualTestCase {
	p := new(int)

	return []comparableEqualTestCase{
		newComparableEqualTestCase("equal ints", 42, 42, true, true),
		newComparableEqualTestCase("unequal ints", 42, 43, false, true),
		newComparableEqualTestCase("different types", 42, "42", false, true),
		newComparableEqualTestCase("both nil", nil, nil, true, true),
		newComparableEqualTestCase("nil vs typed nil",
			nil, (*int)(nil), false, true),
		newComparableEqualTestCase("NaN never equals",
			math.NaN(), math.NaN(), false, true),
		newComparableEqualTestCase("same pointer", p, p, true, true),
		newComparableEqualTestCase("distinct pointers",
			p, new(int), false, true),
		newComparableEqualTestCase("slice vs int", S(1), 1, false, true),
		newComparableEqualTestCase("slices panic", S(1), S(1), false, false),
		newComparableEqualTestCase("held slices reached",
			HeldValuesForTest{1, S(1)}, HeldValuesForTest{1, S(1)}, false, false),
		newComparableEqualTestCase("held slices behind a difference",
			HeldValuesForTest{1, S(1)}, HeldValuesForTest{2, S(1)}, false, true),
	}
}

func TestComparableEqual(t *testing.T) {
	RunTestCases(t, comparableEqualTestCases())
}

// TestComparableEqualTyped instantiates comparableEqual with the concrete
// types the assertions use, where == cannot panic.
func TestComparableEqualTyped(t *testing.T) {
	equal, known := comparableEqual(true, true)
	plainEqual(t, true, equal, "true, true equal")
	plainEqual(t, true, known, "true, true known")

	equal, known = comparableEqual(true, false)
	plainEqual(t, false, equal, "true, false equal")
	plainEqual(t, true, known, "true, false known")

	equal, known = comparableEqual("a", "b")
	plainEqual(t, false, equal, "a, b equal")
	plainEqual(t, true, known, "a, b known")
}
