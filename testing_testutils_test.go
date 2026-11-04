package core_test

import (
	"strings"
	"testing"

	"darvaza.org/core"
)

// assertErrorContains checks that the last error from MockT contains the expected substring.
// An empty expected substring matches any error, so it is a mistake at the call site.
// The precondition is reported directly: this helper backs the tests for the assertions
// it would otherwise route through.
func assertErrorContains(t *testing.T, mock *core.MockT, expected, desc string) {
	t.Helper()
	if expected == "" {
		t.Fatalf("%s: expected a non-empty substring", desc)
	}

	lastErr, ok := mock.LastError()
	core.AssertMustTrue(t, ok, "%s recorded", desc)
	core.AssertTrue(t, strings.Contains(lastErr, expected),
		"%s: %q contains %q", desc, lastErr, expected)
}

// mustMessageAt returns the message at position i from one of MockT's At
// accessors, aborting the test when there is none.
func mustMessageAt(t *testing.T, at func(int) (string, bool), i int, desc string) string {
	t.Helper()
	msg, ok := at(i)
	core.AssertMustTrue(t, ok, "%s recorded", desc)
	return msg
}

// assertPassed checks that an assertion reported success: it returned
// true, logged the value, and recorded no error.
func assertPassed(t *testing.T, mock *core.MockT, result bool, desc string) {
	t.Helper()
	core.AssertTrue(t, result, "%s result", desc)
	core.AssertFalse(t, mock.HasErrors(), "%s errors", desc)
	core.AssertTrue(t, mock.HasLogs(), "%s logs", desc)
}

// assertFailed checks that an assertion reported failure: it returned
// false, recorded an error containing expected, and logged nothing.
func assertFailed(t *testing.T, mock *core.MockT, result bool, expected, desc string) {
	t.Helper()
	core.AssertFalse(t, result, "%s result", desc)
	core.AssertTrue(t, mock.HasErrors(), "%s errors", desc)
	core.AssertFalse(t, mock.HasLogs(), "%s logs", desc)
	assertErrorContains(t, mock, expected, desc+" message")
}
