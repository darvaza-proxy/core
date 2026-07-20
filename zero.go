package core

import "reflect"

// Zero returns the zero value of type T.
// It takes a pointer to T as a parameter to infer the type,
// but the pointer value itself is ignored.
//
// This function is useful when you need to obtain the zero value
// of a generic type without having to declare a variable.
//
// Example:
//
//	var p *int
//	zero := Zero(p)  // returns 0
//
//	var s *string
//	empty := Zero(s) // returns ""
//
//	type Person struct { Name string; Age int }
//	var person *Person
//	zeroPerson := Zero(person) // returns Person{Name: "", Age: 0}
func Zero[T any](_ *T) T {
	var zero T
	return zero
}

// IsZero reports whether vi is in a state where it can/should be initialized.
// It answers the question: "Is this value uninitialized and ready to be set?"
//
// IsZero returns true for values that are in their uninitialized state:
//   - nil (untyped nil)
//   - Zero values of basic types (0, "", false)
//   - nil pointers, slices, maps, channels, and functions
//   - Zero-valued structs (all fields are zero)
//   - Values whose IsZero() method returns true
//
// IsZero returns false for values that have been explicitly initialized:
//   - Non-zero basic values (42, "hello", true)
//   - Initialized empty collections ([]int{}, map[string]int{})
//   - Non-nil pointers (even if pointing to zero values)
//   - Non-zero structs (any field is non-zero)
//
// The key insight: something set to nil but assignable is zero;
// something explicitly initialized (even if empty) is not zero.
//
// Initialization semantics:
//   - var slice []int        // nil slice, IsZero = true (needs initialization)
//   - slice := []int{}       // empty but initialized, IsZero = false (already set)
//   - var m map[string]int   // nil map, IsZero = true (needs initialization)
//   - m := make(map[string]int) // empty but initialized, IsZero = false (already set)
//   - var ptr *int           // nil pointer, IsZero = true (can be assigned)
//   - ptr := new(int)        // non-nil pointer, IsZero = false (already assigned)
//
// Example:
//
//	IsZero(nil)              // true  - uninitialized
//	IsZero(0)                // true  - zero basic value
//	IsZero(42)               // false - initialized basic value
//	IsZero("")               // true  - zero string
//	IsZero("hello")          // false - initialized string
//	IsZero([]int(nil))       // true  - nil slice (uninitialized)
//	IsZero([]int{})          // false - empty slice (initialized)
//	IsZero([]int{1, 2})      // false - non-empty slice (initialized)
//	IsZero((*int)(nil))      // true  - nil pointer (uninitialized)
//	IsZero(new(int))         // false - non-nil pointer (initialized)
func IsZero(vi any) bool {
	switch p := vi.(type) {
	case nil:
		// nil
		return true
	case reflect.Value:
		// Special handling for reflect.Value to avoid double-wrapping
		return isReflectValueZero(p)
	case interface {
		IsZero() bool
	}:
		// interface
		return p.IsZero()
	default:
		// reflection
		v := reflect.ValueOf(vi)
		return isReflectValueZero(v)
	}
}

// IsNil reports whether vi is nil or a typed nil value.
// It answers the question: "Is this value nil (typed or untyped)?"
//
// IsNil returns true for:
//   - nil (untyped nil)
//   - nil pointers, slices, maps, channels, functions, and interfaces
//   - reflect.Value that is invalid or has a nil underlying value
//
// IsNil returns false for:
//   - Zero values of basic types (0, "", false) - these are not nil
//   - Non-nil pointers (even if pointing to zero values)
//   - Initialized empty collections ([]int{}, map[string]int{})
//   - Non-nil interfaces containing zero values
//   - Zero-valued structs (structs cannot be nil)
//
// The key distinction from IsZero: IsNil only checks for nil state,
// while IsZero checks for uninitialized state (which includes nil).
//
// Comparison with IsZero:
//   - IsNil(nil)              // true  - untyped nil
//   - IsNil(0)                // false - zero int is not nil
//   - IsNil("")               // false - zero string is not nil
//   - IsNil([]int(nil))       // true  - nil slice
//   - IsNil([]int{})          // false - empty slice is not nil
//   - IsNil((*int)(nil))      // true  - nil pointer
//   - IsNil(new(int))         // false - non-nil pointer
//   - IsNil(struct{}{})       // false - structs cannot be nil
//
// Example:
//
//	var ptr *int
//	IsNil(ptr)                   // true  - nil pointer
//	IsZero(ptr)                  // true  - nil pointer is also zero
//
//	var slice []int
//	IsNil(slice)                 // true  - nil slice
//	IsZero(slice)                // true  - nil slice is also zero
//
//	slice = []int{}
//	IsNil(slice)                 // false - empty slice is not nil
//	IsZero(slice)                // false - empty slice is initialized
//
//	var num int
//	IsNil(num)                   // false - integers cannot be nil
//	IsZero(num)                  // true  - zero integer is uninitialized
func IsNil(vi any) bool {
	if vi == nil {
		return true
	}

	// Use reflection to check for nil
	v := asReflectValue(vi)
	return isReflectValueNil(v)
}

// asReflectValue returns a reflect.Value for the given value.
// If the value is already a reflect.Value, it returns it directly.
// Otherwise, it calls reflect.ValueOf().
func asReflectValue(v any) reflect.Value {
	if rv, ok := v.(reflect.Value); ok {
		return rv
	}
	return reflect.ValueOf(v)
}

// isReflectValueZero reports whether a reflect.Value is zero.
// This helper reduces code duplication in IsZero.
func isReflectValueZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	return v.IsZero()
}

// isReflectValueNil reports whether a reflect.Value is nil.
// This helper reduces code duplication in IsNil.
func isReflectValueNil(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	// Only certain types can be nil
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	default:
		// Basic types, structs, arrays cannot be nil
		return false
	}
}

// isReflectValueSame reports whether two reflect.Values are the same.
// This helper reduces code duplication in IsSame.
func isReflectValueSame(va, vb reflect.Value) bool {
	// Invalid reflect.Values cannot be compared
	if !va.IsValid() || !vb.IsValid() {
		return false
	}

	if va.Type() != vb.Type() {
		// different types
		return false
	}

	if same, ok := isSameTypedNil(va, vb); ok {
		// typed nil decision
		return same
	}

	return isSamePointer(va, vb)
}

// IsSame reports whether two values are the same.
// It answers the question: "Are these values the same?"
//
// For reference types, IsSame returns true when they point to the same
// underlying data:
//   - Two slices pointing to the same backing array.
//   - Two maps pointing to the same map data.
//   - Two channels pointing to the same channel.
//   - Two function values pointing to the same function.
//   - Two pointers pointing to the same address.
//   - Two interfaces containing the same pointer value.
//
// For value types, IsSame returns true when they have equal values:
//   - Numbers with the same value (42 == 42).
//   - Strings with the same content ("hello" == "hello").
//   - Booleans with the same value (true == true).
//
// IsSame returns false for:
//   - Different backing arrays/maps/channels/functions/pointers.
//   - Different values for basic types.
//   - One nil and one non-nil value.
//   - Different types.
//   - Arrays, structs, and other composite value types: being the
//     same thing means a change to one would reach both, but two
//     equal aggregates are distinct storage that can diverge. Compare
//     them by value with [AreEqual].
//   - Values that cannot be unwrapped via Interface(), as reached
//     through unexported struct fields.
//
// Special case for slices: Go's runtime optimises zero-capacity slices
// (make([]T, 0)) to share a common zero-sized allocation. IsSame treats
// these as distinct to preserve expected semantics.
//
// Example:
//
//	// Value types - compared by value
//	IsSame(42, 42)           // true  - same value
//	IsSame(42, 43)           // false - different values
//	IsSame("hello", "hello") // true  - same string content
//	IsSame("hello", "world") // false - different strings
//
//	// Nil handling
//	IsSame(nil, nil)         // true  - both untyped nil
//	var s1, s2 []int         // both nil slices of same type
//	IsSame(s1, s2)           // true  - both nil of same type
//	IsSame(s1, []int{})      // false - one nil, one empty slice
//
//	// Reference types - compared by reference
//	slice1 := []int{1, 2, 3}
//	slice2 := slice1         // same backing array
//	slice3 := []int{1, 2, 3} // different backing array
//	IsSame(slice1, slice2)   // true  - same backing array
//	IsSame(slice1, slice3)   // false - different backing arrays
//
//	map1 := make(map[string]int)
//	map2 := map1             // same map
//	map3 := make(map[string]int) // different map
//	IsSame(map1, map2)       // true  - same map
//	IsSame(map1, map3)       // false - different maps
//
//	x := 42
//	ptr1 := &x
//	ptr2 := ptr1             // same pointer
//	ptr3 := &x               // different pointer variable, same address
//	IsSame(ptr1, ptr2)       // true  - same pointer
//	IsSame(ptr1, ptr3)       // true  - both point to same address
func IsSame(a, b any) bool {
	if same, ok := isSameNil(a, b); ok {
		// untyped nil decision
		return same
	}

	va := asReflectValue(a)
	vb := asReflectValue(b)

	return isReflectValueSame(va, vb)
}

func isSameNil(a, b any) (same, ok bool) {
	switch {
	case a == nil && b == nil:
		// both nil
		return true, true
	case a == nil || b == nil:
		return false, true
		// one nil, one not
	default:
		// neither is nil
		return false, false
	}
}

func isSameTypedNil(va, vb reflect.Value) (result, handled bool) {
	aNil := isReflectValueNil(va)
	bNil := isReflectValueNil(vb)

	switch {
	case aNil && bNil:
		// both nil
		return true, true
	case aNil || bNil:
		// one nil, one not
		return false, true
	default:
		// neither is nil
		return false, false
	}
}

// isReflectSliceZero reports whether a slice has zero length and capacity.
func isReflectSliceZero(v reflect.Value) bool {
	return v.Len() == 0 && v.Cap() == 0
}

// isSamePointer compares two non-nil values of the same type for sameness.
//
// For reference types (pointers, slices, maps, channels, functions), it checks
// if they reference the same underlying memory. For value types (strings,
// numbers, booleans), it checks value equality. For interfaces, it recursively
// compares the contained values.
//
// Special handling for slices: Go's runtime optimisation causes zero-capacity
// slices to share memory. To preserve the semantics that separate make() calls
// create distinct slices, zero-capacity slices are treated as not same.
//
// Preconditions:
//   - va and vb have the same type (checked by caller)
//   - Neither value is nil (nil cases handled by isSameTypedNil)
//
// Returns false for unhandled types (arrays, structs, etc.), and for
// values that cannot be unwrapped via Interface(), as reached through
// unexported struct fields.
func isSamePointer(va, vb reflect.Value) bool {
	var ok bool
	switch va.Kind() {
	case reflect.Slice:
		// Empty slices from separate make() calls may share the same address
		// due to runtime optimisation. Check that at least one has content
		// before comparing pointers, and reject zero pointers to avoid
		// false positives.
		if !isReflectSliceZero(va) || !isReflectSliceZero(vb) {
			pa, pb := va.Pointer(), vb.Pointer()
			ok = pa == pb && pa != 0
		}
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		ok = va.Pointer() == vb.Pointer()
	case reflect.Interface:
		// Extract concrete values and compare them recursively
		ok = isSameInterface(va, vb)
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Bool:
		ok = isSameValue(va, vb)
	default:
	}
	return ok
}

// isSameInterface extracts the values held by two interfaces and
// compares them recursively. Interfaces that cannot be unwrapped —
// unexported struct fields reached by reflection — are never same.
func isSameInterface(va, vb reflect.Value) bool {
	if !va.CanInterface() || !vb.CanInterface() {
		return false
	}
	return IsSame(va.Elem().Interface(), vb.Elem().Interface())
}

// isSameValue compares two primitive values by ==. Values that
// cannot be unwrapped via Interface() — unexported struct fields
// reached by reflection — are never same.
func isSameValue(va, vb reflect.Value) bool {
	return va.CanInterface() && vb.CanInterface() &&
		va.Interface() == vb.Interface()
}

// AreComparable reports whether the given values are safe operands of
// the == operator: a true result guarantees no comparison between
// them can panic. Untyped nils are safe — a nil interface compares
// against anything — while values with a non-comparable dynamic type
// are not. It returns false when no values are given.
func AreComparable(vvi ...any) bool {
	if len(vvi) == 0 {
		return false
	}

	for _, vi := range vvi {
		if !isComparableValue(asReflectValue(vi)) {
			return false
		}
	}

	return true
}

// isComparableValue reports whether a value can be an operand of ==
// without panicking. The invalid value — an untyped nil — compares
// safely against anything.
func isComparableValue(v reflect.Value) bool {
	return !v.IsValid() || v.Comparable()
}

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
// Equal method standing in for ==.
//
// Slices without a decisive Equal method are compared element by
// element, one level deep: lengths must match, and each element pair
// is decided by the same rules, except that nested slices are not
// walked — those settle only by nil, identity, or their own Equal
// method. A nil slice equals only nil — unlike slices.Equal, the
// empty slice is not its equal.
//
// known reports whether the answer is settled. A pair that is neither
// comparable, nil, identical, decided by an Equal method, nor a slice
// settled element by element leaves the list undecided — deciding
// would take the deep comparison AreEqual deliberately avoids —
// unless a later pair settles the whole list as unequal. Callers that
// need a decision anyway can fall back to [reflect.DeepEqual] when
// known is false.
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

		switch is2, known2 := areEqual2(prev, next, true); {
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

// areEqual2 decides equality for a single pair. known reports whether
// the answer is settled. deep allows one level of element-wise slice
// comparison; element pairs pass false so nested slices are not
// walked.
func areEqual2(a, b comparableValue, deep bool) (is, known bool) {
	switch {
	case a.t != b.t:
		// different types are never equal
		return false, true
	case a.t == nil:
		// both untyped nil
		return true, true
	case a.ok && b.ok:
		// == decides; a false == still defers to an Equal method
		return areEqualComparable(a.v, b.v)
	default:
		return areEqualFallback(a.v, b.v, deep)
	}
}

// areEqualComparable settles two operands that both support ==. A true
// == is authoritative and skips the Equal method. A false == defers to
// the Equal method, following the Equal(T) bool convention, so a type
// whose == tests more than its own notion of equality still settles
// correctly. Without a decisive Equal method, == stands.
func areEqualComparable(va, vb reflect.Value) (is, known bool) {
	if va.Interface() == vb.Interface() {
		// authoritative
		return true, true
	}

	if is, known = equalMethod(va, vb); known {
		// Equal rescues a pair == calls unequal
		return is, true
	}

	// == is authoritative when no Equal method decides
	return false, true
}

// areEqualFallback decides equality when == is unavailable: typed
// nils and identity settle the question — nil only equals nil, and
// identity proves equality — and the value's own Equal method stands
// in for ==. When deep, slices left undecided get one level of
// element-wise comparison. Anything else stays unknown — two distinct
// values may still be equal, and deciding that would take a deep
// comparison.
//
//revive:disable-next-line:flag-parameter // deep bounds recursion, not two behaviours
func areEqualFallback(va, vb reflect.Value, deep bool) (is, known bool) {
	if same, ok := isSameTypedNil(va, vb); ok {
		// nil only equals nil
		return same, true
	}

	if isSamePointer(va, vb) {
		// identity proves equality
		return true, true
	}

	if is, known = equalMethod(va, vb); known || !deep {
		return is, known
	}

	return areEqualSlice(va, vb)
}

// areEqualSlice compares two slices of the same type element by
// element. Lengths must match; anything that isn't a slice stays
// unknown.
func areEqualSlice(va, vb reflect.Value) (is, known bool) {
	switch {
	case va.Kind() != reflect.Slice:
		// only slices are walked
		return false, false
	case va.Len() != vb.Len():
		// different lengths are never equal
		return false, true
	default:
		return areEqualSliceElements(va, vb)
	}
}

// areEqualSliceElements walks the element pairs of two equal-length
// slices, aggregating like [AreEqual] does over its list: one unequal
// element settles the pair, an undecided element leaves it unknown
// unless a later element settles it.
func areEqualSliceElements(va, vb reflect.Value) (is, known bool) {
	known = true
	for i := range va.Len() {
		switch is2, known2 := areEqualElement(va.Index(i), vb.Index(i)); {
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
func areEqualElement(va, vb reflect.Value) (is, known bool) {
	a := newComparableValue(unwrapInterface(va))
	b := newComparableValue(unwrapInterface(vb))
	return areEqual2(a, b, false)
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

// comparableValue carries the reflection state of one [AreEqual]
// operand so each value is reflected only once.
type comparableValue struct {
	t reflect.Type
	v reflect.Value
	// ok reports whether v supports direct == comparison via
	// Interface().
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
		ok: v.Comparable() && v.CanInterface(),
	}
}
