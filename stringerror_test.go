package core_test

import (
	"errors"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase.
var (
	_ core.TestCase = stringErrorTestCase{}
	_ core.TestCase = stringErrorIsTestCase{}
	_ core.TestCase = stringErrorCtorTestCase{}
)

// errStringSentinel exercises the constant-sentinel pattern StringError
// exists for.
const errStringSentinel core.StringError = "sentinel"

// stringErrorTestCase checks Error() and AsError() for a StringError.
type stringErrorTestCase struct {
	name    string
	input   core.StringError
	wantMsg string
	wantNil bool
}

func newStringErrorTestCase(name string, input core.StringError, wantMsg string,
	wantNil bool) stringErrorTestCase {
	return stringErrorTestCase{
		name:    name,
		input:   input,
		wantMsg: wantMsg,
		wantNil: wantNil,
	}
}

func (tc stringErrorTestCase) Name() string {
	return tc.name
}

func (tc stringErrorTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, tc.wantMsg, tc.input.Error(), "Error")
	// OK(), IsZero() and a nil AsError() all track emptiness together.
	core.AssertEqual(t, tc.wantNil, tc.input.OK(), "OK")
	core.AssertEqual(t, tc.wantNil, tc.input.IsZero(), "IsZero")

	err := tc.input.AsError()
	if tc.wantNil {
		core.AssertNil(t, err, "AsError")
		return
	}

	core.AssertNotNil(t, err, "AsError")
	core.AssertEqual(t, tc.wantMsg, err.Error(), "AsError message")
}

func stringErrorTestCases() []stringErrorTestCase {
	return []stringErrorTestCase{
		newStringErrorTestCase("empty", "", "", true),
		newStringErrorTestCase("simple", "boom", "boom", false),
		newStringErrorTestCase("with spaces", "file not found",
			"file not found", false),
		newStringErrorTestCase("blank is not empty", " ", " ", false),
	}
}

func TestStringError(t *testing.T) {
	core.RunTestCases(t, stringErrorTestCases())
}

// stringErrorIsTestCase checks StringError identity matching through
// both errors.Is and the package's IsError, which must agree.
type stringErrorIsTestCase struct {
	err    error
	target error
	name   string
	want   bool
}

func newStringErrorIsTestCase(name string, err, target error,
	want bool) stringErrorIsTestCase {
	return stringErrorIsTestCase{
		name:   name,
		err:    err,
		target: target,
		want:   want,
	}
}

func (tc stringErrorIsTestCase) Name() string {
	return tc.name
}

func (tc stringErrorIsTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, tc.want, errors.Is(tc.err, tc.target), "errors.Is")
	core.AssertEqual(t, tc.want, core.IsError(tc.err, tc.target), "IsError")
}

func stringErrorIsTestCases() []stringErrorIsTestCase {
	return []stringErrorIsTestCase{
		newStringErrorIsTestCase("same value",
			core.StringError("boom"), core.StringError("boom"), true),
		newStringErrorIsTestCase("sentinel constant",
			errStringSentinel, core.StringError("sentinel"), true),
		newStringErrorIsTestCase("different value",
			core.StringError("boom"), core.StringError("bang"), false),
		newStringErrorIsTestCase("wrapped sentinel",
			core.Wrap(errStringSentinel, "context"), errStringSentinel, true),
	}
}

func TestStringErrorIs(t *testing.T) {
	core.RunTestCases(t, stringErrorIsTestCases())
}

// stringErrorCtorTestCase checks NewStringError formatting and its
// empty-to-nil behaviour.
type stringErrorCtorTestCase struct {
	format  string
	name    string
	wantMsg string
	args    []any
	wantNil bool
}

func newStringErrorCtorTestCase(name, format string, args []any,
	wantMsg string, wantNil bool) stringErrorCtorTestCase {
	return stringErrorCtorTestCase{
		args:    args,
		format:  format,
		name:    name,
		wantMsg: wantMsg,
		wantNil: wantNil,
	}
}

func (tc stringErrorCtorTestCase) Name() string {
	return tc.name
}

func (tc stringErrorCtorTestCase) Test(t *testing.T) {
	t.Helper()

	err := core.NewStringError(tc.format, tc.args...)
	if tc.wantNil {
		core.AssertNil(t, err, "NewStringError")
		return
	}

	core.AssertNotNil(t, err, "NewStringError")
	core.AssertEqual(t, tc.wantMsg, err.Error(), "message")
	core.AssertErrorIs(t, err, core.StringError(tc.wantMsg), "matches sentinel")
}

func stringErrorCtorTestCases() []stringErrorCtorTestCase {
	return []stringErrorCtorTestCase{
		newStringErrorCtorTestCase("plain", "boom", nil, "boom", false),
		newStringErrorCtorTestCase("empty", "", nil, "", true),
		newStringErrorCtorTestCase("formatted", "%s: %d",
			core.S[any]("code", 42), "code: 42", false),
		newStringErrorCtorTestCase("literal percent", "50% off", nil,
			"50% off", false),
	}
}

func TestNewStringError(t *testing.T) {
	core.RunTestCases(t, stringErrorCtorTestCases())
}

// TestStringErrorGenericHelpers confirms StringError's methods are picked
// up by the generic AsError and IsZero helpers.
func TestStringErrorGenericHelpers(t *testing.T) {
	core.AssertNil(t, core.AsError(core.StringError("")), "empty via AsError")
	core.AssertError(t, core.AsError(core.StringError("boom")), "non-empty via AsError")
	core.AssertTrue(t, core.IsZero(core.StringError("")), "empty via IsZero")
	core.AssertFalse(t, core.IsZero(core.StringError("boom")), "non-empty via IsZero")
}
