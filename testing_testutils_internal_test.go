package core

import "testing"

// AssertLastLogForTest checks the message an assertion logged on success.
func AssertLastLogForTest(t *testing.T, mock *MockT, expected, desc string) {
	t.Helper()
	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "%s last log", desc)
	AssertComparableForTest(t, expected, lastLog, "%s log message", desc)
}

// AssertMustContinuedForTest checks that a Must assertion passed and let the
// caller carry on. The continuation log is matched rather than counted,
// so a case that asserts on a returned value in between still fits.
func AssertMustContinuedForTest(t *testing.T, mock *MockT, ok bool) {
	t.Helper()
	AssertTrue(t, ok, "Success case should not abort")
	AssertFalse(t, mock.Failed(), "Should not be marked as failed")
	AssertFalse(t, mock.HasErrors(), "Should have no errors")

	lastLog, hasLog := mock.LastLog()
	AssertTrue(t, hasLog, "Should have logged the assertion")
	AssertComparableForTest(t, MustContinuationLogForTest, lastLog, "Should reach continuation log")
}

// MustContinuationLogForTest is what every Must success case logs after the
// assertion, marking that execution was not cut short.
const MustContinuationLogForTest = "execution continues"

// AssertMustAbortedForTest checks that a Must assertion failed and cut the
// caller short before its continuation ran.
func AssertMustAbortedForTest(t *testing.T, mock *MockT, ok bool) {
	t.Helper()
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertComparableForTest(t, 1, mock.NumErrors(), "Should have error from failed assertion")
	AssertComparableForTest(t, 0, mock.NumLogs(), "Should not reach continuation log")
}
