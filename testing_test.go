package core_test

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = assertErrorAsTestCase[*core.WrappedError]{}
	_ core.TestCase = assertTypeIsTestCase[string]{}
	_ core.TestCase = assertPanicTestCase{}
	_ core.TestCase = assertAbortTestCase{}
	_ core.TestCase = isMockTAbortTestCase{}
	_ core.TestCase = (*mockTestCase)(nil)
	_ core.TestCase = mockTMessageAtTestCase{}
	_ core.TestCase = mockTRunTestCase{}
	_ core.TestCase = mockTRunCarryOverTestCase{}
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
	t.Run("counts", testMockTCounts)
	t.Run("message at", testMockTMessageAt)
	t.Run("concurrent safety", testMockTConcurrent)
}

// Test S helper function
func TestS(t *testing.T) {
	// Test with integers
	intSlice := core.S(1, 2, 3)
	core.AssertSliceEqual(t, []int{1, 2, 3}, intSlice, "S with integers")

	// Test with strings
	strSlice := core.S("a", "b", "c")
	core.AssertSliceEqual(t, []string{"a", "b", "c"}, strSlice, "S with strings")

	// Test empty slice
	emptySlice := core.S[string]()
	core.AssertSliceEqual(t, []string{}, emptySlice, "S empty slice")

	// Test single element
	singleSlice := core.S(42)
	core.AssertSliceEqual(t, []int{42}, singleSlice, "S single element")
}

// undecidedPair returns two values AreEqual cannot settle: maps are
// not comparable, and nil, identity and Equal methods all fail to
// decide them.
func undecidedPair() (a, b map[string]int) {
	return map[string]int{"a": 1}, map[string]int{"a": 1}
}

// Test AssertEqual
func TestAssertEqual(t *testing.T) {
	mock := &core.MockT{}

	assertPassed(t, mock, core.AssertEqual(mock, 42, 42, "equal test"), "equal")
	core.AssertLastLogForTest(t, mock, "equal test: 42", "equal")

	mock.Reset()
	assertFailed(t, mock, core.AssertEqual(mock, 42, 24, "not equal test"),
		"expected 42, got 24", "not equal")

	// The point of the widened constraint: values == cannot compare at
	// all, decided by AreEqual instead of rejected by the compiler.
	mock.Reset()
	assertPassed(t, mock, core.AssertEqual(mock, core.S(1, 2), core.S(1, 2), "slice test"),
		"slice")

	mock.Reset()
	assertPassed(t, mock, core.AssertEqual(mock, core.EqSliceForTest{1}, core.EqSliceForTest{1},
		"Equal method test"), "Equal method")

	mock.Reset()
	a, b := undecidedPair()
	assertFailed(t, mock, core.AssertEqual(mock, a, b, "undecided test"),
		"undecided for map[string]int, needs a deep comparison", "undecided")
}

// Test AssertNotEqual
func TestAssertNotEqual(t *testing.T) {
	mock := &core.MockT{}

	assertPassed(t, mock, core.AssertNotEqual(mock, 42, 24, "not equal test"),
		"not equal")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok on success")
	core.AssertEqual(t, "not equal test: 24", lastLog, "log message on success")

	mock.Reset()
	assertFailed(t, mock, core.AssertNotEqual(mock, 42, 42, "equal test"),
		"expected a different value, got 42", "equal")

	// Undecided must fail rather than count as different: AreEqual
	// reports (false, false) here, and reading only the first result
	// would pass the assertion by accident.
	mock.Reset()
	assertPassed(t, mock, core.AssertNotEqual(mock, core.S(1, 2), core.S(1, 3), "slice test"),
		"slice")

	mock.Reset()
	a, b := undecidedPair()
	assertFailed(t, mock, core.AssertNotEqual(mock, a, b, "undecided test"),
		"undecided for map[string]int, needs a deep comparison", "undecided")
}

// Test AssertSliceEqual
func TestAssertSliceEqual(t *testing.T) {
	mock := &core.MockT{}

	assertPassed(t, mock, core.AssertSliceEqual(mock, core.S(1, 2, 3), core.S(1, 2, 3),
		"slice equal test"), "equal")

	mock.Reset()
	// The differing element is named, not left to be spotted.
	assertFailed(t, mock, core.AssertSliceEqual(mock, core.S(1, 2, 3), core.S(1, 2, 4),
		"slice not equal test"), "index 2: expected 3, got 4", "not equal")

	// Length is settled before the elements, and reported as a count.
	mock.Reset()
	assertFailed(t, mock, core.AssertSliceEqual(mock, core.S(1, 2, 3), core.S(1, 2),
		"length test"), "expected 3 elements, got 2", "length")

	// An empty glass and no glass are not the same: only nil equals nil.
	// Both render as "[]", so the message names the nil-ness instead.
	mock.Reset()
	assertFailed(t, mock, core.AssertSliceEqual(mock, nil, core.S[int](),
		"nil against empty test"), "expected nil slice, got empty slice",
		"nil against empty")

	mock.Reset()
	assertFailed(t, mock, core.AssertSliceEqual(mock, core.S[int](), nil,
		"empty against nil test"), "expected empty slice, got nil slice",
		"empty against nil")

	// Matched nil-ness passes on either side of that distinction, an
	// empty slice having no elements to disagree over and nil being
	// equal to the only thing it can be.
	mock.Reset()
	assertPassed(t, mock, core.AssertSliceEqual(mock, core.S[int](), core.S[int](),
		"both empty test"), "both empty")

	mock.Reset()
	assertPassed(t, mock, core.AssertSliceEqual[int](mock, nil, nil,
		"both nil test"), "both nil")

	// AreEqual walks one level deep, so the inner slices are undecided.
	// The message names the element type, the size being already settled.
	mock.Reset()
	assertFailed(t, mock, core.AssertSliceEqual(mock, core.S(core.S(1)), core.S(core.S(1)),
		"undecided test"), "undecided for []int elements", "undecided")

	// An undecided element ahead of the mismatch must not be mistaken
	// for it: what settles the slice is the pair that differs, and that
	// is the index worth reporting.
	mock.Reset()
	a, b := undecidedPair()
	assertFailed(t, mock, core.AssertSliceEqual(mock, core.S[any](a, 1), core.S[any](b, 2),
		"mixed test"), "index 1: expected 1, got 2", "mixed")
}

// Test AssertDeepEqual
func TestAssertDeepEqual(t *testing.T) {
	mock := &core.MockT{}

	// Succeeds on a pair AssertEqual leaves undecided.
	assertPassed(t, mock, core.AssertDeepEqual(mock, core.S(core.S(1)), core.S(core.S(1)),
		"deep equal test"), "equal")

	mock.Reset()
	assertFailed(t, mock, core.AssertDeepEqual(mock, core.S(core.S(1)), core.S(core.S(2)),
		"deep not equal test"), "expected [[1]], got [[2]]", "not equal")
}

// taggedValue is non-comparable — the slice sees to that — and its Equal
// considers only the tag, so AreEqual and reflect.DeepEqual reach
// different verdicts whenever the payloads differ.
type taggedValue struct {
	tag     string
	payload []int
}

func (v taggedValue) Equal(o taggedValue) bool { return v.tag == o.tag }

// TestEqualMethodVersusDeepEqual pins the divergence between the two
// forms: AreEqual asks the type, reflect.DeepEqual never does. The deep
// form is a fallback for what AreEqual cannot decide, not a stronger
// version of it, and swapping one for the other can flip the verdict.
func TestEqualMethodVersusDeepEqual(t *testing.T) {
	mock := &core.MockT{}
	a := taggedValue{payload: core.S(1), tag: "x"}
	b := taggedValue{payload: core.S(2), tag: "x"}

	assertPassed(t, mock, core.AssertEqual(mock, a, b, "Equal method test"),
		"Equal method")

	mock.Reset()
	assertFailed(t, mock, core.AssertDeepEqual(mock, a, b, "deep test"),
		"expected {x [1]}, got {x [2]}", "deep")
}

// Test AssertNotDeepEqual
func TestAssertNotDeepEqual(t *testing.T) {
	mock := &core.MockT{}

	assertPassed(t, mock, core.AssertNotDeepEqual(mock, core.S(core.S(1)), core.S(core.S(2)),
		"deep not equal test"), "not equal")

	mock.Reset()
	assertFailed(t, mock, core.AssertNotDeepEqual(mock, core.S(core.S(1)), core.S(core.S(1)),
		"deep equal test"), "expected a different value, got [[1]]", "equal")
}

// Test AssertContains
func TestAssertContains(t *testing.T) {
	mock := &core.MockT{}

	assertPassed(t, mock, core.AssertContains(mock, "hello world", "world",
		"contains test"), "contains")
	core.AssertLastLogForTest(t, mock, "contains test: contains \"world\"", "contains")

	mock.Reset()
	assertFailed(t, mock, core.AssertContains(mock, "hello world", "xyz",
		"not contains test"), "expected \"hello world\" to contain \"xyz\"", "missing")

	mock.Reset()
	assertFailed(t, mock, core.AssertContains(mock, "hello world", "",
		"empty substring test"), "expected a non-empty substring", "empty")
}

// Test AssertNotContain
func TestAssertNotContain(t *testing.T) {
	mock := &core.MockT{}

	assertPassed(t, mock, core.AssertNotContain(mock, "hello world", "xyz",
		"not contains test"), "not contains")
	core.AssertLastLogForTest(t, mock, "not contains test: does not contain \"xyz\"", "not contains")

	mock.Reset()
	assertFailed(t, mock, core.AssertNotContain(mock, "hello world", "world",
		"contains test"), "expected \"hello world\" not to contain \"world\"", "present")

	mock.Reset()
	assertFailed(t, mock, core.AssertNotContain(mock, "hello world", "",
		"empty substring test"), "expected a non-empty substring", "empty")
}

// Test AssertError and AssertNoError
func TestAssertError(t *testing.T) {
	mock := &core.MockT{}
	testErr := errors.New("test error")

	// Test AssertError with error (success)
	result := core.AssertError(mock, testErr, "error test")
	core.AssertTrue(t, result, "AssertError result with error")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertError with nil (failure)
	result = core.AssertError(mock, nil, "nil error test")
	core.AssertFalse(t, result, "AssertError result with nil")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertNoError with nil (success)
	result = core.AssertNoError(mock, nil, "no error test")
	core.AssertTrue(t, result, "AssertNoError result with nil")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNoError with error (failure)
	result = core.AssertNoError(mock, testErr, "unexpected error test")
	core.AssertFalse(t, result, "AssertNoError result with error")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertTrue and AssertFalse
func TestAssertBool(t *testing.T) {
	mock := &core.MockT{}

	// Test AssertTrue with true (success)
	result := core.AssertTrue(mock, true, "true test")
	core.AssertEqual(t, true, result, "AssertTrue(%v)", true)
	core.AssertEqual(t, false, mock.HasErrors(), "HasErrors")
	core.AssertEqual(t, true, mock.HasLogs(), "HasLogs")

	mock.Reset()

	// Test AssertTrue with false (failure)
	result = core.AssertTrue(mock, false, "false test")
	core.AssertEqual(t, false, result, "AssertTrue(%v)", false)
	core.AssertEqual(t, true, mock.HasErrors(), "HasErrors")

	mock.Reset()

	// Test AssertFalse with false (success)
	result = core.AssertFalse(mock, false, "false test")
	core.AssertEqual(t, true, result, "AssertFalse(%v)", false)
	core.AssertEqual(t, false, mock.HasErrors(), "HasErrors")
	core.AssertEqual(t, true, mock.HasLogs(), "HasLogs")

	mock.Reset()

	// Test AssertFalse with true (failure)
	result = core.AssertFalse(mock, true, "true test")
	core.AssertEqual(t, false, result, "AssertFalse(%v)", true)
	core.AssertEqual(t, true, mock.HasErrors(), "HasErrors")
}

// Test AssertNil and AssertNotNil
func TestAssertNil(t *testing.T) {
	mock := &core.MockT{}

	// Test AssertNil with nil (success)
	result := core.AssertNil(mock, nil, "nil test")
	core.AssertTrue(t, result, "AssertNil result with nil")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNil with non-nil (failure)
	result = core.AssertNil(mock, "not nil", "not nil test")
	core.AssertFalse(t, result, "AssertNil result with non-nil")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertNotNil with non-nil (success)
	result = core.AssertNotNil(mock, "not nil", "not nil test")
	core.AssertTrue(t, result, "AssertNotNil result with non-nil")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNotNil with a pointer (success): the log names the
	// type, not the address, so it reads the same on every run.
	result = core.AssertNotNil(mock, new(int), "pointer test")
	core.AssertTrue(t, result, "AssertNotNil result with pointer")
	core.AssertLastLogForTest(t, mock, "pointer test: *int", "pointer")

	mock.Reset()

	// Test AssertNotNil with nil (failure)
	result = core.AssertNotNil(mock, nil, "nil test")
	core.AssertFalse(t, result, "AssertNotNil result with nil")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertErrorIs
func TestAssertErrorIs(t *testing.T) {
	mock := &core.MockT{}
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(baseErr, errors.New("wrapped"))

	// Test AssertErrorIs with matching error (success)
	result := core.AssertErrorIs(mock, wrappedErr, baseErr, "error is test")
	core.AssertTrue(t, result, "AssertErrorIs result with matching error")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertErrorIs with non-matching error (failure)
	otherErr := errors.New("other error")
	result = core.AssertErrorIs(mock, wrappedErr, otherErr, "error is not test")
	core.AssertFalse(t, result, "AssertErrorIs result with non-matching error")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertNotErrorIs
func TestAssertNotErrorIs(t *testing.T) {
	mock := &core.MockT{}
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(baseErr, errors.New("wrapped"))
	otherErr := errors.New("other error")

	// Test AssertNotErrorIs with non-matching error (success)
	result := core.AssertNotErrorIs(mock, wrappedErr, otherErr, "no match test")
	core.AssertTrue(t, result, "AssertNotErrorIs result with non-matching error")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertNotErrorIs with matching error (failure)
	result = core.AssertNotErrorIs(mock, wrappedErr, baseErr, "match test")
	core.AssertFalse(t, result, "AssertNotErrorIs result with matching error")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertNotErrorIs with nil error (success, nil matches no
	// non-nil target)
	result = core.AssertNotErrorIs(mock, nil, otherErr, "nil error test")
	core.AssertTrue(t, result, "AssertNotErrorIs result with nil error")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")

	mock.Reset()

	// Test AssertNotErrorIs with nil error and nil target (failure,
	// errors.Is(nil, nil) matches)
	result = core.AssertNotErrorIs(mock, nil, nil, "nil both test")
	core.AssertFalse(t, result, "AssertNotErrorIs result with nil error and target")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// Test AssertErrorIsFn
func TestAssertErrorIsFn(t *testing.T) {
	mock := &core.MockT{}
	baseErr := errors.New("base error")
	wrappedErr := core.Wrap(baseErr, "wrapped")
	isBase := func(err error) bool { return err == baseErr }

	// Test AssertErrorIsFn with matching function (success)
	result := core.AssertErrorIsFn(mock, wrappedErr, isBase, "error fn test")
	core.AssertTrue(t, result, "AssertErrorIsFn result with matching function")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	mock.Reset()

	// Test AssertErrorIsFn with non-matching error (failure)
	result = core.AssertErrorIsFn(mock, errors.New("other error"), isBase, "error fn mismatch test")
	core.AssertFalse(t, result, "AssertErrorIsFn result with non-matching error")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertErrorIsFn with nil function (failure)
	result = core.AssertErrorIsFn(mock, wrappedErr, nil, "nil fn test")
	core.AssertFalse(t, result, "AssertErrorIsFn result with nil function")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertErrorIsFn with nil error (failure)
	result = core.AssertErrorIsFn(mock, nil, isBase, "nil error test")
	core.AssertFalse(t, result, "AssertErrorIsFn result with nil error")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
}

// assertErrorAsTestCase states what AssertErrorAs answers for one
// target type: the match itself and true when the chain holds one,
// the zero value and a report naming the target when it does not, the
// target included when it is an interface and the zero value has no
// dynamic type to print.
type assertErrorAsTestCase[V error] struct {
	err       error
	want      V
	wantError string
	name      string
	wantOK    bool
}

func newAssertErrorAsTestCase[V error](name string, err error, want V) core.TestCase {
	return assertErrorAsTestCase[V]{
		err:    err,
		want:   want,
		name:   name,
		wantOK: true,
	}
}

func newAssertErrorAsTestCaseFail[V error](name string, err error,
	wantError string) core.TestCase {
	return assertErrorAsTestCase[V]{
		err:       err,
		wantError: wantError,
		name:      name,
	}
}

func (tc assertErrorAsTestCase[V]) Name() string {
	return tc.name
}

func (tc assertErrorAsTestCase[V]) Test(t *testing.T) {
	mock := &core.MockT{}

	out, ok := core.AssertErrorAs[V](mock, tc.err, "type")
	if !tc.wantOK {
		var zero V
		core.AssertEqual(t, zero, out, "value")
		assertFailed(t, mock, ok, tc.wantError, "type")
		return
	}

	core.AssertSame(t, tc.want, out, "value")
	assertPassed(t, mock, ok, "type")
}

func assertErrorAsTestCases() []core.TestCase {
	baseErr := errors.New("base error")
	wrapped := core.MustT[*core.WrappedError](core.Wrap(baseErr, "note"))
	joined := errors.Join(errors.New("other error"), wrapped)

	return core.S(
		newAssertErrorAsTestCase("direct match", wrapped, wrapped),
		newAssertErrorAsTestCase("nested match", joined, wrapped),
		newAssertErrorAsTestCase("value match", core.StringError("value"),
			core.StringError("value")),
		newAssertErrorAsTestCaseFail[*core.WrappedError]("no match", baseErr,
			"expected error of type *core.WrappedError"),
		newAssertErrorAsTestCaseFail[core.Recovered]("interface target", baseErr,
			"expected error of type core.Recovered"),
		newAssertErrorAsTestCaseFail[core.StringError]("value target", baseErr,
			"expected error of type core.StringError"),
	)
}

func TestAssertErrorAs(t *testing.T) {
	core.RunTestCases(t, assertErrorAsTestCases())
}

// assertTypeIsTestCase states what AssertTypeIs answers for one target
// type: the value and true when it is one, the zero value and a report
// naming both types when it is not, the target included when it is an
// interface and the zero value has no dynamic type to print.
type assertTypeIsTestCase[V any] struct {
	input     any
	want      V
	wantError string
	name      string
	wantOK    bool
}

func newAssertTypeIsTestCase[V any](name string, input any, want V) core.TestCase {
	return assertTypeIsTestCase[V]{
		input:  input,
		want:   want,
		name:   name,
		wantOK: true,
	}
}

func newAssertTypeIsTestCaseFail[V any](name string, input any,
	wantError string) core.TestCase {
	return assertTypeIsTestCase[V]{
		input:     input,
		wantError: wantError,
		name:      name,
	}
}

func (tc assertTypeIsTestCase[V]) Name() string {
	return tc.name
}

func (tc assertTypeIsTestCase[V]) Test(t *testing.T) {
	mock := &core.MockT{}

	got, ok := core.AssertTypeIs[V](mock, tc.input, "type")
	core.AssertEqual(t, tc.want, got, "value")
	if tc.wantOK {
		assertPassed(t, mock, ok, "type")
	} else {
		assertFailed(t, mock, ok, tc.wantError, "type")
	}
}

func assertTypeIsTestCases() []core.TestCase {
	err := errors.New("test error")

	return core.S(
		newAssertTypeIsTestCase("string", testHello, testHello),
		newAssertTypeIsTestCase("error", err, err),
		newAssertTypeIsTestCaseFail[string]("int to string", 42,
			"expected type string, got int"),
		newAssertTypeIsTestCaseFail[error]("int to error", 42,
			"expected type error, got int"),
	)
}

func TestAssertTypeIs(t *testing.T) {
	core.RunTestCases(t, assertTypeIsTestCases())
}

// assertPanicTestCase for table-driven tests
type assertPanicTestCase struct {
	panicFn  func()
	expected any
	// message is a substring of what the assertion is expected to
	// report: the log on acceptance, the error on rejection.
	message      string
	name         string
	desc         string
	expectResult bool
}

func (tc assertPanicTestCase) Name() string {
	return tc.name
}

func (tc assertPanicTestCase) Test(t *testing.T) {
	mock := &core.MockT{}

	result := core.AssertPanic(mock, tc.panicFn, tc.expected, tc.desc)
	core.AssertEqual(t, !result, mock.HasErrors(), "errors follow result")

	if !tc.expectResult {
		assertFailed(t, mock, result, tc.message, tc.desc)
		return
	}

	assertPassed(t, mock, result, tc.desc)

	lastLog, ok := mock.LastLog()
	core.AssertMustTrue(t, ok, "logged")
	core.AssertContains(t, lastLog, tc.message, "log content")
}

// newAssertPanicTestCase builds the part of a row that does not depend on the
// outcome. Every row says what the assertion reports, so neither variant may
// leave message empty.
func newAssertPanicTestCase(name string, panicFn func(), expected any,
	desc, message string) assertPanicTestCase {
	if message == "" {
		panic("assertPanicTestCase: every row must say what it reports")
	}

	return assertPanicTestCase{
		name:     name,
		panicFn:  panicFn,
		expected: expected,
		desc:     desc,
		message:  message,
	}
}

// newAssertPanicTestCaseAccept declares a row the assertion accepts: it
// returns true, records no error, and logs a message containing wantLog.
func newAssertPanicTestCaseAccept(name string, panicFn func(), expected any,
	desc, wantLog string) assertPanicTestCase {
	tc := newAssertPanicTestCase(name, panicFn, expected, desc, wantLog)
	tc.expectResult = true
	return tc
}

// newAssertPanicTestCaseReject declares a row the assertion rejects: it
// returns false, logs nothing, and records an error containing wantError.
func newAssertPanicTestCaseReject(name string, panicFn func(), expected any,
	desc, wantError string) assertPanicTestCase {
	tc := newAssertPanicTestCase(name, panicFn, expected, desc, wantError)
	tc.expectResult = false
	return tc
}

func assertPanicTestCases() []assertPanicTestCase {
	testErr := errors.New("test error")
	otherErr := errors.New("other error")
	panicErr := core.NewPanicError(0, "recovered panic")
	wrappedPanic := core.NewPanicError(0, 99)

	return []assertPanicTestCase{
		// Any panic tests
		newAssertPanicTestCaseAccept("any panic accepted",
			func() { panic("test panic") }, nil, "panic test", "test panic"),
		newAssertPanicTestCaseReject("no panic fails",
			func() {}, nil, "no panic test", "expected panic but got nil"),
		newAssertPanicTestCaseReject("nil function",
			nil, nil, "nil function test", "expected a function"),

		// String matching tests
		newAssertPanicTestCaseAccept("string substring match",
			func() { panic("specific panic message") }, "specific", "string test", "contains"),
		newAssertPanicTestCaseReject("string mismatch",
			func() { panic("wrong") }, "expected", "wrong string test",
			`expected panic to contain "expected"`),
		newAssertPanicTestCaseReject("empty substring",
			func() { panic("any message") }, "", "empty substring test",
			"expected a non-empty substring"),
		// The call-site mistake is reported whatever fn does, not
		// as "expected panic but got nil".
		newAssertPanicTestCaseReject("empty substring without panic",
			func() {}, "", "empty substring no panic test",
			"expected a non-empty substring"),
		newAssertPanicTestCaseAccept("non-string panic with string expected",
			func() { panic(123) }, "123", "non-string test", "contains"),
		// Only a string is searched as itself; every other payload
		// is rendered as a *PanicError, an error included.
		newAssertPanicTestCaseReject("string panic carries no prefix",
			func() { panic("boom") }, "panic: boom",
			"string prefix test",
			`expected panic to contain "panic: boom"`),
		newAssertPanicTestCaseAccept("StringError panic renders as an error",
			func() { panic(core.StringError("boom")) }, "panic: boom",
			"StringError test", "contains"),
		newAssertPanicTestCaseAccept("PanicError panic keeps its own prefix",
			func() { panic(panicErr) }, "panic:",
			"PanicError prefix test", "contains"),

		// Error matching tests
		newAssertPanicTestCaseAccept("error match",
			func() { panic(testErr) }, testErr, "error test", "panic error"),
		newAssertPanicTestCaseReject("error mismatch",
			func() { panic(testErr) }, otherErr, "wrong error test",
			"expected panic error other error"),

		// Exact matching tests
		newAssertPanicTestCaseAccept("integer match",
			func() { panic(42) }, 42, "int test", "panic: 42"),
		newAssertPanicTestCaseReject("integer mismatch",
			func() { panic(42) }, 43, "wrong int test", "expected panic 43"),

		// Recovered type tests
		newAssertPanicTestCaseAccept("Recovered type match",
			func() { panic(panicErr) }, panicErr, "recovered test", "panic:"),
		newAssertPanicTestCaseAccept("Recovered unwrapping",
			func() { panic(wrappedPanic) }, 99, "unwrapped test", "panic: 99"),
	}
}

// Test AssertPanic
func TestAssertPanic(t *testing.T) {
	core.RunTestCases(t, assertPanicTestCases())
}

// assertAbortTestCase states what AssertPanic and AssertNoPanic do when
// fn cuts the test short on the T the assertion reports to: the abort
// passes through and is not mistaken for a panic, the test is failed,
// skipped or both by fn alone, and the assertion neither logs nor adds
// an error of its own. expectPanic is what the assertion under test
// expects of fn.
type assertAbortTestCase struct {
	abort      func(core.T)
	name       string
	wantErrors int
	wantLogs   int

	expectPanic bool
	wantFailed  bool
	wantSkipped bool
}

func (tc assertAbortTestCase) Name() string {
	return tc.name
}

func (tc assertAbortTestCase) Test(t *testing.T) {
	mock := &core.MockT{}
	ok := mock.Run(tc.name, func(mt core.T) {
		fn := func() { tc.abort(mt) }
		if tc.expectPanic {
			core.AssertPanic(mt, fn, nil, "aborted")
		} else {
			core.AssertNoPanic(mt, fn, "aborted")
		}
	})

	core.AssertEqual(t, !tc.wantFailed, ok, "passed")
	core.AssertEqual(t, tc.wantFailed, mock.Failed(), "failed")
	core.AssertEqual(t, tc.wantSkipped, mock.Skipped(), "skipped")
	core.AssertEqual(t, tc.wantErrors, mock.NumErrors(), "errors")
	core.AssertEqual(t, tc.wantLogs, mock.NumLogs(), "logs")
}

// newAssertPanicAbortTestCase is an AssertPanic row whose abort fails
// the test, recording wantErrors on the way.
func newAssertPanicAbortTestCase(name string, abort func(core.T),
	wantErrors int) assertAbortTestCase {
	return assertAbortTestCase{
		abort:       abort,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    0,
		expectPanic: true,
		wantFailed:  true,
		wantSkipped: false,
	}
}

// newAssertPanicAbortTestCaseSkips is an AssertPanic row whose abort
// skips the test, recording wantLogs on the way.
func newAssertPanicAbortTestCaseSkips(name string, abort func(core.T),
	wantLogs int) assertAbortTestCase {
	return assertAbortTestCase{
		abort:       abort,
		name:        name,
		wantErrors:  0,
		wantLogs:    wantLogs,
		expectPanic: true,
		wantFailed:  false,
		wantSkipped: true,
	}
}

// newAssertPanicAbortTestCaseFailsSkips is an AssertPanic row whose
// abort fails the test and then skips it, recording wantErrors and
// wantLogs on the way.
func newAssertPanicAbortTestCaseFailsSkips(name string, abort func(core.T),
	wantErrors, wantLogs int) assertAbortTestCase {
	return assertAbortTestCase{
		abort:       abort,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    wantLogs,
		expectPanic: true,
		wantFailed:  true,
		wantSkipped: true,
	}
}

var assertPanicAbortTestCases = core.S(
	newAssertPanicAbortTestCase("FailNow", func(mt core.T) { mt.FailNow() }, 0),
	newAssertPanicAbortTestCase("Fatal", func(mt core.T) { mt.Fatal("fatal") }, 1),
	newAssertPanicAbortTestCase("AssertMust", func(mt core.T) {
		core.AssertMustTrue(mt, false, "must")
	}, 1),
	newAssertPanicAbortTestCaseSkips("SkipNow", func(mt core.T) { mt.SkipNow() }, 0),
	newAssertPanicAbortTestCaseSkips("Skip", func(mt core.T) { mt.Skip("skip") }, 1),
	newAssertPanicAbortTestCaseFailsSkips("Fail then SkipNow",
		func(mt core.T) { mt.Fail(); mt.SkipNow() }, 0, 0),
)

func TestAssertPanicAbort(t *testing.T) {
	core.RunTestCases(t, assertPanicAbortTestCases)
}

// newAssertNoPanicAbortTestCase is an AssertNoPanic row whose abort
// fails the test, recording wantErrors on the way.
func newAssertNoPanicAbortTestCase(name string, abort func(core.T),
	wantErrors int) assertAbortTestCase {
	return assertAbortTestCase{
		abort:       abort,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    0,
		expectPanic: false,
		wantFailed:  true,
		wantSkipped: false,
	}
}

// newAssertNoPanicAbortTestCaseSkips is an AssertNoPanic row whose abort
// skips the test, recording wantLogs on the way.
func newAssertNoPanicAbortTestCaseSkips(name string, abort func(core.T),
	wantLogs int) assertAbortTestCase {
	return assertAbortTestCase{
		abort:       abort,
		name:        name,
		wantErrors:  0,
		wantLogs:    wantLogs,
		expectPanic: false,
		wantFailed:  false,
		wantSkipped: true,
	}
}

// newAssertNoPanicAbortTestCaseFailsSkips is an AssertNoPanic row whose
// abort fails the test and then skips it, recording wantErrors and
// wantLogs on the way.
func newAssertNoPanicAbortTestCaseFailsSkips(name string, abort func(core.T),
	wantErrors, wantLogs int) assertAbortTestCase {
	return assertAbortTestCase{
		abort:       abort,
		name:        name,
		wantErrors:  wantErrors,
		wantLogs:    wantLogs,
		expectPanic: false,
		wantFailed:  true,
		wantSkipped: true,
	}
}

var assertNoPanicAbortTestCases = core.S(
	newAssertNoPanicAbortTestCase("FailNow", func(mt core.T) { mt.FailNow() }, 0),
	newAssertNoPanicAbortTestCase("Fatal", func(mt core.T) { mt.Fatal("fatal") }, 1),
	newAssertNoPanicAbortTestCase("AssertMust", func(mt core.T) {
		core.AssertMustTrue(mt, false, "must")
	}, 1),
	newAssertNoPanicAbortTestCaseSkips("SkipNow", func(mt core.T) { mt.SkipNow() }, 0),
	newAssertNoPanicAbortTestCaseSkips("Skip", func(mt core.T) { mt.Skip("skip") }, 1),
	newAssertNoPanicAbortTestCaseFailsSkips("Fail then SkipNow",
		func(mt core.T) { mt.Fail(); mt.SkipNow() }, 0, 0),
)

func TestAssertNoPanicAbort(t *testing.T) {
	core.RunTestCases(t, assertNoPanicAbortTestCases)
}

// Test AssertNoPanic
func TestAssertNoPanic(t *testing.T) {
	mock := &core.MockT{}

	// Test AssertNoPanic without panic (success)
	result := core.AssertNoPanic(mock, func() {}, "no panic test")
	core.AssertTrue(t, result, "AssertNoPanic result without panic")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok on success")
	core.AssertTrue(t, strings.Contains(lastLog, "no panic"), "log contains no panic message")

	mock.Reset()

	// Test AssertNoPanic with panic (failure)
	result = core.AssertNoPanic(mock, func() { panic("unexpected") }, "panic test")
	core.AssertFalse(t, result, "AssertNoPanic result with panic")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")

	mock.Reset()

	// Test AssertNoPanic with a nil function (failure): the call-site
	// mistake is reported as such, not as a panic.
	result = core.AssertNoPanic(mock, nil, "nil function test")
	assertFailed(t, mock, result, "expected a function", "nil function")
}

// Test RunConcurrentTest
func TestRunConcurrentTest(t *testing.T) {
	mock := &core.MockT{}
	var counter atomic.Int64

	// Test successful concurrent execution
	err := core.RunConcurrentTest(mock, 5, func(_ int) error {
		counter.Add(1)
		return nil
	})
	core.AssertNoError(t, err, "RunConcurrentTest successful")
	core.AssertEqual(t, int64(5), counter.Load(), "all workers executed")

	// Test concurrent execution with error
	err = core.RunConcurrentTest(mock, 3, func(id int) error {
		if id == 1 {
			return errors.New("worker error")
		}
		return nil
	})
	core.AssertError(t, err, "RunConcurrentTest with error")
	core.AssertTrue(t, strings.Contains(err.Error(), "worker error"), "error message")
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

func TestRunTestCases(t *testing.T) {
	tc1 := newMockTestCase("test1")
	tc2 := newMockTestCase("test2")
	tc3 := newMockTestCase("test3")

	cases := []core.TestCase{tc1, tc2, tc3}
	core.RunTestCases(t, cases)

	// Verify all test cases were called
	core.AssertTrue(t, tc1.called, "test case 1 called")
	core.AssertTrue(t, tc2.called, "test case 2 called")
	core.AssertTrue(t, tc3.called, "test case 3 called")
}

// Test RunBenchmark
func TestRunBenchmark(t *testing.T) {
	setupCalled := false
	execCount := 0

	// Create a mock benchmark
	b := &testing.B{N: 10}
	b.ResetTimer() // Simulate benchmark setup

	core.RunBenchmark(b, func() any {
		setupCalled = true
		return "test data"
	}, func(data any) {
		core.AssertEqual(t, "test data", data, "benchmark data")
		execCount++
	})

	core.AssertTrue(t, setupCalled, "setup function called")
	core.AssertEqual(t, 10, execCount, "execution function called N times")
}

func testMockTInitialState(t *testing.T) {
	mock := &core.MockT{}

	core.AssertFalse(t, mock.HasErrors(), "initial HasErrors")
	core.AssertFalse(t, mock.HasLogs(), "initial HasLogs")
	core.AssertEqual(t, 0, mock.NumHelperCalls(), "initial NumHelperCalls")
	core.AssertFalse(t, mock.Failed(), "initial Failed")
	core.AssertFalse(t, mock.Skipped(), "initial Skipped")
}

func testMockTHelper(t *testing.T) {
	mock := &core.MockT{}

	mock.Helper()
	core.AssertEqual(t, 1, mock.NumHelperCalls(), "NumHelperCalls after Helper()")

	// Test multiple calls increment the counter
	mock.Helper()
	mock.Helper()
	core.AssertEqual(t, 3, mock.NumHelperCalls(), "NumHelperCalls after multiple calls")
}

func testMockTErrors(t *testing.T) {
	mock := &core.MockT{}

	mock.Error("test error")
	core.AssertTrue(t, mock.HasErrors(), "HasErrors after Error")
	core.AssertTrue(t, mock.Failed(), "Failed after Error")
	core.AssertEqual(t, 1, mock.NumErrors(), "NumErrors")

	lastErr, ok := mock.LastError()
	core.AssertTrue(t, ok, "LastError ok")
	core.AssertEqual(t, "test error", lastErr, "LastError value")
}

func testMockTLogs(t *testing.T) {
	mock := &core.MockT{}

	mock.Log("test log")
	core.AssertTrue(t, mock.HasLogs(), "HasLogs after Log")
	core.AssertFalse(t, mock.Failed(), "Failed after Log should be false")
	core.AssertEqual(t, 1, mock.NumLogs(), "NumLogs")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok")
	core.AssertEqual(t, "test log", lastLog, "LastLog value")
}

func testMockTFail(t *testing.T) {
	mock := &core.MockT{}

	// Test Fail() marks test as failed
	core.AssertFalse(t, mock.Failed(), "initial Failed state")
	mock.Fail()
	core.AssertTrue(t, mock.Failed(), "Failed after Fail()")

	// Test that Error() also marks as failed
	mock.Reset()
	core.AssertFalse(t, mock.Failed(), "Failed after Reset")
	mock.Error("test error")
	core.AssertTrue(t, mock.Failed(), "Failed after Error")

	// Test that Errorf() also marks as failed
	mock.Reset()
	core.AssertFalse(t, mock.Failed(), "Failed after Reset")
	mock.Errorf("formatted error: %s", "test")
	core.AssertTrue(t, mock.Failed(), "Failed after Errorf")

	// Test that Log() does not mark as failed
	mock.Reset()
	mock.Log("test log")
	core.AssertFalse(t, mock.Failed(), "Failed after Log should be false")

	// Test that Logf() does not mark as failed
	mock.Reset()
	mock.Logf("formatted log: %s", "test")
	core.AssertFalse(t, mock.Failed(), "Failed after Logf should be false")
}

func testMockTFormatted(t *testing.T) {
	mock := &core.MockT{}

	// Test Errorf
	mock.Errorf("error %d: %s", 42, "test message")
	core.AssertTrue(t, mock.HasErrors(), "HasErrors after Errorf")
	core.AssertTrue(t, mock.Failed(), "Failed after Errorf")
	core.AssertEqual(t, 1, mock.NumErrors(), "NumErrors after Errorf")

	lastErr, ok := mock.LastError()
	core.AssertTrue(t, ok, "LastError ok after Errorf")
	core.AssertEqual(t, "error 42: test message", lastErr, "LastError value after Errorf")

	mock.Reset()

	// Test Logf
	mock.Logf("log %d: %s", 24, "test message")
	core.AssertTrue(t, mock.HasLogs(), "HasLogs after Logf")
	core.AssertFalse(t, mock.Failed(), "Failed after Logf should be false")
	core.AssertEqual(t, 1, mock.NumLogs(), "NumLogs after Logf")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok after Logf")
	core.AssertEqual(t, "log 24: test message", lastLog, "LastLog value after Logf")
}

func testMockTMultiple(t *testing.T) {
	mock := &core.MockT{}

	mock.Error("first error")
	mock.Log("first log")
	mock.Error("second error")
	mock.Log("second log")

	core.AssertEqual(t, 2, mock.NumErrors(), "multiple errors length")
	core.AssertEqual(t, 2, mock.NumLogs(), "multiple logs length")

	lastErr, ok := mock.LastError()
	core.AssertTrue(t, ok, "LastError ok after multiple")
	core.AssertEqual(t, "second error", lastErr, "LastError after multiple")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok after multiple")
	core.AssertEqual(t, "second log", lastLog, "LastLog after multiple")
}

func testMockTReset(t *testing.T) {
	mock := &core.MockT{}

	mock.Error("test error")
	mock.Log("test log")
	mock.Helper()
	mock.Helper()
	mock.Fail()
	mock.Run("skip", func(mt core.T) { mt.SkipNow() })
	core.AssertMustTrue(t, mock.Skipped(), "Skipped before Reset")

	mock.Reset()
	core.AssertFalse(t, mock.HasErrors(), "HasErrors after Reset")
	core.AssertFalse(t, mock.HasLogs(), "HasLogs after Reset")
	core.AssertEqual(t, 0, mock.NumErrors(), "NumErrors after Reset")
	core.AssertEqual(t, 0, mock.NumLogs(), "NumLogs after Reset")
	core.AssertEqual(t, 0, mock.NumHelperCalls(), "NumHelperCalls after Reset")
	core.AssertFalse(t, mock.Failed(), "Failed after Reset")
	core.AssertFalse(t, mock.Skipped(), "Skipped after Reset")
}

func testMockTEmptyQueries(t *testing.T) {
	mock := &core.MockT{}

	lastErr, ok := mock.LastError()
	core.AssertFalse(t, ok, "LastError ok when empty")
	core.AssertEqual(t, "", lastErr, "LastError value when empty")

	lastLog, ok := mock.LastLog()
	core.AssertFalse(t, ok, "LastLog ok when empty")
	core.AssertEqual(t, "", lastLog, "LastLog value when empty")

	core.AssertEqual(t, 0, mock.NumErrors(), "NumErrors when empty")
	core.AssertEqual(t, 0, mock.NumLogs(), "NumLogs when empty")
	core.AssertEqual(t, 0, mock.NumHelperCalls(), "NumHelperCalls when empty")

	msg, ok := mock.ErrorAt(0)
	core.AssertFalse(t, ok, "ErrorAt(0) ok when empty")
	core.AssertEqual(t, "", msg, "ErrorAt(0) value when empty")

	msg, ok = mock.LogAt(-1)
	core.AssertFalse(t, ok, "LogAt(-1) ok when empty")
	core.AssertEqual(t, "", msg, "LogAt(-1) value when empty")
}

func testMockTCounts(t *testing.T) {
	mock := &core.MockT{}

	mock.Error("first error")
	mock.Errorf("second %s", "error")
	mock.Log("first log")
	mock.Logf("second %s", "log")
	mock.Log("third log")
	mock.Helper()
	mock.Helper()
	mock.Helper()
	mock.Helper()

	core.AssertEqual(t, 2, mock.NumErrors(), "NumErrors")
	core.AssertEqual(t, 3, mock.NumLogs(), "NumLogs")
	core.AssertEqual(t, 4, mock.NumHelperCalls(), "NumHelperCalls")
}

// mockTMessageAtTestCase exercises the index rule shared by ErrorAt and
// LogAt against a mock holding the same three messages in both lists. The
// middle one is empty, so a recorded empty message and an index out of
// range come back with the same string and differ only in the bool.
type mockTMessageAtTestCase struct {
	name   string
	want   string
	index  int
	wantOK bool
}

func newMockTMessageAtTestCase(name string, index int, want string,
	wantOK bool) mockTMessageAtTestCase {
	return mockTMessageAtTestCase{
		name:   name,
		index:  index,
		want:   want,
		wantOK: wantOK,
	}
}

func (tc mockTMessageAtTestCase) Name() string {
	return tc.name
}

func (tc mockTMessageAtTestCase) Test(t *testing.T) {
	mock := &core.MockT{}
	for _, msg := range core.S("first", "", "third") {
		mock.Error(msg)
		mock.Log(msg)
	}

	msg, ok := mock.ErrorAt(tc.index)
	core.AssertEqual(t, tc.wantOK, ok, "ErrorAt ok")
	core.AssertEqual(t, tc.want, msg, "ErrorAt value")

	msg, ok = mock.LogAt(tc.index)
	core.AssertEqual(t, tc.wantOK, ok, "LogAt ok")
	core.AssertEqual(t, tc.want, msg, "LogAt value")
}

func testMockTMessageAt(t *testing.T) {
	core.RunTestCases(t, core.S(
		newMockTMessageAtTestCase("first", 0, "first", true),
		newMockTMessageAtTestCase("empty message", 1, "", true),
		newMockTMessageAtTestCase("last", 2, "third", true),
		newMockTMessageAtTestCase("past the end", 3, "", false),
		newMockTMessageAtTestCase("last from the end", -1, "third", true),
		newMockTMessageAtTestCase("first from the end", -3, "first", true),
		newMockTMessageAtTestCase("before the start", -4, "", false),
	))
}

func testMockTConcurrent(t *testing.T) {
	mock := &core.MockT{}

	// Test concurrent access to MockT methods
	err := core.RunConcurrentTest(t, 10, func(id int) error {
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
		default:
		}
		// Read while the other workers write. The values are in flight
		// and not asserted; the race detector is what this exercises.
		mock.NumErrors()
		mock.ErrorAt(-1)
		mock.NumLogs()
		mock.LogAt(0)
		mock.NumHelperCalls()
		return nil
	})
	core.AssertNoError(t, err, "concurrent MockT operations")

	// Verify that all operations completed without race conditions
	core.AssertTrue(t, mock.HasErrors(), "has errors after concurrent operations")
	core.AssertTrue(t, mock.HasLogs(), "has logs after concurrent operations")
	core.AssertTrue(t, mock.Failed(), "failed after concurrent operations")

	// Check that we have the expected number of operations
	// We should have 2 errors per 5 operations (case 1 and 3)
	// We should have 2 logs per 5 operations (case 2 and 4)
	// We should have 1 helper call per 5 operations (case 0)
	expectedErrors := 4  // 10 workers, 2 out of every 5 operations
	expectedLogs := 4    // 10 workers, 2 out of every 5 operations
	expectedHelpers := 2 // 10 workers, 1 out of every 5 operations

	core.AssertEqual(t, expectedErrors, mock.NumErrors(), "concurrent error count")
	core.AssertEqual(t, expectedLogs, mock.NumLogs(), "concurrent log count")
	core.AssertEqual(t, expectedHelpers, mock.NumHelperCalls(), "concurrent helper count")
}

// Test AssertSame
func TestAssertSame(t *testing.T) {
	mock := &core.MockT{}

	// Test successful assertion with value types
	result := core.AssertSame(mock, 42, 42, "same value test")
	core.AssertTrue(t, result, "AssertSame result when same values")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok on success")
	core.AssertEqual(t, "same value test: same value or reference", lastLog, "log message on success")

	mock.Reset()

	// Test successful assertion with reference types
	slice1 := core.S(1, 2, 3)
	slice2 := slice1
	result = core.AssertSame(mock, slice1, slice2, "same slice test")
	core.AssertTrue(t, result, "AssertSame result when same slices")
	core.AssertFalse(t, mock.HasErrors(), "no errors on slice success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on slice success")

	mock.Reset()

	// Test failed assertion with value types
	result = core.AssertSame(mock, 42, 43, "different value test")
	core.AssertFalse(t, result, "AssertSame result when different values")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
	core.AssertFalse(t, mock.HasLogs(), "no logs on failure")

	assertErrorContains(t, mock, "expected same value or reference, got different", "error message on failure")

	mock.Reset()

	// Test failed assertion with reference types
	slice3 := core.S(1, 2, 3)
	result = core.AssertSame(mock, slice1, slice3, "different slice test")
	core.AssertFalse(t, result, "AssertSame result when different slices")
	core.AssertTrue(t, mock.HasErrors(), "has errors on slice failure")
	core.AssertFalse(t, mock.HasLogs(), "no logs on slice failure")
}

// Test AssertNotSame
func TestAssertNotSame(t *testing.T) {
	mock := &core.MockT{}

	// Test successful assertion with value types
	result := core.AssertNotSame(mock, 42, 43, "different value test")
	core.AssertTrue(t, result, "AssertNotSame result when different values")
	core.AssertFalse(t, mock.HasErrors(), "no errors on success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on success")

	lastLog, ok := mock.LastLog()
	core.AssertTrue(t, ok, "LastLog ok on success")
	core.AssertEqual(t, "different value test: different values or references", lastLog, "log message on success")

	mock.Reset()

	// Test successful assertion with reference types
	slice1 := core.S(1, 2, 3)
	slice2 := core.S(1, 2, 3) // same contents, different backing arrays
	result = core.AssertNotSame(mock, slice1, slice2, "different slice test")
	core.AssertTrue(t, result, "AssertNotSame result when different slices")
	core.AssertFalse(t, mock.HasErrors(), "no errors on slice success")
	core.AssertTrue(t, mock.HasLogs(), "has logs on slice success")

	mock.Reset()

	// Test failed assertion with value types
	result = core.AssertNotSame(mock, 42, 42, "same value test")
	core.AssertFalse(t, result, "AssertNotSame result when same values")
	core.AssertTrue(t, mock.HasErrors(), "has errors on failure")
	core.AssertFalse(t, mock.HasLogs(), "no logs on failure")

	assertErrorContains(t, mock, "expected different values or references, got same", "error message on failure")

	mock.Reset()

	// Test failed assertion with reference types
	slice3 := slice1 // same reference
	result = core.AssertNotSame(mock, slice1, slice3, "same slice test")
	core.AssertFalse(t, result, "AssertNotSame result when same slices")
	core.AssertTrue(t, mock.HasErrors(), "has errors on slice failure")
	core.AssertFalse(t, mock.HasLogs(), "no logs on slice failure")
}

// mockTRunTestCase tests the result MockT.Run reports for a test body,
// and what the body's calls left recorded in the MockT.
type mockTRunTestCase struct {
	fn            func(core.T)
	name          string
	expectErrors  []string
	expectLogs    []string
	expectFailed  bool
	expectSkipped bool
}

// newMockTRunTestCase declares a body under which Run passes without
// skipping, with the logs it leaves.
func newMockTRunTestCase(name string, fn func(core.T),
	expectLogs []string) mockTRunTestCase {
	return mockTRunTestCase{
		fn:            fn,
		name:          name,
		expectErrors:  nil,
		expectLogs:    expectLogs,
		expectFailed:  false,
		expectSkipped: false,
	}
}

// newMockTRunTestCaseFails declares a body under which Run fails without
// skipping, with the errors and logs it leaves.
func newMockTRunTestCaseFails(name string, fn func(core.T),
	expectErrors, expectLogs []string) mockTRunTestCase {
	return mockTRunTestCase{
		fn:            fn,
		name:          name,
		expectErrors:  expectErrors,
		expectLogs:    expectLogs,
		expectFailed:  true,
		expectSkipped: false,
	}
}

// newMockTRunTestCaseSkips declares a body that skips, under which Run
// passes, with the logs it leaves.
func newMockTRunTestCaseSkips(name string, fn func(core.T),
	expectLogs []string) mockTRunTestCase {
	return mockTRunTestCase{
		fn:            fn,
		name:          name,
		expectErrors:  nil,
		expectLogs:    expectLogs,
		expectFailed:  false,
		expectSkipped: true,
	}
}

// newMockTRunTestCaseFailsSkips declares a body that fails and then
// skips, under which Run fails, with the errors and logs it leaves.
func newMockTRunTestCaseFailsSkips(name string, fn func(core.T),
	expectErrors, expectLogs []string) mockTRunTestCase {
	return mockTRunTestCase{
		fn:            fn,
		name:          name,
		expectErrors:  expectErrors,
		expectLogs:    expectLogs,
		expectFailed:  true,
		expectSkipped: true,
	}
}

func (tc mockTRunTestCase) Name() string {
	return tc.name
}

func (tc mockTRunTestCase) Test(t *testing.T) {
	mock := &core.MockT{}
	ok := mock.Run(tc.name, tc.fn)

	core.AssertEqual(t, !tc.expectFailed, ok, "result")
	core.AssertEqual(t, tc.expectFailed, mock.Failed(), "Failed()")
	core.AssertEqual(t, tc.expectSkipped, mock.Skipped(), "Skipped()")
	core.AssertSliceEqual(t, tc.expectErrors, mock.Errors, "errors")
	core.AssertSliceEqual(t, tc.expectLogs, mock.Logs, "logs")
}

func TestMockTRun(t *testing.T) {
	core.RunTestCases(t, []mockTRunTestCase{
		newMockTRunTestCase("Log", func(mt core.T) { mt.Log("logged") },
			core.S("logged")),
		newMockTRunTestCase("Logf", func(mt core.T) { mt.Logf("logged %d", 42) },
			core.S("logged 42")),
		newMockTRunTestCaseFails("Fail", func(mt core.T) { mt.Fail(); mt.Log("continued") },
			nil, core.S("continued")),
		newMockTRunTestCaseFails("FailNow", func(mt core.T) { mt.FailNow(); mt.Log("continued") },
			nil, nil),
		newMockTRunTestCaseFails("Error", func(mt core.T) { mt.Error("error"); mt.Log("continued") },
			core.S("error"), core.S("continued")),
		newMockTRunTestCaseFails("Errorf", func(mt core.T) { mt.Errorf("error %d", 42); mt.Log("continued") },
			core.S("error 42"), core.S("continued")),
		newMockTRunTestCaseFails("Fatal", func(mt core.T) { mt.Fatal("fatal"); mt.Log("continued") },
			core.S("fatal"), nil),
		newMockTRunTestCaseFails("Fatalf", func(mt core.T) { mt.Fatalf("fatal %d", 42); mt.Log("continued") },
			core.S("fatal 42"), nil),
		newMockTRunTestCaseSkips("SkipNow", func(mt core.T) { mt.SkipNow(); mt.Log("continued") },
			nil),
		newMockTRunTestCaseSkips("Skip", func(mt core.T) { mt.Skip("skip"); mt.Log("continued") },
			core.S("skip")),
		newMockTRunTestCaseSkips("Skipf", func(mt core.T) { mt.Skipf("skip %d", 42); mt.Log("continued") },
			core.S("skip 42")),
		newMockTRunTestCaseFailsSkips("Fail then SkipNow",
			func(mt core.T) { mt.Fail(); mt.SkipNow(); mt.Log("continued") },
			nil, nil),
	})
}

// mockTRunCarryOverTestCase states that Failed and Skipped carry over
// from one Run to the next until Reset: after a first body sets them, a
// passing body leaves them as they were, and Run reports the result
// they give.
type mockTRunCarryOverTestCase struct {
	first       func(core.T)
	name        string
	wantFailed  bool
	wantSkipped bool
}

// newMockTRunCarryOverTestCaseFails declares a first body that fails
// the test without skipping it.
func newMockTRunCarryOverTestCaseFails(name string,
	first func(core.T)) mockTRunCarryOverTestCase {
	return mockTRunCarryOverTestCase{
		first:       first,
		name:        name,
		wantFailed:  true,
		wantSkipped: false,
	}
}

// newMockTRunCarryOverTestCaseSkips declares a first body that skips
// the test without failing it.
func newMockTRunCarryOverTestCaseSkips(name string,
	first func(core.T)) mockTRunCarryOverTestCase {
	return mockTRunCarryOverTestCase{
		first:       first,
		name:        name,
		wantFailed:  false,
		wantSkipped: true,
	}
}

func (tc mockTRunCarryOverTestCase) Name() string {
	return tc.name
}

func (tc mockTRunCarryOverTestCase) Test(t *testing.T) {
	mock := &core.MockT{}
	mock.Run("first", tc.first)
	core.AssertMustEqual(t, tc.wantFailed, mock.Failed(), "failed before")
	core.AssertMustEqual(t, tc.wantSkipped, mock.Skipped(), "skipped before")

	ok := mock.Run("passes", func(mt core.T) { mt.Log("passed") })
	core.AssertEqual(t, !tc.wantFailed, ok, "passed")
	core.AssertEqual(t, tc.wantFailed, mock.Failed(), "failed")
	core.AssertEqual(t, tc.wantSkipped, mock.Skipped(), "skipped")
}

func TestMockTRunCarriesOver(t *testing.T) {
	core.RunTestCases(t, []mockTRunCarryOverTestCase{
		newMockTRunCarryOverTestCaseFails("after FailNow",
			func(mt core.T) { mt.FailNow() }),
		newMockTRunCarryOverTestCaseSkips("after SkipNow",
			func(mt core.T) { mt.SkipNow() }),
	})
}

func TestMockTRunNilChecks(t *testing.T) {
	// Test nil MockT
	var mock *core.MockT
	ok := mock.Run("nil test", func(mt core.T) {
		mt.Log("should not run")
	})
	core.AssertFalse(t, ok, "nil MockT should return false")

	// Test nil function
	mock = &core.MockT{}
	ok = mock.Run("nil func test", nil)
	core.AssertFalse(t, ok, "nil function should return false")
}

func TestMockTRunPanicPropagation(t *testing.T) {
	mock := &core.MockT{}

	// Test that panics other than MockT's own sentinels are propagated
	core.AssertPanic(t, func() {
		mock.Run("panic test", func(_ core.T) {
			panic("custom panic")
		})
	}, "custom panic", "propagated")
}

// isMockTAbortTestCase states which recovered values IsMockTAbort takes
// for MockT stopping a test: the one its stopping methods raise, wrapped
// or not, and nothing else, an error of the same text included.
type isMockTAbortTestCase struct {
	recovered any
	name      string
	want      bool
}

// newIsMockTAbortTestCase declares a value MockT raises to stop a test.
func newIsMockTAbortTestCase(name string,
	recovered any) isMockTAbortTestCase {
	return isMockTAbortTestCase{
		recovered: recovered,
		name:      name,
		want:      true,
	}
}

// newIsMockTAbortTestCaseOther declares a value that is not MockT
// stopping a test.
func newIsMockTAbortTestCaseOther(name string,
	recovered any) isMockTAbortTestCase {
	return isMockTAbortTestCase{
		recovered: recovered,
		name:      name,
		want:      false,
	}
}

func (tc isMockTAbortTestCase) Name() string {
	return tc.name
}

func (tc isMockTAbortTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.want, core.IsMockTAbort(tc.recovered), "aborted")
}

// recoverFrom calls fn and returns what it panicked with, nil if it
// returned.
func recoverFrom(fn func()) (recovered any) {
	defer func() { recovered = recover() }()
	fn()
	return nil
}

func isMockTAbortTestCases() []isMockTAbortTestCase {
	failNow := recoverFrom(func() { (&core.MockT{}).FailNow() })
	skipNow := recoverFrom(func() { (&core.MockT{}).SkipNow() })

	return []isMockTAbortTestCase{
		newIsMockTAbortTestCase("FailNow", failNow),
		newIsMockTAbortTestCase("SkipNow", skipNow),
		newIsMockTAbortTestCase("wrapped",
			core.Wrap(core.MustT[error](failNow), "wrapped")),
		newIsMockTAbortTestCase("recovered", core.AsRecovered(failNow)),
		newIsMockTAbortTestCase("joined",
			errors.Join(errors.New("another"), core.MustT[error](failNow))),
		newIsMockTAbortTestCaseOther("nil", nil),
		newIsMockTAbortTestCaseOther("another error", errors.New("another")),
		newIsMockTAbortTestCaseOther("same text", errors.New("MockT.FailNow")),
	}
}

func TestIsMockTAbort(t *testing.T) {
	core.RunTestCases(t, isMockTAbortTestCases())
}

// Test early abort pattern: if !Assert() { FailNow() }
func TestEarlyAbortPattern(t *testing.T) {
	t.Run("successful assertion continues", testEarlyAbortSuccess)
	t.Run("failed assertion aborts", testEarlyAbortFailure)
	t.Run("multiple assertions with early abort", testEarlyAbortMultiple)
	t.Run("mixed patterns", testEarlyAbortMixed)
}

func testEarlyAbortSuccess(t *testing.T) {
	mock := &core.MockT{}

	// Test successful assertion that continues execution
	ok := mock.Run("early abort success", func(mt core.T) {
		if !core.AssertEqual(mt, 42, 42, "equal values") {
			mt.FailNow()
		}
		// This should execute
		mt.Log("execution continues after successful assertion")
	})

	core.AssertTrue(t, ok, "Test should pass when assertion succeeds")
	core.AssertFalse(t, mock.Failed(), "Should not be marked as failed")
	core.AssertEqual(t, 2, mock.NumLogs(), "Should have 2 log messages")
	// First log from AssertEqual success, second from explicit Log
	core.AssertContains(t, mustMessageAt(t, mock.LogAt, 0, "first log"), "equal values: 42",
		"First log from assertion")
}

func testEarlyAbortFailure(t *testing.T) {
	mock := &core.MockT{}

	// Test failed assertion that aborts execution
	ok := mock.Run("early abort failure", func(mt core.T) {
		if !core.AssertEqual(mt, 42, 24, "different values") {
			mt.FailNow()
		}
		// This should NOT execute
		mt.Log("this should not be reached")
	})

	core.AssertFalse(t, ok, "Test should fail when assertion fails and FailNow is called")
	core.AssertTrue(t, mock.Failed(), "Should be marked as failed")
	core.AssertEqual(t, 1, mock.NumErrors(), "Should have error from failed assertion")
	core.AssertEqual(t, 0, mock.NumLogs(), "Should have no logs after early abort")
	core.AssertContains(t, mustMessageAt(t, mock.ErrorAt, 0, "error"), "expected 42, got 24",
		"Error from failed assertion")
}

func testEarlyAbortMultiple(t *testing.T) {
	mock := &core.MockT{}

	// Test multiple assertions with early abort pattern
	ok := mock.Run("multiple early abort", func(mt core.T) {
		if !core.AssertEqual(mt, "hello", "hello", "first check") {
			mt.FailNow()
		}
		if !core.AssertTrue(mt, true, "second check") {
			mt.FailNow()
		}
		if !core.AssertEqual(mt, 1, 2, "third check") {
			mt.FailNow() // This should abort
		}
		// This should NOT execute
		mt.Log("unreachable code")
	})

	core.AssertFalse(t, ok, "Test should fail on third assertion")
	core.AssertTrue(t, mock.Failed(), "Should be marked as failed")
	core.AssertEqual(t, 1, mock.NumErrors(), "Should have one error from failed assertion")
	core.AssertEqual(t, 2, mock.NumLogs(), "Should have logs from first two successful assertions")
	core.AssertContains(t, mustMessageAt(t, mock.ErrorAt, 0, "error"), "expected 1, got 2",
		"Error from third assertion")
}

func testEarlyAbortMixed(t *testing.T) {
	mock := &core.MockT{}

	// Test mixing early abort pattern with regular assertions
	ok := mock.Run("mixed patterns", func(mt core.T) {
		// Regular assertion (doesn't abort on failure)
		core.AssertEqual(mt, 1, 2, "regular assertion")

		// Early abort pattern (aborts on failure)
		if !core.AssertEqual(mt, "test", "test", "critical assertion") {
			mt.FailNow()
		}

		// This should execute because the critical assertion passed
		mt.Log("continuing after critical assertion")

		// Another early abort that fails
		if !core.AssertFalse(mt, true, "should be false") {
			mt.FailNow() // This should abort
		}

		// This should NOT execute
		mt.Log("unreachable after failed critical assertion")
	})

	core.AssertFalse(t, ok, "Test should fail due to early abort")
	core.AssertTrue(t, mock.Failed(), "Should be marked as failed")
	core.AssertEqual(t, 2, mock.NumErrors(), "Should have two errors")
	core.AssertEqual(t, 2, mock.NumLogs(), "Should have two logs before abort")

	// Check the error messages
	core.AssertContains(t, mustMessageAt(t, mock.ErrorAt, 0, "first error"), "expected 1, got 2",
		"First error from regular assertion")
	core.AssertContains(t, mustMessageAt(t, mock.ErrorAt, 1, "second error"), "expected false, got true",
		"Second error from early abort assertion")

	// Check log messages
	core.AssertContains(t, mustMessageAt(t, mock.LogAt, 0, "first log"), "critical assertion: test",
		"Log from successful critical assertion")
	core.AssertContains(t, mustMessageAt(t, mock.LogAt, 1, "second log"), "continuing after critical assertion",
		"Log showing execution continued")
}

// Test AssertMustFoo() functions that call FailNow() on failure
func TestAssertMustFunctions(t *testing.T) {
	t.Run("AssertMustEqual", testAssertMustEqual)
	t.Run("AssertMustNotEqual", testAssertMustNotEqual)
	t.Run("AssertMustSliceEqual", testAssertMustSliceEqual)
	t.Run("AssertMustDeepEqual", testAssertMustDeepEqual)
	t.Run("AssertMustNotDeepEqual", testAssertMustNotDeepEqual)
	t.Run("AssertMustContains", testAssertMustContains)
	t.Run("AssertMustNotContain", testAssertMustNotContain)
	t.Run("AssertMustError", testAssertMustError)
	t.Run("AssertMustNoError", testAssertMustNoError)
	t.Run("AssertMustPanic", testAssertMustPanic)
	t.Run("AssertMustNoPanic", testAssertMustNoPanic)
	t.Run("AssertMustTrue", testAssertMustTrue)
	t.Run("AssertMustFalse", testAssertMustFalse)
	t.Run("AssertMustErrorIs", testAssertMustErrorIs)
	t.Run("AssertMustNotErrorIs", testAssertMustNotErrorIs)
	t.Run("AssertMustErrorIsFn", testAssertMustErrorIsFn)
	t.Run("AssertMustErrorAs", testAssertMustErrorAs)
	t.Run("AssertMustTypeIs", testAssertMustTypeIs)
	t.Run("AssertMustNil", testAssertMustNil)
	t.Run("AssertMustNotNil", testAssertMustNotNil)
	t.Run("AssertMustSame", testAssertMustSame)
	t.Run("AssertMustNotSame", testAssertMustNotSame)
}

func testAssertMustEqual(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustEqual(mt, 42, 42, "equal values")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustEqual(mt, 42, 24, "different values")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNotEqual(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNotEqual(mt, 42, 24, "different values")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNotEqual(mt, 42, 42, "equal values")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustSliceEqual(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustSliceEqual(mt, core.S(1, 2, 3), core.S(1, 2, 3), "equal slices")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustSliceEqual(mt, core.S(1, 2, 3), core.S(1, 2, 4), "different slices")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustDeepEqual(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustDeepEqual(mt, core.S(core.S(1)), core.S(core.S(1)), "equal values")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustDeepEqual(mt, core.S(core.S(1)), core.S(core.S(2)), "different values")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNotDeepEqual(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNotDeepEqual(mt, core.S(core.S(1)), core.S(core.S(2)), "different values")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNotDeepEqual(mt, core.S(core.S(1)), core.S(core.S(1)), "equal values")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustContains(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustContains(mt, "hello world", "world", "contains substring")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustContains(mt, "hello world", "xyz", "missing substring")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNotContain(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNotContain(mt, "hello world", "xyz", "no substring")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNotContain(mt, "hello world", "world", "has substring")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustError(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustError(mt, errors.New("test error"), "expects error")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustError(mt, nil, "expects error but got nil")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNoError(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNoError(mt, nil, "expects no error")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNoError(mt, errors.New("unexpected error"), "expects no error")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustPanic(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustPanic(mt, func() { panic("test panic") }, nil, "expects panic")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustPanic(mt, func() {}, nil, "expects panic but got none")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNoPanic(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNoPanic(mt, func() {}, "expects no panic")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNoPanic(mt, func() { panic("unexpected") }, "expects no panic")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustTrue(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustTrue(mt, true, "expects true")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustTrue(mt, false, "expects true but got false")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustFalse(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustFalse(mt, false, "expects false")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustFalse(mt, true, "expects false but got true")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustErrorIs(t *testing.T) {
	mock := &core.MockT{}
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(baseErr, errors.New("wrapped"))

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustErrorIs(mt, wrappedErr, baseErr, "error should match")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	otherErr := errors.New("other error")
	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustErrorIs(mt, wrappedErr, otherErr, "error should not match")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNotErrorIs(t *testing.T) {
	mock := &core.MockT{}
	baseErr := errors.New("base error")
	wrappedErr := errors.Join(baseErr, errors.New("wrapped"))

	otherErr := errors.New("other error")
	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNotErrorIs(mt, wrappedErr, otherErr, "error should not match")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNotErrorIs(mt, wrappedErr, baseErr, "error should match")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustErrorIsFn(t *testing.T) {
	mock := &core.MockT{}
	baseErr := errors.New("base error")
	wrappedErr := core.Wrap(baseErr, "wrapped")
	isBase := func(err error) bool { return err == baseErr }

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustErrorIsFn(mt, wrappedErr, isBase, "error should match")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustErrorIsFn(mt, errors.New("other error"), isBase, "error should not match")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustErrorAs(t *testing.T) {
	mock := &core.MockT{}
	wrapped := core.Wrap(errors.New("base error"), "note")

	ok := mock.Run("success", func(mt core.T) {
		out := core.AssertMustErrorAs[*core.WrappedError](mt, wrapped, "type should match")
		core.AssertSame(mt, wrapped, out, "matched value")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		_ = core.AssertMustErrorAs[*core.WrappedError](mt, errors.New("plain"), "type should not match")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustTypeIs(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		var val any = "hello"
		result := core.AssertMustTypeIs[string](mt, val, "type should match")
		core.AssertEqual(mt, "hello", result, "should return cast value")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		var val any = 42
		core.AssertMustTypeIs[string](mt, val, "type should not match")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNil(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNil(mt, nil, "expects nil")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNil(mt, "not nil", "expects nil but got value")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNotNil(t *testing.T) {
	mock := &core.MockT{}

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNotNil(mt, "not nil", "expects not nil")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNotNil(mt, nil, "expects not nil but got nil")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustSame(t *testing.T) {
	mock := &core.MockT{}

	slice1 := []int{1, 2, 3}
	slice2 := slice1 // Same reference

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustSame(mt, slice1, slice2, "same slice reference")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	slice3 := []int{1, 2, 3} // Different reference
	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustSame(mt, slice1, slice3, "different slice references")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}

func testAssertMustNotSame(t *testing.T) {
	mock := &core.MockT{}

	slice1 := []int{1, 2, 3}
	slice3 := []int{1, 2, 3} // Different reference

	ok := mock.Run("success", func(mt core.T) {
		core.AssertMustNotSame(mt, slice1, slice3, "different slice references")
		mt.Log(core.MustContinuationLogForTest)
	})
	core.AssertMustContinuedForTest(t, mock, ok)

	mock.Reset()

	slice2 := slice1 // Same reference
	ok = mock.Run("failure", func(mt core.T) {
		core.AssertMustNotSame(mt, slice1, slice2, "same slice reference")
		mt.Log("should not reach here")
	})
	core.AssertMustAbortedForTest(t, mock, ok)
}
