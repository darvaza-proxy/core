package core

import (
	"math"
	"reflect"
	"sync/atomic"
	"testing"
)

// TestCase validations
var (
	_ TestCase = comparableEqualTestCase{}
	_ TestCase = areEqualTestCase{}
)

// panicEqSlice is non-comparable with an Equal method that panics.
type panicEqSlice []int

func (panicEqSlice) Equal(panicEqSlice) bool { panic("Equal called") }

// oddEqSlice is non-comparable with an Equal method that doesn't
// follow the Equal(T) bool convention.
type oddEqSlice []int

func (oddEqSlice) Equal() bool { return true }

// eqNeverSlice is non-comparable with an Equal method that never
// matches, proving Equal wins over element-wise comparison.
type eqNeverSlice []int

func (eqNeverSlice) Equal(eqNeverSlice) bool { return false }

// panicEqMap is non-comparable and not a slice; after its Equal
// panics there is no further rung to consult.
type panicEqMap map[string]int

func (panicEqMap) Equal(panicEqMap) bool { panic("Equal called") }

// eqAlways is comparable with an Equal method that always matches,
// proving a false == defers to Equal for comparable types.
type eqAlways int

func (eqAlways) Equal(eqAlways) bool { return true }

// eqNeverComparable is comparable with an Equal method that never
// matches. The call counter lets a test prove a true == skips Equal
// outright while a false == still invokes it.
type eqNeverComparable int

var eqNeverComparableCalls atomic.Int64

func (eqNeverComparable) Equal(eqNeverComparable) bool {
	eqNeverComparableCalls.Add(1)
	return false
}

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

type areEqualTestCase struct {
	name      string
	vvi       []any
	wantIs    bool
	wantKnown bool
}

func (tc areEqualTestCase) Name() string { return tc.name }

func (tc areEqualTestCase) Test(t *testing.T) {
	is, known := AreEqual(tc.vvi...)
	AssertComparableForTest(t, tc.wantIs, is, "is")
	AssertComparableForTest(t, tc.wantKnown, known, "known")
}

func newAreEqualTestCase(name string, vvi []any,
	wantIs, wantKnown bool) areEqualTestCase {
	return areEqualTestCase{
		name:      name,
		vvi:       vvi,
		wantIs:    wantIs,
		wantKnown: wantKnown,
	}
}

func areEqualComparableTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualTestCase("no values", nil, false, true),
		newAreEqualTestCase("single value", S[any](42), true, true),
		newAreEqualTestCase("equal ints", S[any](42, 42), true, true),
		newAreEqualTestCase("unequal ints", S[any](42, 43), false, true),
		newAreEqualTestCase("different types", S[any](42, "42"), false, true),
		newAreEqualTestCase("int vs int64", S[any](42, int64(42)), false, true),
		newAreEqualTestCase("equal chain", S[any]("a", "a", "a"), true, true),
		newAreEqualTestCase("chain broken at the end",
			S[any]("a", "a", "b"), false, true),
		newAreEqualTestCase("NaN never equals",
			S[any](math.NaN(), math.NaN()), false, true),
		newAreEqualTestCase("Equal rescues a false ==",
			S[any](eqAlways(1), eqAlways(2)), true, true),
		newAreEqualTestCase("true == skips Equal",
			S[any](eqNeverComparable(1), eqNeverComparable(1)), true, true),
		newAreEqualTestCase("false == with false Equal stays unequal",
			S[any](eqNeverComparable(1), eqNeverComparable(2)), false, true),
		newAreEqualTestCase("held slices reached stay unknown",
			S[any](HeldValuesForTest{1, S(1)}, HeldValuesForTest{1, S(1)}), false, false),
		newAreEqualTestCase("held slices behind a difference",
			S[any](HeldValuesForTest{1, S(1)}, HeldValuesForTest{2, S(1)}), false, true),
	}
}

func areEqualNilTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualTestCase("both untyped nil", S[any](nil, nil), true, true),
		newAreEqualTestCase("untyped nil vs zero", S[any](nil, 0), false, true),
		newAreEqualTestCase("untyped nil vs typed nil",
			S[any](nil, (*int)(nil)), false, true),
		newAreEqualTestCase("typed nil pointers",
			S[any]((*int)(nil), (*int)(nil)), true, true),
		newAreEqualTestCase("nil slices",
			S[any]([]int(nil), []int(nil)), true, true),
		newAreEqualTestCase("nil vs empty slice",
			S[any]([]int(nil), S[int]()), false, true),
		newAreEqualTestCase("nil maps",
			S[any](map[string]int(nil), map[string]int(nil)), true, true),
	}
}

func areEqualIdentityTestCases() []areEqualTestCase {
	slice := S(1, 2, 3)
	m := map[string]int{"a": 1}
	o := reflect.ValueOf(NewOpaqueFieldsForTest(42, 1))

	return []areEqualTestCase{
		newAreEqualTestCase("same slice", S[any](slice, slice), true, true),
		newAreEqualTestCase("same map", S[any](m, m), true, true),
		newAreEqualTestCase("unexported fields stay unknown",
			S[any](o.Field(1), o.Field(1)), false, false),
		newAreEqualTestCase("distinct maps",
			S[any](map[string]int{}, map[string]int{}), false, false),
		newAreEqualTestCase("undecided then settled",
			S[any](map[string]int{"a": 1}, map[string]int{"a": 1}, "x"),
			false, true),
	}
}

func areEqualMethodTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualTestCase("Equal method match",
			S[any](EqSliceForTest{1, 2}, EqSliceForTest{1, 2}), true, true),
		newAreEqualTestCase("Equal method mismatch",
			S[any](EqSliceForTest{1}, EqSliceForTest{2}), false, true),
		newAreEqualTestCase("Equal beats elements",
			S[any](eqNeverSlice{1}, eqNeverSlice{1}), false, true),
		newAreEqualTestCase("panicking Equal falls back to elements",
			S[any](panicEqSlice{1}, panicEqSlice{1}), true, true),
		newAreEqualTestCase("mismatched Equal signature falls back to elements",
			S[any](oddEqSlice{1}, oddEqSlice{1}), true, true),
		newAreEqualTestCase("panicking Equal without elements stays unknown",
			S[any](panicEqMap{"a": 1}, panicEqMap{"a": 1}), false, false),
	}
}

func areEqualSliceTestCases() []areEqualTestCase {
	shared := S(1, 2, 3)

	return []areEqualTestCase{
		newAreEqualTestCase("distinct equal slices",
			S[any](S(1, 2, 3), S(1, 2, 3)), true, true),
		newAreEqualTestCase("unequal elements",
			S[any](S(1, 2, 3), S(1, 2, 4)), false, true),
		newAreEqualTestCase("different lengths",
			S[any](S(1, 2), S(1, 2, 3)), false, true),
		newAreEqualTestCase("empty slices",
			S[any](S[int](), S[int]()), true, true),
		newAreEqualTestCase("equal interface elements",
			S[any](S[any](1, "a"), S[any](1, "a")), true, true),
		newAreEqualTestCase("nil interface elements",
			S[any](S[any](nil), S[any](nil)), true, true),
		newAreEqualTestCase("interface elements of different types",
			S[any](S[any](1), S[any]("a")), false, true),
		newAreEqualTestCase("shared nested slice",
			S[any](S[any](shared), S[any](shared)), true, true),
		newAreEqualTestCase("distinct nested slices stay unknown",
			S[any](S(S(1)), S(S(1))), false, false),
		newAreEqualTestCase("unknown element then settled",
			S[any](S[any](S(1), 1), S[any](S(2), 2)), false, true),
	}
}

func TestAreEqual(t *testing.T) {
	t.Run("comparable", func(t *testing.T) {
		RunTestCases(t, areEqualComparableTestCases())
	})
	t.Run("nil", func(t *testing.T) {
		RunTestCases(t, areEqualNilTestCases())
	})
	t.Run("identity", func(t *testing.T) {
		RunTestCases(t, areEqualIdentityTestCases())
	})
	t.Run("Equal method", func(t *testing.T) {
		RunTestCases(t, areEqualMethodTestCases())
	})
	t.Run("slices", func(t *testing.T) {
		RunTestCases(t, areEqualSliceTestCases())
	})
}

// TestAreEqualComparableEqualCalls pins that AreEqual consults a
// comparable type's Equal method only when == is false: a true == skips
// it outright, a false == invokes it. The table rows above cover the
// returned verdict; this covers the call itself.
func TestAreEqualComparableEqualCalls(t *testing.T) {
	eqNeverComparableCalls.Store(0)
	is, _ := AreEqual(eqNeverComparable(1), eqNeverComparable(1))
	AssertTrue(t, is, "true == equal")
	AssertComparableForTest(t, int64(0), eqNeverComparableCalls.Load(), "Equal skipped on true ==")

	eqNeverComparableCalls.Store(0)
	is, _ = AreEqual(eqNeverComparable(1), eqNeverComparable(2))
	AssertFalse(t, is, "false == unequal")
	AssertComparableForTest(t, int64(1), eqNeverComparableCalls.Load(), "Equal consulted on false ==")
}
