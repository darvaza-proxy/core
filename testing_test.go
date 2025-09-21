package core

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = (*mockTestCase)(nil)

// Test MockT implementation
func TestMockT(t *testing.T) {
	t.Run("initial state", runTestMockTInitialState)
	t.Run("helper functionality", runTestMockTHelper)
	t.Run("error handling", runTestMockTErrors)
	t.Run("log handling", runTestMockTLogs)
	t.Run("fail functionality", runTestMockTFail)
	t.Run("formatted logging", runTestMockTFormatted)
	t.Run("multiple messages", runTestMockTMultiple)
	t.Run("reset functionality", runTestMockTReset)
	t.Run("empty state queries", runTestMockTEmptyQueries)
	t.Run("concurrent safety", runTestMockTConcurrent)
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

	assertErrorContains(t, mock, "expected 42, got 24", "error message contains values")
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

	assertErrorContains(t, mock, "expected not 42, got 42", "error message contains values")
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

	assertErrorContains(t, mock, "expected \"hello world\" to contain \"xyz\"", "error message")
}

// Test AssertNotContain
func TestAssertNotContain(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion (substring not present)
	result := AssertNotContain(mock, "hello world", "xyz", "not contains test")
	AssertTrue(t, result, "AssertNotContain result when not contains")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertEqual(t, "not contains test: does not contain \"xyz\"", lastLog, "log message on success")

	mock.Reset()

	// Test failed assertion (substring present)
	result = AssertNotContain(mock, "hello world", "world", "contains test")
	AssertFalse(t, result, "AssertNotContain result when contains")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")

	assertErrorContains(t, mock, "expected \"hello world\" not to contain \"world\"", "error message")

	mock.Reset()

	// Test empty substring assertion (should fail)
	result = AssertNotContain(mock, "hello world", "", "empty substring test")
	AssertFalse(t, result, "AssertNotContain result with empty substring")
	AssertTrue(t, mock.HasErrors(), "has errors on empty substring")
	AssertFalse(t, mock.HasLogs(), "no logs on empty substring")

	assertErrorContains(t, mock, "substring cannot be empty for AssertNotContain", "empty substring error message")
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

// assertPanicTestCase for table-driven tests
type assertPanicTestCase struct {
	panicFn      func()
	expected     any
	logContains  string
	name         string
	desc         string
	expectResult bool
	expectErrors bool
}

// Compile-time verification
var _ TestCase = assertPanicTestCase{}

func (tc assertPanicTestCase) Name() string {
	return tc.name
}

func (tc assertPanicTestCase) Test(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	result := AssertPanic(mock, tc.panicFn, tc.expected, tc.desc)
	AssertEqual(t, tc.expectResult, result, "result")
	AssertEqual(t, tc.expectErrors, mock.HasErrors(), "has errors")

	if tc.logContains != "" && mock.HasLogs() {
		lastLog, _ := mock.LastLog()
		AssertContains(t, lastLog, tc.logContains, "log content")
	}
}

// revive:disable-next-line:argument-limit
func newAssertPanicTestCase(name string, panicFn func(), expected any, desc string,
	expectResult, expectErrors bool, logContains string) TestCase {
	return assertPanicTestCase{
		name:         name,
		panicFn:      panicFn,
		expected:     expected,
		desc:         desc,
		expectResult: expectResult,
		expectErrors: expectErrors,
		logContains:  logContains,
	}
}

func makeAssertPanicTestCases() []TestCase {
	testErr := errors.New("test error")
	otherErr := errors.New("other error")
	panicErr := NewPanicError(0, "recovered panic")
	wrappedPanic := NewPanicError(0, 99)

	return []TestCase{
		// Any panic tests
		newAssertPanicTestCase("any panic accepted",
			func() { panic("test panic") }, nil, "panic test",
			true, false, "test panic"),
		newAssertPanicTestCase("no panic fails",
			func() {}, nil, "no panic test",
			false, true, ""),

		// String matching tests
		newAssertPanicTestCase("string substring match",
			func() { panic("specific panic message") }, "specific", "string test",
			true, false, "contains"),
		newAssertPanicTestCase("string mismatch",
			func() { panic("wrong") }, "expected", "wrong string test",
			false, true, ""),
		newAssertPanicTestCase("non-string panic with string expected",
			func() { panic(123) }, "123", "non-string test",
			true, false, "contains"),

		// Error matching tests
		newAssertPanicTestCase("error match",
			func() { panic(testErr) }, testErr, "error test",
			true, false, "panic error"),
		newAssertPanicTestCase("error mismatch",
			func() { panic(testErr) }, otherErr, "wrong error test",
			false, true, ""),

		// Exact matching tests
		newAssertPanicTestCase("integer match",
			func() { panic(42) }, 42, "int test",
			true, false, "panic: 42"),
		newAssertPanicTestCase("integer mismatch",
			func() { panic(42) }, 43, "wrong int test",
			false, true, ""),

		// Recovered type tests
		newAssertPanicTestCase("Recovered type match",
			func() { panic(panicErr) }, panicErr, "recovered test",
			true, false, "panic:"),
		newAssertPanicTestCase("Recovered unwrapping",
			func() { panic(wrappedPanic) }, 99, "unwrapped test",
			true, false, "panic: 99"),
	}
}

// Test AssertPanic
func TestAssertPanic(t *testing.T) {
	RunTestCases(t, makeAssertPanicTestCases())
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

func newMockTestCase(name string) *mockTestCase {
	return &mockTestCase{
		name:   name,
		called: false,
	}
}

func makeRunTestCasesTestCases() []TestCase {
	tc1 := newMockTestCase("test1")
	tc2 := newMockTestCase("test2")
	tc3 := newMockTestCase("test3")

	return []TestCase{tc1, tc2, tc3}
}

func TestRunTestCases(t *testing.T) {
	cases := makeRunTestCasesTestCases()
	RunTestCases(t, cases)

	// Extract test cases for verification
	tc1 := AssertMustTypeIs[*mockTestCase](t, cases[0], "test case 1")
	tc2 := AssertMustTypeIs[*mockTestCase](t, cases[1], "test case 2")
	tc3 := AssertMustTypeIs[*mockTestCase](t, cases[2], "test case 3")

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

func runTestMockTInitialState(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	AssertFalse(t, mock.HasErrors(), "initial HasErrors")
	AssertFalse(t, mock.HasLogs(), "initial HasLogs")
	AssertEqual(t, 0, mock.HelperCalled, "initial HelperCalled")
	AssertFalse(t, mock.Failed(), "initial Failed")
}

func runTestMockTHelper(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	mock.Helper()
	AssertEqual(t, 1, mock.HelperCalled, "HelperCalled after Helper()")

	// Test multiple calls increment the counter
	mock.Helper()
	mock.Helper()
	AssertEqual(t, 3, mock.HelperCalled, "HelperCalled after multiple calls")
}

func runTestMockTErrors(t *testing.T) {
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

func runTestMockTLogs(t *testing.T) {
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

func runTestMockTFail(t *testing.T) {
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

func runTestMockTFormatted(t *testing.T) {
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

func runTestMockTMultiple(t *testing.T) {
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

func runTestMockTReset(t *testing.T) {
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

func runTestMockTEmptyQueries(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	lastErr, ok := mock.LastError()
	AssertFalse(t, ok, "LastError ok when empty")
	AssertEqual(t, "", lastErr, "LastError value when empty")

	lastLog, ok := mock.LastLog()
	AssertFalse(t, ok, "LastLog ok when empty")
	AssertEqual(t, "", lastLog, "LastLog value when empty")
}

func runTestMockTConcurrent(t *testing.T) {
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

// Test AssertSame
func TestAssertSame(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion with value types
	result := AssertSame(mock, 42, 42, "same value test")
	AssertTrue(t, result, "AssertSame result when same values")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertEqual(t, "same value test: same value or reference", lastLog, "log message on success")

	mock.Reset()

	// Test successful assertion with reference types
	slice1 := S(1, 2, 3)
	slice2 := slice1
	result = AssertSame(mock, slice1, slice2, "same slice test")
	AssertTrue(t, result, "AssertSame result when same slices")
	AssertFalse(t, mock.HasErrors(), "no errors on slice success")
	AssertTrue(t, mock.HasLogs(), "has logs on slice success")

	mock.Reset()

	// Test failed assertion with value types
	result = AssertSame(mock, 42, 43, "different value test")
	AssertFalse(t, result, "AssertSame result when different values")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")

	assertErrorContains(t, mock, "expected same value or reference, got different", "error message on failure")

	mock.Reset()

	// Test failed assertion with reference types
	slice3 := S(1, 2, 3)
	result = AssertSame(mock, slice1, slice3, "different slice test")
	AssertFalse(t, result, "AssertSame result when different slices")
	AssertTrue(t, mock.HasErrors(), "has errors on slice failure")
	AssertFalse(t, mock.HasLogs(), "no logs on slice failure")
}

// Test AssertNotSame
func TestAssertNotSame(t *testing.T) {
	mock := &MockT{}

	// Test successful assertion with value types
	result := AssertNotSame(mock, 42, 43, "different value test")
	AssertTrue(t, result, "AssertNotSame result when different values")
	AssertFalse(t, mock.HasErrors(), "no errors on success")
	AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	AssertTrue(t, ok, "LastLog ok on success")
	AssertEqual(t, "different value test: different values or references", lastLog, "log message on success")

	mock.Reset()

	// Test successful assertion with reference types
	slice1 := S(1, 2, 3)
	slice2 := S(1, 2, 3) // same contents, different backing arrays
	result = AssertNotSame(mock, slice1, slice2, "different slice test")
	AssertTrue(t, result, "AssertNotSame result when different slices")
	AssertFalse(t, mock.HasErrors(), "no errors on slice success")
	AssertTrue(t, mock.HasLogs(), "has logs on slice success")

	mock.Reset()

	// Test failed assertion with value types
	result = AssertNotSame(mock, 42, 42, "same value test")
	AssertFalse(t, result, "AssertNotSame result when same values")
	AssertTrue(t, mock.HasErrors(), "has errors on failure")
	AssertFalse(t, mock.HasLogs(), "no logs on failure")

	assertErrorContains(t, mock, "expected different values or references, got same", "error message on failure")

	mock.Reset()

	// Test failed assertion with reference types
	slice3 := slice1 // same reference
	result = AssertNotSame(mock, slice1, slice3, "same slice test")
	AssertFalse(t, result, "AssertNotSame result when same slices")
	AssertTrue(t, mock.HasErrors(), "has errors on slice failure")
	AssertFalse(t, mock.HasLogs(), "no logs on slice failure")
}

func TestMockTFatal(t *testing.T) {
	mock := &MockT{}

	// Test Fatal panics and records error
	ok := mock.Run("fatal test", func(mt T) {
		mt.Fatal("test fatal message")
	})

	AssertFalse(t, ok, "Fatal should cause test to fail")
	AssertTrue(t, mock.Failed(), "Fatal should mark test as failed")
	AssertEqual(t, 1, len(mock.Errors), "Fatal should record error")
	AssertEqual(t, "test fatal message", mock.Errors[0], "Fatal error message")
}

func TestMockTFatalf(t *testing.T) {
	mock := &MockT{}

	// Test Fatalf panics and records formatted error
	ok := mock.Run("fatalf test", func(mt T) {
		mt.Fatalf("test %s message %d", "fatalf", 42)
	})

	AssertFalse(t, ok, "Fatalf should cause test to fail")
	AssertTrue(t, mock.Failed(), "Fatalf should mark test as failed")
	AssertEqual(t, 1, len(mock.Errors), "Fatalf should record error")
	AssertEqual(t, "test fatalf message 42", mock.Errors[0], "Fatalf error message")
}

func TestMockTFailNow(t *testing.T) {
	mock := &MockT{}

	// Test FailNow panics and marks as failed
	ok := mock.Run("FailNow test", func(mt T) {
		mt.FailNow()
	})

	AssertFalse(t, ok, "FailNow should cause test to fail")
	AssertTrue(t, mock.Failed(), "FailNow should mark test as failed")
	AssertEqual(t, 0, len(mock.Errors), "FailNow should not record error")
}

func TestMockTRunSuccess(t *testing.T) {
	mock := &MockT{}

	// Test successful run
	ok := mock.Run("success test", func(mt T) {
		mt.Log("test passed")
	})

	AssertTrue(t, ok, "Successful test should return true")
	AssertFalse(t, mock.Failed(), "Successful test should not be marked as failed")
	AssertEqual(t, 1, len(mock.Logs), "Should record log message")
	AssertEqual(t, "test passed", mock.Logs[0], "Log message content")
}

func TestMockTRunNilChecks(t *testing.T) {
	// Test nil MockT
	var mock *MockT
	ok := mock.Run("nil test", func(mt T) {
		mt.Log("should not run")
	})
	AssertFalse(t, ok, "nil MockT should return false")

	// Test nil function
	mock = &MockT{}
	ok = mock.Run("nil func test", nil)
	AssertFalse(t, ok, "nil function should return false")
}

func TestMockTRunPanicPropagation(t *testing.T) {
	mock := &MockT{}

	// Test that non-FailNow panics are propagated
	AssertPanic(t, func() {
		mock.Run("panic test", func(_ T) {
			panic("custom panic")
		})
	}, "custom panic", "Non-FailNow panics should be propagated")
}

// Test early abort pattern: if !Assert() { FailNow() }
func TestEarlyAbortPattern(t *testing.T) {
	t.Run("successful assertion continues", runTestEarlyAbortSuccess)
	t.Run("failed assertion aborts", runTestEarlyAbortFailure)
	t.Run("multiple assertions with early abort", runTestEarlyAbortMultiple)
	t.Run("mixed patterns", runTestEarlyAbortMixed)
}

func runTestEarlyAbortSuccess(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test successful assertion that continues execution
	ok := mock.Run("early abort success", func(mt T) {
		if !AssertEqual(mt, 42, 42, "equal values") {
			mt.FailNow()
		}
		// This should execute
		mt.Log("execution continues after successful assertion")
	})

	AssertTrue(t, ok, "Test should pass when assertion succeeds")
	AssertFalse(t, mock.Failed(), "Should not be marked as failed")
	AssertEqual(t, 2, len(mock.Logs), "Should have 2 log messages")
	// First log from AssertEqual success, second from explicit Log
	AssertTrue(t, strings.Contains(mock.Logs[0], "equal values: 42"), "First log from assertion")
}

func runTestEarlyAbortFailure(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test failed assertion that aborts execution
	ok := mock.Run("early abort failure", func(mt T) {
		if !AssertEqual(mt, 42, 24, "different values") {
			mt.FailNow()
		}
		// This should NOT execute
		mt.Log("this should not be reached")
	})

	AssertFalse(t, ok, "Test should fail when assertion fails and FailNow is called")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 1, len(mock.Errors), "Should have error from failed assertion")
	AssertEqual(t, 0, len(mock.Logs), "Should have no logs after early abort")
	AssertTrue(t, strings.Contains(mock.Errors[0], "expected 42, got 24"), "Error from failed assertion")
}

func runTestEarlyAbortMultiple(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test multiple assertions with early abort pattern
	ok := mock.Run("multiple early abort", func(mt T) {
		if !AssertEqual(mt, "hello", "hello", "first check") {
			mt.FailNow()
		}
		if !AssertTrue(mt, true, "second check") {
			mt.FailNow()
		}
		if !AssertEqual(mt, 1, 2, "third check") {
			mt.FailNow() // This should abort
		}
		// This should NOT execute
		mt.Log("unreachable code")
	})

	AssertFalse(t, ok, "Test should fail on third assertion")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 1, len(mock.Errors), "Should have one error from failed assertion")
	AssertEqual(t, 2, len(mock.Logs), "Should have logs from first two successful assertions")
	AssertTrue(t, strings.Contains(mock.Errors[0], "expected 1, got 2"), "Error from third assertion")
}

func runTestEarlyAbortMixed(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test mixing early abort pattern with regular assertions
	ok := mock.Run("mixed patterns", func(mt T) {
		// Regular assertion (doesn't abort on failure)
		AssertEqual(mt, 1, 2, "regular assertion")

		// Early abort pattern (aborts on failure)
		if !AssertEqual(mt, "test", "test", "critical assertion") {
			mt.FailNow()
		}

		// This should execute because the critical assertion passed
		mt.Log("continuing after critical assertion")

		// Another early abort that fails
		if !AssertFalse(mt, true, "should be false") {
			mt.FailNow() // This should abort
		}

		// This should NOT execute
		mt.Log("unreachable after failed critical assertion")
	})

	AssertFalse(t, ok, "Test should fail due to early abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 2, len(mock.Errors), "Should have two errors")
	AssertEqual(t, 2, len(mock.Logs), "Should have two logs before abort")

	// Check the error messages
	AssertTrue(t, strings.Contains(mock.Errors[0], "expected 1, got 2"),
		"First error from regular assertion")
	AssertTrue(t, strings.Contains(mock.Errors[1], "expected false, got true"),
		"Second error from early abort assertion")

	// Check log messages
	AssertTrue(t, strings.Contains(mock.Logs[0], "critical assertion: test"),
		"Log from successful critical assertion")
	AssertTrue(t, strings.Contains(mock.Logs[1], "continuing after critical assertion"),
		"Log showing execution continued")
}

// Test AssertMustFoo() functions that call FailNow() on failure
func TestAssertMustFunctions(t *testing.T) {
	t.Run("AssertMustEqual", runTestAssertMustEqual)
	t.Run("AssertMustNotEqual", runTestAssertMustNotEqual)
	t.Run("AssertMustSliceEqual", runTestAssertMustSliceEqual)
	t.Run("AssertMustContains", runTestAssertMustContains)
	t.Run("AssertMustNotContain", runTestAssertMustNotContain)
	t.Run("AssertMustError", runTestAssertMustError)
	t.Run("AssertMustNoError", runTestAssertMustNoError)
	t.Run("AssertMustPanic", runTestAssertMustPanic)
	t.Run("AssertMustNoPanic", runTestAssertMustNoPanic)
	t.Run("AssertMustTrue", runTestAssertMustTrue)
	t.Run("AssertMustFalse", runTestAssertMustFalse)
	t.Run("AssertMustErrorIs", runTestAssertMustErrorIs)
	t.Run("AssertMustTypeIs", runTestAssertMustTypeIs)
	t.Run("AssertMustNil", runTestAssertMustNil)
	t.Run("AssertMustNotNil", runTestAssertMustNotNil)
	t.Run("AssertMustSame", runTestAssertMustSame)
	t.Run("AssertMustNotSame", runTestAssertMustNotSame)
}

func runTestAssertMustEqual(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case - should not call FailNow
	ok := mock.Run("success", func(mt T) {
		AssertMustEqual(mt, 42, 42, "equal values")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")
	AssertFalse(t, mock.Failed(), "Should not be marked as failed")
	AssertEqual(t, 2, len(mock.Logs), "Should have logs from assertion and continuation")

	mock.Reset()

	// Test failure case - should call FailNow and abort
	ok = mock.Run("failure", func(mt T) {
		AssertMustEqual(mt, 42, 24, "different values")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 1, len(mock.Errors), "Should have error from failed assertion")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNotEqual(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustNotEqual(mt, 42, 24, "different values")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustNotEqual(mt, 42, 42, "equal values")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustSliceEqual(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustSliceEqual(mt, S(1, 2, 3), S(1, 2, 3), "equal slices")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustSliceEqual(mt, S(1, 2, 3), S(1, 2, 4), "different slices")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustContains(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustContains(mt, "hello world", "world", "contains substring")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustContains(mt, "hello world", "xyz", "missing substring")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNotContain(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case (substring not present)
	ok := mock.Run("success", func(mt T) {
		AssertMustNotContain(mt, "hello world", "xyz", "no substring")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case (substring present)
	ok = mock.Run("failure", func(mt T) {
		AssertMustNotContain(mt, "hello world", "world", "has substring")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustError(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustError(mt, errors.New("test error"), "expects error")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustError(mt, nil, "expects error but got nil")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNoError(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustNoError(mt, nil, "expects no error")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustNoError(mt, errors.New("unexpected error"), "expects no error")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustPanic(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustPanic(mt, func() { panic("test panic") }, nil, "expects panic")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustPanic(mt, func() {}, nil, "expects panic but got none")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNoPanic(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustNoPanic(mt, func() {}, "expects no panic")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustNoPanic(mt, func() { panic("unexpected") }, "expects no panic")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustTrue(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustTrue(mt, true, "expects true")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustTrue(mt, false, "expects true but got false")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustFalse(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustFalse(mt, false, "expects false")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustFalse(mt, true, "expects false but got true")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustErrorIs(t *testing.T) {
	t.Helper()
	mock := &MockT{}
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(baseErr, errors.New("wrapped"))

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustErrorIs(mt, wrappedErr, baseErr, "error should match")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	otherErr := errors.New("other error")
	ok = mock.Run("failure", func(mt T) {
		AssertMustErrorIs(mt, wrappedErr, otherErr, "error should not match")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustTypeIs(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		var val any = "hello"
		result := AssertMustTypeIs[string](mt, val, "type should match")
		mt.Log("execution continues")
		AssertEqual(mt, "hello", result, "should return cast value")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		var val any = 42
		result := AssertMustTypeIs[string](mt, val, "type should not match")
		mt.Log("should not reach here")
		// result should be zero value but we shouldn't reach here
		_ = result
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNil(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustNil(mt, nil, "expects nil")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustNil(mt, "not nil", "expects nil but got value")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNotNil(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	// Test success case
	ok := mock.Run("success", func(mt T) {
		AssertMustNotNil(mt, "not nil", "expects not nil")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case
	ok = mock.Run("failure", func(mt T) {
		AssertMustNotNil(mt, nil, "expects not nil but got nil")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustSame(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	slice1 := []int{1, 2, 3}
	slice2 := slice1 // Same reference

	// Test success case - same reference
	ok := mock.Run("success", func(mt T) {
		AssertMustSame(mt, slice1, slice2, "same slice reference")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case - different references
	slice3 := []int{1, 2, 3} // Different reference
	ok = mock.Run("failure", func(mt T) {
		AssertMustSame(mt, slice1, slice3, "different slice references")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

func runTestAssertMustNotSame(t *testing.T) {
	t.Helper()
	mock := &MockT{}

	slice1 := []int{1, 2, 3}
	slice3 := []int{1, 2, 3} // Different reference

	// Test success case - different references
	ok := mock.Run("success", func(mt T) {
		AssertMustNotSame(mt, slice1, slice3, "different slice references")
		mt.Log("execution continues")
	})
	AssertTrue(t, ok, "Success case should not abort")

	mock.Reset()

	// Test failure case - same reference
	slice2 := slice1 // Same reference
	ok = mock.Run("failure", func(mt T) {
		AssertMustNotSame(mt, slice1, slice2, "same slice reference")
		mt.Log("should not reach here")
	})
	AssertFalse(t, ok, "Failure case should abort")
	AssertTrue(t, mock.Failed(), "Should be marked as failed")
	AssertEqual(t, 0, len(mock.Logs), "Should not reach continuation log")
}

// assertErrorContains checks that the last error from MockT contains the expected substring
func assertErrorContains(t *testing.T, mock *MockT, expected, desc string) {
	t.Helper()
	lastErr, ok := mock.LastError()
	AssertTrue(t, ok, "LastError ok for "+desc)
	AssertTrue(t, strings.Contains(lastErr, expected), desc)
}
