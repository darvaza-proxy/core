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
	switch p := vi.(type) {
	case nil:
		// untyped nil
		return true
	case reflect.Value:
		// Special handling for reflect.Value to avoid double-wrapping
		return isReflectValueNil(p)
	default:
		// Use reflection to check for nil
		v := reflect.ValueOf(vi)
		return isReflectValueNil(v)
	}
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
