package core

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

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

// extractMsgAndArgs safely extracts format string and arguments from msgAndArgs.
// Returns format string, remaining args, and ok flag indicating success.
// This prevents panic from unsafe type assertions when msgAndArgs[0] is not a string.
//
// Example usage:
//
//	format, args, ok := extractMsgAndArgs(msgAndArgs)
//	if ok {
//		msg = fmt.Sprintf(format, args...)
//	}
func extractMsgAndArgs(msgAndArgs []any) (format string, args []any, ok bool) {
	if len(msgAndArgs) == 0 {
		return "", nil, false
	}

	format, ok = msgAndArgs[0].(string)
	if !ok {
		return "", nil, false
	}

	return format, msgAndArgs[1:], true
}

// TestCase represents a test case that can be executed.
type TestCase interface {
	Name() string
	Test(t *testing.T)
}

// RunTestCases runs a slice of test cases that implement the TestCase interface.
// This eliminates the boilerplate of looping through test cases and calling t.Run.
//
// Example usage:
//
//	RunTestCases(t, []TestCase{tc1, tc2, tc3})
func RunTestCases(t *testing.T, cases []TestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.Name(), func(t *testing.T) {
			tc.Test(t)
		})
	}
}

// AssertEqual compares two values and reports differences.
// This is a generic helper that works with any comparable type.
//
// Example usage:
//
//	AssertEqual(t, 42, result)
//	AssertEqual(t, "hello", str, "string comparison failed")
func AssertEqual[T comparable](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if actual != expected {
		msg := fmt.Sprintf("values not equal: expected %v, got %v", expected, actual)
		if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
			msg = fmt.Sprintf(format, args...) + ": " + msg
		}
		t.Error(msg)
	}
}

// AssertSliceEqual compares two slices and reports differences.
// This uses reflect.DeepEqual for comprehensive comparison.
//
// Example usage:
//
//	AssertSliceEqual(t, S(1, 2, 3), result)
//	AssertSliceEqual(t, S("a", "b"), strings, "string slice comparison")
func AssertSliceEqual[T any](t *testing.T, expected, actual []T, msgAndArgs ...any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := fmt.Sprintf("slices not equal: expected %v, got %v", expected, actual)
		if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
			msg = fmt.Sprintf(format, args...) + ": " + msg
		}
		t.Error(msg)
	}
}

// AssertError checks error expectations.
// This standardizes error checking patterns across tests.
//
// Example usage:
//
//	AssertError(t, err, true)   // expect error
//	AssertError(t, err, false)  // expect no error
//	AssertError(t, err, true, "operation should fail")
//
// revive:disable-next-line:flag-parameter
func AssertError(t *testing.T, err error, expectError bool, msgAndArgs ...any) {
	t.Helper()
	if expectError {
		assertExpectedError(t, err, msgAndArgs...)
		return
	}
	assertNoError(t, err, msgAndArgs...)
}

func assertExpectedError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		var msg string
		if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
			msg = fmt.Sprintf(format, args...)
		} else {
			msg = "expected error but got nil"
		}
		t.Error(msg)
	}
}

func assertNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		msg := "expected no error but got: %v"
		if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
			msg = fmt.Sprintf(format, args...) + ": %v"
		}
		t.Errorf(msg, err)
	}
}

// AssertPanic runs a function expecting it to panic and optionally validates the panic value.
// This standardizes panic testing patterns.
//
// Example usage:
//
//	AssertPanic(t, func() { someFunctionThatPanics() }, nil)
//	AssertPanic(t, func() { divide(1, 0) }, "division by zero")
func AssertPanic(t *testing.T, fn func(), expectedPanic any, msgAndArgs ...any) {
	t.Helper()
	defer func() {
		r := recover()
		validatePanicRecovery(t, r, expectedPanic, msgAndArgs...)
	}()
	fn()
}

func validatePanicRecovery(t *testing.T, recovered, expectedPanic any, msgAndArgs ...any) {
	t.Helper()
	if recovered == nil {
		var msg string
		if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
			msg = fmt.Sprintf(format, args...)
		} else {
			msg = "expected panic but got nil"
		}
		t.Error(msg)
		return
	}
	if expectedPanic != nil && !reflect.DeepEqual(recovered, expectedPanic) {
		t.Errorf("expected panic %v, got %v", expectedPanic, recovered)
	}
}

// AssertNoPanic runs a function expecting it not to panic.
// This is useful for testing that functions handle edge cases gracefully.
//
// Example usage:
//
//	AssertNoPanic(t, func() { safeFunction() })
//	AssertNoPanic(t, func() { handleNilInput(nil) }, "nil input handling")
func AssertNoPanic(t *testing.T, fn func(), msgAndArgs ...any) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("expected no panic but got: %v", r)
			if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
				msg = fmt.Sprintf(format, args...) + ": " + msg
			}
			t.Error(msg)
		}
	}()
	fn()
}

// AssertBool checks boolean expectations with custom messages.
// This is a specialized helper for boolean assertions.
//
// Example usage:
//
//	AssertBool(t, result, true, "operation should succeed")
//	AssertBool(t, isEmpty, false, "container should not be empty")
//
// revive:disable-next-line:flag-parameter
func AssertBool(t *testing.T, actual, expected bool, msgAndArgs ...any) {
	t.Helper()
	if actual != expected {
		msg := fmt.Sprintf("boolean assertion failed: expected %v, got %v", expected, actual)
		if format, args, ok := extractMsgAndArgs(msgAndArgs); ok {
			msg = fmt.Sprintf(format, args...) + ": " + msg
		}
		t.Error(msg)
	}
}

// RunConcurrentTest runs multiple goroutines and waits for completion.
// This standardizes concurrent testing patterns.
//
// Example usage:
//
//	err := RunConcurrentTest(t, 10, func(id int) error {
//		// worker logic here
//		return nil
//	})
//	AssertError(t, err, false, "concurrent test should not fail")
func RunConcurrentTest(t *testing.T, numWorkers int, worker func(int) error) error {
	t.Helper()
	errors := make(chan error, numWorkers)

	runWorkers(numWorkers, worker, errors)
	return collectErrors(errors)
}

func runWorkers(numWorkers int, worker func(int) error, errors chan error) {
	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := worker(id); err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	close(errors)
}

func collectErrors(errors chan error) error {
	for err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

// RunBenchmark runs a benchmark with setup and execution phases.
// This standardizes benchmark patterns with proper timer management.
//
// Example usage:
//
//	RunBenchmark(b, func() interface{} {
//		return setupTestData()
//	}, func(data interface{}) {
//		processData(data)
//	})
func RunBenchmark(b *testing.B, setup func() any, fn func(any)) {
	b.Helper()
	data := setup()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(data)
	}
}
