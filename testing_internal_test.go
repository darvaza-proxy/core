package core

import "testing"

var _ TestCase = assertComparableTestCase{}

// assertComparableTestCase reads its results with plain checks:
// assertComparable sits under AssertTrue and AssertFalse, which would
// otherwise read them. A row states whether the assertion passes, and
// the essence of what it logs or reports.
type assertComparableTestCase struct {
	expected any
	actual   any
	name     string
	message  string
	wantPass bool
}

// newAssertComparableTestCase declares a row the assertion passes, with
// the essence of what it logs.
func newAssertComparableTestCase(name string, expected, actual any,
	message string) assertComparableTestCase {
	return assertComparableTestCase{
		name:     name,
		expected: expected,
		actual:   actual,
		message:  message,
		wantPass: true,
	}
}

// newAssertComparableTestCaseFails declares a row the assertion fails,
// with the essence of the failure it reports.
func newAssertComparableTestCaseFails(name string, expected, actual any,
	message string) assertComparableTestCase {
	return assertComparableTestCase{
		name:     name,
		expected: expected,
		actual:   actual,
		message:  message,
		wantPass: false,
	}
}

func (tc assertComparableTestCase) Name() string { return tc.name }

func (tc assertComparableTestCase) Test(t *testing.T) {
	mock := &MockT{}
	ok := assertComparable(mock, tc.expected, tc.actual, "value %d", 1)

	plainEqual(t, tc.wantPass, ok, "result")
	plainEqual(t, !tc.wantPass, mock.HasErrors(), "errors")
	plainEqual(t, tc.wantPass, mock.HasLogs(), "logs")

	msg, _ := mock.LastLog()
	if !tc.wantPass {
		msg, _ = mock.LastError()
	}
	plainContains(t, msg, tc.message, "message")
}

func assertComparableTestCases() []assertComparableTestCase {
	return S(
		newAssertComparableTestCase("equal ints", 42, 42, "value 1: 42"),
		newAssertComparableTestCase("equal bools", true, true, "value 1: true"),
		newAssertComparableTestCase("both nil", nil, nil, "value 1: <nil>"),
		newAssertComparableTestCaseFails("unequal ints", 42, 24,
			"value 1: expected 42, got 24"),
		newAssertComparableTestCaseFails("unequal bools", false, true,
			"value 1: expected false, got true"),
		newAssertComparableTestCaseFails("different types", 42, "42",
			"value 1: expected 42, got 42"),
		newAssertComparableTestCaseFails("slices undecided", S(1), S(1),
			"value 1: undecided for []int"),
	)
}

func TestAssertComparable(t *testing.T) {
	RunTestCases(t, assertComparableTestCases())
}

// TestSliceDiffErrorWithoutKnownDifference exercises the arm
// AssertSliceEqual cannot reach: a slice it settles as unequal always
// leaves an element known to differ for the walk to find. Called
// directly with a pair that has none, doSliceDiffError must still
// report — the assertion has already failed by then, and failing
// without a message would be worse than naming the whole slice.
func TestSliceDiffErrorWithoutKnownDifference(t *testing.T) {
	mock := &MockT{}

	doSliceDiffError(mock, S(1, 2), S(1, 2), "no known difference", nil)

	lastErr, ok := mock.LastError()
	AssertMustTrue(t, ok, "error reported")
	AssertContains(t, lastErr, "expected [1 2], got [1 2]", "whole-slice message")
}
