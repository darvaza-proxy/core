package core_test

import (
	"errors"
	"fmt"
	"testing"

	"darvaza.org/core"
)

var (
	_ core.TestCase = asTestCase[string]{}
	_ core.TestCase = asFnTestCase{}
	_ core.TestCase = sliceAsTestCase{}
	_ core.TestCase = sliceAsFnTestCase{}
	_ core.TestCase = asErrorTestCase{}
	_ core.TestCase = asErrorsTestCase{}
)

// asTestCase states what As does for one target type V: ok says whether
// the input already held a V, and the value returned is that V when it
// did, the zero value when it did not.
type asTestCase[V any] struct {
	input any
	want  V
	name  string

	wantOK bool
}

// newAsTestCase declares a row whose input holds a V, which comes back
// as want with true.
func newAsTestCase[V any](name string, input any, want V) asTestCase[V] {
	return asTestCase[V]{
		input:  input,
		want:   want,
		name:   name,
		wantOK: true,
	}
}

// newAsTestCaseFails declares a row whose input holds no V, answered by
// the zero value and false.
func newAsTestCaseFails[V any](name string, input any) asTestCase[V] {
	var zero V

	return asTestCase[V]{
		input:  input,
		want:   zero,
		name:   name,
		wantOK: false,
	}
}

func (tc asTestCase[V]) Name() string {
	return tc.name
}

func (tc asTestCase[V]) Test(t *testing.T) {
	got, ok := core.As[any, V](tc.input)
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	core.AssertEqual(t, tc.want, got, "value")
}

// asTestCases compounds the rows of every target type into one table,
// grouped by target: string, then int, then error. A zero value of the
// target and a typed nil pointer come back with true, which is what
// makes the ok column say something the value alone does not.
func asTestCases() []core.TestCase {
	testErr := errors.New("test error")
	var typedNil *typedNilError

	return []core.TestCase{
		newAsTestCase("string to string", testHello, testHello),
		newAsTestCase("empty string to string", "", ""),
		newAsTestCaseFails[string]("int to string fails", 42),
		newAsTestCaseFails[string]("nil to string", nil),
		newAsTestCase("int to int", 42, 42),
		newAsTestCase("zero to int", 0, 0),
		newAsTestCaseFails[int]("string to int fails", testHello),
		newAsTestCaseFails[int]("nil to int", nil),
		newAsTestCase[error]("error to error", testErr, testErr),
		newAsTestCase[error]("typed nil to error", typedNil, typedNil),
		newAsTestCaseFails[error]("string to error fails", testHello),
		newAsTestCaseFails[error]("nil to error", nil),
	}
}

func TestAs(t *testing.T) {
	core.RunTestCases(t, asTestCases())
}

// asFnTestCase tests AsFn: the function's answer comes back as given,
// and a nil function answers the zero value and false.
type asFnTestCase struct {
	fn     func(any) (string, bool)
	input  any
	want   string
	name   string
	wantOK bool
}

func newAsFnTestCase(name string, fn func(any) (string, bool), input any, want string, wantOK bool) asFnTestCase {
	return asFnTestCase{
		fn:     fn,
		input:  input,
		want:   want,
		name:   name,
		wantOK: wantOK,
	}
}

func (tc asFnTestCase) Name() string {
	return tc.name
}

func (tc asFnTestCase) Test(t *testing.T) {
	got, ok := core.AsFn(tc.fn, tc.input)
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	core.AssertEqual(t, tc.want, got, "value")
}

func TestAsFn(t *testing.T) {
	intToString := func(v any) (string, bool) {
		if i, ok := v.(int); ok {
			return fmt.Sprintf("%d", i), true
		}
		return "", false
	}
	ignored := func(any) (string, bool) { return "ignored", false }

	testCases := core.S(
		newAsFnTestCase("converted", intToString, 42, "42", true),
		newAsFnTestCase("wrong type", intToString, "not an int", "", false),
		newAsFnTestCase("function returning false", ignored, 42, "ignored", false),
		newAsFnTestCase("nil function", nil, 42, "", false),
	)

	core.RunTestCases(t, testCases)
}

// sliceAsTestCase tests SliceAs: the elements of the target type, in
// order, and nil when there are none.
type sliceAsTestCase struct {
	name  string
	input []any
	want  []string
}

func newSliceAsTestCase(name string, input []any, want []string) sliceAsTestCase {
	return sliceAsTestCase{
		name:  name,
		input: input,
		want:  want,
	}
}

func (tc sliceAsTestCase) Name() string {
	return tc.name
}

func (tc sliceAsTestCase) Test(t *testing.T) {
	got := core.SliceAs[any, string](tc.input)
	core.AssertSliceEqual(t, tc.want, got, "slice")
}

func TestSliceAs(t *testing.T) {
	testCases := core.S(
		newSliceAsTestCase("all strings", core.S[any]("a", "b", "c"), core.S("a", "b", "c")),
		newSliceAsTestCase("mixed types", core.S[any](testHello, 42, "world", 3.14, "!"),
			core.S(testHello, "world", "!")),
		newSliceAsTestCase("with nil values", core.S[any](testHello, nil, "world"), core.S(testHello, "world")),
		newSliceAsTestCase("no strings", core.S[any](1, 2, 3, 4.5, true), nil),
		newSliceAsTestCase("empty slice", core.S[any](), nil),
		newSliceAsTestCase("nil slice", nil, nil),
	)

	core.RunTestCases(t, testCases)
}

// sliceAsFnTestCase tests SliceAsFn: the elements the function accepts,
// as it returns them, and nil when there are none or no function.
type sliceAsFnTestCase struct {
	name  string
	fn    func(any) (string, bool)
	input []any
	want  []string
}

func newSliceAsFnTestCase(name string, fn func(any) (string, bool), input []any, want []string) sliceAsFnTestCase {
	return sliceAsFnTestCase{
		name:  name,
		fn:    fn,
		input: input,
		want:  want,
	}
}

func (tc sliceAsFnTestCase) Name() string {
	return tc.name
}

func (tc sliceAsFnTestCase) Test(t *testing.T) {
	got := core.SliceAsFn(tc.fn, tc.input)
	core.AssertSliceEqual(t, tc.want, got, "slice")
}

func TestSliceAsFn(t *testing.T) {
	prefixString := func(v any) (string, bool) {
		if s, ok := v.(string); ok {
			return "prefix:" + s, true
		}
		return "", false
	}

	testCases := core.S(
		newSliceAsFnTestCase("converted", prefixString, core.S[any]("a", 1, "b", 2, "c"),
			core.S("prefix:a", "prefix:b", "prefix:c")),
		newSliceAsFnTestCase("all filtered out", prefixString, core.S[any](1, 2, 3), nil),
		newSliceAsFnTestCase("empty slice", prefixString, core.S[any](), nil),
		newSliceAsFnTestCase("nil slice", prefixString, nil, nil),
		newSliceAsFnTestCase("nil function", nil, core.S[any]("a", "b", "c"), nil),
	)

	core.RunTestCases(t, testCases)
}

func TestSliceAsFnPanic(t *testing.T) {
	panicFn := func(_ any) (string, bool) {
		panic("test panic")
	}

	core.AssertPanic(t, func() {
		_ = core.SliceAsFn(panicFn, core.S[any]("will panic"))
	}, "test panic", "panic")
}

// errorWithAsError answers AsError with the error it holds, nil
// included.
type errorWithAsError struct {
	err error
}

func (e errorWithAsError) AsError() error {
	return e.err
}

// errorWithOK is an error that reports through OK whether it is one.
type errorWithOK struct {
	msg string
	ok  bool
}

func (e errorWithOK) Error() string {
	return e.msg
}

func (e errorWithOK) OK() bool {
	return e.ok
}

// asErrorTestCase tests AsError: the error a value stands for, or nil
// when it stands for none.
type asErrorTestCase struct {
	input any
	want  error
	name  string
}

func newAsErrorTestCase(name string, input any, want error) asErrorTestCase {
	return asErrorTestCase{
		input: input,
		want:  want,
		name:  name,
	}
}

func (tc asErrorTestCase) Name() string {
	return tc.name
}

func (tc asErrorTestCase) Test(t *testing.T) {
	got := core.AsError(tc.input)
	core.AssertEqual(t, tc.want, got, "error")
}

func asErrorTestCases() []asErrorTestCase {
	standard := errors.New("standard error")
	custom := errors.New("custom error")
	notOK := errorWithOK{msg: "not ok error", ok: false}
	var typedNil *typedNilError

	return core.S(
		newAsErrorTestCase("error", standard, standard),
		newAsErrorTestCase("AsError returning error", errorWithAsError{err: custom}, custom),
		newAsErrorTestCase("AsError returning nil", errorWithAsError{}, nil),
		newAsErrorTestCase("OK returning false", notOK, notOK),
		newAsErrorTestCase("OK returning true", errorWithOK{msg: "ok error", ok: true}, nil),
		newAsErrorTestCase("typed-nil error", error(typedNil), nil),
		newAsErrorTestCase("string", "not an error", nil),
		newAsErrorTestCase("int", 42, nil),
		newAsErrorTestCase("nil", nil, nil),
	)
}

func TestAsError(t *testing.T) {
	core.RunTestCases(t, asErrorTestCases())
}

// asErrorsTestCase tests AsErrors: the errors the elements stand for,
// in order, and nil when there are none.
type asErrorsTestCase struct {
	name  string
	input []any
	want  []error
}

func newAsErrorsTestCase(name string, input []any, want []error) asErrorsTestCase {
	return asErrorsTestCase{
		name:  name,
		input: input,
		want:  want,
	}
}

func (tc asErrorsTestCase) Name() string {
	return tc.name
}

func (tc asErrorsTestCase) Test(t *testing.T) {
	got := core.AsErrors(tc.input)
	core.AssertSliceEqual(t, tc.want, got, "errors")
}

func asErrorsTestCases() []asErrorsTestCase {
	errA, errB, errC := errors.New("a"), errors.New("b"), errors.New("c")
	err1, err2, err3 := errors.New("error1"), errors.New("error2"), errors.New("error3")
	fail := errorWithOK{msg: "fail", ok: false}
	fail2 := errorWithOK{msg: "fail2", ok: false}
	pass := errorWithOK{msg: "pass", ok: true}

	return core.S(
		newAsErrorsTestCase("all errors", core.S[any](errA, errB, errC), core.S(errA, errB, errC)),
		newAsErrorsTestCase("mixed values", core.S[any](
			err1,
			"not an error",
			err2,
			42,
			errorWithAsError{err: err3},
			nil,
		), core.S(err1, err2, err3)),
		newAsErrorsTestCase("OK interface", core.S[any](fail, pass, fail2), core.S[error](fail, fail2)),
		newAsErrorsTestCase("no errors", core.S[any]("a", 1, true, nil), nil),
		newAsErrorsTestCase("empty slice", core.S[any](), nil),
		newAsErrorsTestCase("nil slice", nil, nil),
	)
}

func TestAsErrors(t *testing.T) {
	core.RunTestCases(t, asErrorsTestCases())
}

// TestAsWithConcreteTypes states As over source types the typed tables,
// which take any, cannot: no conversion between numeric types, and a
// pointer coming back as itself.
func TestAsWithConcreteTypes(t *testing.T) {
	i := 42

	v, ok := core.As[int, int64](i)
	core.AssertFalse(t, ok, "int to int64 ok")
	core.AssertEqual(t, 0, v, "int to int64 value")

	pi := &i
	p, ok := core.As[*int, *int](pi)
	core.AssertTrue(t, ok, "*int to *int ok")
	core.AssertSame(t, pi, p, "*int to *int value")
}

// Benchmark tests
func BenchmarkAs(b *testing.B) {
	input := "test string"
	for b.Loop() {
		_, _ = core.As[any, string](input)
	}
}

func BenchmarkAsFn(b *testing.B) {
	fn := func(v any) (string, bool) {
		s, ok := v.(string)
		return s, ok
	}
	var input any = "test string"

	for b.Loop() {
		_, _ = core.AsFn(fn, input)
	}
}

func BenchmarkSliceAs(b *testing.B) {
	input := core.S[any]("a", 1, "b", 2, "c", 3, "d", 4, "e", 5)

	for b.Loop() {
		_ = core.SliceAs[any, string](input)
	}
}

func BenchmarkAsError(b *testing.B) {
	err := errors.New("test error")

	for b.Loop() {
		_ = core.AsError(err)
	}
}

// typedNilError is an error whose nil pointer still satisfies error,
// for the typed-nil rows.
type typedNilError struct{}

func (*typedNilError) Error() string { return "typed-nil" }
