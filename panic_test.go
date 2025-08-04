package core

import (
	"errors"
	"fmt"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = asRecoveredTestCase{}
var _ TestCase = catcherDoTestCase{}
var _ TestCase = catcherTryTestCase{}
var _ TestCase = catchTestCase{}
var _ TestCase = catchWithPanicRecoveryTestCase{}
var _ TestCase = mustSuccessTestCase{}
var _ TestCase = mustPanicTestCase{}
var _ TestCase = maybeTestCase{}
var _ TestCase = mustOKSuccessTestCase{}
var _ TestCase = mustOKPanicTestCase{}
var _ TestCase = maybeOKTestCase{}

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

// testMust is a helper to test Must function by catching panics.
// It wraps Must calls in panic recovery to allow testing both success
// and panic scenarios. Returns the value and any recovered panic as an error.
func testMust[T any](v0 T, e0 error) (v1 T, e1 error) {
	defer func() {
		if e2 := AsRecovered(recover()); e2 != nil {
			e1 = e2
		}
	}()

	v1 = Must(v0, e0)
	return v1, nil
}

// mustSuccessTestCase tests Must function success scenarios where no panic should occur.
type mustSuccessTestCase struct {
	// Large fields first - interfaces (8 bytes on 64-bit)
	value any
	err   error

	// Small fields last - strings (16 bytes on 64-bit)
	name string
}

// newMustSuccessTestCase creates a new mustSuccessTestCase with the given parameters.
// For success cases, err is always nil.
func newMustSuccessTestCase(name string, value any) mustSuccessTestCase {
	return mustSuccessTestCase{
		value: value,
		err:   nil,
		name:  name,
	}
}

// test validates that Must returns the value unchanged when err is nil.
func (tc mustSuccessTestCase) Name() string {
	return tc.name
}

func (tc mustSuccessTestCase) Test(t *testing.T) {
	t.Helper()

	tc.testMustWithValue(t)
}

// testMustT is a generic test helper for Must function with comparable types.
// It handles the common pattern of testing Must with a value and verifying
// the result matches expectations.
func testMustT[V comparable](t *testing.T, tc mustSuccessTestCase, value V) {
	t.Helper()

	got, err := testMust(value, tc.err)
	AssertNoError(t, err, "Must success")
	AssertEqual(t, value, got, "Must value")
}

// testMustSlice is a specialized test helper for Must function with slice types.
func testMustSlice[V any](t *testing.T, tc mustSuccessTestCase, value []V) {
	t.Helper()

	got, err := testMust(value, tc.err)
	AssertNoError(t, err, "Must success")
	AssertSliceEqual(t, value, got, "Must slice")
}

// testMustWithValue dispatches to the appropriate test helper.
func (tc mustSuccessTestCase) testMustWithValue(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		testMustT(t, tc, v)
	case int:
		testMustT(t, tc, v)
	case bool:
		testMustT(t, tc, v)
	case []int:
		testMustSlice(t, tc, v)
	case *int:
		testMustT(t, tc, v)
	case struct{ Name string }:
		testMustT(t, tc, v)
	default:
		t.Errorf("unsupported test value type: %T", tc.value)
	}
}

func TestMustSuccess(t *testing.T) {
	testCases := []mustSuccessTestCase{
		newMustSuccessTestCase("string success", "hello"),
		newMustSuccessTestCase("int success", 42),
		newMustSuccessTestCase("bool success", true),
		newMustSuccessTestCase("slice success", S(1, 2, 3)),
		newMustSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}

	RunTestCases(t, testCases)
}

// mustPanicTestCase tests Must function panic scenarios where Must should panic.
type mustPanicTestCase struct {
	// Large fields first - error interface (8 bytes)
	err error

	// Small fields last - string (16 bytes)
	name string
}

// test validates that Must panics with proper PanicError when err is not nil.
func newMustPanicTestCase(name string, err error) mustPanicTestCase {
	return mustPanicTestCase{
		name: name,
		err:  err,
	}
}

func (tc mustPanicTestCase) Name() string {
	return tc.name
}

func (tc mustPanicTestCase) Test(t *testing.T) {
	t.Helper()

	_, err := testMust("value", tc.err)
	AssertError(t, err, "Must panic")

	// Verify the panic contains our original error
	AssertTrue(t, errors.Is(err, tc.err), "panic wraps original")

	// Verify it's a proper PanicError
	panicErr, ok := AssertTypeIs[*PanicError](t, err, "panic type")
	if ok {
		// Verify stack trace exists
		stack := panicErr.CallStack()
		AssertTrue(t, len(stack) > 0, "has stack trace")
	}
}

func TestMustPanic(t *testing.T) {
	testCases := []mustPanicTestCase{
		newMustPanicTestCase("simple error", errors.New("test error")),
		newMustPanicTestCase("formatted error", fmt.Errorf("formatted error: %d", 42)),
		newMustPanicTestCase("wrapped error", fmt.Errorf("wrapped: %w", errors.New("inner"))),
	}

	RunTestCases(t, testCases)
}

type maybeTestCase struct {
	// Large fields first - interfaces (8 bytes)
	value any
	err   error

	// Small fields last - string (16 bytes)
	name string
}

func newMaybeTestCase(name string, value any, err error) maybeTestCase {
	return maybeTestCase{
		name:  name,
		value: value,
		err:   err,
	}
}

func (tc maybeTestCase) Name() string {
	return tc.name
}

func (tc maybeTestCase) Test(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe string")
	case int:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe int")
	case *int:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe pointer")
	case struct{ Name string }:
		got := Maybe(v, tc.err)
		AssertEqual(t, v, got, "Maybe struct")
	default:
		t.Errorf("unsupported test value type: %T", tc.value)
	}
}

func TestMaybe(t *testing.T) {
	testCases := []maybeTestCase{
		newMaybeTestCase("string with nil error", "hello", nil),
		newMaybeTestCase("string with error", "world", errors.New("ignored error")),
		newMaybeTestCase("int with nil error", 42, nil),
		newMaybeTestCase("int with error", 0, errors.New("another ignored error")),
		newMaybeTestCase("nil pointer with error", (*int)(nil), errors.New("pointer error")),
		newMaybeTestCase("struct with error", struct{ Name string }{"test"}, fmt.Errorf("formatted: %d", 123)),
	}

	RunTestCases(t, testCases)
}

// testMustOK is a helper to test MustOK function by catching panics.
// It wraps MustOK calls in panic recovery to allow testing both success
// and panic scenarios. Returns the value and any recovered panic as an error.
func testMustOK[T any](v0 T, ok bool) (v1 T, e1 error) {
	defer func() {
		if e2 := AsRecovered(recover()); e2 != nil {
			e1 = e2
		}
	}()

	v1 = MustOK(v0, ok)
	return v1, nil
}

// mustOKSuccessTestCase tests MustOK function success scenarios where no panic should occur.
type mustOKSuccessTestCase struct {
	// Large fields first - interfaces (8 bytes on 64-bit)
	value any

	// Small fields last - strings (16 bytes on 64-bit), bool (1 byte)
	name string
	ok   bool
}

// newMustOKSuccessTestCase creates a new mustOKSuccessTestCase with the given parameters.
// For success cases, ok is always true.
func newMustOKSuccessTestCase(name string, value any) mustOKSuccessTestCase {
	return mustOKSuccessTestCase{
		value: value,
		ok:    true,
		name:  name,
	}
}

func (tc mustOKSuccessTestCase) Name() string {
	return tc.name
}

func (tc mustOKSuccessTestCase) Test(t *testing.T) {
	t.Helper()

	tc.testMustOKWithValue(t)
}

// testMustOKT is a generic test helper for MustOK function with comparable types.
// It handles the common pattern of testing MustOK with a value and verifying
// the result matches expectations.
func testMustOKT[V comparable](t *testing.T, tc mustOKSuccessTestCase, value V) {
	t.Helper()

	got, err := testMustOK(value, tc.ok)
	AssertNoError(t, err, "MustOK success")
	AssertEqual(t, value, got, "MustOK value")
}

// testMustOKSlice is a specialized test helper for MustOK function with slice types.
func testMustOKSlice[V any](t *testing.T, tc mustOKSuccessTestCase, value []V) {
	t.Helper()

	got, err := testMustOK(value, tc.ok)
	AssertNoError(t, err, "MustOK success")
	AssertSliceEqual(t, value, got, "MustOK slice")
}

// testMustOKWithValue dispatches to the appropriate test helper.
func (tc mustOKSuccessTestCase) testMustOKWithValue(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		testMustOKT(t, tc, v)
	case int:
		testMustOKT(t, tc, v)
	case bool:
		testMustOKT(t, tc, v)
	case []int:
		testMustOKSlice(t, tc, v)
	case *int:
		testMustOKT(t, tc, v)
	case struct{ Name string }:
		testMustOKT(t, tc, v)
	default:
		t.Errorf("unsupported test value type: %T", tc.value)
	}
}

func TestMustOKSuccess(t *testing.T) {
	testCases := []mustOKSuccessTestCase{
		newMustOKSuccessTestCase("string success", "hello"),
		newMustOKSuccessTestCase("int success", 42),
		newMustOKSuccessTestCase("bool success", true),
		newMustOKSuccessTestCase("slice success", S(1, 2, 3)),
		newMustOKSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustOKSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}

	RunTestCases(t, testCases)
}

// mustOKPanicTestCase tests MustOK function panic scenarios where MustOK should panic.
type mustOKPanicTestCase struct {
	// Large fields first - interfaces (8 bytes)
	value any

	// Small fields last - string (16 bytes), bool (1 byte)
	name string
	ok   bool
}

// newMustOKPanicTestCase creates a new mustOKPanicTestCase with the given parameters.
// For panic cases, ok is always false.
func newMustOKPanicTestCase(name string, value any) mustOKPanicTestCase {
	return mustOKPanicTestCase{
		name:  name,
		value: value,
		ok:    false,
	}
}

func (tc mustOKPanicTestCase) Name() string {
	return tc.name
}

func (tc mustOKPanicTestCase) Test(t *testing.T) {
	t.Helper()

	_, err := testMustOK(tc.value, tc.ok)
	AssertError(t, err, "MustOK panic")

	// Verify it's a proper PanicError
	panicErr, ok := AssertTypeIs[*PanicError](t, err, "panic type")
	if ok {
		// Verify stack trace exists
		stack := panicErr.CallStack()
		AssertTrue(t, len(stack) > 0, "has stack trace")

		// Verify the error message contains our expected text
		AssertContains(t, panicErr.Error(), "core.MustOK: operation failed", "panic message")
	}
}

func TestMustOKPanic(t *testing.T) {
	testCases := []mustOKPanicTestCase{
		newMustOKPanicTestCase("string panic", "hello"),
		newMustOKPanicTestCase("int panic", 42),
		newMustOKPanicTestCase("bool panic", false),
		newMustOKPanicTestCase("slice panic", S(1, 2, 3)),
		newMustOKPanicTestCase("nil pointer panic", (*int)(nil)),
		newMustOKPanicTestCase("struct panic", struct{ Name string }{"test"}),
	}

	RunTestCases(t, testCases)
}

type maybeOKTestCase struct {
	// Large fields first - interfaces (8 bytes)
	value any

	// Small fields last - string (16 bytes), bool (1 byte)
	name string
	ok   bool
}

func newMaybeOKTestCase(name string, value any, ok bool) maybeOKTestCase {
	return maybeOKTestCase{
		name:  name,
		value: value,
		ok:    ok,
	}
}

func (tc maybeOKTestCase) Name() string {
	return tc.name
}

func (tc maybeOKTestCase) Test(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		got := MaybeOK(v, tc.ok)
		AssertEqual(t, v, got, "MaybeOK string")
	case int:
		got := MaybeOK(v, tc.ok)
		AssertEqual(t, v, got, "MaybeOK int")
	case *int:
		got := MaybeOK(v, tc.ok)
		AssertEqual(t, v, got, "MaybeOK pointer")
	case struct{ Name string }:
		got := MaybeOK(v, tc.ok)
		AssertEqual(t, v, got, "MaybeOK struct")
	default:
		t.Errorf("unsupported test value type: %T", tc.value)
	}
}

func TestMaybeOK(t *testing.T) {
	testCases := []maybeOKTestCase{
		newMaybeOKTestCase("string with true", "hello", true),
		newMaybeOKTestCase("string with false", "world", false),
		newMaybeOKTestCase("int with true", 42, true),
		newMaybeOKTestCase("int with false", 0, false),
		newMaybeOKTestCase("nil pointer with false", (*int)(nil), false),
		newMaybeOKTestCase("struct with false", struct{ Name string }{"test"}, false),
	}

	RunTestCases(t, testCases)
}
