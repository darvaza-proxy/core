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

var panicErrorMethodsTestCases = []panicErrorMethodsTestCase{
	newPanicErrorMethodsTestCase("string payload", "test error", "test error"),
	newPanicErrorMethodsTestCase("error payload", errors.New("wrapped error"), "wrapped error"),
	newPanicErrorMethodsTestCase("int payload", 42, "42"),
	newPanicErrorMethodsTestCase("nil payload", nil, "<nil>"),
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
func newPanicErrorMethodsTestCase(name string, payload any, expected string) panicErrorMethodsTestCase {
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

var panicErrorUnwrapTestCases = []panicErrorUnwrapTestCase{
	newPanicErrorUnwrapTestCase("error payload", errors.New("test error"), true, "test error"),
	newPanicErrorUnwrapTestCase("string payload converts to error", "string error", true, "string error"),
	newPanicErrorUnwrapTestCase("non-error payload", 42, false, ""),
	newPanicErrorUnwrapTestCase("nil payload", nil, false, ""),
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
			t.Fatal("expected unwrapped error, got nil")
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
	expectUnwrap bool, expectedError string) panicErrorUnwrapTestCase {
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

var newPanicErrorfTestCases = []newPanicErrorfTestCase{
	newNewPanicErrorfTestCase("no args", "simple error", nil, "simple error"),
	newNewPanicErrorfTestCase("with args", "error %d: %s", S[any](42, "test"), "error 42: test"),
	newNewPanicErrorfTestCase("with wrapped error", "wrapped: %w", S[any](errors.New("original")), "wrapped: original"),
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
func newNewPanicErrorfTestCase(name, format string, args []any, expected string) newPanicErrorfTestCase {
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
		t.Fatal("expected unwrapped error, got nil")
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
		t.Fatal("expected unwrapped error, got nil")
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
func newPanicFunctionsTestCase(name string, fn func(), expected any) panicFunctionsTestCase {
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
func newPanicWrapFunctionsTestCase(name string, fn func(error), originalErr error) panicWrapFunctionsTestCase {
	return panicWrapFunctionsTestCase{
		name:        name,
		fn:          fn,
		originalErr: originalErr,
	}
}

// newUnreachableErrorTestCase states NewUnreachableError's contract in
// the terms callers use: what [errors.Is] finds in the returned chain,
// what the message carries, and which errors the chain bottoms out in.
// How the payload is assembled to hold them is not asserted — a caller
// matching the sentinel or the cause cannot tell, and pinning the
// concrete type only breaks the rows when the assembly is rearranged.
//
// wantCause is the distinct cause the row expects at the bottom of the
// chain beside ErrUnreachable, or nil for the rows where there is none
// — passing nil, or passing ErrUnreachable itself, which the
// constructor normalises to the same shape: the sentinel alone, not
// the sentinel paired with itself. Every row declares wantMsg outright
// rather than deriving it from note, so the no-note rows state what the
// message carries instead.
type newUnreachableErrorTestCase struct {
	err       error
	wantCause error
	name      string
	note      string
	wantMsg   string
}

var newUnreachableErrorTestCases = []newUnreachableErrorTestCase{
	newNewUnreachableErrorTestCase("nil error, empty note",
		nil, "", nil, "unreachable"),
	newNewUnreachableErrorTestCase("nil error, with note",
		nil, "test note", nil, "test note"),
	newNewUnreachableErrorTestCase("ErrUnreachable, empty note",
		ErrUnreachable, "", nil, "unreachable"),
	newNewUnreachableErrorTestCase("ErrUnreachable, with note",
		ErrUnreachable, "test note", nil, "test note"),
	newNewUnreachableErrorTestCase("other error, no note",
		errSentinel, "", errSentinel, "sentinel error"),
	newNewUnreachableErrorTestCase("other error, with note",
		errSentinel, "test note", errSentinel, "test note"),
}

func (tc newUnreachableErrorTestCase) Name() string {
	return tc.name
}

func (tc newUnreachableErrorTestCase) Test(t *testing.T) {
	t.Helper()
	result := NewUnreachableError(0, tc.err, tc.note)

	AssertMustNotNil(t, result, "result")
	pe := AssertMustTypeIs[*PanicError](t, result, "result is *PanicError")
	AssertTrue(t, len(pe.CallStack()) > 0, "stack captured")

	AssertErrorIs(t, result, ErrUnreachable, "ErrUnreachable in chain")
	AssertNotErrorIs(t, result, errUnrelated, "unrelated error absent")
	AssertContains(t, result.Error(), tc.wantMsg, "message")
	AssertSliceEqual(t, tc.wantLeaves(), errorLeaves(result), "chain leaves")
}

// wantLeaves is the exact list of errors the chain bottoms out in: the
// sentinel alone, or the sentinel followed by the declared cause.
func (tc newUnreachableErrorTestCase) wantLeaves() []error {
	if tc.wantCause == nil {
		return S(ErrUnreachable)
	}
	return S(ErrUnreachable, tc.wantCause)
}

// errorLeaves follows Unwrap through every layer of err and returns the
// errors that unwrap no further, in order.
func errorLeaves(err error) []error {
	errs := Unwrap(err)
	if len(errs) == 0 {
		return S(err)
	}

	var leaves []error
	for _, e := range errs {
		leaves = append(leaves, errorLeaves(e)...)
	}
	return leaves
}

// Factory function for newUnreachableErrorTestCase
func newNewUnreachableErrorTestCase(name string, err error, note string,
	wantCause error, wantMsg string) newUnreachableErrorTestCase {
	return newUnreachableErrorTestCase{
		name:      name,
		err:       err,
		note:      note,
		wantCause: wantCause,
		wantMsg:   wantMsg,
	}
}

func runNewUnreachableErrorfTest(t *testing.T) {
	err := errors.New("test error")
	format := "formatted %s: %d"
	args := S[any]("note", 42)

	result := NewUnreachableErrorf(0, err, format, args...)

	if result == nil {
		t.Fatal("expected non-nil error, got nil")
	}

	// Test that it's a PanicError
	pe, ok := result.(*PanicError)
	if !ok {
		t.Fatalf("expected PanicError, got %T", result)
	}

	// Test that ErrUnreachable is in the chain
	if !errors.Is(result, ErrUnreachable) {
		t.Fatal("expected ErrUnreachable in error chain")
	}

	// Test that original error is in the chain
	if !errors.Is(result, err) {
		t.Fatal("expected original error in error chain")
	}

	// Test formatted message
	errorStr := result.Error()
	if !strings.Contains(errorStr, "formatted note: 42") {
		t.Fatalf("expected formatted message in error, got '%s'", errorStr)
	}

	// Test stack trace
	stack := pe.CallStack()
	if len(stack) == 0 {
		t.Fatal("expected non-empty stack trace")
	}
}

// Main test functions that call the helpers
func TestPanicErrorMethods(t *testing.T) {
	RunTestCases(t, panicErrorMethodsTestCases)
}

func TestPanicErrorUnwrap(t *testing.T) {
	RunTestCases(t, panicErrorUnwrapTestCases)
}

func TestNewPanicErrorf(t *testing.T) {
	RunTestCases(t, newPanicErrorfTestCases)
}

func TestNewPanicWrap(t *testing.T) {
	t.Run("NewPanicWrap", runNewPanicWrapTest)
}

func TestNewPanicWrapf(t *testing.T) {
	t.Run("NewPanicWrapf", runNewPanicWrapfTest)
}

func TestPanicFunctions(t *testing.T) {
	RunTestCases(t, panicFunctionsTestCases)
}

func TestPanicWrapFunctions(t *testing.T) {
	originalErr := errors.New("original error")

	testCases := []panicWrapFunctionsTestCase{
		newPanicWrapFunctionsTestCase("PanicWrap", func(err error) {
			PanicWrap(err, "wrap note")
		}, originalErr),
		newPanicWrapFunctionsTestCase("PanicWrapf", func(err error) {
			PanicWrapf(err, "wrap %s: %d", "note", 42)
		}, originalErr),
	}

	RunTestCases(t, testCases)
}

func TestNewUnreachableError(t *testing.T) {
	RunTestCases(t, newUnreachableErrorTestCases)
}

func TestNewUnreachableErrorf(t *testing.T) {
	t.Run("NewUnreachableErrorf", runNewUnreachableErrorfTest)
}

var _ TestCase = deeperTestCase{}

// deeperTestCase pins the skip normalisation every panic constructor
// applies. A negative skip is clamped to 1 rather than passed down to
// getCallers, which rejects it and yields an empty stack.
type deeperTestCase struct {
	name string
	skip int
	want int
}

func newDeeperTestCase(name string, skip, want int) deeperTestCase {
	return deeperTestCase{
		name: name,
		skip: skip,
		want: want,
	}
}

func (tc deeperTestCase) Name() string {
	return tc.name
}

func (tc deeperTestCase) Test(t *testing.T) {
	t.Helper()
	AssertEqual(t, tc.want, deeper(tc.skip), "deeper(%d)", tc.skip)
}

// The clamped rows all state 1, the same answer the zero row states, so
// a negative skip is indistinguishable from no skip at all. The far
// positive row declares that only the negative side is clamped.
func deeperTestCases() []deeperTestCase {
	return []deeperTestCase{
		newDeeperTestCase("no skip", 0, 1),
		newDeeperTestCase("one frame", 1, 2),
		newDeeperTestCase("far positive", 9999, 10000),
		newDeeperTestCase("minus one", -1, 1),
		newDeeperTestCase("far negative", -9999, 1),
	}
}

func TestDeeper(t *testing.T) {
	RunTestCases(t, deeperTestCases())
}

// The seven wrappers below give each constructor a stable, named caller
// for TestPanicConstructorsNegativeSkip, in the manner of
// callMustNoError in the MustNoError tests. Each passes a negative
// skip, so the captured top frame should resolve to the wrapper itself.

func negativeSkipPanicError() *PanicError {
	return NewPanicError(-1, "boom")
}

func negativeSkipPanicErrorf() *PanicError {
	return NewPanicErrorf(-1, "boom %d", 42)
}

func negativeSkipPanicWrap() *PanicError {
	return NewPanicWrap(-1, errSentinel, "note")
}

func negativeSkipPanicWrapf() *PanicError {
	return NewPanicWrapf(-1, errSentinel, "note %d", 42)
}

func negativeSkipUnreachableError() error {
	return NewUnreachableError(-1, errSentinel, "note")
}

// NewUnreachableError calls deeper once per arm, so the note-less arm
// needs a caller of its own; without it a revert there goes unnoticed.
func negativeSkipUnreachableErrorNoNote() error {
	return NewUnreachableError(-1, errSentinel, "")
}

func negativeSkipUnreachableErrorf() error {
	return NewUnreachableErrorf(-1, errSentinel, "note %d", 42)
}

var _ TestCase = negativeSkipTestCase{}

// negativeSkipTestCase states what the clamp is worth: 1 means the
// constructor's immediate caller, so a negative skip attributes there
// rather than to a frame inside panicerror.go. recovered is the value
// one of the wrappers above returned, and wantFunc names that wrapper.
// TestDeeper pins the arithmetic; these rows pin the wiring.
type negativeSkipTestCase struct {
	recovered any
	name      string
	wantFunc  string
}

func newNegativeSkipTestCase(name string, recovered any, wantFunc string) negativeSkipTestCase {
	return negativeSkipTestCase{
		name:      name,
		recovered: recovered,
		wantFunc:  wantFunc,
	}
}

func (tc negativeSkipTestCase) Name() string {
	return tc.name
}

func (tc negativeSkipTestCase) Test(t *testing.T) {
	t.Helper()
	assertTopFrameIs(t, tc.recovered, tc.wantFunc, 2)
}

// negativeSkipTestCases covers every one of the seven deeper call
// sites — verified by reverting each to skip+1 in turn and checking a
// row failed. NewUnreachableError holds two of the seven, one per arm,
// hence its two rows.
func negativeSkipTestCases() []negativeSkipTestCase {
	return []negativeSkipTestCase{
		newNegativeSkipTestCase("NewPanicError",
			negativeSkipPanicError(), "negativeSkipPanicError"),
		newNegativeSkipTestCase("NewPanicErrorf",
			negativeSkipPanicErrorf(), "negativeSkipPanicErrorf"),
		newNegativeSkipTestCase("NewPanicWrap",
			negativeSkipPanicWrap(), "negativeSkipPanicWrap"),
		newNegativeSkipTestCase("NewPanicWrapf",
			negativeSkipPanicWrapf(), "negativeSkipPanicWrapf"),
		newNegativeSkipTestCase("NewUnreachableError with note",
			negativeSkipUnreachableError(), "negativeSkipUnreachableError"),
		newNegativeSkipTestCase("NewUnreachableError without note",
			negativeSkipUnreachableErrorNoNote(), "negativeSkipUnreachableErrorNoNote"),
		newNegativeSkipTestCase("NewUnreachableErrorf",
			negativeSkipUnreachableErrorf(), "negativeSkipUnreachableErrorf"),
	}
}

func TestPanicConstructorsNegativeSkip(t *testing.T) {
	RunTestCases(t, negativeSkipTestCases())
}
