package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

var errMockTFailNow = errors.New("MockT.FailNow")

// Compile-time verification that our types implement the T interface
var (
	_ T = (*testing.T)(nil)
	_ T = (*MockT)(nil)
)

// T is an interface that abstracts the testing functionality we need.
// This allows our testing utilities to work with both *testing.T and mock implementations.
type T interface {
	Helper()
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
}

// MockT is a mock implementation of the T interface for testing purposes.
// It collects error and log messages instead of reporting them to the testing framework.
//
// MockT supports all standard testing methods including Fatal/Fatalf which panic
// with a special error that can be caught by the Run method. This allows testing
// of assertion functions and other utilities that may call Fatal methods.
//
// The Run method executes test functions and recovers from FailNow/Fatal panics,
// making it ideal for testing assertion functions where you need to verify both
// success and failure scenarios without terminating the test runner.
type MockT struct {
	Errors       []string
	Logs         []string
	HelperCalled int
	mu           sync.RWMutex
	failed       bool
}

// Helper implements the T interface and tracks that it was called.
func (m *MockT) Helper() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HelperCalled++
}

// Error implements the T interface and collects error messages.
// It also marks the test as failed.
func (m *MockT) Error(args ...any) {
	msg := fmt.Sprint(args...)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors = append(m.Errors, msg)
	m.failed = true
}

// Errorf implements the T interface and collects formatted error messages.
// It also marks the test as failed.
func (m *MockT) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors = append(m.Errors, msg)
	m.failed = true
}

// Log implements the T interface and collects log messages.
func (m *MockT) Log(args ...any) {
	msg := fmt.Sprint(args...)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Logs = append(m.Logs, msg)
}

// Logf implements the T interface and collects formatted log messages.
func (m *MockT) Logf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Logs = append(m.Logs, msg)
}

// Fatal implements the T interface and collects error messages, then panics.
// It combines Error and FailNow functionality.
func (m *MockT) Fatal(args ...any) {
	msg := fmt.Sprint(args...)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors = append(m.Errors, msg)
	m.failed = true
	panic(errMockTFailNow)
}

// Fatalf implements the T interface and collects formatted error messages, then panics.
// It combines Errorf and FailNow functionality.
func (m *MockT) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors = append(m.Errors, msg)
	m.failed = true
	panic(errMockTFailNow)
}

// Fail implements the T interface and marks the test as failed.
func (m *MockT) Fail() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failed = true
}

// FailNow implements the T interface and marks the test as failed, then panics.
func (m *MockT) FailNow() {
	m.Fail()
	panic(errMockTFailNow)
}

// Failed implements the T interface and returns whether the test has been marked as failed.
func (m *MockT) Failed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failed
}

// HasErrors returns true if any errors were recorded.
func (m *MockT) HasErrors() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Errors) > 0
}

// LastError returns the last error message and whether one was found.
func (m *MockT) LastError() (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.Errors) == 0 {
		return "", false
	}
	return m.Errors[len(m.Errors)-1], true
}

// HasLogs returns true if any log messages were recorded.
func (m *MockT) HasLogs() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Logs) > 0
}

// LastLog returns the last log message and whether one was found.
func (m *MockT) LastLog() (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.Logs) == 0 {
		return "", false
	}
	return m.Logs[len(m.Logs)-1], true
}

// Reset clears all recorded errors, logs, failed state, and helper state.
func (m *MockT) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors = nil
	m.Logs = nil
	m.HelperCalled = 0
	m.failed = false
}

// Run runs the test function f with the MockT instance and returns whether it passed.
// It recovers from FailNow/Fatal panics and returns false if the test failed or panicked.
// Non-FailNow panics are re-thrown. Returns false for nil MockT or nil function.
//
// This method is ideal for testing assertion functions that may call Fatal/FailNow:
//
//	mock := &MockT{}
//	ok := mock.Run("test assertion", func(t T) {
//		AssertEqual(t, 1, 2, "should fail") // This calls t.Fatal internally
//	})
//	// ok == false, mock.Failed() == true, mock.Errors contains failure message
//
// Unlike testing.T.Run, this method uses the same MockT instance throughout,
// allowing inspection of all collected errors, logs, and failure state after execution.
func (m *MockT) Run(_ string, f func(T)) (ok bool) {
	if m == nil || f == nil {
		return false
	}

	defer func() {
		if r := recover(); r != nil && r != errMockTFailNow {
			// Re-panic if it's not our FailNow error
			panic(r)
		}
	}()

	f(m)

	return !m.Failed()
}

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

// TestCase represents a test case that can be executed.
// This interface is used by RunTestCases to standardize test case execution.
type TestCase interface {
	Name() string
	Test(t *testing.T)
}

// RunTestCases runs a slice of test cases that implement the TestCase interface.
// This eliminates the boilerplate of looping through test cases and calling t.Run.
//
// Example usage:
//
//	RunTestCases(t, testCases)
func RunTestCases[T TestCase](t *testing.T, cases []T) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.Name(), func(t *testing.T) {
			tc.Test(t)
		})
	}
}

// AssertEqual compares two values and reports differences.
// This is a generic helper that works with any comparable type.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertEqual(t, 42, result, "result value")
//	AssertEqual(t, "hello", str, "string %d comparison", 1)
func AssertEqual[U comparable](t T, expected, actual U, name string, args ...any) bool {
	t.Helper()
	ok := actual == expected
	if !ok {
		doError(t, name, args, "expected %v, got %v", expected, actual)
	} else {
		doLog(t, name, args, "%v", actual)
	}
	return ok
}

// AssertNotEqual compares two values and ensures they are different.
// This is a generic helper that works with any comparable type.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertNotEqual(t, 42, result, "result value")
//	AssertNotEqual(t, "hello", str, "string %d comparison", 1)
func AssertNotEqual[U comparable](t T, expected, actual U, name string, args ...any) bool {
	t.Helper()
	ok := actual != expected
	if !ok {
		doError(t, name, args, "expected not %v, got %v", expected, actual)
	} else {
		doLog(t, name, args, "%v", actual)
	}
	return ok
}

// AssertSliceEqual compares two slices and reports differences.
// This uses reflect.DeepEqual for comprehensive comparison.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertSliceEqual(t, S(1, 2, 3), result, "result slice")
//	AssertSliceEqual(t, S("a", "b"), strings, "string slice %s", "test")
func AssertSliceEqual[U any](t T, expected, actual []U, name string, args ...any) bool {
	t.Helper()
	ok := reflect.DeepEqual(expected, actual)
	if !ok {
		doError(t, name, args, "expected %v, got %v", expected, actual)
	} else {
		doLog(t, name, args, "%v", actual)
	}
	return ok
}

// AssertContains fails the test if the string does not contain the substring.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertContains(t, "hello world", "world", "substring check")
//	AssertContains(t, output, "success", "command output for %s", cmd)
func AssertContains(t T, s, substr, name string, args ...any) bool {
	t.Helper()
	ok := strings.Contains(s, substr)
	if !ok {
		doError(t, name, args, "expected %q to contain %q", s, substr)
	} else {
		doLog(t, name, args, "contains %q", substr)
	}
	return ok
}

// AssertError fails the test if error is nil.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertError(t, err, "parse error")
//	AssertError(t, err, "operation %s", "save")
func AssertError(t T, err error, name string, args ...any) bool {
	t.Helper()
	return AssertNotNil(t, err, name, args...)
}

// AssertNoError fails the test if error is not nil.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertNoError(t, err, "initialization")
//	AssertNoError(t, err, "loading %s", filename)
func AssertNoError(t T, err error, name string, args ...any) bool {
	t.Helper()
	return AssertNil(t, err, name, args...)
}

// AssertPanic runs a function expecting it to panic and optionally validates the panic value.
// This standardizes panic testing patterns.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertPanic(t, func() { someFunctionThatPanics() }, nil, "panic test")
//	AssertPanic(t, func() { divide(1, 0) }, "division by zero", "divide %d by zero", 1)
func AssertPanic(t T, fn func(), expectedPanic any, name string, args ...any) (ok bool) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			doError(t, name, args, "expected panic but got nil")
			return
		}
		if expectedPanic != nil && !reflect.DeepEqual(r, expectedPanic) {
			doError(t, name, args, "expected panic %v, got %v", expectedPanic, r)
			return
		}
		ok = true
		doLog(t, name, args, "%v", r)
	}()
	fn()
	return ok
}

// AssertNoPanic runs a function expecting it not to panic.
// This is useful for testing that functions handle edge cases gracefully.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertNoPanic(t, func() { safeFunction() }, "safe function")
//	AssertNoPanic(t, func() { handleNilInput(nil) }, "nil input %s", "handling")
func AssertNoPanic(t T, fn func(), name string, args ...any) (ok bool) {
	t.Helper()
	ok = true
	defer func() {
		if r := recover(); r != nil {
			doError(t, name, args, "expected no panic but got: %v", r)
			ok = false
			return
		}
		doLog(t, name, args, "%v", "no panic")
	}()
	fn()
	return ok
}

// AssertTrue fails the test if value is not true.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertTrue(t, result, "operation succeeded")
//	AssertTrue(t, isValid, "validation for %s", field)
//
// revive:disable-next-line:flag-parameter
func AssertTrue(t T, value bool, name string, args ...any) bool {
	t.Helper()
	return AssertEqual(t, true, value, name, args...)
}

// AssertFalse fails the test if value is not false.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertFalse(t, hasError, "no errors expected")
//	AssertFalse(t, isEmpty, "container %s should not be empty", name)
//
// revive:disable-next-line:flag-parameter
func AssertFalse(t T, value bool, name string, args ...any) bool {
	t.Helper()
	return AssertEqual(t, false, value, name, args...)
}

// AssertErrorIs fails the test if the error does not match the target error.
// Uses errors.Is to check if the error matches.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertErrorIs(t, err, ErrNotFound, "lookup error")
//	AssertErrorIs(t, err, ErrInvalid, "validation for %s", field)
func AssertErrorIs(t T, err, target error, name string, args ...any) bool {
	t.Helper()
	ok := errors.Is(err, target)
	if !ok {
		doError(t, name, args, "expected error %v, got %v", target, err)
	} else {
		doLog(t, name, args, "%v", err)
	}
	return ok
}

// AssertTypeIs fails the test if value is not of the expected type.
// It returns the value cast to the expected type and a boolean indicating success.
// The name parameter can include printf-style formatting.
//
// Example usage:
//
//	val, ok := AssertTypeIs[*MyError](t, err, "error type")
//	config, ok := AssertTypeIs[*Config](t, result, "config type for %s", name)
func AssertTypeIs[U any](t T, value any, name string, args ...any) (U, bool) {
	t.Helper()
	result, ok := value.(U)
	if !ok {
		doError(t, name, args, "expected type %T, got %T", result, value)
	} else {
		doLog(t, name, args, "%T", value)
	}
	return result, ok
}

// AssertNil asserts that a value is nil.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertNil(t, err, "error should be nil")
//	AssertNil(t, ptr, "pointer %s should be nil", ptrName)
func AssertNil(t T, value any, name string, args ...any) bool {
	t.Helper()
	ok := IsNil(value)
	if !ok {
		doError(t, name, args, "expected nil, got %v", value)
	} else {
		doLog(t, name, args, "%v", value)
	}
	return ok
}

// AssertNotNil asserts that a value is not nil.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertNotNil(t, result, "result should not be nil")
//	AssertNotNil(t, m, "map %s should not be nil", mapName)
func AssertNotNil(t T, value any, name string, args ...any) bool {
	t.Helper()
	ok := !IsNil(value)
	if !ok {
		doError(t, name, args, "expected non-nil value, got nil")
	} else {
		doLog(t, name, args, "%v", value)
	}
	return ok
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
//	AssertNoError(t, err, "concurrent test should not fail")
func RunConcurrentTest(t T, numWorkers int, worker func(int) error) error {
	t.Helper()
	errCh := make(chan error, numWorkers)

	runWorkers(numWorkers, worker, errCh)
	return collectErrors(errCh)
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
	for range b.N {
		fn(data)
	}
}

// doMessage formats a message with optional prefix and sends it using the provided function
// revive:disable-next-line:argument-limit
func doMessage(
	t T, outputFunc func(...any), prefixFormat string, prefixArgs []any,
	messageFormat string, messageArgs ...any,
) {
	t.Helper()

	// Format the message
	msg := fmt.Sprintf(messageFormat, messageArgs...)

	// Add prefix if provided
	if prefixFormat != "" {
		var prefix string
		if len(prefixArgs) > 0 {
			prefix = fmt.Sprintf(prefixFormat, prefixArgs...)
		} else {
			prefix = prefixFormat
		}
		msg = fmt.Sprintf("%s: %s", prefix, msg)
	}

	outputFunc(msg)
}

// doError reports a test error with optional prefix formatting
func doError(t T, prefixFormat string, prefixArgs []any, messageFormat string, messageArgs ...any) {
	doMessage(t, t.Error, prefixFormat, prefixArgs, messageFormat, messageArgs...)
}

// doLog reports a test log message with optional prefix formatting
// revive:disable-next-line:argument-limit
func doLog(t T, prefixFormat string, prefixArgs []any, messageFormat string, messageArgs ...any) {
	doMessage(t, t.Log, prefixFormat, prefixArgs, messageFormat, messageArgs...)
}

func runWorkers(numWorkers int, worker func(int) error, errCh chan error) {
	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := worker(id); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
}

func collectErrors(errCh chan error) error {
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
