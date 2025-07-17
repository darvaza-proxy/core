package core

// S is a helper function for creating test slices in a more concise way.
// It takes variadic arguments and returns a slice of the same type.
// This is particularly useful in table-driven tests where many slice literals are used.
// The function accepts any type, including structs with non-comparable fields.
//
// Example usage:
//
//	S(1, 2, 3)           // []int{1, 2, 3}
//	S("a", "b", "c")     // []string{"a", "b", "c"}
//	S[int]()             // []int{}
//	S[string]()          // []string{}
//	S(testCase{...})     // []testCase{...} (works with any struct)
func S[T any](v ...T) []T {
	if len(v) == 0 {
		return []T{}
	}
	return v
}
