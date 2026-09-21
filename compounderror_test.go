package core

import (
	"errors"
	"fmt"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = newCompoundErrorTestCase{}
	_ TestCase = compoundErrorErrorTestCase{}
	_ TestCase = compoundErrorOKTestCase{}
	_ TestCase = compoundErrorAsErrorTestCase{}
	_ TestCase = compoundErrorAppendErrorTestCase{}
	_ TestCase = compoundErrorAppendTestCase{}
	_ TestCase = compoundErrorAppendUnchangedTestCase{}
)

// newCompoundErrorTestCase states what NewCompoundError collects from
// its arguments. want is the exact member list, in order; a row that
// expects nothing passes nil, since a constructor with nothing to
// collect appends nothing.
type newCompoundErrorTestCase struct {
	name string
	errs []error
	want []error
}

func newNewCompoundErrorTestCase(name string, errs, want []error) newCompoundErrorTestCase {
	return newCompoundErrorTestCase{
		name: name,
		errs: errs,
		want: want,
	}
}

func (tc newCompoundErrorTestCase) Name() string {
	return tc.name
}

func (tc newCompoundErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ce := NewCompoundError(tc.errs...)

	AssertMustNotNil(t, ce, "result")
	AssertSliceEqual(t, tc.want, ce.Errs, "members")
	AssertEqual(t, len(ce.Errs) == 0, ce.OK(), "OK matches emptiness")
}

func newCompoundErrorTestCases() []newCompoundErrorTestCase {
	first := errors.New("first error")
	second := errors.New("second error")
	third := errors.New("third error")
	nested := error(&CompoundError{Errs: S(second, third)})

	return []newCompoundErrorTestCase{
		newNewCompoundErrorTestCase("no errors", nil, nil),
		newNewCompoundErrorTestCase("single error",
			S(first), S(first)),
		newNewCompoundErrorTestCase("several errors",
			S(first, second, third), S(first, second, third)),
		newNewCompoundErrorTestCase("nil members dropped",
			S(first, nil, third), S(first, third)),
		newNewCompoundErrorTestCase("all nil", S[error](nil, nil), nil),
		newNewCompoundErrorTestCase("nested list kept as one member",
			S(first, nested), S(first, nested)),
	}
}

func TestNewCompoundError(t *testing.T) {
	RunTestCases(t, newCompoundErrorTestCases())
}

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

// A nil receiver holds no errors, so the three emptiness accessors agree on
// it. The table above cannot state this: its Test body builds the receiver.
// AsError dereferenced it before OK did the guarding for both.
func TestCompoundErrorNilReceiver(t *testing.T) {
	var ce *CompoundError

	AssertTrue(t, ce.OK(), "OK")
	AssertTrue(t, ce.Ok(), "Ok")
	AssertNil(t, ce.AsError(), "AsError")
	AssertNil(t, AsError(ce), "core.AsError")
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

// compoundErrorAppendTestCase tests Append by the one error it appends:
// the row's error wrapped in the note rendered with its arguments, or
// the note alone as an error of its own. A row with neither an error
// nor a rendered note appends nothing.
type compoundErrorAppendTestCase struct {
	err  error
	name string
	note string
	want string
	args []any
}

func (tc compoundErrorAppendTestCase) Name() string {
	return tc.name
}

func (tc compoundErrorAppendTestCase) Test(t *testing.T) {
	t.Helper()
	ce := new(CompoundError)
	result := ce.Append(tc.err, tc.note, tc.args...)

	AssertSame(t, ce, result, "CompoundError instance")
	// With neither an error nor a rendered note, nothing is appended.
	if tc.err == nil && renderAppendNote(tc.note, tc.args) == "" {
		AssertEqual(t, 0, len(ce.Errs), "error count")
		return
	}

	AssertMustEqual(t, 1, len(ce.Errs), "error count")
	AssertMustError(t, ce.Errs[0], "appended error")
	// A note alone is an error of its own, carrying no cause.
	if tc.err != nil {
		AssertErrorIs(t, ce.Errs[0], tc.err, "cause")
	}
	AssertEqual(t, tc.want, ce.Errs[0].Error(), "message")
}

// renderAppendNote renders a note as Append does, formatting it only
// when arguments are given.
func renderAppendNote(note string, args []any) string {
	if len(args) > 0 {
		return fmt.Sprintf(note, args...)
	}
	return note
}

func newCompoundErrorAppendTestCase(name string, err error, note string, args []any,
	want string) compoundErrorAppendTestCase {
	if err == nil {
		panic("compoundErrorAppendTestCase: every wrapping row must pass an error")
	}
	if renderAppendNote(note, args) == "" {
		panic("compoundErrorAppendTestCase: every wrapping row must render a note")
	}
	if want == "" {
		panic("compoundErrorAppendTestCase: every wrapping row must state its message")
	}

	return compoundErrorAppendTestCase{
		err:  err,
		name: name,
		note: note,
		want: want,
		args: args,
	}
}

func newCompoundErrorAppendTestCaseNote(name, note string, args []any,
	want string) compoundErrorAppendTestCase {
	if renderAppendNote(note, args) == "" {
		panic("compoundErrorAppendTestCase: every note row must render a note")
	}
	if want == "" {
		panic("compoundErrorAppendTestCase: every note row must state its message")
	}

	return compoundErrorAppendTestCase{
		err:  nil,
		name: name,
		note: note,
		want: want,
		args: args,
	}
}

func newCompoundErrorAppendTestCaseNothing(name, note string, args []any) compoundErrorAppendTestCase {
	if renderAppendNote(note, args) != "" {
		panic("compoundErrorAppendTestCase: every row appending nothing must render an empty note")
	}

	return compoundErrorAppendTestCase{
		err:  nil,
		name: name,
		note: note,
		want: "",
		args: args,
	}
}

// compoundErrorAppendUnchangedTestCase tests Append with an error and
// no rendered note, which appends the error as it is.
type compoundErrorAppendUnchangedTestCase struct {
	err  error
	name string
	note string
	args []any
}

func (tc compoundErrorAppendUnchangedTestCase) Name() string {
	return tc.name
}

func (tc compoundErrorAppendUnchangedTestCase) Test(t *testing.T) {
	t.Helper()
	ce := new(CompoundError)
	result := ce.Append(tc.err, tc.note, tc.args...)

	AssertSame(t, ce, result, "CompoundError instance")
	AssertMustEqual(t, 1, len(ce.Errs), "error count")
	AssertEqual(t, tc.err, ce.Errs[0], "appended error")
}

func newCompoundErrorAppendUnchangedTestCase(name string, err error, note string,
	args []any) compoundErrorAppendUnchangedTestCase {
	if err == nil {
		panic("compoundErrorAppendUnchangedTestCase: every row must pass an error")
	}
	if renderAppendNote(note, args) != "" {
		panic("compoundErrorAppendUnchangedTestCase: every row must render an empty note")
	}

	return compoundErrorAppendUnchangedTestCase{
		err:  err,
		name: name,
		note: note,
		args: args,
	}
}

func compoundErrorAppendTestCases() []TestCase {
	err := errors.New("test error")
	// skipcq: GO-W1024 - Testing an error with an empty message
	empty := errors.New("")

	return S[TestCase](
		newCompoundErrorAppendUnchangedTestCase("error without note", err, "", nil),
		newCompoundErrorAppendUnchangedTestCase("errors.New(\"\") without note", empty, "", nil),
		newCompoundErrorAppendUnchangedTestCase("error, note rendering empty", err, "%s", S[any]("")),
		newCompoundErrorAppendTestCase("error with note", err, "wrapped note", nil, "wrapped note: test error"),
		newCompoundErrorAppendTestCase("error with formatted note", err, "wrapped %s: %d", S[any]("note", 42),
			"wrapped note: 42: test error"),
		newCompoundErrorAppendTestCaseNote("note only", "note only", nil, "note only"),
		newCompoundErrorAppendTestCaseNote("note with verb, no arguments", "note %d", nil, "note %d"),
		newCompoundErrorAppendTestCaseNote("formatted note only", "note %d", S[any](7), "note 7"),
		newCompoundErrorAppendTestCaseNothing("nil error, empty note", "", nil),
		newCompoundErrorAppendTestCaseNothing("nil error, note rendering empty", "%s", S[any]("")),
	)
}

func TestCompoundErrorAppend(t *testing.T) {
	RunTestCases(t, compoundErrorAppendTestCases())
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

// Test OK() on a nil receiver — should report true.
func TestCompoundErrorNilOK(t *testing.T) {
	var ce *CompoundError
	AssertTrue(t, ce.OK(), "nil CompoundError.OK")
}
