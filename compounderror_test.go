package core

import (
	"errors"
	"strings"
	"testing"
)

type compoundErrorErrorTestCase struct {
	expected string
	name     string
	errs     []error
}

var compoundErrorErrorTestCases = []compoundErrorErrorTestCase{
	{
		name:     "empty errors",
		errs:     []error{},
		expected: "",
	},
	{
		name:     "single error",
		errs:     []error{errors.New("first error")},
		expected: "first error",
	},
	{
		name:     "multiple errors",
		errs:     []error{errors.New("first error"), errors.New("second error")},
		expected: "first error\nsecond error",
	},
	{
		name:     "with nil errors",
		errs:     []error{errors.New("first error"), nil, errors.New("third error")},
		expected: "first error\nthird error",
	},
	{
		name:     "all nil errors",
		errs:     []error{nil, nil, nil},
		expected: "",
	},
}

func (tc compoundErrorErrorTestCase) test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.errs}
	result := ce.Error()

	if result != tc.expected {
		t.Fatalf("expected '%s', got '%s'", tc.expected, result)
	}
}

func TestCompoundErrorError(t *testing.T) {
	for _, tc := range compoundErrorErrorTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestCompoundErrorErrors(t *testing.T) {
	errs := []error{
		errors.New("first error"),
		errors.New("second error"),
	}

	ce := &CompoundError{Errs: errs}
	result := ce.Errors()

	if len(result) != len(errs) {
		t.Fatalf("expected %d errors, got %d", len(errs), len(result))
	}

	for i, err := range result {
		if err != errs[i] {
			t.Fatalf("expected error %d to be %v, got %v", i, errs[i], err)
		}
	}
}

func TestCompoundErrorUnwrap(t *testing.T) {
	errs := []error{
		errors.New("first error"),
		errors.New("second error"),
	}

	ce := &CompoundError{Errs: errs}
	result := ce.Unwrap()

	if len(result) != len(errs) {
		t.Fatalf("expected %d errors, got %d", len(errs), len(result))
	}

	for i, err := range result {
		if err != errs[i] {
			t.Fatalf("expected error %d to be %v, got %v", i, errs[i], err)
		}
	}
}

type compoundErrorOkTestCase struct {
	name     string
	errs     []error
	expected bool
}

var compoundErrorOkTestCases = []compoundErrorOkTestCase{
	{
		name:     "empty errors",
		errs:     []error{},
		expected: true,
	},
	{
		name:     "nil slice",
		errs:     nil,
		expected: true,
	},
	{
		name:     "single error",
		errs:     []error{errors.New("error")},
		expected: false,
	},
	{
		name:     "multiple errors",
		errs:     []error{errors.New("first"), errors.New("second")},
		expected: false,
	},
}

func (tc compoundErrorOkTestCase) test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.errs}
	result := ce.Ok()

	if result != tc.expected {
		t.Fatalf("expected %t, got %t", tc.expected, result)
	}
}

func TestCompoundErrorOk(t *testing.T) {
	for _, tc := range compoundErrorOkTestCases {
		t.Run(tc.name, tc.test)
	}
}

type compoundErrorAsErrorTestCase struct {
	name      string
	errs      []error
	expectNil bool
}

var compoundErrorAsErrorTestCases = []compoundErrorAsErrorTestCase{
	{
		name:      "empty errors",
		errs:      []error{},
		expectNil: true,
	},
	{
		name:      "nil slice",
		errs:      nil,
		expectNil: true,
	},
	{
		name:      "single error",
		errs:      []error{errors.New("error")},
		expectNil: false,
	},
	{
		name:      "multiple errors",
		errs:      []error{errors.New("first"), errors.New("second")},
		expectNil: false,
	},
}

func (tc compoundErrorAsErrorTestCase) test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.errs}
	result := ce.AsError()

	if tc.expectNil {
		if result != nil {
			t.Fatalf("expected nil, got %v", result)
		}
	} else {
		if result == nil {
			t.Fatalf("expected non-nil error, got nil")
		}
		if result != ce {
			t.Fatalf("expected same CompoundError instance, got different")
		}
	}
}

func TestCompoundErrorAsError(t *testing.T) {
	for _, tc := range compoundErrorAsErrorTestCases {
		t.Run(tc.name, tc.test)
	}
}

type compoundErrorAppendErrorTestCase struct {
	name        string
	initial     []error
	toAppend    []error
	expectedLen int
}

var compoundErrorAppendErrorTestCases = []compoundErrorAppendErrorTestCase{
	{
		name:        "append to empty",
		initial:     []error{},
		toAppend:    []error{errors.New("first")},
		expectedLen: 1,
	},
	{
		name:        "append multiple",
		initial:     []error{errors.New("existing")},
		toAppend:    []error{errors.New("first"), errors.New("second")},
		expectedLen: 3,
	},
	{
		name:        "append with nils",
		initial:     []error{errors.New("existing")},
		toAppend:    []error{nil, errors.New("valid"), nil},
		expectedLen: 2,
	},
	{
		name:        "append all nils",
		initial:     []error{errors.New("existing")},
		toAppend:    []error{nil, nil},
		expectedLen: 1,
	},
}

func (tc compoundErrorAppendErrorTestCase) test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.initial}
	result := ce.AppendError(tc.toAppend...)

	// Test method chaining
	if result != ce {
		t.Fatalf("expected same CompoundError instance for chaining")
	}

	// Test length
	if len(ce.Errs) != tc.expectedLen {
		t.Fatalf("expected %d errors, got %d", tc.expectedLen, len(ce.Errs))
	}

	// Test no nil errors were added
	for i, err := range ce.Errs {
		if err == nil {
			t.Fatalf("unexpected nil error at index %d", i)
		}
	}
}

func TestCompoundErrorAppendError(t *testing.T) {
	for _, tc := range compoundErrorAppendErrorTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestCompoundErrorAppendErrorWithCompoundError(t *testing.T) {
	ce1 := &CompoundError{Errs: []error{errors.New("first")}}
	ce2 := &CompoundError{Errs: []error{errors.New("second"), errors.New("third")}}

	result := ce1.AppendError(ce2)

	// Test method chaining
	if result != ce1 {
		t.Fatalf("expected same CompoundError instance for chaining")
	}

	// Should unwrap the compound error and append individual errors
	expectedLen := 3 // original 1 + unwrapped 2
	if len(ce1.Errs) != expectedLen {
		t.Fatalf("expected %d errors, got %d", expectedLen, len(ce1.Errs))
	}

	// Check error messages
	errorStr := ce1.Error()
	if !strings.Contains(errorStr, "first") {
		t.Fatalf("expected 'first' in error string, got '%s'", errorStr)
	}
	if !strings.Contains(errorStr, "second") {
		t.Fatalf("expected 'second' in error string, got '%s'", errorStr)
	}
	if !strings.Contains(errorStr, "third") {
		t.Fatalf("expected 'third' in error string, got '%s'", errorStr)
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
	ce := &CompoundError{Errs: []error{errors.New("initial")}}
	mock := &mockUnwrappable{
		errs: []error{errors.New("unwrapped1"), errors.New("unwrapped2")},
	}

	result := ce.AppendError(mock)

	// Test method chaining
	if result != ce {
		t.Fatalf("expected same CompoundError instance for chaining")
	}

	// Should unwrap and append individual errors
	expectedLen := 3 // original 1 + unwrapped 2
	if len(ce.Errs) != expectedLen {
		t.Fatalf("expected %d errors, got %d", expectedLen, len(ce.Errs))
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

var compoundErrorAppendTestCases = []compoundErrorAppendTestCase{
	{
		name:        "nil error, empty note",
		initial:     []error{},
		err:         nil,
		note:        "",
		args:        nil,
		expectedLen: 0,
		expectNote:  false,
	},
	{
		name:        "nil error, with note",
		initial:     []error{},
		err:         nil,
		note:        "note only",
		args:        nil,
		expectedLen: 1,
		expectNote:  true,
	},
	{
		name:        "error without note",
		initial:     []error{},
		err:         errors.New("test error"),
		note:        "",
		args:        nil,
		expectedLen: 1,
		expectNote:  false,
	},
	{
		name:        "error with note",
		initial:     []error{},
		err:         errors.New("test error"),
		note:        "wrapped note",
		args:        nil,
		expectedLen: 1,
		expectNote:  true,
	},
	{
		name:        "formatted note",
		initial:     []error{},
		err:         errors.New("test error"),
		note:        "wrapped %s: %d",
		args:        []any{"note", 42},
		expectedLen: 1,
		expectNote:  true,
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc compoundErrorAppendTestCase) test(t *testing.T) {
	t.Helper()
	ce := &CompoundError{Errs: tc.initial}
	result := ce.Append(tc.err, tc.note, tc.args...)

	// Test method chaining
	if result != ce {
		t.Fatalf("expected same CompoundError instance for chaining")
	}

	// Test length
	if len(ce.Errs) != tc.expectedLen {
		t.Fatalf("expected %d errors, got %d", tc.expectedLen, len(ce.Errs))
	}

	if tc.expectedLen > 0 {
		lastErr := ce.Errs[len(ce.Errs)-1]
		if lastErr == nil {
			t.Fatalf("expected non-nil error")
		}

		errorStr := lastErr.Error()
		if tc.expectNote {
			if tc.note != "" {
				expectedNote := tc.note
				if len(tc.args) > 0 {
					expectedNote = "wrapped note: 42" // for the formatted case
				}
				if !strings.Contains(errorStr, expectedNote) && !strings.Contains(errorStr, tc.note) {
					t.Fatalf("expected note in error string, got '%s'", errorStr)
				}
			}
		}
	}
}

func TestCompoundErrorAppend(t *testing.T) {
	for _, tc := range compoundErrorAppendTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestCompoundErrorAppendChaining(t *testing.T) {
	ce := &CompoundError{}

	result := ce.
		Append(errors.New("first"), "").
		Append(nil, "second note").
		AppendError(errors.New("third")).
		Append(errors.New("fourth"), "wrapped %s", "note")

	// Test method chaining returns same instance
	if result != ce {
		t.Fatalf("expected same CompoundError instance for chaining")
	}

	// Test final length
	expectedLen := 4
	if len(ce.Errs) != expectedLen {
		t.Fatalf("expected %d errors, got %d", expectedLen, len(ce.Errs))
	}

	// Test that all errors are non-nil
	for i, err := range ce.Errs {
		if err == nil {
			t.Fatalf("unexpected nil error at index %d", i)
		}
	}
}

func TestCompoundErrorNilHandling(t *testing.T) {
	// Test that nil errors are properly filtered out
	ce := &CompoundError{}

	_ = ce.AppendError(nil, errors.New("valid"), nil)

	if len(ce.Errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(ce.Errs))
	}

	if ce.Errs[0].Error() != "valid" {
		t.Fatalf("expected 'valid' error, got '%s'", ce.Errs[0].Error())
	}
}

func TestCompoundErrorIsInterface(t *testing.T) {
	ce := &CompoundError{Errs: []error{errors.New("test")}}

	// Test Errors interface
	var _ Errors = ce

	// Test that it implements error interface
	var _ error = ce

	// Test that Error() method works
	if ce.Error() == "" {
		t.Fatalf("expected non-empty error string")
	}

	// Test that Errors() method works
	errs := ce.Errors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}
