package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type panicErrorMethodsTestCase struct {
	name     string
	payload  any
	expected string
}

var panicErrorMethodsTestCases = []panicErrorMethodsTestCase{
	{
		name:     "string payload",
		payload:  "test error",
		expected: "test error",
	},
	{
		name:     "error payload",
		payload:  errors.New("wrapped error"),
		expected: "wrapped error",
	},
	{
		name:     "int payload",
		payload:  42,
		expected: "%!s(int=42)", // Go's %s format for non-string
	},
	{
		name:     "nil payload",
		payload:  nil,
		expected: "%!s(<nil>)", // Go's %s format for nil
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc panicErrorMethodsTestCase) test(t *testing.T) {
	t.Helper()
	pe := NewPanicError(0, tc.payload)

	// Test Error method
	errorStr := pe.Error()
	expectedError := fmt.Sprintf("panic: %s", tc.expected)
	if errorStr != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, errorStr)
	}

	// Test Recovered method
	recovered := pe.Recovered()
	if tc.payload == nil {
		if recovered != nil {
			t.Fatalf("expected nil recovered, got %v", recovered)
		}
	} else {
		// For strings, they get converted to errors in NewPanicError
		if s, ok := tc.payload.(string); ok {
			if err, ok := recovered.(error); ok {
				if err.Error() != s {
					t.Fatalf("expected recovered error '%s', got '%s'", s, err.Error())
				}
			} else {
				t.Fatalf("expected error type for string payload, got %T", recovered)
			}
		} else {
			if recovered != tc.payload {
				t.Fatalf("expected recovered %v, got %v", tc.payload, recovered)
			}
		}
	}

	// Test CallStack method
	stack := pe.CallStack()
	if len(stack) == 0 {
		t.Fatalf("expected non-empty stack trace")
	}
}

type panicErrorUnwrapTestCase struct {
	// Large fields (16 bytes) - string headers and interface
	name          string
	payload       any
	expectedError string

	// Small fields (1 byte) - boolean flags
	expectUnwrap bool
}

var panicErrorUnwrapTestCases = []panicErrorUnwrapTestCase{
	{
		name:          "error payload",
		payload:       errors.New("test error"),
		expectUnwrap:  true,
		expectedError: "test error",
	},
	{
		name:          "string payload converts to error",
		payload:       "string error",
		expectUnwrap:  true,
		expectedError: "string error",
	},
	{
		name:         "non-error payload",
		payload:      42,
		expectUnwrap: false,
	},
	{
		name:         "nil payload",
		payload:      nil,
		expectUnwrap: false,
	},
}

func (tc panicErrorUnwrapTestCase) test(t *testing.T) {
	t.Helper()
	pe := NewPanicError(0, tc.payload)
	unwrapped := pe.Unwrap()

	if tc.expectUnwrap {
		if unwrapped == nil {
			t.Fatalf("expected unwrapped error, got nil")
		}
		if unwrapped.Error() != tc.expectedError {
			t.Fatalf("expected unwrapped error '%s', got '%s'", tc.expectedError, unwrapped.Error())
		}
	} else if unwrapped != nil {
		t.Fatalf("expected nil unwrapped, got %v", unwrapped)
	}
}

type newPanicErrorfTestCase struct {
	expected string
	format   string
	name     string
	args     []any
}

var newPanicErrorfTestCases = []newPanicErrorfTestCase{
	{
		name:     "no args",
		format:   "simple error",
		args:     nil,
		expected: "simple error",
	},
	{
		name:     "with args",
		format:   "error %d: %s",
		args:     []any{42, "test"},
		expected: "error 42: test",
	},
	{
		name:     "with wrapped error",
		format:   "wrapped: %w",
		args:     []any{errors.New("original")},
		expected: "wrapped: original",
	},
}

func (tc newPanicErrorfTestCase) test(t *testing.T) {
	t.Helper()
	pe := NewPanicErrorf(0, tc.format, tc.args...)

	// Test Error method
	errorStr := pe.Error()
	expectedError := fmt.Sprintf("panic: %s", tc.expected)
	if errorStr != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, errorStr)
	}

	// Test that payload is an error
	if _, ok := pe.Recovered().(error); !ok {
		t.Fatalf("expected error payload, got %T", pe.Recovered())
	}
}

func runNewPanicWrapTest(t *testing.T) {
	originalErr := errors.New("original error")
	note := "wrapped note"

	pe := NewPanicWrap(0, originalErr, note)

	// Test that it wraps the error
	unwrapped := pe.Unwrap()
	if unwrapped == nil {
		t.Fatalf("expected unwrapped error, got nil")
	}

	// Test error message contains both note and original
	errorStr := pe.Error()
	if !strings.Contains(errorStr, note) {
		t.Fatalf("expected error to contain note '%s', got '%s'", note, errorStr)
	}
	if !strings.Contains(errorStr, originalErr.Error()) {
		t.Fatalf("expected error to contain original error '%s', got '%s'", originalErr.Error(), errorStr)
	}
}

func runNewPanicWrapfTest(t *testing.T) {
	originalErr := errors.New("original error")
	format := "wrapped %s: %d"
	args := []any{"note", 42}

	pe := NewPanicWrapf(0, originalErr, format, args...)

	// Test that it wraps the error
	unwrapped := pe.Unwrap()
	if unwrapped == nil {
		t.Fatalf("expected unwrapped error, got nil")
	}

	// Test error message contains formatted note and original
	errorStr := pe.Error()
	if !strings.Contains(errorStr, "wrapped note: 42") {
		t.Fatalf("expected error to contain formatted note, got '%s'", errorStr)
	}
	if !strings.Contains(errorStr, originalErr.Error()) {
		t.Fatalf("expected error to contain original error '%s', got '%s'", originalErr.Error(), errorStr)
	}
}

type panicFunctionsTestCase struct {
	expected any
	fn       func()
	name     string
}

var panicFunctionsTestCases = []panicFunctionsTestCase{
	{
		name: "Panic with string",
		fn: func() {
			Panic("test panic")
		},
		expected: "test panic",
	},
	{
		name: "Panic with error",
		fn: func() {
			Panic(errors.New("test error"))
		},
		expected: "test error", // Compare error content as string
	},
	{
		name: "Panicf without args",
		fn: func() {
			Panicf("simple panic")
		},
		expected: "simple panic",
	},
	{
		name: "Panicf with args",
		fn: func() {
			Panicf("panic %d: %s", 42, "test")
		},
		expected: "panic 42: test",
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc panicFunctionsTestCase) test(t *testing.T) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic, got nil")
		}

		pe, ok := r.(*PanicError)
		if !ok {
			t.Fatalf("expected PanicError, got %T", r)
		}

		// Handle payload comparison based on type
		panicValue := pe.Recovered()
		if s, ok := tc.expected.(string); ok {
			if err, ok := panicValue.(error); ok {
				if err.Error() != s {
					t.Fatalf("expected panic payload error '%s', got '%s'", s, err.Error())
				}
			} else {
				t.Fatalf("expected error payload for string, got %T", panicValue)
			}
		} else {
			// For non-string expected values, compare directly
			if panicValue != tc.expected {
				t.Fatalf("expected panic payload %v, got %v", tc.expected, panicValue)
			}
		}
	}()

	tc.fn()
}

type panicWrapFunctionsTestCase struct {
	fn   func()
	name string
}

//revive:disable-next-line:cognitive-complexity
func (tc panicWrapFunctionsTestCase) test(t *testing.T, originalErr error) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic, got nil")
		}

		pe, ok := r.(*PanicError)
		if !ok {
			t.Fatalf("expected PanicError, got %T", r)
		}

		// Test that it unwraps to something that contains the original error
		unwrapped := pe.Unwrap()
		if unwrapped == nil {
			t.Fatalf("expected unwrapped error, got nil")
		}

		// Check if we can find the original error in the chain
		if !errors.Is(unwrapped, originalErr) {
			t.Fatalf("expected to find original error in chain")
		}
	}()

	tc.fn()
}

type newUnreachableErrorTestCase struct {
	name       string
	err        error
	note       string
	expectType string
}

var newUnreachableErrorTestCases = []newUnreachableErrorTestCase{
	{
		name:       "nil error, empty note",
		err:        nil,
		note:       "",
		expectType: "PanicError",
	},
	{
		name:       "nil error, with note",
		err:        nil,
		note:       "test note",
		expectType: "PanicError",
	},
	{
		name:       "ErrUnreachable, with note",
		err:        ErrUnreachable,
		note:       "test note",
		expectType: "PanicError",
	},
	{
		name:       "other error, no note",
		err:        errors.New("other error"),
		note:       "",
		expectType: "PanicError",
	},
	{
		name:       "other error, with note",
		err:        errors.New("other error"),
		note:       "test note",
		expectType: "PanicError",
	},
}

func (tc newUnreachableErrorTestCase) test(t *testing.T) {
	t.Helper()
	result := NewUnreachableError(0, tc.err, tc.note)

	if result == nil {
		t.Fatalf("expected non-nil error, got nil")
	}

	// Test that it's a PanicError
	pe, ok := result.(*PanicError)
	if !ok {
		t.Fatalf("expected PanicError, got %T", result)
	}

	// Test that ErrUnreachable is somewhere in the chain
	if !errors.Is(result, ErrUnreachable) {
		t.Fatalf("expected ErrUnreachable in error chain")
	}

	// Test error message
	errorStr := result.Error()
	if errorStr == "" {
		t.Fatalf("expected non-empty error message")
	}

	// Test stack trace
	stack := pe.CallStack()
	if len(stack) == 0 {
		t.Fatalf("expected non-empty stack trace")
	}
}

func runNewUnreachableErrorfTest(t *testing.T) {
	err := errors.New("test error")
	format := "formatted %s: %d"
	args := []any{"note", 42}

	result := NewUnreachableErrorf(0, err, format, args...)

	if result == nil {
		t.Fatalf("expected non-nil error, got nil")
	}

	// Test that it's a PanicError
	pe, ok := result.(*PanicError)
	if !ok {
		t.Fatalf("expected PanicError, got %T", result)
	}

	// Test that ErrUnreachable is in the chain
	if !errors.Is(result, ErrUnreachable) {
		t.Fatalf("expected ErrUnreachable in error chain")
	}

	// Test that original error is in the chain
	if !errors.Is(result, err) {
		t.Fatalf("expected original error in error chain")
	}

	// Test formatted message
	errorStr := result.Error()
	if !strings.Contains(errorStr, "formatted note: 42") {
		t.Fatalf("expected formatted message in error, got '%s'", errorStr)
	}

	// Test stack trace
	stack := pe.CallStack()
	if len(stack) == 0 {
		t.Fatalf("expected non-empty stack trace")
	}
}

// Main test functions that call the helpers
func TestPanicErrorMethods(t *testing.T) {
	for _, tc := range panicErrorMethodsTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestPanicErrorUnwrap(t *testing.T) {
	for _, tc := range panicErrorUnwrapTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestNewPanicErrorf(t *testing.T) {
	for _, tc := range newPanicErrorfTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestNewPanicWrap(t *testing.T) {
	t.Run("NewPanicWrap", runNewPanicWrapTest)
}

func TestNewPanicWrapf(t *testing.T) {
	t.Run("NewPanicWrapf", runNewPanicWrapfTest)
}

func TestPanicFunctions(t *testing.T) {
	for _, tc := range panicFunctionsTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestPanicWrapFunctions(t *testing.T) {
	originalErr := errors.New("original error")

	testCases := []panicWrapFunctionsTestCase{
		{
			name: "PanicWrap",
			fn: func() {
				PanicWrap(originalErr, "wrap note")
			},
		},
		{
			name: "PanicWrapf",
			fn: func() {
				PanicWrapf(originalErr, "wrap %s: %d", "note", 42)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, originalErr)
		})
	}
}

func TestNewUnreachableError(t *testing.T) {
	for _, tc := range newUnreachableErrorTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestNewUnreachableErrorf(t *testing.T) {
	t.Run("NewUnreachableErrorf", runNewUnreachableErrorfTest)
}
