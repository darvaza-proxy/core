package core

import (
	"errors"
	"fmt"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = quietWrapTestCase{}
var _ TestCase = coalesceErrorTestCase{}
var _ TestCase = isErrorTestCase{}
var _ TestCase = isErrorFnTestCase{}
var _ TestCase = isErrorFn2TestCase{}
var _ TestCase = temporaryErrorTestCase{}
var _ TestCase = checkIsTemporaryTestCase{}
var _ TestCase = isTemporaryTestCase{}
var _ TestCase = checkIsTimeoutTestCase{}
var _ TestCase = isTimeoutTestCase{}

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
	result := IsErrorFn(tc.checkFn, tc.errs...)
	AssertEqual(t, tc.expected, result, "IsErrorFn result")
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

	RunTestCases(t, testCases)
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
	is, known := IsErrorFn2(tc.checkFn, tc.errs...)
	AssertEqual(t, tc.expectedIs, is, "IsErrorFn2 is result")
	AssertEqual(t, tc.expectedKnown, known, "IsErrorFn2 known result")
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

	RunTestCases(t, testCases)
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

	var result error
	if len(tc.args) > 0 {
		result = QuietWrap(tc.err, tc.format, tc.args...)
	} else {
		result = QuietWrap(tc.err, tc.format)
	}

	if tc.expected == "" {
		AssertEqual(t, tc.err, result, "QuietWrap original")
	} else {
		AssertEqual(t, tc.expected, result.Error(), "QuietWrap message")

		// Test that it's wrappable
		if wrapped, ok := result.(Unwrappable); ok {
			AssertEqual(t, tc.err, wrapped.Unwrap(), "QuietWrap unwrap")
		}
	}
}

func TestQuietWrap(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := S(
		newQuietWrapTestCase("nil error", nil, "message", ""),
		newQuietWrapTestCase("empty format", baseErr, "", ""),
		newQuietWrapTestCase("simple message", baseErr, "wrapped", "wrapped"),
		newQuietWrapTestCase("formatted message", baseErr, "wrapped: %s", "wrapped: test", "test"),
		newQuietWrapTestCase("multiple args", baseErr, "error %d: %s", "error 42: test", 42, "test"),
		newQuietWrapTestCase("nil error with format", nil, "message %s", "", "test"),
	)

	RunTestCases(t, testCases)
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

	result := CoalesceError(tc.errs...)
	AssertEqual(t, tc.expected, result, "CoalesceError")
}

func TestCoalesceError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	testCases := S(
		newCoalesceErrorTestCase("no errors", nil),
		newCoalesceErrorTestCase("single nil error", nil, nil),
		newCoalesceErrorTestCase("single non-nil error", err1, err1),
		newCoalesceErrorTestCase("first non-nil wins", err1, nil, err1, err2),
		newCoalesceErrorTestCase("first error wins", err1, err1, err2, err3),
		newCoalesceErrorTestCase("all nil", nil, nil, nil, nil),
		newCoalesceErrorTestCase("nil then non-nil", err2, nil, nil, err2),
		newCoalesceErrorTestCase("empty slice", nil),
	)

	RunTestCases(t, testCases)
}

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

	result := IsError(tc.err, tc.errs...)
	AssertEqual(t, tc.expected, result, "IsError result")
}

func TestIsError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	wrappedErr := fmt.Errorf("wrapped: %w", err1)

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
	}

	RunTestCases(t, testCases)
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
	AssertEqual(t, tc.expected, result.Error(), "message")

	// Test interface methods
	if tempErr, ok := result.(interface{ IsTemporary() bool }); ok {
		AssertTrue(t, tempErr.IsTemporary(), "IsTemporary result")
	}

	if timeoutErr, ok := result.(interface{ IsTimeout() bool }); ok {
		AssertEqual(t, tc.isTimeout, timeoutErr.IsTimeout(), "IsTimeout result")
	}

	// Test legacy methods
	if tempErr, ok := result.(interface{ Temporary() bool }); ok {
		AssertTrue(t, tempErr.Temporary(), "Temporary result")
	}

	if timeoutErr, ok := result.(interface{ Timeout() bool }); ok {
		AssertEqual(t, tc.isTimeout, timeoutErr.Timeout(), "Timeout result")
	}
}

func TestNewTimeoutError(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := S(
		newTemporaryErrorTestCase("nil error", nil, "time-out", true, NewTimeoutError),
		newTemporaryErrorTestCase("with cause", baseErr, "time-out: base error", true, NewTimeoutError),
		newTemporaryErrorTestCase("empty cause", errors.New(emptyString), "time-out", true, NewTimeoutError),
	)

	RunTestCases(t, testCases)
}

func TestNewTemporaryError(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := S(
		newTemporaryErrorTestCase("nil error", nil, "", false, NewTemporaryError),
		newTemporaryErrorTestCase("with cause", baseErr, "base error", false, NewTemporaryError),
		newTemporaryErrorTestCase("empty cause", errors.New(emptyString), "", false, NewTemporaryError),
	)

	RunTestCases(t, testCases)
}

// Test for TemporaryError with nil receiver
func TestTemporaryErrorNilReceiver(t *testing.T) {
	var tempErr *TemporaryError

	AssertEqual(t, "", tempErr.Error(), "nil TemporaryError should return empty string")
	AssertFalse(t, tempErr.IsTimeout(), "nil TemporaryError IsTimeout")
	AssertFalse(t, tempErr.Timeout(), "nil TemporaryError Timeout")
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
	is, known := CheckIsTemporary(tc.err)
	AssertEqual(t, tc.expectedIs, is, "CheckIsTemporary is result")
	AssertEqual(t, tc.expectedKnown, known, "CheckIsTemporary known result")
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
	tempErr := NewTemporaryError(errors.New("temp error"))
	timeoutErr := NewTimeoutError(errors.New("timeout error"))
	regularErr := errors.New("regular error")

	return S(
		newCheckIsTemporaryTestCase("nil error", nil, false, true),
		newCheckIsTemporaryTestCase("temporary error", tempErr, true, true),
		newCheckIsTemporaryTestCase("timeout error", timeoutErr, true, true),
		newCheckIsTemporaryTestCase("regular error", regularErr, false, false),
	)
}

// Test CheckIsTemporary function (0% coverage)
func TestCheckIsTemporary(t *testing.T) {
	RunTestCases(t, checkIsTemporaryTestCases())
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
	result := IsTemporary(tc.err)
	AssertEqual(t, tc.expected, result, "IsTemporary result")
}

func newIsTemporaryTestCase(name string, err error, expected bool) isTemporaryTestCase {
	return isTemporaryTestCase{
		err:      err,
		name:     name,
		expected: expected,
	}
}

func isTemporaryTestCases() []isTemporaryTestCase {
	tempErr := NewTemporaryError(errors.New("temp error"))
	timeoutErr := NewTimeoutError(errors.New("timeout error"))
	wrappedTempErr := fmt.Errorf("wrapped: %w", tempErr)
	regularErr := errors.New("regular error")

	return S(
		newIsTemporaryTestCase("nil error", nil, false),
		newIsTemporaryTestCase("temporary error", tempErr, true),
		newIsTemporaryTestCase("timeout error", timeoutErr, true),
		newIsTemporaryTestCase("wrapped temporary error", wrappedTempErr, true),
		newIsTemporaryTestCase("regular error", regularErr, false),
	)
}

// Test IsTemporary function (0% coverage)
func TestIsTemporary(t *testing.T) {
	RunTestCases(t, isTemporaryTestCases())
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
	is, known := CheckIsTimeout(tc.err)
	AssertEqual(t, tc.expectedIs, is, "CheckIsTimeout is result")
	AssertEqual(t, tc.expectedKnown, known, "CheckIsTimeout known result")
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
	tempErr := NewTemporaryError(errors.New("temp error"))
	timeoutErr := NewTimeoutError(errors.New("timeout error"))
	regularErr := errors.New("regular error")

	return S(
		newCheckIsTimeoutTestCase("nil error", nil, false, true),
		newCheckIsTimeoutTestCase("temporary error", tempErr, false, true),
		newCheckIsTimeoutTestCase("timeout error", timeoutErr, true, true),
		newCheckIsTimeoutTestCase("regular error", regularErr, false, false),
	)
}

// Test CheckIsTimeout function (0% coverage)
func TestCheckIsTimeout(t *testing.T) {
	RunTestCases(t, checkIsTimeoutTestCases())
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
	result := IsTimeout(tc.err)
	AssertEqual(t, tc.expected, result, "IsTimeout result")
}

func newIsTimeoutTestCase(name string, err error, expected bool) isTimeoutTestCase {
	return isTimeoutTestCase{
		err:      err,
		name:     name,
		expected: expected,
	}
}

func isTimeoutTestCases() []isTimeoutTestCase {
	tempErr := NewTemporaryError(errors.New("temp error"))
	timeoutErr := NewTimeoutError(errors.New("timeout error"))
	wrappedTimeoutErr := fmt.Errorf("wrapped: %w", timeoutErr)
	regularErr := errors.New("regular error")

	return S(
		newIsTimeoutTestCase("nil error", nil, false),
		newIsTimeoutTestCase("temporary error", tempErr, false),
		newIsTimeoutTestCase("timeout error", timeoutErr, true),
		newIsTimeoutTestCase("wrapped timeout error", wrappedTimeoutErr, true),
		newIsTimeoutTestCase("regular error", regularErr, false),
	)
}

// Test IsTimeout function (0% coverage)
func TestIsTimeout(t *testing.T) {
	RunTestCases(t, isTimeoutTestCases())
}
