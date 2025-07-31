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

func (tc panicErrorMethodsTestCase) test(t *testing.T) {
	t.Helper()
	pe := NewPanicError(0, tc.payload)

	tc.testErrorMethod(t, pe)
	tc.testRecoveredMethod(t, pe)
	tc.testCallStackMethod(t, pe)
}

func (tc panicErrorMethodsTestCase) testErrorMethod(t *testing.T, pe *PanicError) {
	t.Helper()
	errorStr := pe.Error()
	expectedError := fmt.Sprintf("panic: %s", tc.expected)
	AssertEqual(t, expectedError, errorStr, "error message")
}

func (tc panicErrorMethodsTestCase) testRecoveredMethod(t *testing.T, pe *PanicError) {
	t.Helper()
	recovered := pe.Recovered()
	if tc.payload == nil {
		AssertNil(t, recovered, "nil payload recovered")
	} else {
		tc.validateRecoveredPayload(t, recovered)
	}
}

func (tc panicErrorMethodsTestCase) validateRecoveredPayload(t *testing.T, recovered any) {
	t.Helper()
	// For strings, they get converted to errors in NewPanicError
	if s, ok := tc.payload.(string); ok {
		tc.validateStringPayload(t, recovered, s)
	} else {
		AssertEqual(t, tc.payload, recovered, "recovered payload")
	}
}

func (panicErrorMethodsTestCase) validateStringPayload(t *testing.T, recovered any, expectedStr string) {
	t.Helper()
	if err, ok := recovered.(error); ok {
		AssertEqual(t, expectedStr, err.Error(), "string payload error")
	} else {
		t.Fatalf("expected error type for string payload, got %T", recovered)
	}
}

func (panicErrorMethodsTestCase) testCallStackMethod(t *testing.T, pe *PanicError) {
	t.Helper()
	stack := pe.CallStack()
	AssertTrue(t, len(stack) > 0, "non-empty stack trace")
}

type panicErrorUnwrapTestCase struct {
	// Large fields - string headers and interface
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
		args:     S[any](42, "test"),
		expected: "error 42: test",
	},
	{
		name:     "with wrapped error",
		format:   "wrapped: %w",
		args:     S[any](errors.New("original")),
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
	args := S[any]("note", 42)

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

func (tc panicFunctionsTestCase) test(t *testing.T) {
	t.Helper()
	defer func() {
		r := recover()
		tc.validatePanicFunction(t, r)
	}()
	tc.fn()
}

func (tc panicFunctionsTestCase) validatePanicFunction(t *testing.T, r any) {
	t.Helper()
	AssertNotNil(t, r, "panic occurred")

	pe := tc.extractPanicError(t, r)
	tc.validatePanicPayload(t, pe)
}

func (panicFunctionsTestCase) extractPanicError(t *testing.T, r any) *PanicError {
	t.Helper()
	pe, ok := r.(*PanicError)
	if !ok {
		t.Fatalf("expected PanicError, got %T", r)
	}
	return pe
}

func (tc panicFunctionsTestCase) validatePanicPayload(t *testing.T, pe *PanicError) {
	t.Helper()
	panicValue := pe.Recovered()
	if s, ok := tc.expected.(string); ok {
		tc.validateStringPayload(t, panicValue, s)
	} else {
		AssertEqual(t, tc.expected, panicValue, "panic payload")
	}
}

func (panicFunctionsTestCase) validateStringPayload(t *testing.T, panicValue any, expectedStr string) {
	t.Helper()
	if err, ok := panicValue.(error); ok {
		AssertEqual(t, expectedStr, err.Error(), "panic error message")
	} else {
		t.Fatalf("expected error payload for string, got %T", panicValue)
	}
}

type panicWrapFunctionsTestCase struct {
	fn   func()
	name string
}

func (tc panicWrapFunctionsTestCase) test(t *testing.T, originalErr error) {
	t.Helper()
	defer func() {
		r := recover()
		tc.validateWrapFunction(t, r, originalErr)
	}()
	tc.fn()
}

func (tc panicWrapFunctionsTestCase) validateWrapFunction(t *testing.T, r any, originalErr error) {
	t.Helper()
	AssertNotNil(t, r, "panic occurred")

	pe := tc.extractWrapPanicError(t, r)
	tc.validateUnwrapChain(t, pe, originalErr)
}

func (panicWrapFunctionsTestCase) extractWrapPanicError(t *testing.T, r any) *PanicError {
	t.Helper()
	pe, ok := r.(*PanicError)
	if !ok {
		t.Fatalf("expected PanicError, got %T", r)
	}
	return pe
}

func (panicWrapFunctionsTestCase) validateUnwrapChain(t *testing.T, pe *PanicError, originalErr error) {
	t.Helper()
	unwrapped := pe.Unwrap()
	AssertNotNil(t, unwrapped, "unwrapped error")
	AssertTrue(t, errors.Is(unwrapped, originalErr), "original error in chain")
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
	args := S[any]("note", 42)

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
