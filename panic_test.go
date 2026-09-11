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
var _ TestCase = mustTestCase[int]{}
var _ TestCase = maybeTestCase[int]{}
var _ TestCase = mustOKTestCase[int]{}
var _ TestCase = maybeOKTestCase[int]{}
var _ TestCase = mustTTestCase[int]{}
var _ TestCase = maybeTTestCase[int]{}

// asRecoveredTestCase tests the payload AsRecovered's result reports
// for a recovered value.
type asRecoveredTestCase struct {
	input any
	want  any
	name  string
}

func newAsRecoveredTestCase(name string, input, want any) asRecoveredTestCase {
	return asRecoveredTestCase{
		input: input,
		want:  want,
		name:  name,
	}
}

func (tc asRecoveredTestCase) Name() string {
	return tc.name
}

func (tc asRecoveredTestCase) Test(t *testing.T) {
	t.Helper()

	recovered := AsRecovered(tc.input)
	AssertMustNotNil(t, recovered, "recovered")
	assertPayload(t, tc.want, recovered)
}

// assertPayload states what a Recovered error reports for its payload,
// by the kind of expectation: an error is found through the chain, a
// string is a substring of the message, anything else is the payload
// itself.
func assertPayload(t *testing.T, want any, recovered Recovered) {
	t.Helper()

	switch want := want.(type) {
	case error:
		AssertErrorIs(t, recovered, want, "payload")
	case string:
		AssertContains(t, recovered.Error(), want, "message")
	default:
		AssertEqual(t, want, recovered.Recovered(), "payload")
	}
}

func asRecoveredTestCases() []asRecoveredTestCase {
	panicErr := NewPanicError(1, errSentinel)

	return S(
		newAsRecoveredTestCase("int", 42, 42),
		newAsRecoveredTestCase("string", testHello, testHello),
		newAsRecoveredTestCase("error", errSentinel, errSentinel),
		newAsRecoveredTestCase("recovered", panicErr, errSentinel),
	)
}

func TestAsRecovered(t *testing.T) {
	RunTestCases(t, asRecoveredTestCases())
	t.Run("nil", testAsRecoveredNil)
	t.Run("pass-through", testAsRecoveredPassThrough)
}

func testAsRecoveredNil(t *testing.T) {
	t.Helper()
	AssertNil(t, AsRecovered(nil), "recovered")
}

func testAsRecoveredPassThrough(t *testing.T) {
	t.Helper()

	panicErr := NewPanicError(1, errSentinel)
	AssertSame(t, panicErr, AsRecovered(panicErr), "recovered")
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
	AssertNil(t, recovered, "initially nil recovered")

	// After panic
	_ = catcher.Try(func() error {
		panic("test panic")
	})

	recovered = catcher.Recovered()
	AssertMustNotNil(t, recovered, "recovered panic after Try")

	// A string payload comes back as an error carrying it
	err := AssertMustTypeIs[error](t, recovered.Recovered(), "payload")
	AssertEqual(t, "test panic", err.Error(), "message")
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
	AssertMustNotNil(t, recovered, "concurrent recovered panic")

	// Whichever panic was stored, it was one of the two
	err := AssertMustTypeIs[error](t, recovered.Recovered(), "payload")
	msg := err.Error()
	AssertTrue(t, msg == "first panic" || msg == "second panic", "first or second panic")
}

func TestCatcherFirstPanicWins(t *testing.T) {
	var catcher Catcher

	_ = catcher.Try(func() error {
		panic("first panic")
	})
	first := catcher.Recovered()
	AssertMustNotNil(t, first, "first recovered")

	_ = catcher.Try(func() error {
		panic("second panic")
	})
	AssertSame(t, first, catcher.Recovered(), "recovered after the second panic")
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

// catchWithPanicRecoveryTestCase tests the payload Catch's error
// reports for a panic value.
type catchWithPanicRecoveryTestCase struct {
	value any
	want  any
	name  string
}

func catchWithPanicRecoveryTestCases() []catchWithPanicRecoveryTestCase {
	panicErr := NewPanicError(1, errSentinel)

	return S(
		newCatchWithPanicRecoveryTestCase("int", 42, 42),
		newCatchWithPanicRecoveryTestCase("float", 3.14, 3.14),
		newCatchWithPanicRecoveryTestCase("string", testHello, testHello),
		newCatchWithPanicRecoveryTestCase("error", errSentinel, errSentinel),
		newCatchWithPanicRecoveryTestCase("recovered", panicErr, errSentinel),
	)
}

func newCatchWithPanicRecoveryTestCase(name string, value, want any) catchWithPanicRecoveryTestCase {
	return catchWithPanicRecoveryTestCase{
		value: value,
		want:  want,
		name:  name,
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

	recovered := AssertMustTypeIs[Recovered](t, err, "Recovered")
	assertPayload(t, tc.want, recovered)
}

func TestCatchWithPanicRecovery(t *testing.T) {
	RunTestCases(t, catchWithPanicRecoveryTestCases())
}

// callMust calls Must and returns what came back: the value, or the
// recovered panic as an error.
func callMust[V any](value V, err error) (got V, recovered error) {
	defer func() {
		if e := AsRecovered(recover()); e != nil {
			recovered = e
		}
	}()

	got = Must(value, err)
	return got, nil
}

// mustTestCase tests Must over one value type: a nil error returns the
// value, an error panics.
type mustTestCase[V any] struct {
	value V
	err   error
	name  string
}

func newMustTestCase[V any](name string, value V, err error) mustTestCase[V] {
	return mustTestCase[V]{
		value: value,
		err:   err,
		name:  name,
	}
}

func (tc mustTestCase[V]) Name() string {
	return tc.name
}

func (tc mustTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, recovered := callMust(tc.value, tc.err)
	if tc.err == nil {
		AssertNoError(t, recovered, "Must")
		AssertEqual(t, tc.value, got, "value")
		return
	}

	panicErr := AssertMustTypeIs[*PanicError](t, recovered, "panic")
	AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
	AssertErrorIs(t, panicErr, tc.err, "error in chain")
	AssertTrue(t, len(panicErr.CallStack()) > 0, "stack")
}

func mustTestCases[V any](value V) []mustTestCase[V] {
	return S(
		newMustTestCase("nil error", value, nil),
		newMustTestCase("error", value, errSentinel),
	)
}

func runMustTestCases[V any](t *testing.T, name string, value V) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		RunTestCases(t, mustTestCases(value))
	})
}

func TestMust(t *testing.T) {
	runMustTestCases(t, "string", testHello)
	runMustTestCases(t, "int", 42)
	runMustTestCases(t, "bool", true)
	runMustTestCases(t, "slice", S(1, 2, 3))
	runMustTestCases(t, "nil pointer", (*int)(nil))
	runMustTestCases(t, "struct", testStruct{Value: testHello, Count: 1})
}

// maybeTestCase tests Maybe over one value type: the value comes back
// whatever the error.
type maybeTestCase[V any] struct {
	value V
	err   error
	name  string
}

func newMaybeTestCase[V any](name string, value V, err error) maybeTestCase[V] {
	return maybeTestCase[V]{
		value: value,
		err:   err,
		name:  name,
	}
}

func (tc maybeTestCase[V]) Name() string {
	return tc.name
}

func (tc maybeTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got := Maybe(tc.value, tc.err)
	AssertEqual(t, tc.value, got, "value")
}

func maybeTestCases[V any](value V) []maybeTestCase[V] {
	return S(
		newMaybeTestCase("nil error", value, nil),
		newMaybeTestCase("error", value, errSentinel),
	)
}

func runMaybeTestCases[V any](t *testing.T, name string, value V) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		RunTestCases(t, maybeTestCases(value))
	})
}

func TestMaybe(t *testing.T) {
	runMaybeTestCases(t, "string", testHello)
	runMaybeTestCases(t, "int", 42)
	runMaybeTestCases(t, "nil pointer", (*int)(nil))
	runMaybeTestCases(t, "struct", testStruct{Value: testHello, Count: 1})
}

// callMustOK calls MustOK and returns what came back: the value, or the
// recovered panic as an error.
func callMustOK[V any](value V, ok bool) (got V, recovered error) {
	defer func() {
		if e := AsRecovered(recover()); e != nil {
			recovered = e
		}
	}()

	got = MustOK(value, ok)
	return got, nil
}

// mustOKTestCase tests MustOK over one value type: ok returns the value,
// not ok panics.
type mustOKTestCase[V any] struct {
	value V
	name  string
	ok    bool
}

func newMustOKTestCase[V any](name string, value V, ok bool) mustOKTestCase[V] {
	return mustOKTestCase[V]{
		value: value,
		name:  name,
		ok:    ok,
	}
}

func (tc mustOKTestCase[V]) Name() string {
	return tc.name
}

func (tc mustOKTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, recovered := callMustOK(tc.value, tc.ok)
	if tc.ok {
		AssertNoError(t, recovered, "MustOK")
		AssertEqual(t, tc.value, got, "value")
		return
	}

	panicErr := AssertMustTypeIs[*PanicError](t, recovered, "panic")
	AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
	AssertContains(t, panicErr.Error(), "operation failed", "reason")
	AssertTrue(t, len(panicErr.CallStack()) > 0, "stack")
}

func mustOKTestCases[V any](value V) []mustOKTestCase[V] {
	return S(
		newMustOKTestCase("ok", value, true),
		newMustOKTestCase("not ok", value, false),
	)
}

func runMustOKTestCases[V any](t *testing.T, name string, value V) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		RunTestCases(t, mustOKTestCases(value))
	})
}

func TestMustOK(t *testing.T) {
	runMustOKTestCases(t, "string", testHello)
	runMustOKTestCases(t, "int", 42)
	runMustOKTestCases(t, "bool", true)
	runMustOKTestCases(t, "slice", S(1, 2, 3))
	runMustOKTestCases(t, "nil pointer", (*int)(nil))
	runMustOKTestCases(t, "struct", testStruct{Value: testHello, Count: 1})
}

// maybeOKTestCase tests MaybeOK over one value type: the value comes
// back whatever the flag.
type maybeOKTestCase[V any] struct {
	value V
	name  string
	ok    bool
}

func newMaybeOKTestCase[V any](name string, value V, ok bool) maybeOKTestCase[V] {
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

	got := MaybeOK(tc.value, tc.ok)
	AssertEqual(t, tc.value, got, "value")
}

func maybeOKTestCases[V any](value V) []maybeOKTestCase[V] {
	return S(
		newMaybeOKTestCase("ok", value, true),
		newMaybeOKTestCase("not ok", value, false),
	)
}

func runMaybeOKTestCases[V any](t *testing.T, name string, value V) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		RunTestCases(t, maybeOKTestCases(value))
	})
}

func TestMaybeOK(t *testing.T) {
	runMaybeOKTestCases(t, "string", testHello)
	runMaybeOKTestCases(t, "int", 42)
	runMaybeOKTestCases(t, "nil pointer", (*int)(nil))
	runMaybeOKTestCases(t, "struct", testStruct{Value: testHello, Count: 1})
}

// callMustT calls MustT and returns what came back: the value, or the
// recovered panic as an error.
func callMustT[T any](value any) (got T, recovered error) {
	defer func() {
		if e := AsRecovered(recover()); e != nil {
			recovered = e
		}
	}()

	got = MustT[T](value)
	return got, nil
}

// mustTTestCase tests MustT over one target type: an input of that type
// comes back as it, anything else panics with a reason naming both
// types, the target included when it is an interface and the zero
// result has no dynamic type to print.
type mustTTestCase[T any] struct {
	input  any
	want   T
	reason string
	name   string
	panics bool
}

func newMustTTestCase[T any](name string, input any, want T) mustTTestCase[T] {
	return mustTTestCase[T]{
		input: input,
		want:  want,
		name:  name,
	}
}

func newMustTTestCasePanic[T any](name string, input any, reason string) mustTTestCase[T] {
	return mustTTestCase[T]{
		input:  input,
		reason: reason,
		name:   name,
		panics: true,
	}
}

func (tc mustTTestCase[T]) Name() string {
	return tc.name
}

func (tc mustTTestCase[T]) Test(t *testing.T) {
	t.Helper()

	got, recovered := callMustT[T](tc.input)
	if !tc.panics {
		AssertNoError(t, recovered, "MustT")
		AssertEqual(t, tc.want, got, "value")
		return
	}

	panicErr := AssertMustTypeIs[*PanicError](t, recovered, "panic")
	AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
	AssertContains(t, panicErr.Error(), tc.reason, "reason")
	AssertTrue(t, len(panicErr.CallStack()) > 0, "stack")
}

func mustTStringTestCases() []mustTTestCase[string] {
	return S(
		newMustTTestCase("string", testHello, testHello),
		newMustTTestCase("empty string", "", ""),
		newMustTTestCasePanic[string]("int", 42,
			"failed to convert int to string"),
		newMustTTestCasePanic[string]("nil", nil,
			"failed to convert <nil> to string"),
	)
}

func mustTIntTestCases() []mustTTestCase[int] {
	return S(
		newMustTTestCase("int", 42, 42),
		newMustTTestCase("zero", 0, 0),
		newMustTTestCasePanic[int]("string", testHello,
			"failed to convert string to int"),
		newMustTTestCasePanic[int]("nil", nil,
			"failed to convert <nil> to int"),
	)
}

func mustTErrorTestCases() []mustTTestCase[error] {
	return S(
		newMustTTestCase("error", errSentinel, errSentinel),
		newMustTTestCasePanic[error]("string", "not an error",
			"failed to convert string to error"),
		newMustTTestCasePanic[error]("nil", nil,
			"failed to convert <nil> to error"),
	)
}

func mustTStringerTestCases() []mustTTestCase[fmt.Stringer] {
	stringer := mockStringer{value: testHello}

	return S(
		newMustTTestCase[fmt.Stringer]("stringer", stringer, stringer),
		newMustTTestCasePanic[fmt.Stringer]("int", 42,
			"failed to convert int to fmt.Stringer"),
		newMustTTestCasePanic[fmt.Stringer]("nil", nil,
			"failed to convert <nil> to fmt.Stringer"),
	)
}

func TestMustT(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		RunTestCases(t, mustTStringTestCases())
	})
	t.Run("int", func(t *testing.T) {
		RunTestCases(t, mustTIntTestCases())
	})
	t.Run("error", func(t *testing.T) {
		RunTestCases(t, mustTErrorTestCases())
	})
	t.Run("fmt.Stringer", func(t *testing.T) {
		RunTestCases(t, mustTStringerTestCases())
	})
}

// maybeTTestCase tests MaybeT over one target type: an input of that
// type comes back as it, anything else as the zero value.
type maybeTTestCase[T any] struct {
	input any
	want  T
	name  string
}

func newMaybeTTestCase[T any](name string, input any, want T) maybeTTestCase[T] {
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

	got := MaybeT[T](tc.input)
	AssertEqual(t, tc.want, got, "value")
}

func maybeTStringTestCases() []maybeTTestCase[string] {
	return S(
		newMaybeTTestCase("string", testHello, testHello),
		newMaybeTTestCase("empty string", "", ""),
		newMaybeTTestCase("int", 42, ""),
		newMaybeTTestCase("nil", nil, ""),
	)
}

func maybeTIntTestCases() []maybeTTestCase[int] {
	return S(
		newMaybeTTestCase("int", 42, 42),
		newMaybeTTestCase("zero", 0, 0),
		newMaybeTTestCase("string", testHello, 0),
		newMaybeTTestCase("nil", nil, 0),
	)
}

func maybeTErrorTestCases() []maybeTTestCase[error] {
	return S(
		newMaybeTTestCase("error", errSentinel, errSentinel),
		newMaybeTTestCase[error]("string", "not an error", nil),
		newMaybeTTestCase[error]("nil", nil, nil),
	)
}

func maybeTStringerTestCases() []maybeTTestCase[fmt.Stringer] {
	stringer := mockStringer{value: testHello}

	return S(
		newMaybeTTestCase[fmt.Stringer]("stringer", stringer, stringer),
		newMaybeTTestCase[fmt.Stringer]("int", 42, nil),
		newMaybeTTestCase[fmt.Stringer]("nil", nil, nil),
	)
}

func TestMaybeT(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		RunTestCases(t, maybeTStringTestCases())
	})
	t.Run("int", func(t *testing.T) {
		RunTestCases(t, maybeTIntTestCases())
	})
	t.Run("error", func(t *testing.T) {
		RunTestCases(t, maybeTErrorTestCases())
	})
	t.Run("fmt.Stringer", func(t *testing.T) {
		RunTestCases(t, maybeTStringerTestCases())
	})
}

// Helper type for testing fmt.Stringer interface
type mockStringer struct {
	value string
}

func (ms mockStringer) String() string {
	return ms.value
}
