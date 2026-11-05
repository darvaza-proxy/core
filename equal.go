package core

import "reflect"

// AreEqual reports whether every given value equals the next one,
// without ever panicking. Values of the same comparable type are tested
// with == first: a true == is authoritative, and a false == defers to
// the value's own Equal method, following the Equal(T) bool convention.
// A comparable type whose == tests more than its own notion of equality
// still settles correctly this way:
//   - time.Time compares its monotonic reading under ==, its instant
//     under Equal.
//   - a pointer type that defines Equal compares identity under ==, its
//     contents under Equal.
//
// When == is unavailable, the question is settled by typed nils, by
// identity (the same underlying data, as [IsSame] sees it), and by the
// Equal method standing in for ==. Functions those leave undecided
// are unequal.
//
// A comparable type can still hold a value == panics on, as a struct
// with an interface field holding a slice does. == is attempted all the
// same: a difference found before that value settles the pair as
// unequal, and a panic leaves the pair to the rules for values without
// ==.
//
// Slices without a decisive Equal method are compared element by
// element, one level deep: lengths must match, and each element pair
// is decided by the same rules, except that nested slices are not
// walked — those settle only by nil, identity, or their own Equal
// method. A nil slice equals only nil — unlike slices.Equal, the
// empty slice is not its equal.
//
// known reports whether the answer is settled. A pair that is neither
// comparable, nil, identical, decided by an Equal method, a function,
// nor a slice settled element by element leaves the list undecided —
// deciding would take the deep comparison AreEqual deliberately
// avoids — unless a later pair settles the whole list as unequal.
// Callers that need a decision anyway can fall back to
// [reflect.DeepEqual] when known is false.
//
// Untyped nil only equals untyped nil, values of different types are
// never equal, and a single value is vacuously equal. AreEqual
// returns (false, true) when no values are given.
func AreEqual(vvi ...any) (is, known bool) {
	if len(vvi) == 0 {
		return false, true
	}

	known = true
	prev := newComparableValue(vvi[0])
	for _, vi := range vvi[1:] {
		next := newComparableValue(vi)

		switch is2, known2 := areEqual2(prev, next, nil); {
		case known2 && !is2:
			// one unequal pair settles the whole list
			return false, true
		case !known2:
			// undecided; a later pair may still settle the list
			known = false
		default:
			// pair equal; keep walking
		}

		prev = next
	}

	if !known {
		// unknown
		return false, false
	}
	return true, true
}

// areEqualPath holds a fingerprint of each slice the walk is inside,
// outermost first.
type areEqualPath []areEqualFingerprint

// areEqualFingerprint describes a slice on an [areEqualPath].
type areEqualFingerprint struct {
	t reflect.Type
}

// newAreEqualFingerprint returns the fingerprint of the slice v.
func newAreEqualFingerprint(v reflect.Value) areEqualFingerprint {
	return areEqualFingerprint{t: v.Type()}
}

// Append returns the path with the fingerprint of v appended.
func (p areEqualPath) Append(v reflect.Value) areEqualPath {
	return append(p, newAreEqualFingerprint(v))
}

// areEqual2 decides equality for a single pair. known reports whether
// the answer is settled. path fingerprints the slices the walk is
// inside, none for the pairs of [AreEqual]'s list.
func areEqual2(a, b comparableValue, path areEqualPath) (is, known bool) {
	switch {
	case a.t != b.t:
		// different types are never equal
		return false, true
	case a.t == nil:
		// both untyped nil
		return true, true
	case a.ok && b.ok:
		// == is attempted; a false == still defers to an Equal
		// method, and a panic to the rules for values without ==
		return areEqualComparable(a.v, b.v, path)
	default:
		return areEqualFallback(a.v, b.v, path)
	}
}

// areEqualComparable settles two operands that both support ==. A true
// == is authoritative and skips the Equal method. A false == defers to
// the Equal method, following the Equal(T) bool convention, so a type
// whose == tests more than its own notion of equality still settles
// correctly. Without a decisive Equal method, == stands. Operands of a
// comparable type holding a value == panics on are settled as those
// without == are.
func areEqualComparable(va, vb reflect.Value, path areEqualPath) (is, known bool) {
	switch equal, settled := comparableEqual(va.Interface(), vb.Interface()); {
	case !settled:
		// == panicked
		return areEqualFallback(va, vb, path)
	case equal:
		// authoritative
		return true, true
	default:
		if is, known = equalMethod(va, vb); known {
			// Equal rescues a pair == calls unequal
			return is, true
		}

		// == is authoritative when no Equal method decides
		return false, true
	}
}

// areEqualFallback decides equality when == is unavailable: typed
// nils and identity settle the question — nil only equals nil, and
// identity proves equality — and the value's own Equal method stands
// in for ==. A function left undecided is unequal. A slice left
// undecided outside any other is compared element by element.
// Anything else stays unknown — two distinct values may still be
// equal, and deciding that would take a deep comparison.
func areEqualFallback(va, vb reflect.Value, path areEqualPath) (is, known bool) {
	if same, ok := isSameTypedNil(va, vb); ok {
		// nil only equals nil
		return same, true
	}

	if isSamePointer(va, vb) {
		// identity proves equality
		return true, true
	}

	if is, known = equalMethod(va, vb); known {
		return is, true
	}

	if va.Kind() == reflect.Func {
		// a function left undecided is unequal
		return false, true
	}

	return areEqualSlice(va, vb, path)
}

// areEqualSlice compares two slices of the same type element by
// element. Lengths must match; anything that isn't a slice, and a
// slice inside another, stays unknown.
func areEqualSlice(va, vb reflect.Value, path areEqualPath) (is, known bool) {
	switch {
	case va.Kind() != reflect.Slice, len(path) > 0:
		// only a slice outside any other is walked
		return false, false
	case va.Len() != vb.Len():
		// different lengths are never equal
		return false, true
	default:
		return areEqualSliceElements(va, vb, path.Append(va))
	}
}

// areEqualSliceElements walks the element pairs of two equal-length
// slices, aggregating like [AreEqual] does over its list: one unequal
// element settles the pair, an undecided element leaves it unknown
// unless a later element settles it.
func areEqualSliceElements(va, vb reflect.Value, path areEqualPath) (is, known bool) {
	known = true
	for i := range va.Len() {
		switch is2, known2 := areEqualElement(va.Index(i), vb.Index(i), path); {
		case known2 && !is2:
			// one unequal element settles the whole pair
			return false, true
		case !known2:
			// undecided; a later element may still settle the pair
			known = false
		default:
			// element equal; keep walking
		}
	}

	if !known {
		// unknown
		return false, false
	}
	return true, true
}

// areEqualElement decides equality for one pair of slice elements,
// judged like top-level operands — interface elements are unwrapped
// first — but without walking nested slices.
func areEqualElement(va, vb reflect.Value, path areEqualPath) (is, known bool) {
	a := newComparableValue(unwrapInterface(va))
	b := newComparableValue(unwrapInterface(vb))
	return areEqual2(a, b, path)
}

// unwrapInterface returns the value held by an interface so its
// content can be judged on its own; anything else passes through
// unchanged.
func unwrapInterface(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		return v.Elem()
	}
	return v
}

// equalMethod consults the value's own Equal method, following the
// Equal(T) bool convention, to decide a pair == cannot settle, whether
// because == is unavailable or because it called the pair unequal.
// Without such a method, or if the call panics, the question stays
// unknown.
func equalMethod(va, vb reflect.Value) (is, known bool) {
	// a panicking Equal leaves the zero (false, false)
	defer func() {
		_ = recover()
	}()

	m := va.MethodByName("Equal")
	if !m.IsValid() || !isEqualMethodType(m.Type(), vb.Type()) {
		// unknown
		return false, false
	}

	return m.Call([]reflect.Value{vb})[0].Bool(), true
}

// isEqualMethodType reports whether a method signature matches
// Equal(T) bool for the given operand type.
func isEqualMethodType(mt, arg reflect.Type) bool {
	return mt.NumIn() == 1 && mt.In(0) == arg &&
		mt.NumOut() == 1 && mt.Out(0).Kind() == reflect.Bool
}

// comparableEqual compares two values with ==. known is false when the
// comparison panicked, which an interface holding a value of a
// non-comparable type does.
func comparableEqual[U comparable](a, b U) (equal, known bool) {
	defer func() {
		if recover() != nil {
			equal, known = false, false
		}
	}()

	return a == b, true
}

// comparableValue carries the reflection state of one [AreEqual]
// operand so each value is reflected only once.
type comparableValue struct {
	t reflect.Type
	v reflect.Value
	// ok reports whether == can be attempted on v via Interface():
	// its type is comparable and it can be unwrapped. A comparable
	// type may still hold a value == panics on, which
	// [comparableEqual] reports as unknown.
	ok bool
}

// newComparableValue captures the reflection state of one value.
// Untyped nil yields a nil type and no direct == support; its
// equality is decided by type alone.
func newComparableValue(vi any) comparableValue {
	v := asReflectValue(vi)
	if !v.IsValid() {
		// untyped nil
		return comparableValue{v: v}
	}

	return comparableValue{
		t:  v.Type(),
		v:  v,
		ok: v.Type().Comparable() && v.CanInterface(),
	}
}
