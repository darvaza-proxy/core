package core_test

import (
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

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

// boxed compares its contents under Equal and its identity under ==.
// Equal panics on a nil receiver.
type boxed struct{ n int }

func (b *boxed) Equal(o *boxed) bool { return b.n == o.n }

// newAdder returns a closure adding n, so every closure it returns
// shares one function literal's code. It is kept out of line, as an
// inlined call can carry a copy of the literal of its own.
//
//go:noinline
func newAdder(n int) func(int) int {
	return func(x int) int { return x + n }
}

// makeReached returns v as reflection reaches it through an unexported
// field, where it cannot be unwrapped via Interface().
func makeReached[T any](v T) reflect.Value {
	return reflect.ValueOf(struct{ v T }{v}).Field(0)
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

// newAreEqualSettledTestCase builds a row whose values AreEqual settles,
// as equal or not by wantIs.
func newAreEqualSettledTestCase(name string, vvi []any,
	wantIs bool) areEqualTestCase {
	return newAreEqualTestCase(name, vvi, wantIs, true)
}

func areEqualSettledComparableTestCases() []areEqualTestCase {
	p := new(int)
	epoch := time.Unix(0, 0)
	elsewhere := time.FixedZone("", 3600)
	now := time.Now()

	return []areEqualTestCase{
		newAreEqualSettledTestCase("same pointer", core.S[any](p, p), true),
		newAreEqualSettledTestCase("distinct pointers",
			core.S[any](p, new(int)), false),
		newAreEqualSettledTestCase("equal arrays",
			core.S[any]([2]int{1, 2}, [2]int{1, 2}), true),
		newAreEqualSettledTestCase("unequal arrays",
			core.S[any]([2]int{1, 2}, [2]int{1, 3}), false),
		newAreEqualSettledTestCase("equal structs",
			core.S[any](struct{ A int }{1}, struct{ A int }{1}), true),
		newAreEqualSettledTestCase("unequal structs",
			core.S[any](struct{ A int }{1}, struct{ A int }{2}), false),
		newAreEqualSettledTestCase("Equal compares pointer contents",
			core.S[any](&boxed{1}, &boxed{1}), true),
		newAreEqualSettledTestCase("Equal tells pointer contents apart",
			core.S[any](&boxed{1}, &boxed{2}), false),
		newAreEqualSettledTestCase("panicking Equal lets a false == stand",
			core.S[any]((*boxed)(nil), &boxed{1}), false),
		newAreEqualSettledTestCase("Equal compares instants across zones",
			core.S[any](epoch, epoch.In(elsewhere)), true),
		newAreEqualSettledTestCase("Equal ignores the monotonic reading",
			core.S[any](now, now.Round(0)), true),
		newAreEqualSettledTestCase("settled before an undecided pair",
			core.S[any](1, 2, map[string]int{}), false),
	}
}

func areEqualSettledNilTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualSettledTestCase("nil functions",
			core.S[any]((func())(nil), (func())(nil)), true),
		newAreEqualSettledTestCase("nil function against a set one",
			core.S[any]((func(float64) float64)(nil), math.Sin), false),
		newAreEqualSettledTestCase("nil map against a set one",
			core.S[any](map[string]int(nil), map[string]int{}), false),
		newAreEqualSettledTestCase("nil ahead of a matching Equal",
			core.S[any](core.EqSliceForTest(nil), core.EqSliceForTest{}), false),
		newAreEqualSettledTestCase("nil ahead of a failing Equal",
			core.S[any](eqNeverSlice(nil), eqNeverSlice(nil)), true),
		newAreEqualSettledTestCase("nil ahead of a panicking Equal",
			core.S[any](panicEqMap(nil), panicEqMap{}), false),
	}
}

func areEqualSettledIdentityTestCases() []areEqualTestCase {
	never := eqNeverSlice{1}
	m := panicEqMap{"a": 1}

	return []areEqualTestCase{
		newAreEqualSettledTestCase("same function",
			core.S[any](math.Sin, math.Sin), true),
		newAreEqualSettledTestCase("closures of one literal",
			core.S[any](newAdder(1), newAdder(2)), true),
		newAreEqualSettledTestCase("identity ahead of a failing Equal",
			core.S[any](never, never), true),
		newAreEqualSettledTestCase("identity ahead of a panicking Equal",
			core.S[any](m, m), true),
	}
}

func areEqualSettledSliceTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualSettledTestCase("elements rescued by Equal",
			core.S[any]([]eqAlways{1}, []eqAlways{2}), true),
		newAreEqualSettledTestCase("nested slices settled by Equal",
			core.S[any]([]core.EqSliceForTest{{1}}, []core.EqSliceForTest{{1}}), true),
		newAreEqualSettledTestCase("nested slices told apart by Equal",
			core.S[any]([]core.EqSliceForTest{{1}}, []core.EqSliceForTest{{2}}), false),
		newAreEqualSettledTestCase("nested nil slices",
			core.S[any]([][]int{nil}, [][]int{nil}), true),
		newAreEqualSettledTestCase("nested nil against an empty slice",
			core.S[any]([][]int{nil}, [][]int{{}}), false),
		newAreEqualSettledTestCase("lengths differ around undecided elements",
			core.S[any]([]map[string]int{{}}, []map[string]int{{}, {}}), false),
	}
}

func areEqualSettledReachedTestCases() []areEqualTestCase {
	p := new(int)
	s := core.S(1)
	m := map[string]int{}

	return []areEqualTestCase{
		newAreEqualSettledTestCase("reached nil pointer against a set one",
			core.S[any](makeReached((*int)(nil)), makeReached(p)), false),
		newAreEqualSettledTestCase("reached same pointer",
			core.S[any](makeReached(p), makeReached(p)), true),
		newAreEqualSettledTestCase("reached nil slices",
			core.S[any](makeReached([]int(nil)), makeReached([]int(nil))), true),
		newAreEqualSettledTestCase("reached same slice",
			core.S[any](makeReached(s), makeReached(s)), true),
		newAreEqualSettledTestCase("reached slices of different lengths",
			core.S[any](makeReached(core.S(1)), makeReached(core.S(1, 2))), false),
		newAreEqualSettledTestCase("reached slices with Equal, lengths differ",
			core.S[any](makeReached(eqNeverSlice{1}),
				makeReached(eqNeverSlice{1, 2})), false),
		newAreEqualSettledTestCase("reached nil element against a set one",
			core.S[any](makeReached([]*int{nil}), makeReached([]*int{p})), false),
		newAreEqualSettledTestCase("reached same pointer elements",
			core.S[any](makeReached([]*int{p}), makeReached([]*int{p})), true),
		newAreEqualSettledTestCase("reached same map",
			core.S[any](makeReached(m), makeReached(m)), true),
		newAreEqualSettledTestCase("reached nil interface against a set one",
			core.S[any](makeReached[any](nil), makeReached[any](1)), false),
	}
}

// TestAreEqualSettled pins pairs AreEqual settles, so a change that
// leaves one of them unknown fails.
func TestAreEqualSettled(t *testing.T) {
	t.Run("comparable", func(t *testing.T) {
		core.RunTestCases(t, areEqualSettledComparableTestCases())
	})
	t.Run("nil", func(t *testing.T) {
		core.RunTestCases(t, areEqualSettledNilTestCases())
	})
	t.Run("identity", func(t *testing.T) {
		core.RunTestCases(t, areEqualSettledIdentityTestCases())
	})
	t.Run("slices", func(t *testing.T) {
		core.RunTestCases(t, areEqualSettledSliceTestCases())
	})
	t.Run("reached through unexported fields", func(t *testing.T) {
		core.RunTestCases(t, areEqualSettledReachedTestCases())
	})
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
