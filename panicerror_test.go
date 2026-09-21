package core

import (
	"errors"
	"fmt"
	"testing"
)

// TestCase interface validations
var (
	_ TestCase = panicErrorMethodsTestCase{}
	_ TestCase = panicErrorUnwrapTestCase{}
	_ TestCase = newPanicErrorfTestCase{}
	_ TestCase = panicTestCase{}
	_ TestCase = panicStackTestCase{}
	_ TestCase = panicFromTestCase{}
	_ TestCase = panicfTestCase{}
	_ TestCase = panicfFromTestCase{}
	_ TestCase = newUnreachableErrorTestCase{}
	_ TestCase = newUnreachableErrorfTestCase{}
	_ TestCase = panicUnreachableFromTestCase{}
	_ TestCase = panicUnreachablefFromTestCase{}
	_ TestCase = deeperTestCase{}
	_ TestCase = negativeSkipTestCase{}
)

// panicErrorMethodsTestCase states what NewPanicError keeps and what it
// changes: the payload comes back from Recovered as it went in, the
// message renders it behind a "panic: " prefix, and a stack is captured
// either way. A string payload is the exception — it becomes an error
// carrying the same text — and the rows where that happens say so
// through a factory of their own.
type panicErrorMethodsTestCase struct {
	payload any
	name    string
	wantMsg string

	wantConverted bool
}

var panicErrorMethodsTestCases = []panicErrorMethodsTestCase{
	newPanicErrorMethodsTestCaseString("string payload", "test error"),
	newPanicErrorMethodsTestCaseString("verb in string payload", "test %d error"),
	newPanicErrorMethodsTestCaseString("empty string payload", ""),
	newPanicErrorMethodsTestCase("named string payload",
		namedString("test error"), "test error"),
	newPanicErrorMethodsTestCase("error payload", errors.New("wrapped error"),
		"wrapped error"),
	newPanicErrorMethodsTestCase("int payload", 42, "42"),
	newPanicErrorMethodsTestCase("stringer payload",
		mockStringer{value: testHello}, testHello),
	newPanicErrorMethodsTestCase("zero payload", false, "false"),
	newPanicErrorMethodsTestCase("nil payload", nil, "<nil>"),
}

func newPanicErrorMethodsTestCase(name string, payload any,
	wantMsg string) panicErrorMethodsTestCase {
	return panicErrorMethodsTestCase{
		payload:       payload,
		name:          name,
		wantMsg:       wantMsg,
		wantConverted: false,
	}
}

// newPanicErrorMethodsTestCaseString declares a row whose payload is a
// string, so the recovered value is an error carrying that same text.
func newPanicErrorMethodsTestCaseString(name, payload string) panicErrorMethodsTestCase {
	return panicErrorMethodsTestCase{
		payload:       payload,
		name:          name,
		wantMsg:       payload,
		wantConverted: true,
	}
}

func (tc panicErrorMethodsTestCase) Name() string {
	return tc.name
}

func (tc panicErrorMethodsTestCase) Test(t *testing.T) {
	t.Helper()
	pe := NewPanicError(0, tc.payload)

	AssertEqual(t, "panic: "+tc.wantMsg, pe.Error(), "message")
	AssertTrue(t, len(pe.CallStack()) > 0, "stack captured")

	got := pe.Recovered()
	if tc.wantConverted {
		err := AssertMustTypeIs[error](t, got, "payload is an error")
		AssertSame(t, err, pe.Unwrap(), "Unwrap")
		got = err.Error()
	}

	AssertEqual(t, tc.payload, got, "payload")
}

// namedString is a string by kind only, which NewPanicError does not
// convert to an error.
type namedString string

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
	newPanicErrorUnwrapTestCase("named string payload", namedString("string error"), false, ""),
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

	if !tc.expectUnwrap {
		AssertNil(t, unwrapped, "Unwrap")
		return
	}

	AssertMustError(t, unwrapped, "Unwrap")
	AssertEqual(t, tc.expectedError, unwrapped.Error(), "unwrapped")
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
	expectedError := fmt.Sprintf("panic: %s", tc.expected)
	AssertEqual(t, expectedError, pe.Error(), "Error")

	// Test that the payload is an error carrying the formatted message
	payload := AssertMustTypeIs[error](t, pe.Recovered(), "payload")
	AssertEqual(t, tc.expected, payload.Error(), "payload message")
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
	t.Helper()
	originalErr := errors.New("original error")
	note := "wrapped note"

	pe := NewPanicWrap(0, originalErr, note)

	// Test that it wraps the error
	AssertError(t, pe.Unwrap(), "Unwrap")

	// Test error message contains both note and original
	errorStr := pe.Error()
	AssertContains(t, errorStr, note, "note")
	AssertContains(t, errorStr, originalErr.Error(), "original error")
}

func runNewPanicWrapfTest(t *testing.T) {
	t.Helper()
	originalErr := errors.New("original error")
	format := "wrapped %s: %d"
	args := S[any]("note", 42)

	pe := NewPanicWrapf(0, originalErr, format, args...)

	// Test that it wraps the error
	AssertError(t, pe.Unwrap(), "Unwrap")

	// Test error message contains formatted note and original
	errorStr := pe.Error()
	AssertContains(t, errorStr, "wrapped note: 42", "formatted note")
	AssertContains(t, errorStr, originalErr.Error(), "original error")
}

// panicTestCase states what Panic raises: a *PanicError over a captured
// stack carrying the payload it was given, with a string converted to
// an error of the same text.
type panicTestCase struct {
	payload any
	name    string

	wantConverted bool
}

var panicTestCases = []panicTestCase{
	newPanicTestCaseString("string payload", "test panic"),
	newPanicTestCase("named string payload", namedString("test panic")),
	newPanicTestCase("error payload", errors.New("test error")),
	newPanicTestCase("int payload", 42),
}

func newPanicTestCase(name string, payload any) panicTestCase {
	return panicTestCase{
		payload:       payload,
		name:          name,
		wantConverted: false,
	}
}

// newPanicTestCaseString declares a row whose payload is a string, so
// the value recovered is an error carrying that same text.
func newPanicTestCaseString(name, payload string) panicTestCase {
	return panicTestCase{
		payload:       payload,
		name:          name,
		wantConverted: true,
	}
}

func (tc panicTestCase) Name() string {
	return tc.name
}

func (tc panicTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := AssertMustTypeIs[*PanicError](t, recover(), "recovered")
		AssertTrue(t, len(pe.CallStack()) > 0, "stack captured")

		got := pe.Recovered()
		if tc.wantConverted {
			err := AssertMustTypeIs[error](t, got, "payload is an error")
			got = err.Error()
		}

		AssertEqual(t, tc.payload, got, "payload")
	}()

	Panic(tc.payload)
}

// callPanic gives Panic a stable, named caller, so a stack Panic
// captures itself starts here.
func callPanic(payload any) {
	Panic(payload)
}

// newPanicPayload builds a *PanicError whose stack starts here, apart
// from any stack Panic would capture.
func newPanicPayload() *PanicError {
	return NewPanicError(0, errSentinel)
}

// panicStackTestCase states where the stack of the value Panic raises
// starts, and what payload it carries. A *PanicError payload is raised
// as it is, so its own stack and payload come back; a nil one counts as
// no payload.
type panicStackTestCase struct {
	payload     any
	wantPayload any
	name        string
	wantFunc    string
}

func newPanicStackTestCase(name string, payload, wantPayload any,
	wantFunc string) panicStackTestCase {
	return panicStackTestCase{
		payload:     payload,
		wantPayload: wantPayload,
		name:        name,
		wantFunc:    wantFunc,
	}
}

func (tc panicStackTestCase) Name() string {
	return tc.name
}

func (tc panicStackTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := assertTopFrameIs(t, recover(), tc.wantFunc, 2)
		AssertEqual(t, tc.wantPayload, pe.Recovered(), "payload")
	}()

	callPanic(tc.payload)
}

func panicStackTestCases() []panicStackTestCase {
	return []panicStackTestCase{
		newPanicStackTestCase("error payload",
			errSentinel, errSentinel, "callPanic"),
		newPanicStackTestCase("PanicError payload",
			newPanicPayload(), errSentinel, "newPanicPayload"),
		newPanicStackTestCase("nil PanicError payload",
			(*PanicError)(nil), nil, "callPanic"),
	}
}

// callPanicFromOuter and callPanicFrom give PanicFrom two stable, named
// frames: a skip of 0 starts the stack at the inner one, 1 at the outer.
func callPanicFromOuter(skip int, payload any) {
	callPanicFrom(skip, payload)
}

func callPanicFrom(skip int, payload any) {
	PanicFrom(skip, payload)
}

// panicFromTestCase states where the stack of the value PanicFrom raises
// starts for a given skip, and what payload it carries.
type panicFromTestCase struct {
	payload     any
	wantPayload any
	name        string
	wantFunc    string
	skip        int
}

func newPanicFromTestCase(name string, skip int, payload, wantPayload any,
	wantFunc string) panicFromTestCase {
	return panicFromTestCase{
		payload:     payload,
		wantPayload: wantPayload,
		name:        name,
		wantFunc:    wantFunc,
		skip:        skip,
	}
}

func (tc panicFromTestCase) Name() string {
	return tc.name
}

func (tc panicFromTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := assertTopFrameIs(t, recover(), tc.wantFunc, 2)
		AssertEqual(t, tc.wantPayload, pe.Recovered(), "payload")
	}()

	callPanicFromOuter(tc.skip, tc.payload)
}

func panicFromTestCases() []panicFromTestCase {
	return []panicFromTestCase{
		newPanicFromTestCase("no skip", 0,
			errSentinel, errSentinel, "callPanicFrom"),
		newPanicFromTestCase("one frame", 1,
			errSentinel, errSentinel, "callPanicFromOuter"),
		newPanicFromTestCase("negative skip", -1,
			errSentinel, errSentinel, "callPanicFrom"),
		newPanicFromTestCase("PanicError payload", 1,
			newPanicPayload(), errSentinel, "newPanicPayload"),
		newPanicFromTestCase("nil PanicError payload", 1,
			(*PanicError)(nil), nil, "callPanicFromOuter"),
	}
}

// panicfTestCase states the same for Panicf, where the payload is always
// the formatted message as an error.
type panicfTestCase struct {
	format  string
	name    string
	wantMsg string
	args    []any
}

var panicfTestCases = []panicfTestCase{
	newPanicfTestCase("no args", "simple panic", nil, "simple panic"),
	newPanicfTestCase("with args", "panic %d: %s", S[any](42, "test"),
		"panic 42: test"),
}

func newPanicfTestCase(name, format string, args []any,
	wantMsg string) panicfTestCase {
	return panicfTestCase{
		format:  format,
		name:    name,
		wantMsg: wantMsg,
		args:    args,
	}
}

func (tc panicfTestCase) Name() string {
	return tc.name
}

func (tc panicfTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := AssertMustTypeIs[*PanicError](t, recover(), "recovered")
		err := AssertMustTypeIs[error](t, pe.Recovered(), "payload is an error")
		AssertEqual(t, tc.wantMsg, err.Error(), "payload message")
		AssertTrue(t, len(pe.CallStack()) > 0, "stack captured")
	}()

	Panicf(tc.format, tc.args...)
}

// callPanicf gives Panicf a stable, named caller, so the stack Panicf
// captures starts here.
func callPanicf(format string, args ...any) {
	Panicf(format, args...)
}

// callPanicfFromOuter and callPanicfFrom give PanicfFrom two stable,
// named frames: a skip of 0 starts the stack at the inner one, 1 at the
// outer.
func callPanicfFromOuter(skip int, format string, args ...any) {
	callPanicfFrom(skip, format, args...)
}

func callPanicfFrom(skip int, format string, args ...any) {
	PanicfFrom(skip, format, args...)
}

// panicfFromTestCase states where the stack of the value PanicfFrom
// raises starts for a given skip. Every row formats the same message.
type panicfFromTestCase struct {
	name     string
	wantFunc string
	skip     int
}

func newPanicfFromTestCase(name string, skip int,
	wantFunc string) panicfFromTestCase {
	return panicfFromTestCase{
		name:     name,
		wantFunc: wantFunc,
		skip:     skip,
	}
}

func (tc panicfFromTestCase) Name() string {
	return tc.name
}

func (tc panicfFromTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := assertTopFrameIs(t, recover(), tc.wantFunc, 2)
		err := AssertMustTypeIs[error](t, pe.Recovered(), "payload is an error")
		AssertEqual(t, "panic 42", err.Error(), "payload message")
	}()

	callPanicfFromOuter(tc.skip, "panic %d", 42)
}

var panicfFromTestCases = []panicfFromTestCase{
	newPanicfFromTestCase("no skip", 0, "callPanicfFrom"),
	newPanicfFromTestCase("one frame", 1, "callPanicfFromOuter"),
	newPanicfFromTestCase("negative skip", -1, "callPanicfFrom"),
}

func runPanicWrapTest(t *testing.T) {
	t.Helper()
	defer func() {
		pe := AssertMustTypeIs[*PanicError](t, recover(), "recovered")
		AssertErrorIs(t, pe, errSentinel, "original error in chain")
		AssertContains(t, pe.Error(), "wrap note", "message")
	}()

	PanicWrap(errSentinel, "wrap note")
}

func runPanicWrapfTest(t *testing.T) {
	t.Helper()
	defer func() {
		pe := AssertMustTypeIs[*PanicError](t, recover(), "recovered")
		AssertErrorIs(t, pe, errSentinel, "original error in chain")
		AssertContains(t, pe.Error(), "wrap note: 42", "message")
	}()

	PanicWrapf(errSentinel, "wrap %s: %d", "note", 42)
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

// newUnreachableErrorfTestCase states what note NewUnreachableErrorf
// gives the error it builds: the format as it is when there are no
// arguments, formatted otherwise.
type newUnreachableErrorfTestCase struct {
	name    string
	format  string
	wantMsg string
	args    []any
}

var newUnreachableErrorfTestCases = []newUnreachableErrorfTestCase{
	newNewUnreachableErrorfTestCase("no args", "100% sure", nil,
		"100% sure"),
	newNewUnreachableErrorfTestCase("with args", "formatted %s: %d",
		S[any]("note", 42), "formatted note: 42"),
}

func newNewUnreachableErrorfTestCase(name, format string, args []any,
	wantMsg string) newUnreachableErrorfTestCase {
	return newUnreachableErrorfTestCase{
		name:    name,
		format:  format,
		wantMsg: wantMsg,
		args:    args,
	}
}

func (tc newUnreachableErrorfTestCase) Name() string {
	return tc.name
}

func (tc newUnreachableErrorfTestCase) Test(t *testing.T) {
	t.Helper()
	result := NewUnreachableErrorf(0, errSentinel, tc.format, tc.args...)

	assertUnreachablePanicShape(t, result, errSentinel)
	AssertContains(t, result.Error(), tc.wantMsg, "message")
}

// callPanicUnreachable and callPanicUnreachablef give the two functions
// stable, named callers, so the stacks they capture start there.
func callPanicUnreachable(err error, note string) {
	PanicUnreachable(err, note)
}

func callPanicUnreachablef(err error, format string, args ...any) {
	PanicUnreachablef(err, format, args...)
}

// callPanicUnreachableFromOuter and callPanicUnreachableFrom give
// PanicUnreachableFrom two stable, named frames: a skip of 0 starts the
// stack at the inner one, 1 at the outer.
func callPanicUnreachableFromOuter(skip int, err error, note string) {
	callPanicUnreachableFrom(skip, err, note)
}

func callPanicUnreachableFrom(skip int, err error, note string) {
	PanicUnreachableFrom(skip, err, note)
}

// panicUnreachableFromTestCase states where the stack of the value
// PanicUnreachableFrom raises starts for a given skip. Every row passes
// the same error and note.
type panicUnreachableFromTestCase struct {
	name     string
	wantFunc string
	skip     int
}

func newPanicUnreachableFromTestCase(name string, skip int,
	wantFunc string) panicUnreachableFromTestCase {
	return panicUnreachableFromTestCase{
		name:     name,
		wantFunc: wantFunc,
		skip:     skip,
	}
}

func (tc panicUnreachableFromTestCase) Name() string {
	return tc.name
}

func (tc panicUnreachableFromTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := assertTopFrameIs(t, recover(), tc.wantFunc, 2)
		assertUnreachableChain(t, pe, errSentinel)
		AssertContains(t, pe.Error(), "test note", "message")
	}()

	callPanicUnreachableFromOuter(tc.skip, errSentinel, "test note")
}

var panicUnreachableFromTestCases = []panicUnreachableFromTestCase{
	newPanicUnreachableFromTestCase("no skip", 0,
		"callPanicUnreachableFrom"),
	newPanicUnreachableFromTestCase("one frame", 1,
		"callPanicUnreachableFromOuter"),
	newPanicUnreachableFromTestCase("negative skip", -1,
		"callPanicUnreachableFrom"),
}

// callPanicUnreachablefFromOuter and callPanicUnreachablefFrom give
// PanicUnreachablefFrom two stable, named frames: a skip of 0 starts the
// stack at the inner one, 1 at the outer.
func callPanicUnreachablefFromOuter(skip int, err error, format string,
	args ...any) {
	callPanicUnreachablefFrom(skip, err, format, args...)
}

func callPanicUnreachablefFrom(skip int, err error, format string,
	args ...any) {
	PanicUnreachablefFrom(skip, err, format, args...)
}

// panicUnreachablefFromTestCase states the same for
// PanicUnreachablefFrom. Every row passes the same error and formats the
// same note.
type panicUnreachablefFromTestCase struct {
	name     string
	wantFunc string
	skip     int
}

func newPanicUnreachablefFromTestCase(name string, skip int,
	wantFunc string) panicUnreachablefFromTestCase {
	return panicUnreachablefFromTestCase{
		name:     name,
		wantFunc: wantFunc,
		skip:     skip,
	}
}

func (tc panicUnreachablefFromTestCase) Name() string {
	return tc.name
}

func (tc panicUnreachablefFromTestCase) Test(t *testing.T) {
	t.Helper()
	defer func() {
		pe := assertTopFrameIs(t, recover(), tc.wantFunc, 2)
		assertUnreachableChain(t, pe, errSentinel)
		AssertContains(t, pe.Error(), "test note 42", "message")
	}()

	callPanicUnreachablefFromOuter(tc.skip, errSentinel, "test note %d", 42)
}

var panicUnreachablefFromTestCases = []panicUnreachablefFromTestCase{
	newPanicUnreachablefFromTestCase("no skip", 0,
		"callPanicUnreachablefFrom"),
	newPanicUnreachablefFromTestCase("one frame", 1,
		"callPanicUnreachablefFromOuter"),
	newPanicUnreachablefFromTestCase("negative skip", -1,
		"callPanicUnreachablefFrom"),
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

func TestPanic(t *testing.T) {
	RunTestCases(t, panicTestCases)
}

func TestPanicStack(t *testing.T) {
	RunTestCases(t, panicStackTestCases())
}

func TestPanicFrom(t *testing.T) {
	RunTestCases(t, panicFromTestCases())
}

func TestPanicf(t *testing.T) {
	RunTestCases(t, panicfTestCases)
}

// TestPanicfStack states that the stack Panicf captures starts at its
// caller.
func TestPanicfStack(t *testing.T) {
	defer func() {
		_ = assertTopFrameIs(t, recover(), "callPanicf", 2)
	}()
	callPanicf("panic %d", 42)
}

func TestPanicfFrom(t *testing.T) {
	RunTestCases(t, panicfFromTestCases)
}

func TestPanicWrap(t *testing.T) {
	t.Run("PanicWrap", runPanicWrapTest)
}

func TestPanicWrapf(t *testing.T) {
	t.Run("PanicWrapf", runPanicWrapfTest)
}

func TestNewUnreachableError(t *testing.T) {
	RunTestCases(t, newUnreachableErrorTestCases)
}

func TestNewUnreachableErrorf(t *testing.T) {
	RunTestCases(t, newUnreachableErrorfTestCases)
}

// TestPanicUnreachable states that PanicUnreachable raises the
// unreachable error for err and note, its stack starting at the caller.
func TestPanicUnreachable(t *testing.T) {
	defer func() {
		pe := assertTopFrameIs(t, recover(), "callPanicUnreachable", 2)
		assertUnreachableChain(t, pe, errSentinel)
		AssertContains(t, pe.Error(), "test note", "message")
	}()
	callPanicUnreachable(errSentinel, "test note")
}

// TestPanicUnreachablef states the same for PanicUnreachablef, with the
// note formatted.
func TestPanicUnreachablef(t *testing.T) {
	defer func() {
		pe := assertTopFrameIs(t, recover(), "callPanicUnreachablef", 2)
		assertUnreachableChain(t, pe, errSentinel)
		AssertContains(t, pe.Error(), "test note 42", "message")
	}()
	callPanicUnreachablef(errSentinel, "test note %d", 42)
}

func TestPanicUnreachableFrom(t *testing.T) {
	RunTestCases(t, panicUnreachableFromTestCases)
}

func TestPanicUnreachablefFrom(t *testing.T) {
	RunTestCases(t, panicUnreachablefFromTestCases)
}

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
	_ = assertTopFrameIs(t, tc.recovered, tc.wantFunc, 2)
}

// negativeSkipTestCases covers the seven deeper call sites in the
// constructors — verified by reverting each to skip+1 in turn and
// checking a row failed. NewUnreachableError holds two of the seven, one
// per arm, hence its two rows. The From functions have negative rows in
// their own tables.
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
