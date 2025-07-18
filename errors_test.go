package core

import (
	"errors"
	"fmt"
	"testing"
)

const emptyString = ""

func TestIsErrorFn(t *testing.T) {
	testErr := errors.New("test error")
	differentErr := errors.New("different error")
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)

	isTestErr := func(err error) bool {
		return errors.Is(err, testErr)
	}

	for _, tc := range []struct {
		name     string
		checkFn  func(error) bool
		errs     []error
		expected bool
	}{
		{"matching error", isTestErr, S(testErr), true},
		{"non-matching error", isTestErr, S(differentErr), false},
		{"nil check function", nil, S(testErr), false},
		{"no errors", isTestErr, S[error](), false},
		{"wrapped error", isTestErr, S(wrappedErr), true},
		{"nil error in slice", isTestErr, S[error](nil, testErr), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := IsErrorFn(tc.checkFn, tc.errs...)
			AssertBool(t, result, tc.expected, "IsErrorFn() result")
		})
	}
}

//revive:disable:cognitive-complexity
func TestIsErrorFn2(t *testing.T) {
	//revive:enable:cognitive-complexity
	testErr := errors.New("test error")
	differentErr := errors.New("different error")
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)

	isTestErr := func(err error) (bool, bool) {
		if err == nil {
			return false, false
		}
		return errors.Is(err, testErr), true
	}

	for _, tc := range []struct {
		name          string
		checkFn       func(error) (bool, bool)
		errs          []error
		expectedIs    bool
		expectedKnown bool
	}{
		{"matching error", isTestErr, S(testErr), true, true},
		{"non-matching error", isTestErr, S(differentErr), false, true},
		{"nil check function", nil, S(testErr), false, true},
		{"no errors", isTestErr, S[error](), false, true},
		{"wrapped error", isTestErr, S(wrappedErr), true, true},
		{"unknown error type", func(_ error) (bool, bool) { return false, false }, S(testErr), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			is, known := IsErrorFn2(tc.checkFn, tc.errs...)
			AssertBool(t, is, tc.expectedIs, "IsErrorFn2() first return value")
			AssertBool(t, known, tc.expectedKnown, "IsErrorFn2() second return value")
		})
	}
}

// Test cases for QuietWrap function
type quietWrapTestCase struct {
	err      error
	name     string
	format   string
	expected string
	args     []any
}

func quietWrapTest(name string, err error, format, expected string, args ...any) quietWrapTestCase {
	return quietWrapTestCase{
		name:     name,
		err:      err,
		format:   format,
		expected: expected,
		args:     args,
	}
}

func (tc quietWrapTestCase) test(t *testing.T) {
	t.Helper()

	var result error
	if len(tc.args) > 0 {
		result = QuietWrap(tc.err, tc.format, tc.args...)
	} else {
		result = QuietWrap(tc.err, tc.format)
	}

	if tc.expected == "" {
		AssertEqual(t, result, tc.err, "QuietWrap should return original error")
	} else {
		AssertEqual(t, result.Error(), tc.expected, "QuietWrap error message")

		// Test that it's wrappable
		if wrapped, ok := result.(Unwrappable); ok {
			AssertEqual(t, wrapped.Unwrap(), tc.err, "QuietWrap should wrap original error")
		}
	}
}

func TestQuietWrap(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := S(
		quietWrapTest("nil error", nil, "message", ""),
		quietWrapTest("empty format", baseErr, "", ""),
		quietWrapTest("simple message", baseErr, "wrapped", "wrapped"),
		quietWrapTest("formatted message", baseErr, "wrapped: %s", "wrapped: test", "test"),
		quietWrapTest("multiple args", baseErr, "error %d: %s", "error 42: test", 42, "test"),
		quietWrapTest("nil error with format", nil, "message %s", "", "test"),
	)

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for CoalesceError function
type coalesceErrorTestCase struct {
	expected error
	name     string
	errs     []error
}

func coalesceErrorTest(name string, expected error, errs ...error) coalesceErrorTestCase {
	return coalesceErrorTestCase{
		name:     name,
		expected: expected,
		errs:     errs,
	}
}

func (tc coalesceErrorTestCase) test(t *testing.T) {
	t.Helper()

	result := CoalesceError(tc.errs...)
	AssertEqual(t, result, tc.expected, "CoalesceError result")
}

func TestCoalesceError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	testCases := S(
		coalesceErrorTest("no errors", nil),
		coalesceErrorTest("single nil error", nil, nil),
		coalesceErrorTest("single non-nil error", err1, err1),
		coalesceErrorTest("first non-nil wins", err1, nil, err1, err2),
		coalesceErrorTest("first error wins", err1, err1, err2, err3),
		coalesceErrorTest("all nil", nil, nil, nil, nil),
		coalesceErrorTest("nil then non-nil", err2, nil, nil, err2),
		coalesceErrorTest("empty slice", nil),
	)

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for IsError function
type isErrorTestCase struct {
	name     string
	err      error
	errs     []error
	expected bool
}

func (tc isErrorTestCase) test(t *testing.T) {
	t.Helper()

	result := IsError(tc.err, tc.errs...)
	AssertBool(t, result, tc.expected, "IsError result")
}

func TestIsError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	wrappedErr := fmt.Errorf("wrapped: %w", err1)

	testCases := []isErrorTestCase{
		{"nil error", nil, S(err1), false},
		{"no target errors - non-nil", err1, S[error](), true},
		{"no target errors - nil", nil, S[error](), false},
		{"exact match", err1, S(err1, err2), true},
		{"no match", err3, S(err1, err2), false},
		{"wrapped error match", wrappedErr, S(err1), true},
		{"wrapped error no match", wrappedErr, S(err2), false},
		{"multiple targets", err2, S(err1, err2, err3), true},
		{"empty targets", err1, nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for TemporaryError constructors and methods
type temporaryErrorTestCase struct {
	err       error
	testFunc  func(error) error
	name      string
	expected  string
	isTimeout bool
}

func temporaryErrorTest(name string, err error, expected string, isTimeout bool,
	testFunc func(error) error) temporaryErrorTestCase {
	return temporaryErrorTestCase{
		name:      name,
		err:       err,
		expected:  expected,
		isTimeout: isTimeout,
		testFunc:  testFunc,
	}
}

func (tc temporaryErrorTestCase) test(t *testing.T) {
	t.Helper()

	result := tc.testFunc(tc.err)

	// Test error message
	AssertEqual(t, result.Error(), tc.expected, "TemporaryError message")

	// Test interface methods
	if tempErr, ok := result.(interface{ IsTemporary() bool }); ok {
		AssertBool(t, tempErr.IsTemporary(), true, "IsTemporary() should return true")
	}

	if timeoutErr, ok := result.(interface{ IsTimeout() bool }); ok {
		AssertBool(t, timeoutErr.IsTimeout(), tc.isTimeout, "IsTimeout() result")
	}

	// Test legacy methods
	if tempErr, ok := result.(interface{ Temporary() bool }); ok {
		AssertBool(t, tempErr.Temporary(), true, "Temporary() should return true")
	}

	if timeoutErr, ok := result.(interface{ Timeout() bool }); ok {
		AssertBool(t, timeoutErr.Timeout(), tc.isTimeout, "Timeout() result")
	}
}

func TestNewTimeoutError(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := S(
		temporaryErrorTest("nil error", nil, "time-out", true, NewTimeoutError),
		temporaryErrorTest("with cause", baseErr, "time-out: base error", true, NewTimeoutError),
		temporaryErrorTest("empty cause", errors.New(emptyString), "time-out", true, NewTimeoutError),
	)

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestNewTemporaryError(t *testing.T) {
	baseErr := errors.New("base error")

	testCases := S(
		temporaryErrorTest("nil error", nil, "", false, NewTemporaryError),
		temporaryErrorTest("with cause", baseErr, "base error", false, NewTemporaryError),
		temporaryErrorTest("empty cause", errors.New(emptyString), "", false, NewTemporaryError),
	)

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// Test for TemporaryError with nil receiver
func TestTemporaryErrorNilReceiver(t *testing.T) {
	var tempErr *TemporaryError

	AssertEqual(t, tempErr.Error(), "", "nil TemporaryError should return empty string")
	AssertBool(t, tempErr.IsTimeout(), false, "nil TemporaryError IsTimeout should return false")
	AssertBool(t, tempErr.Timeout(), false, "nil TemporaryError Timeout should return false")
}

// Test case for CheckIsTemporary function
type checkIsTemporaryTestCase struct {
	err           error
	name          string
	expectedIs    bool
	expectedKnown bool
}

func (tc checkIsTemporaryTestCase) test(t *testing.T) {
	t.Helper()
	is, known := CheckIsTemporary(tc.err)
	AssertBool(t, is, tc.expectedIs, "CheckIsTemporary is result")
	AssertBool(t, known, tc.expectedKnown, "CheckIsTemporary known result")
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
	for _, tc := range checkIsTemporaryTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// Test case for IsTemporary function
type isTemporaryTestCase struct {
	err      error
	name     string
	expected bool
}

func (tc isTemporaryTestCase) test(t *testing.T) {
	t.Helper()
	result := IsTemporary(tc.err)
	AssertBool(t, result, tc.expected, "IsTemporary result")
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
	for _, tc := range isTemporaryTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// Test case for CheckIsTimeout function
type checkIsTimeoutTestCase struct {
	err           error
	name          string
	expectedIs    bool
	expectedKnown bool
}

func (tc checkIsTimeoutTestCase) test(t *testing.T) {
	t.Helper()
	is, known := CheckIsTimeout(tc.err)
	AssertBool(t, is, tc.expectedIs, "CheckIsTimeout is result")
	AssertBool(t, known, tc.expectedKnown, "CheckIsTimeout known result")
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
	for _, tc := range checkIsTimeoutTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// Test case for IsTimeout function
type isTimeoutTestCase struct {
	err      error
	name     string
	expected bool
}

func (tc isTimeoutTestCase) test(t *testing.T) {
	t.Helper()
	result := IsTimeout(tc.err)
	AssertBool(t, result, tc.expected, "IsTimeout result")
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
	for _, tc := range isTimeoutTestCases() {
		t.Run(tc.name, tc.test)
	}
}
