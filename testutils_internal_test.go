package core

import (
	"errors"
	"strings"
	"testing"
)

var (
	// ErrSentinelForTest is the known error the tests pass in and
	// expect back, matched by identity.
	ErrSentinelForTest = errors.New("sentinel error")

	// ErrUnrelatedForTest is never put into any chain under test. The
	// errors.Is assertions pair a match against it as a non-match, so
	// a chain that answered true to everything would fail.
	ErrUnrelatedForTest = errors.New("unrelated error")
)

// plainEqual and plainContains check on the real t without any
// assertion, for the tests of what the assertions are built on.
//
//revive:disable-next-line:flag-parameter
func plainEqual(t *testing.T, expected, actual bool, desc string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", desc, expected, actual)
	}
}

func plainContains(t *testing.T, s, substr, desc string) {
	t.Helper()
	if substr == "" || !strings.Contains(s, substr) {
		t.Errorf("%s: %q does not contain %q", desc, s, substr)
	}
}
