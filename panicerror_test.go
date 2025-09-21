package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestCase interface validations
var _ TestCase = panicErrorMethodsTestCase{}
var _ TestCase = panicErrorUnwrapTestCase{}
var _ TestCase = newPanicErrorfTestCase{}
var _ TestCase = panicFunctionsTestCase{}
var _ TestCase = panicWrapFunctionsTestCase{}
var _ TestCase = newUnreachableErrorTestCase{}

type panicErrorMethodsTestCase struct {
	name     string
	payload  any
	expected string
}

func makePanicErrorMethodsTestCases() []TestCase {
	return S(
		newPanicErrorMethodsTestCase("string payload", "test error", "test error"),
		newPanicErrorMethodsTestCase("error payload", errors.New("wrapped error"), "wrapped error"),
		newPanicErrorMethodsTestCase("int payload", 42, "%!s(int=42)"),
		newPanicErrorMethodsTestCase("nil payload", nil, "%!s(<nil>)"),
	)
}

func (tc panicErrorMethodsTestCase) Name() string {
	return tc.name
}

func (tc panicErrorMethodsTestCase) Test(t *testing.T) {
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

// Factory function for panicErrorMethodsTestCase
func newPanicErrorMethodsTestCase(name string, payload any, expected string) TestCase {
	return panicErrorMethodsTestCase{
		name:     name,
		payload:  payload,
		expected: expected,
	}
}

type panicErrorUnwrapTestCase struct {
	// Large fields - string headers and interface
	name          string
	payload       any
	expectedError string

	// Small fields (1 byte) - boolean flags
	expectUnwrap bool
}

func makePanicErrorUnwrapTestCases() []TestCase {
	return S(
		newPanicErrorUnwrapTestCase("error payload", errors.New("test error"), true, "test error"),
		newPanicErrorUnwrapTestCase("string payload converts to error", "string error", true, "string error"),
		newPanicErrorUnwrapTestCase("non-error payload", 42, false, ""),
		newPanicErrorUnwrapTestCase("nil payload", nil, false, ""),
	)
}

func (tc panicErrorUnwrapTestCase) Name() string {
	return tc.name
}

func (tc panicErrorUnwrapTestCase) Test(t *testing.T) {
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

// Factory function for panicErrorUnwrapTestCase
func newPanicErrorUnwrapTestCase(name string, payload any,
	expectUnwrap bool, expectedError string) TestCase {
	return panicErrorUnwrapTestCase{
		name:          name,
		payload:       payload,
		expectUnwrap:  expectUnwrap,
		expectedError: expectedError,
	}
}

type newPanicErrorfTestCase struct {
	expected string
	format   string
	name     string
	args     []any
}

func makeNewPanicErrorfTestCases() []TestCase {
	return S(
		newNewPanicErrorfTestCase("no args", "simple error", nil, "simple error"),
		newNewPanicErrorfTestCase("with args", "error %d: %s", S[any](42, "test"), "error 42: test"),
		newNewPanicErrorfTestCase("with wrapped error", "wrapped: %w", S[any](errors.New("original")), "wrapped: original"),
	)
}

func (tc newPanicErrorfTestCase) Name() string {
	return tc.name
}

func (tc newPanicErrorfTestCase) Test(t *testing.T) {
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

// Factory function for newPanicErrorfTestCase
func newNewPanicErrorfTestCase(name, format string, args []any, expected string) TestCase {
	return newPanicErrorfTestCase{
		name:     name,
		format:   format,
		args:     args,
		expected: expected,
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

func makePanicFunctionsTestCases() []TestCase {
	return S(
		newPanicFunctionsTestCase("Panic with string", func() {
			Panic("test panic")
		}, "test panic"),
		newPanicFunctionsTestCase("Panic with error", func() {
			Panic(errors.New("test error"))
		}, "test error"),
		newPanicFunctionsTestCase("Panicf without args", func() {
			Panicf("simple panic")
		}, "simple panic"),
		newPanicFunctionsTestCase("Panicf with args", func() {
			Panicf("panic %d: %s", 42, "test")
		}, "panic 42: test"),
	)
}

func (tc panicFunctionsTestCase) Name() string {
	return tc.name
}

func (tc panicFunctionsTestCase) Test(t *testing.T) {
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

// Factory function for panicFunctionsTestCase
func newPanicFunctionsTestCase(name string, fn func(), expected any) TestCase {
	return panicFunctionsTestCase{
		name:     name,
		fn:       fn,
		expected: expected,
	}
}

type panicWrapFunctionsTestCase struct {
	fn          func(error)
	originalErr error
	name        string
}

func (tc panicWrapFunctionsTestCase) Name() string {
	return tc.name
}

func (tc panicWrapFunctionsTestCase) Test(t *testing.T) {
	tc.runTest(t, tc.originalErr)
}

func (tc panicWrapFunctionsTestCase) runTest(t *testing.T, originalErr error) {
	t.Helper()
	defer func() {
		r := recover()
		tc.validateWrapFunction(t, r, originalErr)
	}()
	tc.fn(originalErr)
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

// Factory function for panicWrapFunctionsTestCase
func newPanicWrapFunctionsTestCase(name string, fn func(error), originalErr error) TestCase {
	return panicWrapFunctionsTestCase{
		name:        name,
		fn:          fn,
		originalErr: originalErr,
	}
}

type newUnreachableErrorTestCase struct {
	name       string
	err        error
	note       string
	expectType string
}

func makeNewUnreachableErrorTestCases() []TestCase {
	return S(
		newNewUnreachableErrorTestCase("nil error, empty note", nil, ""),
		newNewUnreachableErrorTestCase("nil error, with note", nil, "test note"),
		newNewUnreachableErrorTestCase("ErrUnreachable, with note", ErrUnreachable, "test note"),
		newNewUnreachableErrorTestCase("other error, no note", errors.New("other error"), ""),
		newNewUnreachableErrorTestCase("other error, with note", errors.New("other error"), "test note"),
	)
}

func (tc newUnreachableErrorTestCase) Name() string {
	return tc.name
}

func (tc newUnreachableErrorTestCase) Test(t *testing.T) {
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

// Factory function for newUnreachableErrorTestCase
func newNewUnreachableErrorTestCase(name string, err error, note string) TestCase {
	return newUnreachableErrorTestCase{
		name:       name,
		err:        err,
		note:       note,
		expectType: "PanicError",
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
	RunTestCases(t, makePanicErrorMethodsTestCases())
}

func TestPanicErrorUnwrap(t *testing.T) {
	RunTestCases(t, makePanicErrorUnwrapTestCases())
}

func TestNewPanicErrorf(t *testing.T) {
	RunTestCases(t, makeNewPanicErrorfTestCases())
}

func TestNewPanicWrap(t *testing.T) {
	t.Run("NewPanicWrap", runNewPanicWrapTest)
	t.Run("NewPanicWrapf", runNewPanicWrapfTest)
}

func TestPanicFunctions(t *testing.T) {
	RunTestCases(t, makePanicFunctionsTestCases())
}

func TestPanicWrapFunctions(t *testing.T) {
	originalErr := errors.New("original error")

	testCases := S(
		newPanicWrapFunctionsTestCase("PanicWrap", func(err error) {
			PanicWrap(err, "wrap note")
		}, originalErr),
		newPanicWrapFunctionsTestCase("PanicWrapf", func(err error) {
			PanicWrapf(err, "wrap %s: %d", "note", 42)
		}, originalErr),
	)

	RunTestCases(t, testCases)
}

func TestNewUnreachableError(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		RunTestCases(t, makeNewUnreachableErrorTestCases())
	})
	t.Run("NewUnreachableErrorf", runNewUnreachableErrorfTest)
}
