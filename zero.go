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

// isReflectValueZero reports whether a reflect.Value is zero.
// This helper reduces code duplication in IsZero.
func isReflectValueZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	return v.IsZero()
}
