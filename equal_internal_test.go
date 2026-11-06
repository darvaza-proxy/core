package core

import (
	"math"
	"reflect"
	"testing"
	"time"
)

// boxed compares its contents under Equal and its identity under ==.
// Equal panics on a nil receiver.
type boxed struct{ n int }

func (b *boxed) Equal(o *boxed) bool { return b.n == o.n }

// zeroBoxed compares its contents under Equal, a nil pointer counting
// as the zero value.
type zeroBoxed struct{ n int }

func (b *zeroBoxed) Equal(o *zeroBoxed) bool { return b.value() == o.value() }

func (b *zeroBoxed) value() int {
	if b == nil {
		return 0
	}
	return b.n
}

// eqAlwaysFunc is a function type with an Equal method that always
// matches.
type eqAlwaysFunc func(float64) float64

func (eqAlwaysFunc) Equal(eqAlwaysFunc) bool { return true }

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

// makeInterfaceValue returns v as a reflect.Value of interface kind.
func makeInterfaceValue(v any) reflect.Value {
	return reflect.ValueOf(&v).Elem()
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
		newAreEqualSettledTestCase("same pointer", S[any](p, p), true),
		newAreEqualSettledTestCase("distinct pointers",
			S[any](p, new(int)), false),
		newAreEqualSettledTestCase("equal arrays",
			S[any]([2]int{1, 2}, [2]int{1, 2}), true),
		newAreEqualSettledTestCase("unequal arrays",
			S[any]([2]int{1, 2}, [2]int{1, 3}), false),
		newAreEqualSettledTestCase("equal structs",
			S[any](struct{ A int }{1}, struct{ A int }{1}), true),
		newAreEqualSettledTestCase("unequal structs",
			S[any](struct{ A int }{1}, struct{ A int }{2}), false),
		newAreEqualSettledTestCase("Equal compares pointer contents",
			S[any](&boxed{1}, &boxed{1}), true),
		newAreEqualSettledTestCase("Equal tells pointer contents apart",
			S[any](&boxed{1}, &boxed{2}), false),
		newAreEqualSettledTestCase("panicking Equal lets a false == stand",
			S[any]((*boxed)(nil), &boxed{1}), false),
		newAreEqualSettledTestCase("Equal compares instants across zones",
			S[any](epoch, epoch.In(elsewhere)), true),
		newAreEqualSettledTestCase("Equal ignores the monotonic reading",
			S[any](now, now.Round(0)), true),
		newAreEqualSettledTestCase("Equal behind an interface value",
			S[any](makeInterfaceValue(now),
				makeInterfaceValue(now.Round(0))), true),
		newAreEqualSettledTestCase("settled before an undecided pair",
			S[any](1, 2, map[string]int{}), false),
	}
}

func areEqualSettledNilTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualSettledTestCase("nil functions",
			S[any]((func())(nil), (func())(nil)), true),
		newAreEqualSettledTestCase("nil function against a set one",
			S[any]((func(float64) float64)(nil), math.Sin), false),
		newAreEqualSettledTestCase("nil map against a set one",
			S[any](map[string]int(nil), map[string]int{}), false),
		newAreEqualSettledTestCase("nil ahead of a matching Equal",
			S[any](EqSliceForTest(nil), EqSliceForTest{}), false),
		newAreEqualSettledTestCase("nil pointer ahead of a matching Equal",
			S[any]((*zeroBoxed)(nil), &zeroBoxed{}), false),
		newAreEqualSettledTestCase("nil ahead of a failing Equal",
			S[any](eqNeverSlice(nil), eqNeverSlice(nil)), true),
		newAreEqualSettledTestCase("nil ahead of a panicking Equal",
			S[any](panicEqMap(nil), panicEqMap{}), false),
	}
}

func areEqualSettledIdentityTestCases() []areEqualTestCase {
	never := eqNeverSlice{1}
	m := panicEqMap{"a": 1}

	return []areEqualTestCase{
		newAreEqualSettledTestCase("same function",
			S[any](math.Sin, math.Sin), true),
		newAreEqualSettledTestCase("distinct functions",
			S[any](math.Sin, math.Cos), false),
		newAreEqualSettledTestCase("Equal rescues distinct functions",
			S[any](eqAlwaysFunc(math.Sin), eqAlwaysFunc(math.Cos)), true),
		newAreEqualSettledTestCase("closures of one literal",
			S[any](newAdder(1), newAdder(2)), true),
		newAreEqualSettledTestCase("identity ahead of a failing Equal",
			S[any](never, never), true),
		newAreEqualSettledTestCase("identity ahead of a panicking Equal",
			S[any](m, m), true),
	}
}

func areEqualSettledSliceTestCases() []areEqualTestCase {
	return []areEqualTestCase{
		newAreEqualSettledTestCase("elements rescued by Equal",
			S[any]([]eqAlways{1}, []eqAlways{2}), true),
		newAreEqualSettledTestCase("nested slices settled by Equal",
			S[any]([]EqSliceForTest{{1}}, []EqSliceForTest{{1}}), true),
		newAreEqualSettledTestCase("nested slices told apart by Equal",
			S[any]([]EqSliceForTest{{1}}, []EqSliceForTest{{2}}), false),
		newAreEqualSettledTestCase("nested nil slices",
			S[any]([][]int{nil}, [][]int{nil}), true),
		newAreEqualSettledTestCase("nested nil against an empty slice",
			S[any]([][]int{nil}, [][]int{{}}), false),
		newAreEqualSettledTestCase("lengths differ around undecided elements",
			S[any]([]map[string]int{{}}, []map[string]int{{}, {}}), false),
	}
}

func areEqualSettledReachedTestCases() []areEqualTestCase {
	p := new(int)
	s := S(1)
	m := map[string]int{}

	return []areEqualTestCase{
		newAreEqualSettledTestCase("reached equal ints",
			S[any](makeReached(1), makeReached(1)), true),
		newAreEqualSettledTestCase("reached unequal ints",
			S[any](makeReached(1), makeReached(2)), false),
		newAreEqualSettledTestCase("reached equal interfaces",
			S[any](makeReached[any](1), makeReached[any](1)), true),
		newAreEqualSettledTestCase("reached unequal interfaces",
			S[any](makeReached[any](1), makeReached[any](2)), false),
		newAreEqualSettledTestCase("reached nil pointer against a set one",
			S[any](makeReached((*int)(nil)), makeReached(p)), false),
		newAreEqualSettledTestCase("reached same pointer",
			S[any](makeReached(p), makeReached(p)), true),
		newAreEqualSettledTestCase("reached nil ahead of an Equal out of reach",
			S[any](makeReached((*boxed)(nil)), makeReached(&boxed{1})), false),
		newAreEqualSettledTestCase("reached nil slices",
			S[any](makeReached([]int(nil)), makeReached([]int(nil))), true),
		newAreEqualSettledTestCase("reached same slice",
			S[any](makeReached(s), makeReached(s)), true),
		newAreEqualSettledTestCase("reached slices of different lengths",
			S[any](makeReached(S(1)), makeReached(S(1, 2))), false),
		newAreEqualSettledTestCase("reached slices with Equal, lengths differ",
			S[any](makeReached(eqNeverSlice{1}),
				makeReached(eqNeverSlice{1, 2})), false),
		newAreEqualSettledTestCase("reached nil element against a set one",
			S[any](makeReached([]*int{nil}), makeReached([]*int{p})), false),
		newAreEqualSettledTestCase("reached same pointer elements",
			S[any](makeReached([]*int{p}), makeReached([]*int{p})), true),
		newAreEqualSettledTestCase("reached same map",
			S[any](makeReached(m), makeReached(m)), true),
		newAreEqualSettledTestCase("reached nil interface against a set one",
			S[any](makeReached[any](nil), makeReached[any](1)), false),
	}
}

// TestAreEqualSettled pins pairs AreEqual settles, so a change that
// leaves one of them unknown fails.
func TestAreEqualSettled(t *testing.T) {
	t.Run("comparable", func(t *testing.T) {
		RunTestCases(t, areEqualSettledComparableTestCases())
	})
	t.Run("nil", func(t *testing.T) {
		RunTestCases(t, areEqualSettledNilTestCases())
	})
	t.Run("identity", func(t *testing.T) {
		RunTestCases(t, areEqualSettledIdentityTestCases())
	})
	t.Run("slices", func(t *testing.T) {
		RunTestCases(t, areEqualSettledSliceTestCases())
	})
	t.Run("reached through unexported fields", func(t *testing.T) {
		RunTestCases(t, areEqualSettledReachedTestCases())
	})
}
