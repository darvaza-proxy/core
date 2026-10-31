package core_test

// cspell:words Errno ENOENT

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = quietWrapTestCase{}
	_ core.TestCase = coalesceErrorTestCase{}
	_ core.TestCase = isErrorTestCase{}
	_ core.TestCase = isErrorFnTestCase{}
	_ core.TestCase = isErrorFn2TestCase{}
	_ core.TestCase = checkErrorIsInTestCase{}
	_ core.TestCase = checkErrorIsIn2TestCase{}
	_ core.TestCase = temporaryErrorTestCase{}
	_ core.TestCase = permanentErrorTestCase{}
	_ core.TestCase = permanentErrorNilCauseTestCase{}
	_ core.TestCase = checkIsTemporaryTestCase{}
	_ core.TestCase = isTemporaryTestCase{}
	_ core.TestCase = checkIsTimeoutTestCase{}
	_ core.TestCase = isTimeoutTestCase{}
	_ core.TestCase = wrappedErrorTestCase{}
	_ core.TestCase = unwrapBranchesTestCase{}
)

const emptyString = ""

// Test cases for IsErrorFn function
type isErrorFnTestCase struct {
	checkFn  func(error) bool
	name     string
	errs     []error
	expected bool
}

func (tc isErrorFnTestCase) Name() string {
	return tc.name
}

func (tc isErrorFnTestCase) Test(t *testing.T) {
	t.Helper()
	result := core.IsErrorFn(tc.checkFn, tc.errs...)
	core.AssertEqual(t, tc.expected, result, "IsErrorFn result")
}

func newIsErrorFnTestCase(name string, checkFn func(error) bool, expected bool, errs ...error) isErrorFnTestCase {
	return isErrorFnTestCase{
		name:     name,
		checkFn:  checkFn,
		errs:     errs,
		expected: expected,
	}
}

func TestIsErrorFn(t *testing.T) {
	testErr := errors.New("test error")
	differentErr := errors.New("different error")
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)

	isTestErr := func(err error) bool {
		return errors.Is(err, testErr)
	}

	testCases := []isErrorFnTestCase{
		newIsErrorFnTestCase("matching error", isTestErr, true, testErr),
		newIsErrorFnTestCase("non-matching error", isTestErr, false, differentErr),
		newIsErrorFnTestCase("nil check function", nil, false, testErr),
		newIsErrorFnTestCase("no errors", isTestErr, false),
		newIsErrorFnTestCase("wrapped error", isTestErr, true, wrappedErr),
		newIsErrorFnTestCase("nil error in slice", isTestErr, true, nil, testErr),
	}

	core.RunTestCases(t, testCases)
}

// Test cases for IsErrorFn2 function
type isErrorFn2TestCase struct {
	checkFn       func(error) (bool, bool)
	name          string
	errs          []error
	expectedIs    bool
	expectedKnown bool
}

func (tc isErrorFn2TestCase) Name() string {
	return tc.name
}

func (tc isErrorFn2TestCase) Test(t *testing.T) {
	t.Helper()
	is, known := core.IsErrorFn2(tc.checkFn, tc.errs...)
	core.AssertEqual(t, tc.expectedIs, is, "IsErrorFn2 is result")
	core.AssertEqual(t, tc.expectedKnown, known, "IsErrorFn2 known result")
}

func newIsErrorFn2TestCase(name string, checkFn func(error) (bool, bool),
	expectedIs, expectedKnown bool, errs ...error) isErrorFn2TestCase {
	return isErrorFn2TestCase{
		name:          name,
		checkFn:       checkFn,
		errs:          errs,
		expectedIs:    expectedIs,
		expectedKnown: expectedKnown,
	}
}

func TestIsErrorFn2(t *testing.T) {
	testErr := errors.New("test error")
	differentErr := errors.New("different error")
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)

	isTestErr := func(err error) (bool, bool) {
		if err == nil {
			return false, false
		}
		return errors.Is(err, testErr), true
	}

	testCases := []isErrorFn2TestCase{
		newIsErrorFn2TestCase("matching error", isTestErr, true, true, testErr),
		newIsErrorFn2TestCase("non-matching error", isTestErr, false, true, differentErr),
		newIsErrorFn2TestCase("nil check function", nil, false, true, testErr),
		newIsErrorFn2TestCase("no errors", isTestErr, false, true),
		newIsErrorFn2TestCase("wrapped error", isTestErr, true, true, wrappedErr),
		newIsErrorFn2TestCase("unknown error type",
			func(_ error) (bool, bool) { return false, false }, false, false, testErr),
	}

	core.RunTestCases(t, testCases)
}

// Test cases for QuietWrap function
type quietWrapTestCase struct {
	err      error
	name     string
	format   string
	expected string
	args     []any
}

func newQuietWrapTestCase(name string, err error, format, expected string, args ...any) quietWrapTestCase {
	return quietWrapTestCase{
		name:     name,
		err:      err,
		format:   format,
		expected: expected,
		args:     args,
	}
}

func (tc quietWrapTestCase) Name() string {
	return tc.name
}

func (tc quietWrapTestCase) Test(t *testing.T) {
	t.Helper()

	result := core.QuietWrap(tc.err, tc.format, tc.args...)

	if tc.expected == "" {
		core.AssertEqual(t, tc.err, result, "QuietWrap original")
	} else {
		// Test that it's wrappable
		wrapped := core.AssertMustTypeIs[core.Unwrappable](t, result, "QuietWrap")

		core.AssertEqual(t, tc.expected, result.Error(), "QuietWrap message")
		core.AssertEqual(t, tc.err, wrapped.Unwrap(), "QuietWrap unwrap")
	}
}

func TestQuietWrap(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := core.S(
		newQuietWrapTestCase("nil error", nil, "message", ""),
		newQuietWrapTestCase("empty format", baseErr, "", ""),
		newQuietWrapTestCase("simple message", baseErr, "wrapped", "wrapped"),
		newQuietWrapTestCase("formatted message", baseErr, "wrapped: %s", "wrapped: test", "test"),
		newQuietWrapTestCase("multiple args", baseErr, "error %d: %s", "error 42: test", 42, "test"),
		newQuietWrapTestCase("nil error with format", nil, "message %s", "", "test"),
	)

	core.RunTestCases(t, testCases)
}

// Test cases for CoalesceError function
type coalesceErrorTestCase struct {
	expected error
	name     string
	errs     []error
}

func newCoalesceErrorTestCase(name string, expected error, errs ...error) coalesceErrorTestCase {
	return coalesceErrorTestCase{
		name:     name,
		expected: expected,
		errs:     errs,
	}
}

func (tc coalesceErrorTestCase) Name() string {
	return tc.name
}

func (tc coalesceErrorTestCase) Test(t *testing.T) {
	t.Helper()

	result := core.CoalesceError(tc.errs...)
	core.AssertEqual(t, tc.expected, result, "CoalesceError")
}

func TestCoalesceError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	testCases := core.S(
		newCoalesceErrorTestCase("no errors", nil),
		newCoalesceErrorTestCase("single nil error", nil, nil),
		newCoalesceErrorTestCase("single non-nil error", err1, err1),
		newCoalesceErrorTestCase("first non-nil wins", err1, nil, err1, err2),
		newCoalesceErrorTestCase("first error wins", err1, err1, err2, err3),
		newCoalesceErrorTestCase("all nil", nil, nil, nil, nil),
		newCoalesceErrorTestCase("nil then non-nil", err2, nil, nil, err2),
		newCoalesceErrorTestCase("empty slice", nil),
	)

	core.RunTestCases(t, testCases)
}

// errorWithIs matches a single target via its Is method instead of
// identity, exercising the Is(error) bool support in IsError.
type errorWithIs struct {
	target error
}

func (*errorWithIs) Error() string { return "error with Is" }

func (e *errorWithIs) Is(target error) bool { return target == e.target }

// nonComparableError is a value-typed error carrying a slice, so its
// dynamic type does not support ==: comparing two of them panics
// rather than reporting false. It pins that IsError degrades the
// comparison to "no match" instead of panicking.
type nonComparableError struct {
	codes []int
}

func (nonComparableError) Error() string { return "non-comparable error" }

// nonComparableErrorWithIs is non-comparable like nonComparableError,
// and its Is accepts anything: it pins that the Is(error) bool probe
// still runs after the identity comparison degrades, so a
// non-comparable error can match through its own Is method.
type nonComparableErrorWithIs struct {
	codes []int
}

func (nonComparableErrorWithIs) Error() string {
	return "non-comparable error with Is"
}

func (nonComparableErrorWithIs) Is(error) bool { return true }

// Test cases for IsError function
type isErrorTestCase struct {
	name     string
	err      error
	errs     []error
	expected bool
}

func newIsErrorTestCase(name string, err error, expected bool, errs ...error) isErrorTestCase {
	return isErrorTestCase{
		name:     name,
		err:      err,
		errs:     errs,
		expected: expected,
	}
}

func (tc isErrorTestCase) Name() string {
	return tc.name
}

func (tc isErrorTestCase) Test(t *testing.T) {
	t.Helper()

	result := core.IsError(tc.err, tc.errs...)
	core.AssertEqual(t, tc.expected, result, "IsError result")
}

func TestIsError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	wrappedErr := fmt.Errorf("wrapped: %w", err1)
	errIs := &errorWithIs{target: err2}
	pathErr := &fs.PathError{Op: "open", Path: "missing",
		Err: syscall.ENOENT}
	errNonComparable := nonComparableError{codes: []int{1}}
	errNonComparableIs := nonComparableErrorWithIs{codes: []int{2}}

	testCases := []isErrorTestCase{
		newIsErrorTestCase("nil error", nil, false, err1),
		newIsErrorTestCase("no target errors - non-nil", err1, true),
		newIsErrorTestCase("no target errors - nil", nil, false),
		newIsErrorTestCase("exact match", err1, true, err1, err2),
		newIsErrorTestCase("no match", err3, false, err1, err2),
		newIsErrorTestCase("wrapped error match", wrappedErr, true, err1),
		newIsErrorTestCase("wrapped error no match", wrappedErr, false, err2),
		newIsErrorTestCase("multiple targets", err2, true, err1, err2, err3),
		newIsErrorTestCase("empty targets", err1, true),
		newIsErrorTestCase("Is method match", errIs, true, err2),
		newIsErrorTestCase("Is method mismatch", errIs, false, err3),
		newIsErrorTestCase("wrapped Is method match",
			fmt.Errorf("wrapped: %w", errIs), true, err2),
		newIsErrorTestCase("target Is method not consulted",
			err2, false, errIs),
		newIsErrorTestCase("Errno match via Is method",
			pathErr, true, fs.ErrNotExist),
		newIsErrorTestCase("Is method self match", errIs, true, errIs),
		newIsErrorTestCase("Is method match second target",
			errIs, true, err3, err2),
		newIsErrorTestCase("Errno self match",
			syscall.ENOENT, true, syscall.ENOENT),
		newIsErrorTestCase("non-comparable no match",
			errNonComparable, false, errNonComparable),
		newIsErrorTestCase("non-comparable Is method match",
			errNonComparableIs, true, errNonComparableIs),
	}

	core.RunTestCases(t, testCases)
}

// Test cases for the NewCheckErrorIsIn per-node closure
type checkErrorIsInTestCase struct {
	err     error
	name    string
	targets []error

	want bool
}

func newCheckErrorIsInTestCase(name string, err error,
	targets []error, want bool) checkErrorIsInTestCase {
	return checkErrorIsInTestCase{
		name:    name,
		err:     err,
		targets: targets,
		want:    want,
	}
}

func (tc checkErrorIsInTestCase) Name() string { return tc.name }

func (tc checkErrorIsInTestCase) Test(t *testing.T) {
	t.Helper()
	check := core.NewCheckErrorIsIn(tc.targets)
	core.AssertEqual(t, tc.want, check(tc.err), "check result")
}

func checkErrorIsInTestCases() []checkErrorIsInTestCase {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	errIs := &errorWithIs{target: err2}

	return []checkErrorIsInTestCase{
		newCheckErrorIsInTestCase("identity match",
			err1, []error{err1}, true),
		newCheckErrorIsInTestCase("identity mismatch",
			err1, []error{err2}, false),
		newCheckErrorIsInTestCase("Is method match second target",
			errIs, []error{err1, err2}, true),
		newCheckErrorIsInTestCase("nil target never matches",
			err1, []error{nil}, false),
		newCheckErrorIsInTestCase("no targets",
			err1, nil, false),
		newCheckErrorIsInTestCase("no unwrapping",
			fmt.Errorf("wrapped: %w", err1), []error{err1}, false),
	}
}

// TestNewCheckErrorIsIn exercises the per-node closure directly:
// matching is identity-or-Is per node, without unwrapping — traversal
// belongs to IsErrorFn.
func TestNewCheckErrorIsIn(t *testing.T) {
	core.RunTestCases(t, checkErrorIsInTestCases())
}

// Test cases for the NewCheckErrorIsIn2 per-node closure
type checkErrorIsIn2TestCase struct {
	err     error
	name    string
	targets []error

	wantIs    bool
	wantKnown bool
}

func newCheckErrorIsIn2TestCase(name string, err error,
	targets []error, wantIs, wantKnown bool) checkErrorIsIn2TestCase {
	return checkErrorIsIn2TestCase{
		name:      name,
		err:       err,
		targets:   targets,
		wantIs:    wantIs,
		wantKnown: wantKnown,
	}
}

func (tc checkErrorIsIn2TestCase) Name() string { return tc.name }

func (tc checkErrorIsIn2TestCase) Test(t *testing.T) {
	t.Helper()
	is, known := core.NewCheckErrorIsIn2(tc.targets)(tc.err)
	core.AssertEqual(t, tc.wantIs, is, "is")
	core.AssertEqual(t, tc.wantKnown, known, "known")
}

func checkErrorIsIn2TestCases() []checkErrorIsIn2TestCase {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	errIs := &errorWithIs{target: err2}

	return []checkErrorIsIn2TestCase{
		newCheckErrorIsIn2TestCase("identity match",
			err1, []error{err1}, true, true),
		newCheckErrorIsIn2TestCase("miss is not final",
			err1, []error{err2}, false, false),
		newCheckErrorIsIn2TestCase("Is method match second target",
			errIs, []error{err1, err2}, true, true),
		newCheckErrorIsIn2TestCase("Is method miss not final",
			errIs, []error{err1}, false, false),
		newCheckErrorIsIn2TestCase("nil target final miss",
			err1, []error{nil}, false, true),
		newCheckErrorIsIn2TestCase("nil beside real target",
			err1, []error{nil, err2}, false, false),
		newCheckErrorIsIn2TestCase("no targets final miss",
			err1, nil, false, true),
		newCheckErrorIsIn2TestCase("no unwrapping",
			fmt.Errorf("wrapped: %w", err1), []error{err1},
			false, false),
	}
}

// TestNewCheckErrorIsIn2 exercises the two-result per-node closure:
// a match is final, a miss is final only when no target could match
// deeper nodes either.
func TestNewCheckErrorIsIn2(t *testing.T) {
	core.RunTestCases(t, checkErrorIsIn2TestCases())
}

// TestNewCheckErrorIsIn2Compose pins composition with IsErrorFn2: a
// per-node miss is not final, so traversal descends into wrapped
// errors and still finds the match.
func TestNewCheckErrorIsIn2Compose(t *testing.T) {
	target := errors.New("target")
	check := core.NewCheckErrorIsIn2([]error{target})
	is, known := core.IsErrorFn2(check, fmt.Errorf("wrapped: %w", target))
	core.AssertTrue(t, is, "is")
	core.AssertTrue(t, known, "known")
}

// Test cases for TemporaryError constructors and methods
type temporaryErrorTestCase struct {
	err       error
	testFunc  func(error) error
	name      string
	expected  string
	isTimeout bool
}

func newTemporaryErrorTestCase(name string, err error, expected string, isTimeout bool,
	testFunc func(error) error) temporaryErrorTestCase {
	return temporaryErrorTestCase{
		name:      name,
		err:       err,
		expected:  expected,
		isTimeout: isTimeout,
		testFunc:  testFunc,
	}
}

func (tc temporaryErrorTestCase) Name() string {
	return tc.name
}

func (tc temporaryErrorTestCase) Test(t *testing.T) {
	t.Helper()

	result := tc.testFunc(tc.err)

	// Test error message
	core.AssertEqual(t, tc.expected, result.Error(), "message")

	// Test interface methods
	if tempErr, ok := result.(interface{ IsTemporary() bool }); ok {
		core.AssertTrue(t, tempErr.IsTemporary(), "IsTemporary result")
	}

	if timeoutErr, ok := result.(interface{ IsTimeout() bool }); ok {
		core.AssertEqual(t, tc.isTimeout, timeoutErr.IsTimeout(), "IsTimeout result")
	}

	// Test legacy methods
	if tempErr, ok := result.(interface{ Temporary() bool }); ok {
		core.AssertTrue(t, tempErr.Temporary(), "Temporary result")
	}

	if timeoutErr, ok := result.(interface{ Timeout() bool }); ok {
		core.AssertEqual(t, tc.isTimeout, timeoutErr.Timeout(), "Timeout result")
	}
}

func TestNewTimeoutError(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := core.S(
		newTemporaryErrorTestCase("nil error", nil, "time-out", true, core.NewTimeoutError),
		newTemporaryErrorTestCase("with cause", baseErr, "time-out: base error", true, core.NewTimeoutError),
		newTemporaryErrorTestCase("empty cause", errors.New(emptyString), "time-out", true, core.NewTimeoutError),
	)

	core.RunTestCases(t, testCases)
}

func TestNewTemporaryError(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := core.S(
		newTemporaryErrorTestCase("nil error", nil, "", false, core.NewTemporaryError),
		newTemporaryErrorTestCase("with cause", baseErr, "base error", false, core.NewTemporaryError),
		newTemporaryErrorTestCase("empty cause", errors.New(emptyString), "", false, core.NewTemporaryError),
	)

	core.RunTestCases(t, testCases)
}

// Test for TemporaryError with nil receiver
func TestTemporaryErrorNilReceiver(t *testing.T) {
	var tempErr *core.TemporaryError

	core.AssertEqual(t, "", tempErr.Error(), "nil TemporaryError should return empty string")
	core.AssertFalse(t, tempErr.IsTimeout(), "nil TemporaryError IsTimeout")
	core.AssertFalse(t, tempErr.Timeout(), "nil TemporaryError Timeout")
}

// permanentErrorTestCase wraps a cause with NewPermanentError, and checks
// the wrapper answers neither temporary nor a timeout, alone and further
// down a chain, while it still unwraps to the cause.
type permanentErrorTestCase struct {
	cause error
	name  string
}

// newPermanentOverTemporaryCase wraps a cause that reports itself
// temporary but not a timeout.
func newPermanentOverTemporaryCase(name string, cause error) permanentErrorTestCase {
	if !core.IsTemporary(cause) || core.IsTimeout(cause) {
		panic("permanentErrorTestCase: every temporary row must wrap a temporary error that is not a timeout")
	}

	return permanentErrorTestCase{
		cause: cause,
		name:  name,
	}
}

// newPermanentOverTimeoutCase wraps a cause that reports itself a
// timeout.
func newPermanentOverTimeoutCase(name string, cause error) permanentErrorTestCase {
	if !core.IsTimeout(cause) {
		panic("permanentErrorTestCase: every timeout row must wrap a timeout")
	}

	return permanentErrorTestCase{
		cause: cause,
		name:  name,
	}
}

// newPermanentOverPlainCase wraps a cause that reports itself neither
// temporary nor a timeout.
func newPermanentOverPlainCase(name string, cause error) permanentErrorTestCase {
	if cause == nil || core.IsTemporary(cause) || core.IsTimeout(cause) {
		panic("permanentErrorTestCase: every plain row must wrap an error that is neither temporary nor a timeout")
	}

	return permanentErrorTestCase{
		cause: cause,
		name:  name,
	}
}

func (tc permanentErrorTestCase) Name() string {
	return tc.name
}

func (tc permanentErrorTestCase) Test(t *testing.T) {
	t.Helper()

	err := core.NewPermanentError(tc.cause)
	pe := core.AssertMustTypeIs[*core.PermanentError](t, err, "type")
	core.AssertEqual(t, tc.cause.Error(), pe.Error(), "message")
	core.AssertEqual(t, tc.cause, pe.Unwrap(), "Unwrap")
	core.AssertErrorIs(t, err, tc.cause, "errors.Is")

	// errors.As finds in the wrapper what it finds in the cause
	var wantAs, gotAs *core.TemporaryError
	wantOK := errors.As(tc.cause, &wantAs)
	gotOK := errors.As(err, &gotAs)
	core.AssertEqual(t, wantOK, gotOK, "errors.As found")
	core.AssertSame(t, wantAs, gotAs, "errors.As target")

	assertPermanentAnswers(t, pe)

	// further down a chain, the walk still stops at the wrapper
	chained := fmt.Errorf("chained: %w", err)
	core.AssertFalse(t, core.IsTemporary(chained), "chained IsTemporary")
	core.AssertFalse(t, core.IsTimeout(chained), "chained IsTimeout")
}

func permanentErrorTestCases() []permanentErrorTestCase {
	timeoutErr := core.NewTimeoutError(errors.New("i/o timeout"))

	return core.S(
		newPermanentOverTemporaryCase("temporary error",
			core.NewTemporaryError(errors.New("too many open files"))),
		newPermanentOverTimeoutCase("timeout error", timeoutErr),
		newPermanentOverTimeoutCase("wrapped timeout error",
			fmt.Errorf("read: %w", timeoutErr)),
		newPermanentOverTimeoutCase("context deadline", context.DeadlineExceeded),
		newPermanentOverPlainCase("plain error", errors.New("closed")),
		newPermanentOverPlainCase("typed nil error", (*typedNilError)(nil)),
	)
}

func TestNewPermanentError(t *testing.T) {
	core.RunTestCases(t, permanentErrorTestCases())
}

// TestNewPermanentErrorNil checks NewPermanentError returns a nil error,
// not a typed nil, for a nil err.
func TestNewPermanentErrorNil(t *testing.T) {
	core.AssertNoError(t, core.NewPermanentError(nil), "NewPermanentError(nil)")
}

// permanentErrorNilCauseTestCase checks a PermanentError with no cause,
// a nil receiver or the zero value: it has no text, unwraps to nil, and
// answers as any PermanentError does.
type permanentErrorNilCauseTestCase struct {
	pe   *core.PermanentError
	name string
}

// newPermanentNilCauseCase takes a PermanentError that has no cause.
func newPermanentNilCauseCase(name string, pe *core.PermanentError) permanentErrorNilCauseTestCase {
	if pe != nil && pe.Unwrap() != nil {
		panic("permanentErrorNilCauseTestCase: every row must have no cause")
	}

	return permanentErrorNilCauseTestCase{
		pe:   pe,
		name: name,
	}
}

func (tc permanentErrorNilCauseTestCase) Name() string {
	return tc.name
}

func (tc permanentErrorNilCauseTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, "", tc.pe.Error(), "message")
	core.AssertNil(t, tc.pe.Unwrap(), "Unwrap")
	assertPermanentAnswers(t, tc.pe)
}

func TestPermanentErrorNilCause(t *testing.T) {
	core.RunTestCases(t, core.S(
		newPermanentNilCauseCase("nil receiver", nil),
		newPermanentNilCauseCase("zero value", &core.PermanentError{}),
	))
}

// assertPermanentAnswers fails unless pe answers neither temporary nor a
// timeout through any of its methods, and core's checks take that as a
// known answer, alone and walking its chain.
func assertPermanentAnswers(t *testing.T, pe *core.PermanentError) {
	t.Helper()

	core.AssertFalse(t, pe.Temporary(), "Temporary")
	core.AssertFalse(t, pe.IsTemporary(), "IsTemporary method")
	core.AssertFalse(t, pe.Timeout(), "Timeout")
	core.AssertFalse(t, pe.IsTimeout(), "IsTimeout method")

	is, known := core.CheckIsTemporary(pe)
	core.AssertFalse(t, is, "CheckIsTemporary is")
	core.AssertTrue(t, known, "CheckIsTemporary known")
	is, known = core.CheckIsTimeout(pe)
	core.AssertFalse(t, is, "CheckIsTimeout is")
	core.AssertTrue(t, known, "CheckIsTimeout known")

	core.AssertFalse(t, core.IsTemporary(pe), "IsTemporary")
	core.AssertFalse(t, core.IsTimeout(pe), "IsTimeout")
}

// Test case for CheckIsTemporary function
type checkIsTemporaryTestCase struct {
	err           error
	name          string
	expectedIs    bool
	expectedKnown bool
}

func (tc checkIsTemporaryTestCase) Name() string {
	return tc.name
}

func (tc checkIsTemporaryTestCase) Test(t *testing.T) {
	t.Helper()
	is, known := core.CheckIsTemporary(tc.err)
	core.AssertEqual(t, tc.expectedIs, is, "CheckIsTemporary is result")
	core.AssertEqual(t, tc.expectedKnown, known, "CheckIsTemporary known result")
}

func newCheckIsTemporaryTestCase(name string, err error, expectedIs, expectedKnown bool) checkIsTemporaryTestCase {
	return checkIsTemporaryTestCase{
		err:           err,
		name:          name,
		expectedIs:    expectedIs,
		expectedKnown: expectedKnown,
	}
}

func checkIsTemporaryTestCases() []checkIsTemporaryTestCase {
	tempErr := core.NewTemporaryError(errors.New("temp error"))
	timeoutErr := core.NewTimeoutError(errors.New("timeout error"))
	regularErr := errors.New("regular error")

	return core.S(
		newCheckIsTemporaryTestCase("nil error", nil, false, true),
		newCheckIsTemporaryTestCase("temporary error", tempErr, true, true),
		newCheckIsTemporaryTestCase("timeout error", timeoutErr, true, true),
		newCheckIsTemporaryTestCase("regular error", regularErr, false, false),
		newCheckIsTemporaryTestCase("IsTemporary-only interface",
			&isTemporaryOnlyError{}, true, true),
	)
}

// Test CheckIsTemporary function (0% coverage)
func TestCheckIsTemporary(t *testing.T) {
	core.RunTestCases(t, checkIsTemporaryTestCases())
}

// Test case for IsTemporary function
type isTemporaryTestCase struct {
	err      error
	name     string
	expected bool
}

func (tc isTemporaryTestCase) Name() string {
	return tc.name
}

func (tc isTemporaryTestCase) Test(t *testing.T) {
	t.Helper()
	result := core.IsTemporary(tc.err)
	core.AssertEqual(t, tc.expected, result, "IsTemporary result")
}

func newIsTemporaryTestCase(name string, err error, expected bool) isTemporaryTestCase {
	return isTemporaryTestCase{
		err:      err,
		name:     name,
		expected: expected,
	}
}

func isTemporaryTestCases() []isTemporaryTestCase {
	tempErr := core.NewTemporaryError(errors.New("temp error"))
	timeoutErr := core.NewTimeoutError(errors.New("timeout error"))
	wrappedTempErr := fmt.Errorf("wrapped: %w", tempErr)
	regularErr := errors.New("regular error")

	return core.S(
		newIsTemporaryTestCase("nil error", nil, false),
		newIsTemporaryTestCase("temporary error", tempErr, true),
		newIsTemporaryTestCase("timeout error", timeoutErr, true),
		newIsTemporaryTestCase("wrapped temporary error", wrappedTempErr, true),
		newIsTemporaryTestCase("regular error", regularErr, false),
	)
}

// Test IsTemporary function (0% coverage)
func TestIsTemporary(t *testing.T) {
	core.RunTestCases(t, isTemporaryTestCases())
}

// Test case for CheckIsTimeout function
type checkIsTimeoutTestCase struct {
	err           error
	name          string
	expectedIs    bool
	expectedKnown bool
}

func (tc checkIsTimeoutTestCase) Name() string {
	return tc.name
}

func (tc checkIsTimeoutTestCase) Test(t *testing.T) {
	t.Helper()
	is, known := core.CheckIsTimeout(tc.err)
	core.AssertEqual(t, tc.expectedIs, is, "CheckIsTimeout is result")
	core.AssertEqual(t, tc.expectedKnown, known, "CheckIsTimeout known result")
}

func newCheckIsTimeoutTestCase(name string, err error, expectedIs, expectedKnown bool) checkIsTimeoutTestCase {
	return checkIsTimeoutTestCase{
		err:           err,
		name:          name,
		expectedIs:    expectedIs,
		expectedKnown: expectedKnown,
	}
}

func checkIsTimeoutTestCases() []checkIsTimeoutTestCase {
	tempErr := core.NewTemporaryError(errors.New("temp error"))
	timeoutErr := core.NewTimeoutError(errors.New("timeout error"))
	regularErr := errors.New("regular error")

	return core.S(
		newCheckIsTimeoutTestCase("nil error", nil, false, true),
		newCheckIsTimeoutTestCase("temporary error", tempErr, false, true),
		newCheckIsTimeoutTestCase("timeout error", timeoutErr, true, true),
		newCheckIsTimeoutTestCase("regular error", regularErr, false, false),
		newCheckIsTimeoutTestCase("IsTimeout-only interface",
			&isTimeoutOnlyError{}, true, true),
	)
}

// Test CheckIsTimeout function (0% coverage)
func TestCheckIsTimeout(t *testing.T) {
	core.RunTestCases(t, checkIsTimeoutTestCases())
}

// Test case for IsTimeout function
type isTimeoutTestCase struct {
	err      error
	name     string
	expected bool
}

func (tc isTimeoutTestCase) Name() string {
	return tc.name
}

func (tc isTimeoutTestCase) Test(t *testing.T) {
	t.Helper()
	result := core.IsTimeout(tc.err)
	core.AssertEqual(t, tc.expected, result, "IsTimeout result")
}

func newIsTimeoutTestCase(name string, err error, expected bool) isTimeoutTestCase {
	return isTimeoutTestCase{
		err:      err,
		name:     name,
		expected: expected,
	}
}

func isTimeoutTestCases() []isTimeoutTestCase {
	tempErr := core.NewTemporaryError(errors.New("temp error"))
	timeoutErr := core.NewTimeoutError(errors.New("timeout error"))
	wrappedTimeoutErr := fmt.Errorf("wrapped: %w", timeoutErr)
	regularErr := errors.New("regular error")

	return core.S(
		newIsTimeoutTestCase("nil error", nil, false),
		newIsTimeoutTestCase("temporary error", tempErr, false),
		newIsTimeoutTestCase("timeout error", timeoutErr, true),
		newIsTimeoutTestCase("wrapped timeout error", wrappedTimeoutErr, true),
		newIsTimeoutTestCase("regular error", regularErr, false),
	)
}

// Test IsTimeout function (0% coverage)
func TestIsTimeout(t *testing.T) {
	core.RunTestCases(t, isTimeoutTestCases())
}

// emptyMessageError returns "" from Error() so that WrappedError.Error()
// takes the `s == ""` branch and falls back to the wrapper's note.
type emptyMessageError struct{}

func (emptyMessageError) Error() string { return "" }

// Types that implement only one of Unwrap/Errors/IsTemporary/IsTimeout
// to exercise the non-first branches of the type switches.
type errorsOnlyError struct{ errs []error }

func (*errorsOnlyError) Error() string     { return "errors-only" }
func (e *errorsOnlyError) Errors() []error { return e.errs }

type singleUnwrapError struct{ inner error }

func (*singleUnwrapError) Error() string   { return "single-unwrap" }
func (e *singleUnwrapError) Unwrap() error { return e.inner }

type isTemporaryOnlyError struct{}

func (*isTemporaryOnlyError) Error() string     { return "is-temporary-only" }
func (*isTemporaryOnlyError) IsTemporary() bool { return true }

type isTimeoutOnlyError struct{}

func (*isTimeoutOnlyError) Error() string   { return "is-timeout-only" }
func (*isTimeoutOnlyError) IsTimeout() bool { return true }

// wrappedErrorTestCase exercises WrappedError.Error and Unwrap,
// including the nil-receiver, nil-cause and empty-cause fallback
// branches.
type wrappedErrorTestCase struct {
	err        *core.WrappedError
	wantUnwrap error
	name       string
	wantError  string
}

func newWrappedErrorTestCase(name string, err *core.WrappedError,
	wantError string, wantUnwrap error) wrappedErrorTestCase {
	return wrappedErrorTestCase{
		err:        err,
		name:       name,
		wantError:  wantError,
		wantUnwrap: wantUnwrap,
	}
}

// newWrappedErrorTestCaseWrap declares a row over what Wrap builds from
// cause and note, which unwraps to cause.
func newWrappedErrorTestCaseWrap(name string, cause error,
	note, wantError string) wrappedErrorTestCase {
	err := core.MustT[*core.WrappedError](core.Wrap(cause, note))
	return newWrappedErrorTestCase(name, err, wantError, cause)
}

func (tc wrappedErrorTestCase) Name() string { return tc.name }

func (tc wrappedErrorTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.wantError, tc.err.Error(), "WrappedError.Error")
	if tc.wantUnwrap == nil {
		core.AssertNil(t, tc.err.Unwrap(), "WrappedError.Unwrap")
		return
	}
	core.AssertErrorIs(t, tc.err.Unwrap(), tc.wantUnwrap, "WrappedError.Unwrap")
}

func wrappedErrorTestCases() []wrappedErrorTestCase {
	empty := emptyMessageError{}
	var nilError *core.WrappedError
	return core.S(
		newWrappedErrorTestCase("nil receiver", nil, "", nil),
		newWrappedErrorTestCase("nil cause", &core.WrappedError{}, "", nil),
		newWrappedErrorTestCaseWrap("empty cause", empty, "note", "note"),
		newWrappedErrorTestCaseWrap("typed nil cause", nilError, "note", "note"),
	)
}

// Cover WrappedError method behaviour across its main branches.
func TestWrappedError(t *testing.T) {
	core.RunTestCases(t, wrappedErrorTestCases())
}

// unwrapBranchesTestCase covers the three Unwrap type-switch arms plus
// the default branch, asserting that the returned slice contains the
// expected inner error(s).
type unwrapBranchesTestCase struct {
	err       error
	name      string
	wantInner []error
}

func newUnwrapBranchesTestCase(name string, err error,
	wantInner []error) unwrapBranchesTestCase {
	return unwrapBranchesTestCase{err: err, name: name, wantInner: wantInner}
}

func (tc unwrapBranchesTestCase) Name() string { return tc.name }

func (tc unwrapBranchesTestCase) Test(t *testing.T) {
	t.Helper()
	got := core.Unwrap(tc.err)
	core.AssertSliceEqual(t, tc.wantInner, got, "Unwrap")
}

func unwrapBranchesTestCases() []unwrapBranchesTestCase {
	inner := errors.New("inner")
	return core.S(
		newUnwrapBranchesTestCase("Unwrap []error branch",
			&core.CompoundError{Errs: []error{inner}}, []error{inner}),
		newUnwrapBranchesTestCase("Errors() branch",
			&errorsOnlyError{errs: []error{inner}}, []error{inner}),
		newUnwrapBranchesTestCase("Unwrap error branch",
			&singleUnwrapError{inner: inner}, []error{inner}),
		newUnwrapBranchesTestCase("default branch",
			errors.New("plain"), nil),
	)
}

// Test Unwrap(err) hits all three type-switch branches and the default.
func TestUnwrapBranches(t *testing.T) {
	core.RunTestCases(t, unwrapBranchesTestCases())
}

// Test ErrInvalid is an alias of the standard sentinel, so errors.Is
// matches across the boundary and the message text is unchanged.
func TestErrInvalidAlias(t *testing.T) {
	core.AssertSame(t, fs.ErrInvalid, core.ErrInvalid, "fs.ErrInvalid identity")
	core.AssertErrorIs(t, core.ErrInvalid, os.ErrInvalid, "os.ErrInvalid match")
	core.AssertErrorIs(t, os.ErrInvalid, core.ErrInvalid, "reverse match")
	core.AssertEqual(t, "invalid argument", core.ErrInvalid.Error(), "message")
}

// Test ErrUnsupported is an alias of the standard sentinel, so errors.Is
// matches across the boundary in both directions. A gap in the code is
// not an unsupported operation, so ErrNotImplemented and ErrTODO stay
// out of its chain.
func TestErrUnsupportedAlias(t *testing.T) {
	core.AssertSame(t, errors.ErrUnsupported, core.ErrUnsupported, "errors.ErrUnsupported identity")
	core.AssertErrorIs(t, core.Wrap(errors.ErrUnsupported, "note"), core.ErrUnsupported, "match")
	core.AssertErrorIs(t, core.Wrap(core.ErrUnsupported, "note"), errors.ErrUnsupported, "reverse match")

	core.AssertNotErrorIs(t, core.ErrNotImplemented, core.ErrUnsupported, "ErrNotImplemented apart")
	core.AssertNotErrorIs(t, core.ErrTODO, core.ErrUnsupported, "ErrTODO apart")
}

// Test ErrTODO is an ErrNotImplemented, so errors.Is matches it as one.
func TestErrTODO(t *testing.T) {
	core.AssertErrorIs(t, core.ErrTODO, core.ErrNotImplemented, "ErrNotImplemented match")
}
