package core

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

// Test MockT implementation
func TestMockT(t *testing.T) {
	t.Run("initial state", testMockTInitialState)
	t.Run("helper functionality", testMockTHelper)
	t.Run("error handling", testMockTErrors)
	t.Run("log handling", testMockTLogs)
	t.Run("fail functionality", testMockTFail)
	t.Run("formatted logging", testMockTFormatted)
	t.Run("multiple messages", testMockTMultiple)
	t.Run("reset functionality", testMockTReset)
	t.Run("empty state queries", testMockTEmptyQueries)
	t.Run("concurrent safety", testMockTConcurrent)
}

// Test S helper function
func TestS(t *testing.T) {
	// Test with integers
	intSlice := S(1, 2, 3)
	AssertSliceEqual(t, []int{1, 2, 3}, intSlice, "S with integers")

	// Test with strings
	strSlice := S("a", "b", "c")
	AssertSliceEqual(t, []string{"a", "b", "c"}, strSlice, "S with strings")

	// Test empty slice
	emptySlice := S[string]()
	AssertSliceEqual(t, []string{}, emptySlice, "S empty slice")

	// Test single element
	singleSlice := S(42)
	AssertSliceEqual(t, []int{42}, singleSlice, "S single element")
}

// Test AssertEqual
func TestAssertEqual(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion
	result := AssertEqual(mock, 42, 42, "equal test")
	AssertTrue(t, result, "AssertEqual result when equal")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertEqual(t, "equal test: 42", lastLog, "log message on success")

	mock.Reset()

	// Test failed assertion
	result = AssertEqual(mock, 42, 24, "not equal test")
	AssertFalse(t, result, "AssertEqual result when not equal")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")

	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok on failure")
	AssertTrue(t, strings.Contains(lastErr, "expected 42, got 24"), "error message contains values")
}

// Test AssertNotEqual
func TestAssertNotEqual(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion
	result := AssertNotEqual(mock, 42, 24, "not equal test")
	AssertTrue(t, result, "AssertNotEqual result when not equal")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertEqual(t, "not equal test: 24", lastLog, "log message on success")

	mock.Reset()

	// Test failed assertion
	result = AssertNotEqual(mock, 42, 42, "equal test")
	AssertFalse(t, result, "AssertNotEqual result when equal")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")

	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok on failure")
	AssertTrue(t, strings.Contains(lastErr, "expected not 42, got 42"), "error message contains values")
}

// Test AssertSliceEqual
func TestAssertSliceEqual(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion
	result := AssertSliceEqual(mock, S(1, 2, 3), S(1, 2, 3), "slice equal test")
	AssertTrue(t, result, "AssertSliceEqual result when equal")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test failed assertion
	result = AssertSliceEqual(mock, S(1, 2, 3), S(1, 2, 4), "slice not equal test")
	AssertFalse(t, result, "AssertSliceEqual result when not equal")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")
}

// Test AssertContains
func TestAssertContains(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion
	result := AssertContains(mock, "hello world", "world", "contains test")
	AssertTrue(t, result, "AssertContains result when contains")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertEqual(t, "contains test: contains \"world\"", lastLog, "log message on success")

	mock.Reset()

	// Test failed assertion
	result = AssertContains(mock, "hello world", "xyz", "not contains test")
	AssertFalse(t, result, "AssertContains result when not contains")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")

	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok on failure")
	AssertTrue(t, strings.Contains(lastErr, "expected \"hello world\" to contain \"xyz\""), "error message")
}

// Test AssertError and AssertNoError
func TestAssertError(t *testing.T) {
	mock := &MockT{}
	testErr := errors.New("test error")

	// Test AssertError with error (success)
	result := AssertError(mock, testErr, "error test")
	AssertTrue(t, result, "AssertError result with error")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertError with nil (failure)
	result = AssertError(mock, nil, "nil error test")
	AssertFalse(t, result, "AssertError result with nil")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertNoError with nil (success)
	result = AssertNoError(mock, nil, "no error test")
	AssertTrue(t, result, "AssertNoError result with nil")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNoError with error (failure)
	result = AssertNoError(mock, testErr, "unexpected error test")
	AssertFalse(t, result, "AssertNoError result with error")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertTrue and AssertFalse
func TestAssertBool(t *testing.T) {
	mock := &MockT{}

	// Test AssertTrue with true (success)
	result := AssertTrue(mock, true, "true test")
	AssertEqual(t, true, result, "AssertTrue(%v)", true)
	AssertEqual(t, false, mock.HasErrors(), "HasErrors")
	AssertEqual(t, true, mock.HasLogs(), "HasLogs")

	mock.Reset()

	// Test AssertTrue with false (failure)
	result = AssertTrue(mock, false, "false test")
	AssertEqual(t, false, result, "AssertTrue(%v)", false)
	AssertEqual(t, true, mock.HasErrors(), "HasErrors")

	mock.Reset()

	// Test AssertFalse with false (success)
	result = AssertFalse(mock, false, "false test")
	AssertEqual(t, true, result, "AssertFalse(%v)", false)
	AssertEqual(t, false, mock.HasErrors(), "HasErrors")
	AssertEqual(t, true, mock.HasLogs(), "HasLogs")

	mock.Reset()

	// Test AssertFalse with true (failure)
	result = AssertFalse(mock, true, "true test")
	AssertEqual(t, false, result, "AssertFalse(%v)", true)
	AssertEqual(t, true, mock.HasErrors(), "HasErrors")
}

// Test AssertNil and AssertNotNil
func TestAssertNil(t *testing.T) {
	mock := &MockT{}

	// Test AssertNil with nil (success)
	result := AssertNil(mock, nil, "nil test")
	AssertTrue(t, result, "AssertNil result with nil")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNil with non-nil (failure)
	result = AssertNil(mock, "not nil", "not nil test")
	AssertFalse(t, result, "AssertNil result with non-nil")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertNotNil with non-nil (success)
	result = AssertNotNil(mock, "not nil", "not nil test")
	AssertTrue(t, result, "AssertNotNil result with non-nil")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNotNil with nil (failure)
	result = AssertNotNil(mock, nil, "nil test")
	AssertFalse(t, result, "AssertNotNil result with nil")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertErrorIs
func TestAssertErrorIs(t *testing.T) {
	mock := &MockT{}
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(baseErr, errors.New("wrapped"))

	// Test AssertErrorIs with matching error (success)
	result := AssertErrorIs(mock, wrappedErr, baseErr, "error is test")
	AssertTrue(t, result, "AssertErrorIs result with matching error")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertErrorIs with non-matching error (failure)
	otherErr := errors.New("other error")
	result = AssertErrorIs(mock, wrappedErr, otherErr, "error is not test")
	AssertFalse(t, result, "AssertErrorIs result with non-matching error")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertTypeIs
func TestAssertTypeIs(t *testing.T) {
	mock := &MockT{}

	// Test AssertTypeIs with correct type (success)
	var val any = "hello"
	result, ok := AssertTypeIs[string](mock, val, "type is test")
	AssertTrue(t, ok, "AssertTypeIs ok with correct type")
	AssertEqual(t, "hello", result, "AssertTypeIs result with correct type")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertTypeIs with incorrect type (failure)
	val = 42
	result, ok = AssertTypeIs[string](mock, val, "type is not test")
	AssertFalse(t, ok, "AssertTypeIs ok with incorrect type")
	AssertEqual(t, "", result, "AssertTypeIs result with incorrect type")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertPanic
func TestAssertPanic(t *testing.T) {
	mock := &MockT{}

	// Test AssertPanic with panic (success)
	result := AssertPanic(mock, func() { panic("test panic") }, nil, "panic test")
	AssertTrue(t, result, "AssertPanic result with panic")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertTrue(t, strings.Contains(lastLog, "test panic"), "log contains panic value")

	mock.Reset()

	// Test AssertPanic without panic (failure)
	result = AssertPanic(mock, func() {}, nil, "no panic test")
	AssertFalse(t, result, "AssertPanic result without panic")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertPanic with expected panic value (success)
	result = AssertPanic(mock, func() { panic("specific") }, "specific", "specific panic test")
	AssertTrue(t, result, "AssertPanic result with expected panic")
	AssertFalse(t, mock.HasErrors(), "no errors on success")

	mock.Reset()

	// Test AssertPanic with wrong panic value (failure)
	result = AssertPanic(mock, func() { panic("wrong") }, "expected", "wrong panic test")
	AssertFalse(t, result, "AssertPanic result with wrong panic")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertNoPanic
func TestAssertNoPanic(t *testing.T) {
	mock := &MockT{}

	// Test AssertNoPanic without panic (success)
	result := AssertNoPanic(mock, func() {}, "no panic test")
	AssertTrue(t, result, "AssertNoPanic result without panic")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertTrue(t, strings.Contains(lastLog, "no panic"), "log contains no panic message")

	mock.Reset()

	// Test AssertNoPanic with panic (failure)
	result = AssertNoPanic(mock, func() { panic("unexpected") }, "panic test")
	AssertFalse(t, result, "AssertNoPanic result with panic")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test RunConcurrentTest
func TestRunConcurrentTest(t *testing.T) {
	mock := &MockT{}
	var counter int64

	// Test successful concurrent execution
	err := RunConcurrentTest(mock, 5, func(_ int) error {
		atomic.AddInt64(&counter, 1)
		return nil
	})
	AssertNoError(t, err, "RunConcurrentTest successful")
	AssertEqual(t, int64(5), atomic.LoadInt64(&counter), "all workers executed")

	// Test concurrent execution with error
	err = RunConcurrentTest(mock, 3, func(id int) error {
		if id == 1 {
			return errors.New("worker error")
		}
		return nil
	})
	AssertError(t, err, "RunConcurrentTest with error")
	AssertTrue(t, strings.Contains(err.Error(), "worker error"), "error message")
}

// Test TestCase interface and RunTestCases
type mockTestCase struct {
	name   string
	called bool
}

func (tc *mockTestCase) Name() string {
	return tc.name
}

func (tc *mockTestCase) Test(_ *testing.T) {
	tc.called = true
}

func TestRunTestCases(t *testing.T) {
	tc1 := &mockTestCase{name: "test1"}
	tc2 := &mockTestCase{name: "test2"}
	tc3 := &mockTestCase{name: "test3"}

	cases := []TestCase{tc1, tc2, tc3}
	RunTestCases(t, cases)

	// Verify all test cases were called
	AssertTrue(t, tc1.called, "test case 1 called")
	AssertTrue(t, tc2.called, "test case 2 called")
	AssertTrue(t, tc3.called, "test case 3 called")
}

// Test RunBenchmark
func TestRunBenchmark(t *testing.T) {
	setupCalled := false
	execCount := 0

	// Create a mock benchmark
	b := &testing.B{N: 10}
	b.ResetTimer() // Simulate benchmark setup

	RunBenchmark(b, func() any {
		setupCalled = true
		return "test data"
	}, func(data any) {
		AssertEqual(t, "test data", data, "benchmark data")
		execCount++
	})

	AssertTrue(t, setupCalled, "setup function called")
	AssertEqual(t, 10, execCount, "execution function called N times")
}

func testMockTInitialState(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	AssertFalse(t, mock.HasErrors(), "initial HasErrors")
	AssertFalse(t, mock.HasLogs(), "initial HasLogs")
	AssertEqual(t, 0, mock.HelperCalled, "initial HelperCalled")
	AssertFalse(t, mock.Failed(), "initial Failed")
}

func testMockTHelper(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	mock.Helper()
	AssertEqual(t, 1, mock.HelperCalled, "HelperCalled after Helper()")

	// Test multiple calls increment the counter
	mock.Helper()
	mock.Helper()
	AssertEqual(t, 3, mock.HelperCalled, "HelperCalled after multiple calls")
}

func testMockTErrors(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	mock.Error("test error")
	AssertTrue(t, mock.HasErrors(), "HasErrors after Error")
	AssertTrue(t, mock.Failed(), "Failed after Error")
	AssertEqual(t, 1, len(mock.Errors), "Errors length")
	AssertEqual(t, "test error", mock.Errors[0], "first error")

	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok")
	AssertEqual(t, "test error", lastErr, "LastError value")
}

func testMockTLogs(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	mock.Log("test log")
	AssertTrue(t, mock.HasLogs(), "HasLogs after Log")
	AssertFalse(t, mock.Failed(), "Failed after Log should be false")
	AssertEqual(t, 1, len(mock.Logs), "Logs length")
	AssertEqual(t, "test log", mock.Logs[0], "first log")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok")
	AssertEqual(t, "test log", lastLog, "LastLog value")
}

func testMockTFail(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test Fail() marks test as failed
	AssertFalse(t, mock.Failed(), "initial Failed state")
	mock.Fail()
	AssertTrue(t, mock.Failed(), "Failed after Fail()")

	// Test that Error() also marks as failed
	mock.Reset()
	AssertFalse(t, mock.Failed(), "Failed after Reset")
	mock.Error("test error")
	AssertTrue(t, mock.Failed(), "Failed after Error")

	// Test that Errorf() also marks as failed
	mock.Reset()
	AssertFalse(t, mock.Failed(), "Failed after Reset")
	mock.Errorf("formatted error: %s", "test")
	AssertTrue(t, mock.Failed(), "Failed after Errorf")

	// Test that Log() does not mark as failed
	mock.Reset()
	mock.Log("test log")
	AssertFalse(t, mock.Failed(), "Failed after Log should be false")

	// Test that Logf() does not mark as failed
	mock.Reset()
	mock.Logf("formatted log: %s", "test")
	AssertFalse(t, mock.Failed(), "Failed after Logf should be false")
}

func testMockTFormatted(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test Errorf
	mock.Errorf("error %d: %s", 42, "test message")
	AssertTrue(t, mock.HasErrors(), "HasErrors after Errorf")
	AssertTrue(t, mock.Failed(), "Failed after Errorf")
	AssertEqual(t, 1, len(mock.Errors), "Errors length after Errorf")
	AssertEqual(t, "error 42: test message", mock.Errors[0], "formatted error message")

	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok after Errorf")
	AssertEqual(t, "error 42: test message", lastErr, "LastError value after Errorf")

	mock.Reset()

	// Test Logf
	mock.Logf("log %d: %s", 24, "test message")
	AssertTrue(t, mock.HasLogs(), "HasLogs after Logf")
	AssertFalse(t, mock.Failed(), "Failed after Logf should be false")
	AssertEqual(t, 1, len(mock.Logs), "Logs length after Logf")
	AssertEqual(t, "log 24: test message", mock.Logs[0], "formatted log message")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok after Logf")
	AssertEqual(t, "log 24: test message", lastLog, "LastLog value after Logf")
}

func testMockTMultiple(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	mock.Error("first error")
	mock.Log("first log")
	mock.Error("second error")
	mock.Log("second log")

	AssertEqual(t, 2, len(mock.Errors), "multiple errors length")
	AssertEqual(t, 2, len(mock.Logs), "multiple logs length")

	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok after multiple")
	AssertEqual(t, "second error", lastErr, "LastError after multiple")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok after multiple")
	AssertEqual(t, "second log", lastLog, "LastLog after multiple")
}

func testMockTReset(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	mock.Error("test error")
	mock.Log("test log")
	mock.Helper()
	mock.Helper()
	mock.Fail()

	mock.Reset()
	AssertFalse(t, mock.HasErrors(), "HasErrors after Reset")
	AssertFalse(t, mock.HasLogs(), "HasLogs after Reset")
	AssertEqual(t, 0, mock.HelperCalled, "HelperCalled after Reset")
	AssertFalse(t, mock.Failed(), "Failed after Reset")
	AssertEqual(t, 0, len(mock.Errors), "Errors length after Reset")
	AssertEqual(t, 0, len(mock.Logs), "Logs length after Reset")
}

func testMockTEmptyQueries(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	lastErr, ok := mock.LastError()
	AssertFalse(t, ok, "LastError ok when empty")
	AssertEqual(t, "", lastErr, "LastError value when empty")

	lastLog, ok := mock.LastLog()
	AssertFalse(t, ok, "LastLog ok when empty")
	AssertEqual(t, "", lastLog, "LastLog value when empty")
}

func testMockTConcurrent(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test concurrent access to MockT methods
	err := RunConcurrentTest(t, 10, func(id int) error {
		switch id % 5 {
		case 0:
			mock.Helper()
		case 1:
			mock.Error("concurrent error")
		case 2:
			mock.Log("concurrent log")
		case 3:
			mock.Errorf("concurrent error %d", id)
		case 4:
			mock.Logf("concurrent log %d", id)
		}
		return nil
	})
	AssertNoError(t, err, "concurrent MockT operations")

	// Verify that all operations completed without race conditions
	AssertTrue(t, mock.HasErrors(), "has errors after concurrent operations")
	AssertTrue(t, mock.HasLogs(), "has logs after concurrent operations")
	AssertTrue(t, mock.Failed(), "failed after concurrent operations")
	AssertTrue(t, mock.HelperCalled > 0, "helper called during concurrent operations")

	// Check that we have the expected number of operations
	// We should have 2 errors per 5 operations (case 1 and 3)
	// We should have 2 logs per 5 operations (case 2 and 4)
	// We should have 1 helper call per 5 operations (case 0)
	expectedErrors := 4  // 10 workers, 2 out of every 5 operations
	expectedLogs := 4    // 10 workers, 2 out of every 5 operations
	expectedHelpers := 2 // 10 workers, 1 out of every 5 operations

	AssertEqual(t, expectedErrors, len(mock.Errors), "concurrent error count")
	AssertEqual(t, expectedLogs, len(mock.Logs), "concurrent log count")
	AssertEqual(t, expectedHelpers, mock.HelperCalled, "concurrent helper count")
}
