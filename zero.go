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
//   - Arrays, structs, and other composite types (not handled).
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
// Returns false for unhandled types (arrays, structs, etc.)
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
		a := va.Elem().Interface()
		b := vb.Elem().Interface()
		ok = IsSame(a, b)
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Bool:
		ok = va.Interface() == vb.Interface()
	}
	return ok
}
