package core

import (
	"errors"
	"fmt"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = asRecoveredTestCase{}
	_ TestCase = catcherDoTestCase{}
	_ TestCase = catcherTryTestCase{}
	_ TestCase = catchTestCase{}
	_ TestCase = catchWithPanicRecoveryTestCase{}
	_ TestCase = mustSuccessTestCase[int]{}
	_ TestCase = mustPanicTestCase{}
	_ TestCase = maybeTestCase[int]{}
	_ TestCase = mustOKSuccessTestCase[int]{}
	_ TestCase = mustOKPanicTestCase{}
	_ TestCase = maybeOKTestCase[int]{}
	_ TestCase = mustTSuccessTestCase[int]{}
	_ TestCase = mustTPanicTestCase[int]{}
	_ TestCase = mustTReasonTestCase[int]{}
	_ TestCase = maybeTTestCase[int]{}
)

// asRecoveredTestCase states what AsRecovered wraps a payload in: a
// Recovered carrying that same payload, with a string converted to an
// error of the same text. A nil input and an already-recovered value
// are neither of them payloads being wrapped, and have their own tests.
type asRecoveredTestCase struct {
	input any
	name  string

	wantConverted bool
}

func newAsRecoveredTestCase(name string, input any) asRecoveredTestCase {
	return asRecoveredTestCase{
		input:         input,
		name:          name,
		wantConverted: false,
	}
}

// newAsRecoveredTestCaseString declares a row whose payload is a string,
// so the value recovered is an error carrying that same text.
func newAsRecoveredTestCaseString(name, payload string) asRecoveredTestCase {
	return asRecoveredTestCase{
		input:         payload,
		name:          name,
		wantConverted: true,
	}
}

func asRecoveredTestCases() []asRecoveredTestCase {
	testErr := errors.New("test error")

	return []asRecoveredTestCase{
		newAsRecoveredTestCaseString("string payload", "test panic"),
		newAsRecoveredTestCase("named string payload", namedString("test panic")),
		newAsRecoveredTestCase("error payload", testErr),
		newAsRecoveredTestCase("int payload", 42),
	}
}

func (tc asRecoveredTestCase) Name() string { return tc.name }

func (tc asRecoveredTestCase) Test(t *testing.T) {
	t.Helper()
	rec := AssertMustTypeIs[Recovered](t, AsRecovered(tc.input), "result")

	got := rec.Recovered()
	if tc.wantConverted {
		err := AssertMustTypeIs[error](t, got, "payload is an error")
		got = err.Error()
	}

	AssertEqual(t, tc.input, got, "payload")
}

func TestAsRecovered(t *testing.T) {
	RunTestCases(t, asRecoveredTestCases())
}

func TestAsRecoveredNil(t *testing.T) {
	AssertNil(t, AsRecovered(nil), "result")
}

// AsRecovered hands back a value that is already Recovered rather than
// wrapping it a second time.
func TestAsRecoveredPassThrough(t *testing.T) {
	var rec Recovered = NewPanicError(0, "wrapped error")

	AssertSame(t, rec, AsRecovered(rec), "result")
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

// mustSuccessTestCase states that Must returns its value untouched and
// raises nothing when the error is nil. The value's type is the row's
// type parameter.
type mustSuccessTestCase[V any] struct {
	value V
	name  string
}

func newMustSuccessTestCase[V any](name string, value V) TestCase {
	return mustSuccessTestCase[V]{
		value: value,
		name:  name,
	}
}

func (tc mustSuccessTestCase[V]) Name() string {
	return tc.name
}

func (tc mustSuccessTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, err := testMust(tc.value, nil)
	AssertNoError(t, err, "no panic")
	AssertEqual(t, tc.value, got, "value")
}

func mustSuccessTestCases() []TestCase {
	return []TestCase{
		newMustSuccessTestCase("string success", "hello"),
		newMustSuccessTestCase("int success", 42),
		newMustSuccessTestCase("bool success", true),
		newMustSuccessTestCase("slice success", S(1, 2, 3)),
		newMustSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}
}

func TestMustSuccess(t *testing.T) {
	RunTestCases(t, mustSuccessTestCases())
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

	AssertErrorIs(t, err, ErrUnreachable, "ErrUnreachable in chain")
	AssertErrorIs(t, err, tc.err, "original error in chain")

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

// maybeTestCase states that Maybe returns its value whatever the error
// is, so the rows differ only in the value's type and whether an error
// was there to be ignored.
type maybeTestCase[V any] struct {
	err   error
	value V
	name  string
}

func newMaybeTestCase[V any](name string, value V, err error) TestCase {
	return maybeTestCase[V]{
		err:   err,
		value: value,
		name:  name,
	}
}

func (tc maybeTestCase[V]) Name() string {
	return tc.name
}

func (tc maybeTestCase[V]) Test(t *testing.T) {
	t.Helper()

	AssertEqual(t, tc.value, Maybe(tc.value, tc.err), "value")
}

func maybeTestCases() []TestCase {
	return []TestCase{
		newMaybeTestCase("string with nil error", "hello", nil),
		newMaybeTestCase("string with error", "world", errors.New("ignored error")),
		newMaybeTestCase("int with nil error", 42, nil),
		newMaybeTestCase("int with error", 0, errors.New("another ignored error")),
		newMaybeTestCase("nil pointer with error", (*int)(nil),
			errors.New("pointer error")),
		newMaybeTestCase("struct with error", struct{ Name string }{"test"},
			fmt.Errorf("formatted: %d", 123)),
	}
}

func TestMaybe(t *testing.T) {
	RunTestCases(t, maybeTestCases())
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

// mustOKSuccessTestCase states that MustOK returns its value untouched
// and raises nothing when ok is true, with the value's type carried by
// the row.
type mustOKSuccessTestCase[V any] struct {
	value V
	name  string
}

func newMustOKSuccessTestCase[V any](name string, value V) TestCase {
	return mustOKSuccessTestCase[V]{
		value: value,
		name:  name,
	}
}

func (tc mustOKSuccessTestCase[V]) Name() string {
	return tc.name
}

func (tc mustOKSuccessTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, err := testMustOK(tc.value, true)
	AssertNoError(t, err, "no panic")
	AssertEqual(t, tc.value, got, "value")
}

func mustOKSuccessTestCases() []TestCase {
	return []TestCase{
		newMustOKSuccessTestCase("string success", "hello"),
		newMustOKSuccessTestCase("int success", 42),
		newMustOKSuccessTestCase("bool success", true),
		newMustOKSuccessTestCase("slice success", S(1, 2, 3)),
		newMustOKSuccessTestCase("nil pointer success", (*int)(nil)),
		newMustOKSuccessTestCase("struct success", struct{ Name string }{"test"}),
	}
}

func TestMustOKSuccess(t *testing.T) {
	RunTestCases(t, mustOKSuccessTestCases())
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

		AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
		AssertContains(t, panicErr.Error(), "operation failed", "reason")
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

// maybeOKTestCase states that MaybeOK returns its value whatever ok
// says, so the rows differ only in the value's type and the flag being
// ignored.
type maybeOKTestCase[V any] struct {
	value V
	name  string
	ok    bool
}

func newMaybeOKTestCase[V any](name string, value V, ok bool) TestCase {
	return maybeOKTestCase[V]{
		value: value,
		name:  name,
		ok:    ok,
	}
}

func (tc maybeOKTestCase[V]) Name() string {
	return tc.name
}

func (tc maybeOKTestCase[V]) Test(t *testing.T) {
	t.Helper()

	AssertEqual(t, tc.value, MaybeOK(tc.value, tc.ok), "value")
}

func maybeOKTestCases() []TestCase {
	return []TestCase{
		newMaybeOKTestCase("string with true", "hello", true),
		newMaybeOKTestCase("string with false", "world", false),
		newMaybeOKTestCase("int with true", 42, true),
		newMaybeOKTestCase("int with false", 0, false),
		newMaybeOKTestCase("nil pointer with false", (*int)(nil), false),
		newMaybeOKTestCase("struct with false", struct{ Name string }{"test"},
			false),
	}
}

func TestMaybeOK(t *testing.T) {
	RunTestCases(t, maybeOKTestCases())
}

// mustTSuccessTestCase states that MustT hands back the input as the
// target type when it already holds one. The target is the row's type
// parameter.
type mustTSuccessTestCase[T any] struct {
	input any
	want  T
	name  string
}

func newMustTSuccessTestCase[T any](name string, input any, want T) TestCase {
	return mustTSuccessTestCase[T]{
		input: input,
		want:  want,
		name:  name,
	}
}

func (tc mustTSuccessTestCase[T]) Name() string {
	return tc.name
}

func (tc mustTSuccessTestCase[T]) Test(t *testing.T) {
	t.Helper()

	AssertEqual(t, tc.want, MustT[T](tc.input), "value")
}

func mustTSuccessTestCases() []TestCase {
	testErr := errors.New("test")
	stringer := mockStringer{value: "test"}

	return []TestCase{
		newMustTSuccessTestCase[string]("string to string", "hello", "hello"),
		newMustTSuccessTestCase[int]("int to int", 42, 42),
		newMustTSuccessTestCase[float64]("float64 to float64", 3.14, 3.14),
		newMustTSuccessTestCase[error]("error to error", testErr, testErr),
		newMustTSuccessTestCase[fmt.Stringer]("stringer to fmt.Stringer",
			stringer, stringer),
	}
}

func TestMustTSuccess(t *testing.T) {
	RunTestCases(t, mustTSuccessTestCases())
}

// mustTPanicTestCase states that MustT panics with ErrUnreachable when
// the input does not hold the target type. The target is the row's type
// parameter.
type mustTPanicTestCase[T any] struct {
	input any
	name  string
}

func newMustTPanicTestCase[T any](name string, input any) TestCase {
	return mustTPanicTestCase[T]{
		input: input,
		name:  name,
	}
}

func (tc mustTPanicTestCase[T]) Name() string {
	return tc.name
}

func (tc mustTPanicTestCase[T]) Test(t *testing.T) {
	t.Helper()

	AssertPanic(t, func() {
		_ = MustT[T](tc.input)
	}, ErrUnreachable, "panic")
}

func mustTPanicTestCases() []TestCase {
	return []TestCase{
		newMustTPanicTestCase[string]("int to string", 42),
		newMustTPanicTestCase[int]("string to int", "hello"),
		newMustTPanicTestCase[error]("string to error", "not an error"),
		newMustTPanicTestCase[fmt.Stringer]("int to fmt.Stringer", 42),
		newMustTPanicTestCase[string]("nil to string", nil),
	}
}

func TestMustTPanic(t *testing.T) {
	RunTestCases(t, mustTPanicTestCases())
}

// mustTReasonTestCase states the reason MustT panics with for one
// target type: it names the value's type and the target's, the target
// included when it is an interface and the zero result has no dynamic
// type to print.
type mustTReasonTestCase[T any] struct {
	input any
	want  string
	name  string
}

func newMustTReasonTestCase[T any](name string, input any, want string) TestCase {
	return mustTReasonTestCase[T]{
		input: input,
		want:  want,
		name:  name,
	}
}

func (tc mustTReasonTestCase[T]) Name() string {
	return tc.name
}

func (tc mustTReasonTestCase[T]) Test(t *testing.T) {
	t.Helper()
	AssertPanic(t, func() {
		_ = MustT[T](tc.input)
	}, tc.want, "reason")
}

func TestMustTReason(t *testing.T) {
	testCases := S(
		newMustTReasonTestCase[int]("concrete target", testHello,
			"failed to convert string to int"),
		newMustTReasonTestCase[fmt.Stringer]("interface target", 42,
			"failed to convert int to fmt.Stringer"),
	)

	RunTestCases(t, testCases)
}

// maybeTTestCase states what MaybeT returns for one target type: the
// input when it already holds a T, the zero value when it does not, and
// never a panic. The target is the row's type parameter.
type maybeTTestCase[T any] struct {
	input any
	want  T
	name  string
}

func newMaybeTTestCase[T any](name string, input any, want T) TestCase {
	return maybeTTestCase[T]{
		input: input,
		want:  want,
		name:  name,
	}
}

func (tc maybeTTestCase[T]) Name() string {
	return tc.name
}

func (tc maybeTTestCase[T]) Test(t *testing.T) {
	t.Helper()

	AssertEqual(t, tc.want, MaybeT[T](tc.input), "value")
}

func maybeTTestCases() []TestCase {
	testErr := errors.New("test")
	stringer := mockStringer{value: "test"}

	return []TestCase{
		// the input already holds the target type
		newMaybeTTestCase[string]("string to string", "hello", "hello"),
		newMaybeTTestCase[int]("int to int", 42, 42),
		newMaybeTTestCase[error]("error to error", testErr, testErr),
		newMaybeTTestCase[fmt.Stringer]("stringer to fmt.Stringer", stringer,
			stringer),

		// it does not, so the zero value comes back
		newMaybeTTestCase[string]("int to string", 42, ""),
		newMaybeTTestCase[int]("string to int", "hello", 0),
		newMaybeTTestCase[error]("string to error", "not an error", nil),
		newMaybeTTestCase[fmt.Stringer]("int to fmt.Stringer", 42, nil),
		newMaybeTTestCase[string]("nil to string", nil, ""),
		newMaybeTTestCase[int]("nil to int", nil, 0),
	}
}

func TestMaybeT(t *testing.T) {
	RunTestCases(t, maybeTTestCases())
}

// Helper type for testing fmt.Stringer interface
type mockStringer struct {
	value string
}

func (ms mockStringer) String() string {
	return ms.value
}
