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

// HeldValuesForTest is comparable as a type, while == on two of them
// panics once it reaches a field holding values of a non-comparable type.
// Fields compare in order, and the first difference ends the comparison.
type HeldValuesForTest struct {
	First  any
	Second any
}

// OpaqueFieldsForTest provides reflect.Values that cannot be unwrapped
// via Interface(), as reached through unexported struct fields.
type OpaqueFieldsForTest struct {
	v any
	n int
}

// NewOpaqueFieldsForTest returns an OpaqueFieldsForTest holding v and n.
func NewOpaqueFieldsForTest(v any, n int) OpaqueFieldsForTest {
	return OpaqueFieldsForTest{v: v, n: n}
}

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

// AssertComparableForTest makes assertComparable available to every
// test, inside package core or out.
func AssertComparableForTest[U comparable](t T, expected, actual U, name string, args ...any) bool {
	t.Helper()
	return assertComparable(t, expected, actual, name, args...)
}

// MustNotZeroForTest refuses a zero value where a row declares a
// non-zero one.
func MustNotZeroForTest[T any](v T) T {
	if IsZero(v) {
		panic("row declares a zero value where a non-zero one is expected")
	}
	return v
}
