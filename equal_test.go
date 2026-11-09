package core_test

import (
	"math"
	"reflect"
	"sync/atomic"
	"testing"

	"darvaza.org/core"
)

// TestCase validations
var _ core.TestCase = areEqualTestCase{}

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

type areEqualTestCase struct {
	name      string
	vvi       []any
	wantIs    bool
	wantKnown bool
}

func (tc areEqualTestCase) Name() string { return tc.name }

func (tc areEqualTestCase) Test(t *testing.T) {
	is, known := core.AreEqual(tc.vvi...)
	core.AssertComparableForTest(t, tc.wantIs, is, "is")
	core.AssertComparableForTest(t, tc.wantKnown, known, "known")
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
		newAreEqualTestCase("single value", core.S[any](42), true, true),
		newAreEqualTestCase("equal ints", core.S[any](42, 42), true, true),
		newAreEqualTestCase("unequal ints", core.S[any](42, 43), false, true),
		newAreEqualTestCase("different types", core.S[any](42, "42"), false, true),
		newAreEqualTestCase("int vs int64", core.S[any](42, int64(42)), false, true),
		newAreEqualTestCase("equal chain", core.S[any]("a", "a", "a"), true, true),
		newAreEqualTestCase("chain broken at the end",
			core.S[any]("a", "a", "b"), false, true),
		newAreEqualTestCase("NaN never equals",
			core.S[any](math.NaN(), math.NaN()), false, true),
		newAreEqualTestCase("Equal rescues a false ==",
			core.S[any](eqAlways(1), eqAlways(2)), true, true),
		newAreEqualTestCase("true == skips Equal",
			core.S[any](eqNeverComparable(1), eqNeverComparable(1)), true, true),
		newAreEqualTestCase("false == with false Equal stays unequal",
			core.S[any](eqNeverComparable(1), eqNeverComparable(2)), false, true),
		newAreEqualTestCase("held slices reached stay unknown",
			core.S[any](core.HeldValuesForTest{1, core.S(1)}, core.HeldValuesForTest{1, core.S(1)}), false, false),
		newAreEqualTestCase("held slices behind a difference",
			core.S[any](core.HeldValuesForTest{1, core.S(1)}, core.HeldValuesForTest{2, core.S(1)}), false, true),
	}
}

func areEqualNilTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualTestCase("both untyped nil", core.S[any](nil, nil), true, true),
		newAreEqualTestCase("untyped nil vs zero", core.S[any](nil, 0), false, true),
		newAreEqualTestCase("untyped nil vs typed nil",
			core.S[any](nil, (*int)(nil)), false, true),
		newAreEqualTestCase("typed nil pointers",
			core.S[any]((*int)(nil), (*int)(nil)), true, true),
		newAreEqualTestCase("nil slices",
			core.S[any]([]int(nil), []int(nil)), true, true),
		newAreEqualTestCase("nil vs empty slice",
			core.S[any]([]int(nil), core.S[int]()), false, true),
		newAreEqualTestCase("nil maps",
			core.S[any](map[string]int(nil), map[string]int(nil)), true, true),
	}
}

func areEqualIdentityTestCases() []areEqualTestCase {
	slice := core.S(1, 2, 3)
	m := map[string]int{"a": 1}
	o := reflect.ValueOf(core.NewOpaqueFieldsForTest(42, 1))

	return []areEqualTestCase{
		newAreEqualTestCase("same slice", core.S[any](slice, slice), true, true),
		newAreEqualTestCase("same map", core.S[any](m, m), true, true),
		newAreEqualTestCase("unexported fields stay unknown",
			core.S[any](o.Field(1), o.Field(1)), false, false),
		newAreEqualTestCase("distinct maps",
			core.S[any](map[string]int{}, map[string]int{}), false, false),
		newAreEqualTestCase("undecided then settled",
			core.S[any](map[string]int{"a": 1}, map[string]int{"a": 1}, "x"),
			false, true),
	}
}

func areEqualMethodTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualTestCase("Equal method match",
			core.S[any](core.EqSliceForTest{1, 2}, core.EqSliceForTest{1, 2}), true, true),
		newAreEqualTestCase("Equal method mismatch",
			core.S[any](core.EqSliceForTest{1}, core.EqSliceForTest{2}), false, true),
		newAreEqualTestCase("Equal beats elements",
			core.S[any](eqNeverSlice{1}, eqNeverSlice{1}), false, true),
		newAreEqualTestCase("panicking Equal falls back to elements",
			core.S[any](panicEqSlice{1}, panicEqSlice{1}), true, true),
		newAreEqualTestCase("mismatched Equal signature falls back to elements",
			core.S[any](oddEqSlice{1}, oddEqSlice{1}), true, true),
		newAreEqualTestCase("panicking Equal without elements stays unknown",
			core.S[any](panicEqMap{"a": 1}, panicEqMap{"a": 1}), false, false),
	}
}

func areEqualSliceTestCases() []areEqualTestCase {
	shared := core.S(1, 2, 3)

	return []areEqualTestCase{
		newAreEqualTestCase("distinct equal slices",
			core.S[any](core.S(1, 2, 3), core.S(1, 2, 3)), true, true),
		newAreEqualTestCase("unequal elements",
			core.S[any](core.S(1, 2, 3), core.S(1, 2, 4)), false, true),
		newAreEqualTestCase("different lengths",
			core.S[any](core.S(1, 2), core.S(1, 2, 3)), false, true),
		newAreEqualTestCase("empty slices",
			core.S[any](core.S[int](), core.S[int]()), true, true),
		newAreEqualTestCase("equal interface elements",
			core.S[any](core.S[any](1, "a"), core.S[any](1, "a")), true, true),
		newAreEqualTestCase("nil interface elements",
			core.S[any](core.S[any](nil), core.S[any](nil)), true, true),
		newAreEqualTestCase("interface elements of different types",
			core.S[any](core.S[any](1), core.S[any]("a")), false, true),
		newAreEqualTestCase("shared nested slice",
			core.S[any](core.S[any](shared), core.S[any](shared)), true, true),
		newAreEqualTestCase("distinct nested slices stay unknown",
			core.S[any](core.S(core.S(1)), core.S(core.S(1))), false, false),
		newAreEqualTestCase("unknown element then settled",
			core.S[any](core.S[any](core.S(1), 1), core.S[any](core.S(2), 2)), false, true),
	}
}

func TestAreEqual(t *testing.T) {
	t.Run("comparable", func(t *testing.T) {
		core.RunTestCases(t, areEqualComparableTestCases())
	})
	t.Run("nil", func(t *testing.T) {
		core.RunTestCases(t, areEqualNilTestCases())
	})
	t.Run("identity", func(t *testing.T) {
		core.RunTestCases(t, areEqualIdentityTestCases())
	})
	t.Run("Equal method", func(t *testing.T) {
		core.RunTestCases(t, areEqualMethodTestCases())
	})
	t.Run("slices", func(t *testing.T) {
		core.RunTestCases(t, areEqualSliceTestCases())
	})
}

// TestAreEqualComparableEqualCalls pins that AreEqual consults a
// comparable type's Equal method only when == is false: a true == skips
// it outright, a false == invokes it. The table rows above cover the
// returned verdict; this covers the call itself.
func TestAreEqualComparableEqualCalls(t *testing.T) {
	eqNeverComparableCalls.Store(0)
	is, _ := core.AreEqual(eqNeverComparable(1), eqNeverComparable(1))
	core.AssertTrue(t, is, "true == equal")
	core.AssertComparableForTest(t, int64(0), eqNeverComparableCalls.Load(), "Equal skipped on true ==")

	eqNeverComparableCalls.Store(0)
	is, _ = core.AreEqual(eqNeverComparable(1), eqNeverComparable(2))
	core.AssertFalse(t, is, "false == unequal")
	core.AssertComparableForTest(t, int64(1), eqNeverComparableCalls.Load(), "Equal consulted on false ==")
}

// Benchmarks
//
// AreEqual buys its panic-free breadth with reflection, and these measure
// what that costs against a plain ==. They exist to document the gap, not
// to be closed: the Assert family that calls AreEqual runs in tests, where
// a decisive answer and a useful message beat nanoseconds.

// benchEqualSink keeps the compiler from eliding the comparisons below.
var benchEqualSink bool

func BenchmarkEqualOperator(b *testing.B) {
	x, y := 42, 42
	for b.Loop() {
		benchEqualSink = x == y
	}
}

func BenchmarkAreEqualComparable(b *testing.B) {
	x, y := 42, 42
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}

func BenchmarkAreEqualMethod(b *testing.B) {
	x, y := core.EqSliceForTest{1, 2, 3}, core.EqSliceForTest{1, 2, 3}
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}

func BenchmarkAreEqualSlice(b *testing.B) {
	x, y := core.S(1, 2, 3), core.S(1, 2, 3)
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}

// BenchmarkAreEqualUndecided measures the walk that ends in no answer,
// the path Assert functions turn into a failure.
func BenchmarkAreEqualUndecided(b *testing.B) {
	x, y := map[string]int{"a": 1}, map[string]int{"a": 1}
	for b.Loop() {
		benchEqualSink, _ = core.AreEqual(x, y)
	}
}
