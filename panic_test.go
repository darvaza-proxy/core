package core

import (
	"errors"
	"fmt"
	"testing"
)

type asRecoveredTestCase struct {
	// Large fields - interface types and strings
	expected any
	input    any
	name     string

	// Small fields (1 byte) - booleans
	isNil bool
}

var asRecoveredTestCases = []asRecoveredTestCase{
	{
		name:     "nil input",
		input:    nil,
		expected: nil,
		isNil:    true,
	},
	{
		name:     "string panic",
		input:    "test panic",
		expected: "test panic",
		isNil:    false,
	},
	{
		name:     "error panic",
		input:    errors.New("test error"),
		expected: "test error", // String comparison for error content
		isNil:    false,
	},
	{
		name:     "int panic",
		input:    42,
		expected: 42,
		isNil:    false,
	},
	{
		name:     "already recovered",
		input:    NewPanicError(1, "already wrapped"),
		expected: "already wrapped",
		isNil:    false,
	},
}

func (tc asRecoveredTestCase) test(t *testing.T) {
	t.Helper()
	result := AsRecovered(tc.input)

	if tc.isNil {
		tc.checkNilResult(t, result)
		return
	}

	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	recovered := result.Recovered()
	tc.checkRecoveredValue(t, recovered)
	tc.checkErrorString(t, result)
}

func (asRecoveredTestCase) checkNilResult(t *testing.T, result Recovered) {
	t.Helper()
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func (tc asRecoveredTestCase) checkRecoveredValue(t *testing.T, recovered any) {
	t.Helper()
	switch exp := tc.expected.(type) {
	case string:
		tc.checkStringValue(t, recovered, exp)
	case error:
		tc.checkErrorValue(t, recovered, exp)
	default:
		AssertEqual(t, tc.expected, recovered, "recovered value")
	}
}

func (asRecoveredTestCase) checkStringValue(t *testing.T, recovered any, exp string) {
	t.Helper()
	if err, ok := recovered.(error); ok {
		AssertEqual(t, exp, err.Error(), "error message")
	} else if recovered != exp {
		t.Fatalf("expected recovered value %v, got %v", exp, recovered)
	}
}

func (asRecoveredTestCase) checkErrorValue(t *testing.T, recovered any, exp error) {
	t.Helper()
	if err, ok := recovered.(error); ok {
		AssertEqual(t, exp.Error(), err.Error(), "error message")
	} else {
		t.Fatalf("expected error type, got %T", recovered)
	}
}

func (asRecoveredTestCase) checkErrorString(t *testing.T, result Recovered) {
	t.Helper()
	errorStr := result.Error()
	if errorStr == "" {
		t.Fatalf("expected non-empty error string")
	}
}

func TestAsRecovered(t *testing.T) {
	for _, tc := range asRecoveredTestCases {
		t.Run(tc.name, tc.test)
	}
}

type catcherDoTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

var catcherDoTestCases = []catcherDoTestCase{
	{
		name: "successful function",
		fn: func() error {
			return nil
		},
		expectError: false,
		expectPanic: false,
	},
	{
		name: "function returns error",
		fn: func() error {
			return errors.New("test error")
		},
		expectError: true,
		expectPanic: false,
	},
	{
		name: "function panics with string",
		fn: func() error {
			panic("test panic")
		},
		expectError: true,
		expectPanic: true,
	},
	{
		name: "function panics with error",
		fn: func() error {
			panic(errors.New("panic error"))
		},
		expectError: true,
		expectPanic: true,
	},
	{
		name: "function panics with int",
		fn: func() error {
			panic(42)
		},
		expectError: true,
		expectPanic: true,
	},
	{
		name:        "nil function",
		fn:          nil,
		expectError: false,
		expectPanic: false,
	},
}

func (tc catcherDoTestCase) test(t *testing.T) {
	t.Helper()
	var catcher Catcher
	err := catcher.Do(tc.fn)

	tc.checkError(t, err)
	tc.checkPanic(t, err)
}

func (tc catcherDoTestCase) checkError(t *testing.T, err error) {
	t.Helper()
	if tc.expectError {
		AssertError(t, err, "Catcher.Do")
	} else {
		AssertNoError(t, err, "Catcher.Do")
	}
}

func (tc catcherDoTestCase) checkPanic(t *testing.T, err error) {
	t.Helper()
	if !tc.expectError || !tc.expectPanic {
		return
	}

	recovered, ok := err.(Recovered)
	if !ok {
		t.Fatalf("expected Recovered error, got %T", err)
	}

	if recovered.Recovered() == nil {
		t.Fatalf("expected recovered panic value, got nil")
	}
}

func TestCatcherDo(t *testing.T) {
	for _, tc := range catcherDoTestCases {
		t.Run(tc.name, tc.test)
	}
}

type catcherTryTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

var catcherTryTestCases = []catcherTryTestCase{
	{
		name: "successful function",
		fn: func() error {
			return nil
		},
		expectError: false,
		expectPanic: false,
	},
	{
		name: "function returns error",
		fn: func() error {
			return errors.New("test error")
		},
		expectError: true,
		expectPanic: false,
	},
	{
		name: "function panics",
		fn: func() error {
			panic("test panic")
		},
		expectError: false,
		expectPanic: true,
	},
	{
		name:        "nil function",
		fn:          nil,
		expectError: false,
		expectPanic: false,
	},
}

func (tc catcherTryTestCase) test(t *testing.T) {
	t.Helper()
	var catcher Catcher
	err := catcher.Try(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catcher.Try")
	} else {
		AssertNoError(t, err, "Catcher.Try")
	}

	// Check recovered panic
	recovered := catcher.Recovered()
	if tc.expectPanic {
		if recovered == nil {
			t.Fatalf("expected recovered panic, got nil")
		}
	} else if recovered != nil {
		t.Fatalf("expected no recovered panic, got %v", recovered)
	}
}

func TestCatcherTry(t *testing.T) {
	for _, tc := range catcherTryTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestCatcherRecovered(t *testing.T) {
	var catcher Catcher

	// Initially no panic
	if recovered := catcher.Recovered(); recovered != nil {
		t.Fatalf("expected nil recovered, got %v", recovered)
	}

	// After panic
	_ = catcher.Try(func() error {
		panic("test panic")
	})

	recovered := catcher.Recovered()
	if recovered == nil {
		t.Fatalf("expected recovered panic, got nil")
	}

	// String panics get converted to errors by NewPanicError
	if err, ok := recovered.Recovered().(error); ok {
		AssertEqual(t, "test panic", err.Error(), "error message")
	} else {
		t.Fatalf("expected error type for string panic, got %T", recovered.Recovered())
	}
}

func TestCatcherConcurrent(t *testing.T) {
	var catcher Catcher

	// Use a channel to coordinate goroutines
	done := make(chan bool, 2)

	// Test that only the first panic is stored
	go func() {
		_ = catcher.Try(func() error {
			panic("first panic")
		})
		done <- true
	}()

	go func() {
		_ = catcher.Try(func() error {
			panic("second panic")
		})
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	recovered := catcher.Recovered()
	if recovered == nil {
		t.Fatalf("expected recovered panic, got nil")
	}

	// Should be either "first panic" or "second panic" (converted to errors)
	panicValue := recovered.Recovered()
	if err, ok := panicValue.(error); ok {
		errorStr := err.Error()
		if errorStr != "first panic" && errorStr != "second panic" {
			t.Fatalf("unexpected panic value: %v", errorStr)
		}
	} else {
		t.Fatalf("expected error type for string panic, got %T", panicValue)
	}
}

type catchTestCase struct {
	fn          func() error
	name        string
	expectError bool
}

var catchTestCases = []catchTestCase{
	{
		name: "successful function",
		fn: func() error {
			return nil
		},
		expectError: false,
	},
	{
		name: "function returns error",
		fn: func() error {
			return errors.New("test error")
		},
		expectError: true,
	},
	{
		name: "function panics",
		fn: func() error {
			panic("test panic")
		},
		expectError: true,
	},
}

func (tc catchTestCase) test(t *testing.T) {
	t.Helper()
	err := Catch(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catch")
	} else {
		AssertNoError(t, err, "Catch")
	}
}

func TestCatch(t *testing.T) {
	for _, tc := range catchTestCases {
		t.Run(tc.name, tc.test)
	}
}

type catchWithPanicRecoveryTestCase struct {
	value any
	name  string
}

var catchWithPanicRecoveryTestCases = []catchWithPanicRecoveryTestCase{
	{name: "string panic", value: "string panic"},
	{name: "int panic", value: 42},
	{name: "float panic", value: 3.14},
	{name: "error panic", value: errors.New("error panic")},
	{name: "formatted error", value: errors.New("formatted error")},
	// Skip slice and map as they are not comparable
}

func (tc catchWithPanicRecoveryTestCase) test(t *testing.T) {
	t.Helper()
	err := Catch(func() error {
		panic(tc.value)
	})

	AssertError(t, err, "panic recovery")

	if recovered, ok := err.(Recovered); ok {
		panicValue := recovered.Recovered()

		// Handle string conversion to error by NewPanicError
		if s, ok := tc.value.(string); ok {
			if err, ok := panicValue.(error); ok {
				AssertEqual(t, s, err.Error(), "error message")
			} else {
				t.Fatalf("expected error type for string panic, got %T", panicValue)
			}
		} else {
			AssertEqual(t, tc.value, panicValue, "panic value")
		}
	} else {
		t.Fatalf("expected Recovered error, got %T", err)
	}
}

func TestCatchWithPanicRecovery(t *testing.T) {
	for _, tc := range catchWithPanicRecoveryTestCases {
		t.Run(tc.name, tc.test)
	}
}

// testMust is a helper to test Must function by catching panics.
// It wraps Must calls in panic recovery to allow testing both success
// and panic scenarios. Returns the value and any recovered panic as an error.
func testMust[T any](v0 T, e0 error) (v1 T, e1 error) {
	defer func() {
		if e2 := AsRecovered(recover()); e2 != nil {
			e1 = e2
		}
	}()

	v1 = Must(v0, e0)
	return v1, nil
}

// mustSuccessTestCase tests Must function success scenarios where no panic should occur.
type mustSuccessTestCase struct {
	// Large fields first - interfaces (8 bytes on 64-bit)
	value any
	err   error

	// Small fields last - strings (16 bytes on 64-bit)
	name string
}

// newMustSuccessTestCase creates a new mustSuccessTestCase with the given parameters.
// For success cases, err is always nil.
func newMustSuccessTestCase(name string, value any) mustSuccessTestCase {
	return mustSuccessTestCase{
		value: value,
		err:   nil,
		name:  name,
	}
}

// test validates that Must returns the value unchanged when err is nil.
func (tc mustSuccessTestCase) test(t *testing.T) {
	t.Helper()

	tc.testMustWithValue(t)
}

// testMustT is a generic test helper for Must function with comparable types.
// It handles the common pattern of testing Must with a value and verifying
// the result matches expectations.
func testMustT[V comparable](t *testing.T, tc mustSuccessTestCase, value V) {
	t.Helper()

	got, err := testMust(value, tc.err)
	AssertNoError(t, err, "Must success")
	AssertEqual(t, value, got, "Must value")
}

// testMustSlice is a specialized test helper for Must function with slice types.
func testMustSlice[V any](t *testing.T, tc mustSuccessTestCase, value []V) {
	t.Helper()

	got, err := testMust(value, tc.err)
	AssertNoError(t, err, "Must success")
	AssertSliceEqual(t, value, got, "Must slice")
}

// testMustWithValue dispatches to the appropriate test helper.
func (tc mustSuccessTestCase) testMustWithValue(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		testMustT(t, tc, v)
	case int:
		testMustT(t, tc, v)
	case bool:
		testMustT(t, tc, v)
	case []int:
		testMustSlice(t, tc, v)
	case *int:
		testMustT(t, tc, v)
	case struct{ Name string }:
		testMustT(t, tc, v)
	default:
		t.Fatalf("unsupported test value type: %T", tc.value)
	}
}

func TestMustSuccess(t *testing.T) {
	testCases := []mustSuccessTestCase{
		newMustSuccessTestCase("string success", "hello"),
		newMustSuccessTestCase("int success", 42),
		newMustSuccessTestCase("bool success", true),
		newMustSuccessTestCase("slice success", S(1, 2, 3)),
		newMustSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// mustPanicTestCase tests Must function panic scenarios where Must should panic.
type mustPanicTestCase struct {
	// Large fields first - error interface (8 bytes)
	err error

	// Small fields last - string (16 bytes)
	name string
}

// test validates that Must panics with proper PanicError when err is not nil.
func (tc mustPanicTestCase) test(t *testing.T) {
	t.Helper()

	_, err := testMust("value", tc.err)
	AssertError(t, err, "Must panic")

	// Verify the panic contains our original error
	AssertTrue(t, errors.Is(err, tc.err), "panic wraps original")

	// Verify it's a proper PanicError
	panicErr, ok := AssertTypeIs[*PanicError](t, err, "panic type")
	if ok {
		// Verify stack trace exists
		stack := panicErr.CallStack()
		AssertTrue(t, len(stack) > 0, "has stack trace")
	}
}

func TestMustPanic(t *testing.T) {
	testCases := []mustPanicTestCase{
		{
			name: "simple error",
			err:  errors.New("test error"),
		},
		{
			name: "formatted error",
			err:  fmt.Errorf("formatted error: %d", 42),
		},
		{
			name: "wrapped error",
			err:  fmt.Errorf("wrapped: %w", errors.New("inner")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

type maybeTestCase struct {
	// Large fields first - interfaces (8 bytes)
	value any
	err   error

	// Small fields last - string (16 bytes)
	name string
}

func (tc maybeTestCase) test(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe string")
	case int:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe int")
	case *int:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe pointer")
	case struct{ Name string }:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe struct")
	default:
		t.Fatalf("unsupported test value type: %T", tc.value)
	}
}

func TestMaybe(t *testing.T) {
	testCases := []maybeTestCase{
		{
			name:  "string with nil error",
			value: "hello",
			err:   nil,
		},
		{
			name:  "string with error",
			value: "world",
			err:   errors.New("ignored error"),
		},
		{
			name:  "int with nil error",
			value: 42,
			err:   nil,
		},
		{
			name:  "int with error",
			value: 0,
			err:   errors.New("another ignored error"),
		},
		{
			name:  "nil pointer with error",
			value: (*int)(nil),
			err:   errors.New("pointer error"),
		},
		{
			name:  "struct with error",
			value: struct{ Name string }{"test"},
			err:   fmt.Errorf("formatted: %d", 123),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}
