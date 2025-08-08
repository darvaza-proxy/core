package core

import (
	"errors"
	"strings"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = compoundErrorErrorTestCase{}
	_ TestCase = compoundErrorOKTestCase{}
	_ TestCase = compoundErrorAsErrorTestCase{}
	_ TestCase = compoundErrorAppendErrorTestCase{}
	_ TestCase = compoundErrorAppendTestCase{}
)

type compoundErrorErrorTestCase struct {
	expected string
	name     string
	errs     []error
}

// newCompoundErrorErrorTestCase creates a new compoundErrorErrorTestCase
func newCompoundErrorErrorTestCase(name string, errs []error, expected string) compoundErrorErrorTestCase {
	return compoundErrorErrorTestCase{
		name:     name,
		errs:     errs,
		expected: expected,
	}
}

// newCompoundErrorErrorTestCaseEmpty creates a test case for empty errors
func newCompoundErrorErrorTestCaseEmpty(name string) compoundErrorErrorTestCase {
	return newCompoundErrorErrorTestCase(name, S[error](), "")
}

var compoundErrorErrorTestCases = []compoundErrorErrorTestCase{
	newCompoundErrorErrorTestCaseEmpty("empty errors"),
	newCompoundErrorErrorTestCase("single error", S(errors.New("first error")), "first error"),
	newCompoundErrorErrorTestCase("multiple errors",
		S(errors.New("first error"), errors.New("second error")),
		"first error\nsecond error"),
	newCompoundErrorErrorTestCase("with nil errors",
		S(errors.New("first error"), nil, errors.New("third error")),
		"first error\nthird error"),
	newCompoundErrorErrorTestCase("all nil errors", S[error](nil, nil, nil), ""),
}

func (tc compoundErrorErrorTestCase) Name() string {
	return tc.name
}

func (tc compoundErrorErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.errs}
	result := ce.Error()

	if !AssertEqual(t, tc.expected, result, "error string") {
		return
	}
}

func TestCompoundErrorError(t *testing.T) {
	RunTestCases(t, compoundErrorErrorTestCases)
}

func TestCompoundErrorErrors(t *testing.T) {
	errs := S(
		errors.New("first error"),
		errors.New("second error"),
	)

	ce := &CompoundError{Errs: errs}
	result := ce.Errors()

	if !AssertEqual(t, len(errs), len(result), "error count") {
		return
	}

	for i, err := range result {
		if !AssertSame(t, errs[i], err, "error %d", i) {
			return
		}
	}
}

func TestCompoundErrorUnwrap(t *testing.T) {
	errs := S(
		errors.New("first error"),
		errors.New("second error"),
	)

	ce := &CompoundError{Errs: errs}
	result := ce.Unwrap()

	if !AssertEqual(t, len(errs), len(result), "error count") {
		return
	}

	for i, err := range result {
		if !AssertSame(t, errs[i], err, "error %d", i) {
			return
		}
	}
}

type compoundErrorOKTestCase struct {
	name     string
	errs     []error
	expected bool
}

// newCompoundErrorOKTestCase creates a new compoundErrorOKTestCase
func newCompoundErrorOKTestCase(name string, errs []error, expected bool) compoundErrorOKTestCase {
	return compoundErrorOKTestCase{
		name:     name,
		errs:     errs,
		expected: expected,
	}
}

// newCompoundErrorOKTestCaseEmpty creates a test case expecting OK() to return true
func newCompoundErrorOKTestCaseEmpty(name string, errs []error) compoundErrorOKTestCase {
	return newCompoundErrorOKTestCase(name, errs, true)
}

// newCompoundErrorOKTestCaseHasErrors creates a test case expecting OK() to return false
func newCompoundErrorOKTestCaseHasErrors(name string, errs []error) compoundErrorOKTestCase {
	return newCompoundErrorOKTestCase(name, errs, false)
}

var compoundErrorOKTestCases = []compoundErrorOKTestCase{
	newCompoundErrorOKTestCaseEmpty("empty errors", S[error]()),
	newCompoundErrorOKTestCaseEmpty("nil slice", nil),
	newCompoundErrorOKTestCaseHasErrors("single error", S(errors.New("error"))),
	newCompoundErrorOKTestCaseHasErrors("multiple errors",
		S(errors.New("first"), errors.New("second"))),
}

func (tc compoundErrorOKTestCase) Name() string {
	return tc.name
}

func (tc compoundErrorOKTestCase) Test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.errs}

	// Test both OK() and deprecated Ok() methods
	resultOK := ce.OK()
	resultOk := ce.Ok()

	AssertEqual(t, tc.expected, resultOK, "OK() method")
	AssertEqual(t, tc.expected, resultOk, "Ok() method")
	AssertEqual(t, resultOK, resultOk, "OK() and Ok() should return same result")
}

func TestCompoundErrorOK(t *testing.T) {
	RunTestCases(t, compoundErrorOKTestCases)
}

type compoundErrorAsErrorTestCase struct {
	name      string
	errs      []error
	expectNil bool
}

// newCompoundErrorAsErrorTestCase creates a new compoundErrorAsErrorTestCase
func newCompoundErrorAsErrorTestCase(name string, errs []error, expectNil bool) compoundErrorAsErrorTestCase {
	return compoundErrorAsErrorTestCase{
		name:      name,
		errs:      errs,
		expectNil: expectNil,
	}
}

var compoundErrorAsErrorTestCases = []compoundErrorAsErrorTestCase{
	newCompoundErrorAsErrorTestCase("empty errors", S[error](), true),
	newCompoundErrorAsErrorTestCase("nil slice", nil, true),
	newCompoundErrorAsErrorTestCase("single error", S(errors.New("error")), false),
	newCompoundErrorAsErrorTestCase("multiple errors", S(errors.New("first"), errors.New("second")), false),
}

func (tc compoundErrorAsErrorTestCase) Name() string {
	return tc.name
}

func (tc compoundErrorAsErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.errs}
	result := ce.AsError()

	if tc.expectNil {
		if !AssertNil(t, result, "result") {
			return
		}
	} else {
		if !AssertNotNil(t, result, "result") {
			return
		}
		AssertSame(t, ce, result, "CompoundError instance")
	}
}

func TestCompoundErrorAsError(t *testing.T) {
	RunTestCases(t, compoundErrorAsErrorTestCases)
}

type compoundErrorAppendErrorTestCase struct {
	name        string
	initial     []error
	toAppend    []error
	expectedLen int
}

// newCompoundErrorAppendErrorTestCase creates a new compoundErrorAppendErrorTestCase
func newCompoundErrorAppendErrorTestCase(name string, initial, toAppend []error,
	expectedLen int) compoundErrorAppendErrorTestCase {
	return compoundErrorAppendErrorTestCase{
		name:        name,
		initial:     initial,
		toAppend:    toAppend,
		expectedLen: expectedLen,
	}
}

var compoundErrorAppendErrorTestCases = []compoundErrorAppendErrorTestCase{
	newCompoundErrorAppendErrorTestCase("append to empty", S[error](), S(errors.New("first")), 1),
	newCompoundErrorAppendErrorTestCase("append multiple", S(errors.New("existing")),
		S(errors.New("first"), errors.New("second")), 3),
	newCompoundErrorAppendErrorTestCase("append with nils", S(errors.New("existing")),
		S[error](nil, errors.New("valid"), nil), 2),
	newCompoundErrorAppendErrorTestCase("append all nils", S(errors.New("existing")), S[error](nil, nil), 1),
}

func (tc compoundErrorAppendErrorTestCase) Name() string {
	return tc.name
}

func (tc compoundErrorAppendErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.initial}
	result := ce.AppendError(tc.toAppend...)

	// Test method chaining
	if !AssertSame(t, ce, result, "CompoundError instance") {
		return
	}

	// Test length
	if !AssertEqual(t, tc.expectedLen, len(ce.Errs), "error count") {
		return
	}

	// Test no nil errors were added
	for i, err := range ce.Errs {
		if !AssertNotNil(t, err, "error %d", i) {
			return
		}
	}
}

func TestCompoundErrorAppendError(t *testing.T) {
	RunTestCases(t, compoundErrorAppendErrorTestCases)
}

func TestCompoundErrorAppendErrorWithCompoundError(t *testing.T) {
	ce1 := &CompoundError{Errs: S(errors.New("first"))}
	ce2 := &CompoundError{Errs: S(errors.New("second"), errors.New("third"))}

	result := ce1.AppendError(ce2)

	// Test method chaining
	if !AssertSame(t, ce1, result, "CompoundError instance") {
		return
	}

	// Should unwrap the compound error and append individual errors
	expectedLen := 3 // original 1 + unwrapped 2
	if !AssertEqual(t, expectedLen, len(ce1.Errs), "error count") {
		return
	}

	// Check error messages
	errorStr := ce1.Error()
	if !AssertContains(t, errorStr, "first", "error string") {
		return
	}
	if !AssertContains(t, errorStr, "second", "error string") {
		return
	}
	if !AssertContains(t, errorStr, "third", "error string") {
		return
	}
}

type mockUnwrappable struct {
	errs []error
}

func (*mockUnwrappable) Error() string {
	return "mock unwrappable error"
}

func (m *mockUnwrappable) Unwrap() []error {
	return m.errs
}

func TestCompoundErrorAppendErrorWithUnwrappable(t *testing.T) {
	ce := &CompoundError{Errs: S(errors.New("initial"))}
	mock := &mockUnwrappable{
		errs: S(errors.New("unwrapped1"), errors.New("unwrapped2")),
	}

	result := ce.AppendError(mock)

	// Test method chaining
	if !AssertSame(t, ce, result, "CompoundError instance") {
		return
	}

	// Should unwrap and append individual errors
	expectedLen := 3 // original 1 + unwrapped 2
	if !AssertEqual(t, expectedLen, len(ce.Errs), "error count") {
		return
	}
}

type compoundErrorAppendTestCase struct {
	name        string
	initial     []error
	err         error
	note        string
	args        []any
	expectedLen int
	expectNote  bool
}

// newCompoundErrorAppendTestCase creates a new compoundErrorAppendTestCase
//
//revive:disable-next-line:argument-limit
func newCompoundErrorAppendTestCase(name string, initial []error, err error, note string,
	args []any, expectedLen int, expectNote bool) compoundErrorAppendTestCase {
	return compoundErrorAppendTestCase{
		name:        name,
		initial:     initial,
		err:         err,
		note:        note,
		args:        args,
		expectedLen: expectedLen,
		expectNote:  expectNote,
	}
}

var compoundErrorAppendTestCases = []compoundErrorAppendTestCase{
	newCompoundErrorAppendTestCase("nil error, empty note", S[error](), nil, "", nil, 0, false),
	newCompoundErrorAppendTestCase("nil error, with note", S[error](), nil, "note only", nil, 1, true),
	newCompoundErrorAppendTestCase("error without note", S[error](), errors.New("test error"), "", nil, 1, false),
	newCompoundErrorAppendTestCase("error with note", S[error](), errors.New("test error"),
		"wrapped note", nil, 1, true),
	newCompoundErrorAppendTestCase("formatted note", S[error](), errors.New("test error"),
		"wrapped %s: %d", S[any]("note", 42), 1, true),
}

func (tc compoundErrorAppendTestCase) Name() string {
	return tc.name
}

//revive:disable-next-line:cognitive-complexity
func (tc compoundErrorAppendTestCase) Test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.initial}
	result := ce.Append(tc.err, tc.note, tc.args...)

	// Test method chaining
	AssertSame(t, ce, result, "CompoundError instance")

	// Test length
	if !AssertEqual(t, tc.expectedLen, len(ce.Errs), "error count") {
		return
	}

	if tc.expectedLen > 0 {
		lastErr := ce.Errs[len(ce.Errs)-1]
		if !AssertNotNil(t, lastErr, "last error") {
			return
		}

		errorStr := lastErr.Error()
		if tc.expectNote {
			if tc.note != "" {
				expectedNote := tc.note
				if len(tc.args) > 0 {
					expectedNote = "wrapped note: 42" // for the formatted case
				}
				foundExpected := strings.Contains(errorStr, expectedNote)
				foundOriginal := strings.Contains(errorStr, tc.note)
				if !AssertTrue(t, foundExpected || foundOriginal, "note in error string") {
					return
				}
			}
		}
	}
}

func TestCompoundErrorAppend(t *testing.T) {
	RunTestCases(t, compoundErrorAppendTestCases)
}

func TestCompoundErrorAppendChaining(t *testing.T) {
	ce := &CompoundError{}

	result := ce.
		Append(errors.New("first"), "").
		Append(nil, "second note").
		AppendError(errors.New("third")).
		Append(errors.New("fourth"), "wrapped %s", "note")

	// Test method chaining returns same instance
	if !AssertSame(t, ce, result, "CompoundError instance") {
		return
	}

	// Test final length
	expectedLen := 4
	if !AssertEqual(t, expectedLen, len(ce.Errs), "error count") {
		return
	}

	// Test that all errors are non-nil
	for i, err := range ce.Errs {
		if !AssertNotNil(t, err, "error %d", i) {
			return
		}
	}
}

func TestCompoundErrorNilHandling(t *testing.T) {
	// Test that nil errors are properly filtered out
	ce := &CompoundError{}

	_ = ce.AppendError(nil, errors.New("valid"), nil)

	if !AssertEqual(t, 1, len(ce.Errs), "error count") {
		return
	}

	if !AssertEqual(t, "valid", ce.Errs[0].Error(), "error message") {
		return
	}
}

func TestCompoundErrorIsInterface(t *testing.T) {
	ce := &CompoundError{Errs: S(errors.New("test"))}

	// Test Errors interface
	var _ Errors = ce

	// Test that it implements error interface
	var _ error = ce

	// Test that Error() method works
	if !AssertTrue(t, ce.Error() != "", "non-empty error string") {
		return
	}

	// Test that Errors() method works
	errs := ce.Errors()
	if !AssertEqual(t, 1, len(errs), "error count") {
		return
	}
}
