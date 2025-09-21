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
var _ TestCase = mustTSuccessTestCase{}
var _ TestCase = mustTPanicTestCase{}
var _ TestCase = maybeTTestCase{}

type asRecoveredTestCase struct {
	input    any
	expected any
	name     string
}

func (tc asRecoveredTestCase) Name() string { return tc.name }

func (tc asRecoveredTestCase) MustNil() bool { return IsNil(tc.input) }

func newAsRecoveredTestCase(name string, input, expected any) TestCase {
	return asRecoveredTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func makeAsRecoveredConversionTestCases() []TestCase {
	var testError = errors.New("test error")
	var panicError = NewPanicError(1, "wrapped error")

	return []TestCase{
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
	RunTestCases(t, makeAsRecoveredConversionTestCases())
}

type catcherDoTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

func makeCatcherDoExecutionTestCases() []TestCase {
	return []TestCase{
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

func newCatcherDoTestCase(name string, fn func() error, expectError, expectPanic bool) TestCase {
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
	RunTestCases(t, makeCatcherDoExecutionTestCases())
}

type catcherTryTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

func makeCatcherTryCatchTestCases() []TestCase {
	return []TestCase{
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

func newCatcherTryTestCase(name string, fn func() error, expectError, expectPanic bool) TestCase {
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
	RunTestCases(t, makeCatcherTryCatchTestCases())
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

func makeCatchBasicTestCases() []TestCase {
	return []TestCase{
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

func newCatchTestCase(name string, fn func() error, expectError bool) TestCase {
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
	RunTestCases(t, makeCatchBasicTestCases())
}

type catchWithPanicRecoveryTestCase struct {
	value any
	name  string
}

func makeCatchWithPanicRecoveryTestCases() []TestCase {
	return []TestCase{
		newCatchWithPanicRecoveryTestCase("string panic", "string panic"),
		newCatchWithPanicRecoveryTestCase("int panic", 42),
		newCatchWithPanicRecoveryTestCase("float panic", 3.14),
		newCatchWithPanicRecoveryTestCase("error panic", errors.New("error panic")),
		newCatchWithPanicRecoveryTestCase("formatted error", errors.New("formatted error")),
		// Skip slice and map as they are not comparable
	}
}

func newCatchWithPanicRecoveryTestCase(name string, value any) TestCase {
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
	RunTestCases(t, makeCatchWithPanicRecoveryTestCases())
}

// testMust is a helper to test Must function by catching panics.
// It wraps Must calls in panic recovery to allow testing both success
// and panic scenarios. Returns the value and any recovered panic as an error.
func runTestMust[T any](v0 T, e0 error) (v1 T, e1 error) {
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
func newMustSuccessTestCase(name string, value any) TestCase {
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
func runTestMustT[V comparable](t *testing.T, tc mustSuccessTestCase, value V) {
	t.Helper()

	got, err := runTestMust(value, tc.err)
	AssertNoError(t, err, "Must success")
	AssertEqual(t, value, got, "Must value")
}

// testMustSlice is a specialized test helper for Must function with slice types.
func runTestMustSlice[V any](t *testing.T, tc mustSuccessTestCase, value []V) {
	t.Helper()

	got, err := runTestMust(value, tc.err)
	AssertNoError(t, err, "Must success")
	AssertSliceEqual(t, value, got, "Must slice")
}

// testMustWithValue dispatches to the appropriate test helper.
func (tc mustSuccessTestCase) testMustWithValue(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		runTestMustT(t, tc, v)
	case int:
		runTestMustT(t, tc, v)
	case bool:
		runTestMustT(t, tc, v)
	case []int:
		runTestMustSlice(t, tc, v)
	case *int:
		runTestMustT(t, tc, v)
	case struct{ Name string }:
		runTestMustT(t, tc, v)
	default:
		t.Errorf("unsupported test value type: %T", tc.value)
	}
}

func makeMustSuccessTestCases() []TestCase {
	return []TestCase{
		newMustSuccessTestCase("string success", "hello"),
		newMustSuccessTestCase("int success", 42),
		newMustSuccessTestCase("bool success", true),
		newMustSuccessTestCase("slice success", S(1, 2, 3)),
		newMustSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}
}

func TestMust(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		RunTestCases(t, makeMustSuccessTestCases())
	})
	t.Run("panic", func(t *testing.T) {
		RunTestCases(t, makeMustPanicTestCases())
	})
}

// mustPanicTestCase tests Must function panic scenarios where Must should panic.
type mustPanicTestCase struct {
	// Large fields first - error interface (8 bytes)
	err error

	// Small fields last - string (16 bytes)
	name string
}

// test validates that Must panics with proper PanicError when err is not nil.
func newMustPanicTestCase(name string, err error) TestCase {
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

	_, err := runTestMust("value", tc.err)
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

func makeMustPanicTestCases() []TestCase {
	return []TestCase{
		newMustPanicTestCase("simple error", errors.New("test error")),
		newMustPanicTestCase("formatted error", fmt.Errorf("formatted error: %d", 42)),
		newMustPanicTestCase("wrapped error", fmt.Errorf("wrapped: %w", errors.New("inner"))),
	}
}

type maybeTestCase struct {
	// Large fields first - interfaces (8 bytes)
	value any
	err   error

	// Small fields last - string (16 bytes)
	name string
}

func newMaybeTestCase(name string, value any, err error) TestCase {
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

func makeMaybeIgnoreErrorsTestCases() []TestCase {
	return []TestCase{
		newMaybeTestCase("string with nil error", "hello", nil),
		newMaybeTestCase("string with error", "world", errors.New("ignored error")),
		newMaybeTestCase("int with nil error", 42, nil),
		newMaybeTestCase("int with error", 0, errors.New("another ignored error")),
		newMaybeTestCase("nil pointer with error", (*int)(nil), errors.New("pointer error")),
		newMaybeTestCase("struct with error", struct{ Name string }{"test"}, fmt.Errorf("formatted: %d", 123)),
	}
}

func TestMaybe(t *testing.T) {
	t.Run("ignore errors", func(t *testing.T) {
		RunTestCases(t, makeMaybeIgnoreErrorsTestCases())
	})
}

// testMustOK is a helper to test MustOK function by catching panics.
// It wraps MustOK calls in panic recovery to allow testing both success
// and panic scenarios. Returns the value and any recovered panic as an error.
func runTestMustOK[T any](v0 T, ok bool) (v1 T, e1 error) {
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
func newMustOKSuccessTestCase(name string, value any) TestCase {
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
func runTestMustOKT[V comparable](t *testing.T, tc mustOKSuccessTestCase, value V) {
	t.Helper()

	got, err := runTestMustOK(value, tc.ok)
	AssertNoError(t, err, "MustOK success")
	AssertEqual(t, value, got, "MustOK value")
}

// testMustOKSlice is a specialized test helper for MustOK function with slice types.
func runTestMustOKSlice[V any](t *testing.T, tc mustOKSuccessTestCase, value []V) {
	t.Helper()

	got, err := runTestMustOK(value, tc.ok)
	AssertNoError(t, err, "MustOK success")
	AssertSliceEqual(t, value, got, "MustOK slice")
}

// testMustOKWithValue dispatches to the appropriate test helper.
func (tc mustOKSuccessTestCase) testMustOKWithValue(t *testing.T) {
	t.Helper()

	// Test with different types using type switches
	switch v := tc.value.(type) {
	case string:
		runTestMustOKT(t, tc, v)
	case int:
		runTestMustOKT(t, tc, v)
	case bool:
		runTestMustOKT(t, tc, v)
	case []int:
		runTestMustOKSlice(t, tc, v)
	case *int:
		runTestMustOKT(t, tc, v)
	case struct{ Name string }:
		runTestMustOKT(t, tc, v)
	default:
		t.Errorf("unsupported test value type: %T", tc.value)
	}
}

func makeMustOKSuccessTestCases() []TestCase {
	return []TestCase{
		newMustOKSuccessTestCase("string success", "hello"),
		newMustOKSuccessTestCase("int success", 42),
		newMustOKSuccessTestCase("bool success", true),
		newMustOKSuccessTestCase("slice success", S(1, 2, 3)),
		newMustOKSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustOKSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}
}

func TestMustOK(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		RunTestCases(t, makeMustOKSuccessTestCases())
	})
	t.Run("panic", func(t *testing.T) {
		RunTestCases(t, makeMustOKPanicTestCases())
	})
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
func newMustOKPanicTestCase(name string, value any) TestCase {
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

	_, err := runTestMustOK(tc.value, tc.ok)
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

func makeMustOKPanicTestCases() []TestCase {
	return []TestCase{
		newMustOKPanicTestCase("string panic", "hello"),
		newMustOKPanicTestCase("int panic", 42),
		newMustOKPanicTestCase("bool panic", false),
		newMustOKPanicTestCase("slice panic", S(1, 2, 3)),
		newMustOKPanicTestCase("nil pointer panic", (*int)(nil)),
		newMustOKPanicTestCase("struct panic", struct{ Name string }{"test"}),
	}
}

type maybeOKTestCase struct {
	// Large fields first - interfaces (8 bytes)
	value any

	// Small fields last - string (16 bytes), bool (1 byte)
	name string
	ok   bool
}

func newMaybeOKTestCase(name string, value any, ok bool) TestCase {
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

func makeMaybeOKOkHandlingTestCases() []TestCase {
	return []TestCase{
		newMaybeOKTestCase("string with true", "hello", true),
		newMaybeOKTestCase("string with false", "world", false),
		newMaybeOKTestCase("int with true", 42, true),
		newMaybeOKTestCase("int with false", 0, false),
		newMaybeOKTestCase("nil pointer with false", (*int)(nil), false),
		newMaybeOKTestCase("struct with false", struct{ Name string }{"test"}, false),
	}
}

func TestMaybeOK(t *testing.T) {
	t.Run("ok handling", func(t *testing.T) {
		RunTestCases(t, makeMaybeOKOkHandlingTestCases())
	})
}

// Test cases for MustT function
type mustTSuccessTestCase struct {
	input    any
	expected any
	name     string
}

func newMustTSuccessTestCase(name string, input, expected any) TestCase {
	return mustTSuccessTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func (tc mustTSuccessTestCase) Name() string {
	return tc.name
}

func (tc mustTSuccessTestCase) Test(t *testing.T) {
	t.Helper()

	// Test successful type conversions that don't panic
	switch expected := tc.expected.(type) {
	case string:
		got := MustT[string](tc.input)
		AssertEqual(t, expected, got, "MustT to string")
	case int:
		got := MustT[int](tc.input)
		AssertEqual(t, expected, got, "MustT to int")
	case float64:
		got := MustT[float64](tc.input)
		AssertEqual(t, expected, got, "MustT to float64")
	case error:
		got := MustT[error](tc.input)
		AssertEqual(t, expected, got, "MustT to error")
	case fmt.Stringer:
		got := MustT[fmt.Stringer](tc.input)
		AssertEqual(t, expected.String(), got.String(), "MustT to fmt.Stringer")
	default:
		t.Errorf("unsupported expected type: %T", tc.expected)
	}
}

func makeMustTSuccessTestCases() []TestCase {
	testErr := errors.New("test")
	return []TestCase{
		newMustTSuccessTestCase("string to string", "hello", "hello"),
		newMustTSuccessTestCase("int to int", 42, 42),
		newMustTSuccessTestCase("float64 to float64", 3.14, 3.14),
		newMustTSuccessTestCase("error to error", testErr, testErr),
		newMustTSuccessTestCase("string to fmt.Stringer",
			mockStringer{value: "test"}, mockStringer{value: "test"}),
	}
}

func TestMustT(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		RunTestCases(t, makeMustTSuccessTestCases())
	})
	t.Run("panic", func(t *testing.T) {
		RunTestCases(t, makeMustTPanicTestCases())
	})
}

type mustTPanicTestCase struct {
	input  any
	target string // target type description for error message
	name   string
}

func newMustTPanicTestCase(name, target string, input any) TestCase {
	return mustTPanicTestCase{
		name:   name,
		target: target,
		input:  input,
	}
}

func (tc mustTPanicTestCase) Name() string {
	return tc.name
}

func (tc mustTPanicTestCase) Test(t *testing.T) {
	t.Helper()

	// Test type conversions that should panic
	switch tc.target {
	case "string":
		AssertPanic(t, func() {
			_ = MustT[string](tc.input)
		}, nil, "MustT should panic for invalid string conversion")
	case "int":
		AssertPanic(t, func() {
			_ = MustT[int](tc.input)
		}, nil, "MustT should panic for invalid int conversion")
	case "error":
		AssertPanic(t, func() {
			_ = MustT[error](tc.input)
		}, nil, "MustT should panic for invalid error conversion")
	case "fmt.Stringer":
		AssertPanic(t, func() {
			_ = MustT[fmt.Stringer](tc.input)
		}, nil, "MustT should panic for invalid fmt.Stringer conversion")
	default:
		t.Errorf("unsupported target type: %s", tc.target)
	}
}

func makeMustTPanicTestCases() []TestCase {
	return []TestCase{
		newMustTPanicTestCase("int to string", "string", 42),
		newMustTPanicTestCase("string to int", "int", "hello"),
		newMustTPanicTestCase("string to error", "error", "not an error"),
		newMustTPanicTestCase("int to fmt.Stringer", "fmt.Stringer", 42),
		newMustTPanicTestCase("nil to string", "string", nil),
	}
}

// Test cases for MaybeT function
type maybeTTestCase struct {
	input    any
	target   string // target type description
	expected any    // expected result (zero value if conversion fails)
	name     string
}

func newMaybeTTestCase(name, target string, input, expected any) TestCase {
	return maybeTTestCase{
		name:     name,
		target:   target,
		input:    input,
		expected: expected,
	}
}

func (tc maybeTTestCase) Name() string {
	return tc.name
}

func (tc maybeTTestCase) Test(t *testing.T) {
	t.Helper()

	// Test type conversions with MaybeT (never panics)
	switch tc.target {
	case "string":
		tc.testString(t)
	case "int":
		tc.testInt(t)
	case "error":
		tc.testError(t)
	case "fmt.Stringer":
		tc.testStringer(t)
	default:
		t.Errorf("unsupported target type: %s", tc.target)
	}
}

func (tc maybeTTestCase) testString(t *testing.T) {
	t.Helper()
	got := MaybeT[string](tc.input)
	expected, ok := tc.expected.(string)
	if !ok {
		t.Errorf("expected value is not a string: %T", tc.expected)
		return
	}
	AssertEqual(t, expected, got, "MaybeT to string")
}

func (tc maybeTTestCase) testInt(t *testing.T) {
	t.Helper()
	got := MaybeT[int](tc.input)
	expected, ok := tc.expected.(int)
	if !ok {
		t.Errorf("expected value is not an int: %T", tc.expected)
		return
	}
	AssertEqual(t, expected, got, "MaybeT to int")
}

func (tc maybeTTestCase) testError(t *testing.T) {
	t.Helper()
	got := MaybeT[error](tc.input)
	if tc.expected == nil {
		AssertNil(t, got, "MaybeT to error (nil)")
	} else {
		expected, ok := tc.expected.(error)
		if !ok {
			t.Errorf("expected value is not an error: %T", tc.expected)
			return
		}
		AssertEqual(t, expected, got, "MaybeT to error")
	}
}

func (tc maybeTTestCase) testStringer(t *testing.T) {
	t.Helper()
	got := MaybeT[fmt.Stringer](tc.input)
	if tc.expected == nil {
		AssertNil(t, got, "MaybeT to fmt.Stringer (nil)")
	} else {
		expected, ok := tc.expected.(fmt.Stringer)
		if !ok {
			t.Errorf("expected value is not a fmt.Stringer: %T", tc.expected)
			return
		}
		AssertEqual(t, expected.String(), got.String(), "MaybeT to fmt.Stringer")
	}
}

func makeMaybeTTypeConversionTestCases() []TestCase {
	testErr := errors.New("test")
	return []TestCase{
		// Successful conversions
		newMaybeTTestCase("string to string", "string", "hello", "hello"),
		newMaybeTTestCase("int to int", "int", 42, 42),
		newMaybeTTestCase("error to error", "error", testErr, testErr),
		newMaybeTTestCase("stringer to stringer", "fmt.Stringer",
			mockStringer{value: "test"}, mockStringer{value: "test"}),

		// Failed conversions (should return zero values)
		newMaybeTTestCase("int to string", "string", 42, ""),
		newMaybeTTestCase("string to int", "int", "hello", 0),
		newMaybeTTestCase("string to error", "error", "not an error", nil),
		newMaybeTTestCase("int to fmt.Stringer", "fmt.Stringer", 42, nil),
		newMaybeTTestCase("nil to string", "string", nil, ""),
		newMaybeTTestCase("nil to int", "int", nil, 0),
	}
}

func TestMaybeT(t *testing.T) {
	t.Run("type conversion", func(t *testing.T) {
		RunTestCases(t, makeMaybeTTypeConversionTestCases())
	})
}

// Helper type for testing fmt.Stringer interface
type mockStringer struct {
	value string
}

func (ms mockStringer) String() string {
	return ms.value
}
