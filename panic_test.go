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
	_ TestCase = catcherAbortTestCase{}
	_ TestCase = catchTestCase{}
	_ TestCase = catchWithPanicRecoveryTestCase{}
	_ TestCase = mustTestCase[int]{}
	_ TestCase = maybeTestCase[int]{}
	_ TestCase = mustOKTestCase[int]{}
	_ TestCase = maybeOKTestCase[int]{}
	_ TestCase = mustTTestCase[int]{}
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
	AssertNil(t, recovered, "initially nil recovered")

	// After panic
	_ = catcher.Try(func() error {
		panic("test panic")
	})

	recovered = catcher.Recovered()
	AssertMustNotNil(t, recovered, "recovered panic after Try")

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

// callerRecovered is a caller's own type implementing Recovered, which
// AsRecovered passes through as it does a *PanicError.
type callerRecovered struct {
	payload any
}

func (r callerRecovered) Error() string  { return fmt.Sprint(r.payload) }
func (r callerRecovered) Recovered() any { return r.payload }

// The first panic caught survives: a later one, whatever type of
// Recovered it arrives as, does not replace it.
func TestCatcherFirstPanicWins(t *testing.T) {
	var catcher Catcher

	first := NewPanicError(0, "first")
	_ = catcher.Try(func() error { panic(first) })
	AssertSame(t, first, catcher.Recovered(), "first panic")

	later := callerRecovered{payload: "later"}
	AssertNoPanic(t, func() {
		_ = catcher.Try(func() error { panic(later) })
	}, "later panic")
	AssertSame(t, first, catcher.Recovered(), "first panic kept")
}

// catcherAbortTestCase states what Catcher.Try does with what fn does
// to the test through MockT: an abort is not a panic, so it passes
// through to MockT.Run and cuts the body short after Try, and the
// Catcher records nothing either way.
type catcherAbortTestCase struct {
	fn         func(T)
	name       string
	wantErrors int
	wantLogs   int

	wantAborted bool
	wantFailed  bool
	wantSkipped bool
}

func (tc catcherAbortTestCase) Name() string {
	return tc.name
}

func (tc catcherAbortTestCase) Test(t *testing.T) {
	t.Helper()
	var catcher Catcher
	var continued bool
	mock := &MockT{}
	ok := mock.Run(tc.name, func(mt T) {
		_ = catcher.Try(func() error {
			tc.fn(mt)
			return nil
		})
		continued = true
	})

	AssertEqual(t, !tc.wantFailed, ok, "passed")
	AssertEqual(t, tc.wantFailed, mock.Failed(), "failed")
	AssertEqual(t, tc.wantSkipped, mock.Skipped(), "skipped")
	AssertEqual(t, !tc.wantAborted, continued, "continued")
	AssertEqual(t, tc.wantErrors, mock.NumErrors(), "errors")
	AssertEqual(t, tc.wantLogs, mock.NumLogs(), "logs")
	AssertNil(t, catcher.Recovered(), "recovered")
}

// newCatcherAbortTestCase is a row whose abort fails the test and cuts
// it short, recording wantErrors on the way.
func newCatcherAbortTestCase(name string, abort func(T),
	wantErrors int) catcherAbortTestCase {
	return catcherAbortTestCase{
		fn:          abort,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    0,
		wantAborted: true,
		wantFailed:  true,
		wantSkipped: false,
	}
}

// newCatcherAbortTestCaseContinues is a row whose function fails the
// test without cutting it short, recording wantErrors on the way.
func newCatcherAbortTestCaseContinues(name string, fn func(T),
	wantErrors int) catcherAbortTestCase {
	return catcherAbortTestCase{
		fn:          fn,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    0,
		wantAborted: false,
		wantFailed:  true,
		wantSkipped: false,
	}
}

// newCatcherAbortTestCaseSkips is a row whose abort skips the test,
// recording wantLogs on the way.
func newCatcherAbortTestCaseSkips(name string, abort func(T),
	wantLogs int) catcherAbortTestCase {
	return catcherAbortTestCase{
		fn:          abort,
		name:        name,
		wantErrors:  0,
		wantLogs:    wantLogs,
		wantAborted: true,
		wantFailed:  false,
		wantSkipped: true,
	}
}

// newCatcherAbortTestCaseFailsSkips is a row whose abort fails the test
// and then skips it, recording wantErrors and wantLogs on the way.
func newCatcherAbortTestCaseFailsSkips(name string, abort func(T),
	wantErrors, wantLogs int) catcherAbortTestCase {
	return catcherAbortTestCase{
		fn:          abort,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    wantLogs,
		wantAborted: true,
		wantFailed:  true,
		wantSkipped: true,
	}
}

var catcherAbortTestCases = S(
	newCatcherAbortTestCaseContinues("Fail", func(mt T) { mt.Fail() }, 0),
	newCatcherAbortTestCase("FailNow", func(mt T) { mt.FailNow() }, 0),
	newCatcherAbortTestCase("Fatal", func(mt T) { mt.Fatal("fatal") }, 1),
	newCatcherAbortTestCaseSkips("SkipNow", func(mt T) { mt.SkipNow() }, 0),
	newCatcherAbortTestCaseSkips("Skip", func(mt T) { mt.Skip("skip") }, 1),
	newCatcherAbortTestCaseFailsSkips("Fail then SkipNow",
		func(mt T) { mt.Fail(); mt.SkipNow() }, 0, 0),
)

func TestCatcherAbort(t *testing.T) {
	RunTestCases(t, catcherAbortTestCases)
}

// Catcher.Do reaches fn through Try, so an abort inside it passes on to
// MockT.Run as well, and the Catcher records nothing.
func TestCatcherDoAbort(t *testing.T) {
	var catcher Catcher
	var continued bool
	mock := &MockT{}
	ok := mock.Run("Do", func(mt T) {
		_ = catcher.Do(func() error {
			mt.FailNow()
			return nil
		})
		continued = true
	})

	AssertFalse(t, ok, "passed")
	AssertTrue(t, mock.Failed(), "failed")
	AssertFalse(t, continued, "continued")
	AssertEqual(t, 0, mock.NumErrors(), "errors")
	AssertEqual(t, 0, mock.NumLogs(), "logs")
	AssertNil(t, catcher.Recovered(), "recovered")
}

// Catch reaches fn through a Catcher, so an abort inside it passes on to
// MockT.Run as well.
func TestCatchAbort(t *testing.T) {
	var continued bool
	mock := &MockT{}
	ok := mock.Run("Catch", func(mt T) {
		_ = Catch(func() error {
			mt.FailNow()
			return nil
		})
		continued = true
	})

	AssertFalse(t, ok, "passed")
	AssertTrue(t, mock.Failed(), "failed")
	AssertFalse(t, continued, "continued")
	AssertEqual(t, 0, mock.NumErrors(), "errors")
	AssertEqual(t, 0, mock.NumLogs(), "logs")
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

var _ TestCase = catchStackTestCase{}

// catchStackTestCase states where the stack of a panic Catch recovers
// starts: at the function that panicked, past the runtime's frames.
type catchStackTestCase struct {
	fn       func() error
	name     string
	wantFunc string
}

func newCatchStackTestCase(name string, fn func() error,
	wantFunc string) catchStackTestCase {
	return catchStackTestCase{
		fn:       fn,
		name:     name,
		wantFunc: wantFunc,
	}
}

func (tc catchStackTestCase) Name() string {
	return tc.name
}

func (tc catchStackTestCase) Test(t *testing.T) {
	t.Helper()
	_ = assertTopFrameIs(t, Catch(tc.fn), tc.wantFunc, 2)
}

// panicPlain panics with a plain value, not a PanicError.
func panicPlain() error {
	panic(errSentinel)
}

// panicMapAssign writes to m, which a nil m turns into a panic the
// runtime raises.
func panicMapAssign(m map[string]int) error {
	m["key"] = 1
	return nil
}

func catchStackTestCases() []catchStackTestCase {
	return S(
		newCatchStackTestCase("plain panic", panicPlain, "panicPlain"),
		newCatchStackTestCase("runtime panic",
			func() error { return panicMapAssign(nil) }, "panicMapAssign"),
	)
}

func TestCatchStack(t *testing.T) {
	RunTestCases(t, catchStackTestCases())
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

// mustTestCase states what Must does with a value and an error: a nil
// error returns the value untouched, and any other is raised as a panic
// behind ErrUnreachable. The value's type is the row's type parameter.
type mustTestCase[V any] struct {
	err   error
	value V
	name  string

	wantPanic bool
}

func newMustTestCase[V any](name string, value V) TestCase {
	return mustTestCase[V]{
		err:       nil,
		value:     value,
		name:      name,
		wantPanic: false,
	}
}

// newMustTestCasePanic declares a row whose error makes Must panic, with
// that error in the chain of what it raises.
func newMustTestCasePanic[V any](name string, value V, err error) TestCase {
	if err == nil {
		panic("mustTestCase: a panic row states the error it panics with")
	}

	return mustTestCase[V]{
		err:       err,
		value:     value,
		name:      name,
		wantPanic: true,
	}
}

func (tc mustTestCase[V]) Name() string {
	return tc.name
}

func (tc mustTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, recovered := callMust(tc.value, tc.err)
	if !tc.wantPanic {
		AssertNoError(t, recovered, "no panic")
		AssertEqual(t, tc.value, got, "value")
		return
	}

	panicErr := AssertMustTypeIs[*PanicError](t, recovered, "panic")
	AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
	AssertErrorIs(t, panicErr, tc.err, "error in chain")
	AssertTrue(t, len(panicErr.CallStack()) > 0, "stack captured")
}

func mustTestCases() []TestCase {
	return []TestCase{
		newMustTestCase("string", testHello),
		newMustTestCase("int", 42),
		newMustTestCase("bool", true),
		newMustTestCase("slice", S(1, 2, 3)),
		newMustTestCase("nil pointer", (*int)(nil)),
		newMustTestCase("struct", struct{ Name string }{"test"}),

		newMustTestCasePanic("string with error", testHello, errSentinel),
		newMustTestCasePanic("int with error", 42, errSentinel),
		newMustTestCasePanic("bool with error", true, errSentinel),
		newMustTestCasePanic("slice with error", S(1, 2, 3), errSentinel),
		newMustTestCasePanic("nil pointer with error", (*int)(nil), errSentinel),
		newMustTestCasePanic("struct with error", struct{ Name string }{"test"},
			errSentinel),
	}
}

func TestMust(t *testing.T) {
	RunTestCases(t, mustTestCases())
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

// mustOKTestCase states what MustOK does with a value and a flag: ok
// returns the value untouched, and not ok is raised as a panic behind
// ErrUnreachable. The value's type is the row's type parameter.
type mustOKTestCase[V any] struct {
	value V
	name  string
	ok    bool

	wantPanic bool
}

func newMustOKTestCase[V any](name string, value V) TestCase {
	return mustOKTestCase[V]{
		value:     value,
		name:      name,
		ok:        true,
		wantPanic: false,
	}
}

// newMustOKTestCasePanic declares a row whose flag makes MustOK panic.
func newMustOKTestCasePanic[V any](name string, value V) TestCase {
	return mustOKTestCase[V]{
		value:     value,
		name:      name,
		ok:        false,
		wantPanic: true,
	}
}

func (tc mustOKTestCase[V]) Name() string {
	return tc.name
}

func (tc mustOKTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, recovered := callMustOK(tc.value, tc.ok)
	if !tc.wantPanic {
		AssertNoError(t, recovered, "no panic")
		AssertEqual(t, tc.value, got, "value")
		return
	}

	panicErr := AssertMustTypeIs[*PanicError](t, recovered, "panic")
	AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
	AssertContains(t, panicErr.Error(), "operation failed", "reason")
	AssertTrue(t, len(panicErr.CallStack()) > 0, "stack captured")
}

func mustOKTestCases() []TestCase {
	return []TestCase{
		newMustOKTestCase("string", testHello),
		newMustOKTestCase("int", 42),
		newMustOKTestCase("bool", true),
		newMustOKTestCase("slice", S(1, 2, 3)),
		newMustOKTestCase("nil pointer", (*int)(nil)),
		newMustOKTestCase("struct", struct{ Name string }{"test"}),

		newMustOKTestCasePanic("string not ok", testHello),
		newMustOKTestCasePanic("int not ok", 42),
		newMustOKTestCasePanic("bool not ok", false),
		newMustOKTestCasePanic("slice not ok", S(1, 2, 3)),
		newMustOKTestCasePanic("nil pointer not ok", (*int)(nil)),
		newMustOKTestCasePanic("struct not ok", struct{ Name string }{"test"}),
	}
}

func TestMustOK(t *testing.T) {
	RunTestCases(t, mustOKTestCases())
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

// callMustT calls MustT and returns what came back: the value, or the
// recovered panic as an error.
func callMustT[T any](input any) (got T, recovered error) {
	defer func() {
		if e := AsRecovered(recover()); e != nil {
			recovered = e
		}
	}()

	got = MustT[T](input)
	return got, nil
}

// mustTTestCase states what MustT does for one target type: an input
// already holding the target type comes back as that type, and any other
// is raised as a panic behind ErrUnreachable, for a reason naming both
// types, the target named even when it is an interface and the zero
// result has no dynamic type to print. The target is the row's type
// parameter.
type mustTTestCase[T any] struct {
	input  any
	want   T
	name   string
	reason string

	wantPanic bool
}

func newMustTTestCase[T any](name string, input any, want T) TestCase {
	return mustTTestCase[T]{
		input:     input,
		want:      want,
		name:      name,
		reason:    "",
		wantPanic: false,
	}
}

// newMustTTestCasePanic declares a row whose input makes MustT panic,
// and the reason the panic gives.
func newMustTTestCasePanic[T any](name string, input any, reason string) TestCase {
	if reason == "" {
		panic("mustTTestCase: a panic row states the reason it panics with")
	}

	var zero T

	return mustTTestCase[T]{
		input:     input,
		want:      zero,
		name:      name,
		reason:    reason,
		wantPanic: true,
	}
}

func (tc mustTTestCase[T]) Name() string {
	return tc.name
}

func (tc mustTTestCase[T]) Test(t *testing.T) {
	t.Helper()

	got, recovered := callMustT[T](tc.input)
	if !tc.wantPanic {
		AssertNoError(t, recovered, "no panic")
		AssertEqual(t, tc.want, got, "value")
		return
	}

	panicErr := AssertMustTypeIs[*PanicError](t, recovered, "panic")
	AssertErrorIs(t, panicErr, ErrUnreachable, "ErrUnreachable in chain")
	AssertContains(t, panicErr.Error(), tc.reason, "reason")
	AssertTrue(t, len(panicErr.CallStack()) > 0, "stack captured")
}

// mustTValueTestCases are the rows where the input holds the target
// type. The zero value is a result there, not the mark of a failure.
func mustTValueTestCases() []TestCase {
	stringer := mockStringer{value: "test"}

	return []TestCase{
		newMustTTestCase[string]("string to string", testHello, testHello),
		newMustTTestCase[string]("empty string to string", "", ""),
		newMustTTestCase[int]("int to int", 42, 42),
		newMustTTestCase[int]("zero to int", 0, 0),
		newMustTTestCase[float64]("float64 to float64", 3.14, 3.14),
		newMustTTestCase[error]("error to error", errSentinel, errSentinel),
		newMustTTestCase[fmt.Stringer]("stringer to fmt.Stringer",
			stringer, stringer),
	}
}

// mustTPanicTestCases are the rows where it does not.
func mustTPanicTestCases() []TestCase {
	return []TestCase{
		newMustTTestCasePanic[string]("int to string", 42,
			"failed to convert int to string"),
		newMustTTestCasePanic[int]("string to int", testHello,
			"failed to convert string to int"),
		newMustTTestCasePanic[error]("string to error", "not an error",
			"failed to convert string to error"),
		newMustTTestCasePanic[fmt.Stringer]("int to fmt.Stringer", 42,
			"failed to convert int to fmt.Stringer"),
		newMustTTestCasePanic[string]("nil to string", nil,
			"failed to convert <nil> to string"),
		newMustTTestCasePanic[int]("nil to int", nil,
			"failed to convert <nil> to int"),
		newMustTTestCasePanic[error]("nil to error", nil,
			"failed to convert <nil> to error"),
		newMustTTestCasePanic[fmt.Stringer]("nil to fmt.Stringer", nil,
			"failed to convert <nil> to fmt.Stringer"),
	}
}

func TestMustT(t *testing.T) {
	RunTestCases(t, mustTValueTestCases())
	RunTestCases(t, mustTPanicTestCases())
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
		newMaybeTTestCase[string]("empty string to string", "", ""),
		newMaybeTTestCase[int]("int to int", 42, 42),
		newMaybeTTestCase[int]("zero to int", 0, 0),
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
		newMaybeTTestCase[error]("nil to error", nil, nil),
		newMaybeTTestCase[fmt.Stringer]("nil to fmt.Stringer", nil, nil),
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

// Benchmarks
//
// These take a failed Must apart one step at a time. Catch on its own
// sets the floor. A plain panic makes AsRecovered wrap the payload, and
// Panic builds the PanicError at the panic site instead; either way one
// stack trace is captured. Must adds ErrUnreachable and its annotation on
// top of that capture. Every panic carries the same error.

var (
	errBenchPanic = errors.New("benchmark panic")
	// errBenchNil stays nil. Being a variable, it keeps the compiler from
	// folding Must's check away as it would a literal nil.
	errBenchNil error
)

func BenchmarkMust(b *testing.B) {
	for b.Loop() {
		_ = Must(42, errBenchNil)
	}
}

func BenchmarkCatch(b *testing.B) {
	for b.Loop() {
		_ = Catch(benchReturnNil)
	}
}

func BenchmarkCatchPanic(b *testing.B) {
	for b.Loop() {
		_ = Catch(benchPanic)
	}
}

func BenchmarkCatchPanicError(b *testing.B) {
	for b.Loop() {
		_ = Catch(benchPanicError)
	}
}

func BenchmarkCatchMust(b *testing.B) {
	for b.Loop() {
		_ = Catch(benchMustFail)
	}
}

func benchReturnNil() error { return nil }

func benchPanic() error { panic(errBenchPanic) }

func benchPanicError() error {
	Panic(errBenchPanic)
	return nil
}

func benchMustFail() error {
	_ = Must(42, errBenchPanic)
	return nil
}
