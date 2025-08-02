package core

import (
	"errors"
	"testing"
)

// TestCase validations
var _ TestCase = asRecoveredTestCase{}
var _ TestCase = catcherDoTestCase{}
var _ TestCase = catcherTryTestCase{}
var _ TestCase = catchTestCase{}
var _ TestCase = catchWithPanicRecoveryTestCase{}

type asRecoveredTestCase struct {
	input    any
	expected any
	name     string
}

func (tc asRecoveredTestCase) Name() string { return tc.name }

func (tc asRecoveredTestCase) MustNil() bool { return IsNil(tc.input) }

func newAsRecoveredTestCase(name string, input, expected any) asRecoveredTestCase {
	return asRecoveredTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func asRecoveredTestCases() []asRecoveredTestCase {
	var testError = errors.New("test error")
	var panicError = NewPanicError(1, "wrapped error")

	return []asRecoveredTestCase{
		newAsRecoveredTestCase("nil input", nil, nil),
		newAsRecoveredTestCase("string panic", "test panic", "test panic"),
		newAsRecoveredTestCase("error panic", testError, testError),
		newAsRecoveredTestCase("int panic", 42, 42),
		newAsRecoveredTestCase("already recovered", panicError, panicError),
	}
}

func (tc asRecoveredTestCase) Test(t *testing.T) {
	t.Helper()
	result := AsRecovered(tc.input)

	if tc.MustNil() {
		AssertNil(t, result, "nil result")
		return
	}

	if AssertNotNil(t, result, "not nil result") {
		err, ok := AssertTypeIs[Recovered](t, result, "recovered panic error")
		if ok {
			tc.testRecovered(t, err)
		}
	}
}

func (tc asRecoveredTestCase) testRecovered(t *testing.T, result Recovered) {
	if expectedError, asIs := tc.expected.(Recovered); asIs {
		// pass-through
		AssertEqual(t, expectedError, result, "pass-through recovered")
		return
	}

	// recovered
	recovered := result.Recovered()
	if s0, ok := tc.expected.(string); ok {
		if s1, ok := recovered.(string); ok {
			// string vs string
			AssertEqual(t, s0, s1, "recovered string")
		} else if err, ok := AssertTypeIs[error](t, recovered, "recovered error"); ok {
			// special case of strings converted to errors
			AssertEqual(t, s0, err.Error(), "recovered string")
		}

		return
	}

	AssertEqual(t, tc.expected, recovered, "recovered value")
}

func TestAsRecovered(t *testing.T) {
	RunTestCases(t, asRecoveredTestCases())
}

type catcherDoTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

func catcherDoTestCases() []catcherDoTestCase {
	return []catcherDoTestCase{
		newCatcherDoTestCase("successful function", func() error {
			return nil
		}, false, false),
		newCatcherDoTestCase("function returns error", func() error {
			return errors.New("test error")
		}, true, false),
		newCatcherDoTestCase("function panics with string", func() error {
			panic("test panic")
		}, true, true),
		newCatcherDoTestCase("function panics with error", func() error {
			panic(errors.New("panic error"))
		}, true, true),
		newCatcherDoTestCase("function panics with int", func() error {
			panic(42)
		}, true, true),
		newCatcherDoTestCase("nil function", nil, false, false),
	}
}

func newCatcherDoTestCase(name string, fn func() error, expectError, expectPanic bool) catcherDoTestCase {
	return catcherDoTestCase{
		name:        name,
		fn:          fn,
		expectError: expectError,
		expectPanic: expectPanic,
	}
}

func (tc catcherDoTestCase) Name() string {
	return tc.name
}

func (tc catcherDoTestCase) Test(t *testing.T) {
	t.Helper()
	var catcher Catcher

	err := catcher.Do(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catcher.Do error")
	} else {
		AssertNoError(t, err, "Catcher.Do error")
	}

	if tc.expectPanic {
		if recovered, ok := AssertTypeIs[Recovered](t, err, "Recovered error type"); ok {
			AssertNotNil(t, recovered.Recovered(), "recovered panic value")
		}
	}
}

func TestCatcherDo(t *testing.T) {
	RunTestCases(t, catcherDoTestCases())
}

type catcherTryTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

func catcherTryTestCases() []catcherTryTestCase {
	return []catcherTryTestCase{
		newCatcherTryTestCase("successful function", func() error {
			return nil
		}, false, false),
		newCatcherTryTestCase("function returns error", func() error {
			return errors.New("test error")
		}, true, false),
		newCatcherTryTestCase("function panics", func() error {
			panic("test panic")
		}, false, true),
		newCatcherTryTestCase("nil function", nil, false, false),
	}
}

func newCatcherTryTestCase(name string, fn func() error, expectError, expectPanic bool) catcherTryTestCase {
	return catcherTryTestCase{
		name:        name,
		fn:          fn,
		expectError: expectError,
		expectPanic: expectPanic,
	}
}

func (tc catcherTryTestCase) Name() string {
	return tc.name
}

func (tc catcherTryTestCase) Test(t *testing.T) {
	t.Helper()
	var catcher Catcher
	err := catcher.Try(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catcher.Try error")
	} else {
		AssertNoError(t, err, "Catcher.Try error")
	}

	// Check recovered panic
	recovered := catcher.Recovered()
	if tc.expectPanic {
		_, _ = AssertTypeIs[Recovered](t, recovered, "expected recovered panic")
	} else {
		AssertNil(t, recovered, "no recovered panic")
	}
}

func TestCatcherTry(t *testing.T) {
	RunTestCases(t, catcherTryTestCases())
}

func TestCatcherRecovered(t *testing.T) {
	var catcher Catcher

	// Initially no panic
	recovered := catcher.Recovered()
	if !AssertNil(t, recovered, "initially nil recovered") {
		t.Fail()
	}

	// After panic
	_ = catcher.Try(func() error {
		panic("test panic")
	})

	recovered = catcher.Recovered()
	if !AssertNotNil(t, recovered, "recovered panic after Try") {
		t.Fail()
	}

	// String panics get converted to errors by NewPanicError
	if err, ok := recovered.Recovered().(error); ok {
		AssertEqual(t, "test panic", err.Error(), "error message")
	} else {
		AssertTypeIs[string](t, recovered.Recovered(), "string panic type")
	}
}

func TestCatcherConcurrent(t *testing.T) {
	var catcher Catcher

	// Use a channel to coordinate goroutines
	done := make(chan bool, 2)

	// Test that only the first panic is stored
	go func() {
		_ = catcher.Try(func() error {
			panic("first panic")
		})
		done <- true
	}()

	go func() {
		_ = catcher.Try(func() error {
			panic("second panic")
		})
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	recovered := catcher.Recovered()
	AssertNotNil(t, recovered, "concurrent recovered panic")

	// Should be either "first panic" or "second panic" (converted to errors)
	panicValue := recovered.Recovered()
	if err, ok := panicValue.(error); ok {
		errorStr := err.Error()
		AssertTrue(t, errorStr == "first panic" || errorStr == "second panic",
			"panic value is first or second")
	} else {
		_, _ = AssertTypeIs[error](t, panicValue, "concurrent panic type")
	}
}

type catchTestCase struct {
	fn          func() error
	name        string
	expectError bool
}

func catchTestCases() []catchTestCase {
	return []catchTestCase{
		newCatchTestCase("successful function", func() error {
			return nil
		}, false),
		newCatchTestCase("function returns error", func() error {
			return errors.New("test error")
		}, true),
		newCatchTestCase("function panics", func() error {
			panic("test panic")
		}, true),
	}
}

func newCatchTestCase(name string, fn func() error, expectError bool) catchTestCase {
	return catchTestCase{
		name:        name,
		fn:          fn,
		expectError: expectError,
	}
}

func (tc catchTestCase) Name() string {
	return tc.name
}

func (tc catchTestCase) Test(t *testing.T) {
	t.Helper()
	err := Catch(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catch error")
	} else {
		AssertNoError(t, err, "Catch error")
	}
}

func TestCatch(t *testing.T) {
	RunTestCases(t, catchTestCases())
}

type catchWithPanicRecoveryTestCase struct {
	value any
	name  string
}

func catchWithPanicRecoveryTestCases() []catchWithPanicRecoveryTestCase {
	return []catchWithPanicRecoveryTestCase{
		newCatchWithPanicRecoveryTestCase("string panic", "string panic"),
		newCatchWithPanicRecoveryTestCase("int panic", 42),
		newCatchWithPanicRecoveryTestCase("float panic", 3.14),
		newCatchWithPanicRecoveryTestCase("error panic", errors.New("error panic")),
		newCatchWithPanicRecoveryTestCase("formatted error", errors.New("formatted error")),
		// Skip slice and map as they are not comparable
	}
}

func newCatchWithPanicRecoveryTestCase(name string, value any) catchWithPanicRecoveryTestCase {
	return catchWithPanicRecoveryTestCase{
		name:  name,
		value: value,
	}
}

func (tc catchWithPanicRecoveryTestCase) Name() string {
	return tc.name
}

func (tc catchWithPanicRecoveryTestCase) Test(t *testing.T) {
	t.Helper()
	err := Catch(func() error {
		panic(tc.value)
	})

	AssertError(t, err, "expected error from panic")

	if recovered, ok := err.(Recovered); ok {
		panicValue := recovered.Recovered()

		// Handle string conversion to error by NewPanicError
		if s, ok := tc.value.(string); ok {
			if err, ok := panicValue.(error); ok {
				AssertEqual(t, s, err.Error(), "error message")
			} else {
				_, _ = AssertTypeIs[error](t, panicValue, "string panic type")
			}
		} else {
			AssertEqual(t, tc.value, panicValue, "panic value")
		}
	} else {
		_, _ = AssertTypeIs[Recovered](t, err, "Recovered error type")
	}
}

func TestCatchWithPanicRecovery(t *testing.T) {
	RunTestCases(t, catchWithPanicRecoveryTestCases())
}
