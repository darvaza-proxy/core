package core

import (
	"errors"
	"fmt"
	"testing"
)

var _ TestCase = asTestCase[string]{}
var _ TestCase = asFnTestCase{}
var _ TestCase = sliceAsTestCase{}
var _ TestCase = sliceAsFnTestCase{}
var _ TestCase = asErrorTestCase{}
var _ TestCase = asErrorsTestCase{}

const testHello = "hello"

// asTestCase tests As over one target type: a value of the type comes
// back with true, anything else as the zero value with false.
type asTestCase[V any] struct {
	want   V
	input  any
	name   string
	wantOK bool
}

func newAsTestCase[V any](name string, input any, want V, wantOK bool) TestCase {
	return asTestCase[V]{
		want:   want,
		input:  input,
		name:   name,
		wantOK: wantOK,
	}
}

func (tc asTestCase[V]) Name() string {
	return tc.name
}

func (tc asTestCase[V]) Test(t *testing.T) {
	t.Helper()

	got, ok := As[any, V](tc.input)
	AssertEqual(t, tc.wantOK, ok, "ok")
	AssertEqual(t, tc.want, got, "value")
}

func asTestCases() []TestCase {
	err := errors.New("test error")
	var typedNil *typedNilError

	return S(
		newAsTestCase("string to string", testHello, testHello, true),
		newAsTestCase("empty string to string", "", "", true),
		newAsTestCase("int to string", 42, "", false),
		newAsTestCase("nil to string", nil, "", false),
		newAsTestCase("int to int", 42, 42, true),
		newAsTestCase("zero to int", 0, 0, true),
		newAsTestCase("string to int", testHello, 0, false),
		newAsTestCase("nil to int", nil, 0, false),
		newAsTestCase("error to error", err, err, true),
		newAsTestCase[error]("typed nil to error", typedNil, typedNil, true),
		newAsTestCase[error]("string to error", testHello, nil, false),
		newAsTestCase[error]("nil to error", nil, nil, false),
	)
}

func TestAs(t *testing.T) {
	RunTestCases(t, asTestCases())
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
	t.Helper()

	got, ok := AsFn(tc.fn, tc.input)
	AssertEqual(t, tc.wantOK, ok, "ok")
	AssertEqual(t, tc.want, got, "value")
}

func TestAsFn(t *testing.T) {
	intToString := func(v any) (string, bool) {
		if i, ok := v.(int); ok {
			return fmt.Sprintf("%d", i), true
		}
		return "", false
	}
	ignored := func(any) (string, bool) { return "ignored", false }

	testCases := S(
		newAsFnTestCase("converted", intToString, 42, "42", true),
		newAsFnTestCase("wrong type", intToString, "not an int", "", false),
		newAsFnTestCase("function returning false", ignored, 42, "ignored", false),
		newAsFnTestCase("nil function", nil, 42, "", false),
	)

	RunTestCases(t, testCases)
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
	t.Helper()

	got := SliceAs[any, string](tc.input)
	AssertSliceEqual(t, tc.want, got, "slice")
}

func TestSliceAs(t *testing.T) {
	testCases := S(
		newSliceAsTestCase("all strings", S[any]("a", "b", "c"), S("a", "b", "c")),
		newSliceAsTestCase("mixed types", S[any](testHello, 42, "world", 3.14, "!"),
			S(testHello, "world", "!")),
		newSliceAsTestCase("with nil values", S[any](testHello, nil, "world"), S(testHello, "world")),
		newSliceAsTestCase("no strings", S[any](1, 2, 3, 4.5, true), nil),
		newSliceAsTestCase("empty slice", S[any](), nil),
		newSliceAsTestCase("nil slice", nil, nil),
	)

	RunTestCases(t, testCases)
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
	t.Helper()

	got := SliceAsFn(tc.fn, tc.input)
	AssertSliceEqual(t, tc.want, got, "slice")
}

func TestSliceAsFn(t *testing.T) {
	prefixString := func(v any) (string, bool) {
		if s, ok := v.(string); ok {
			return "prefix:" + s, true
		}
		return "", false
	}

	testCases := S(
		newSliceAsFnTestCase("converted", prefixString, S[any]("a", 1, "b", 2, "c"),
			S("prefix:a", "prefix:b", "prefix:c")),
		newSliceAsFnTestCase("all filtered out", prefixString, S[any](1, 2, 3), nil),
		newSliceAsFnTestCase("empty slice", prefixString, S[any](), nil),
		newSliceAsFnTestCase("nil slice", prefixString, nil, nil),
		newSliceAsFnTestCase("nil function", nil, S[any]("a", "b", "c"), nil),
	)

	RunTestCases(t, testCases)
}

func TestSliceAsFnPanic(t *testing.T) {
	panicFn := func(_ any) (string, bool) {
		panic("test panic")
	}

	AssertPanic(t, func() {
		_ = SliceAsFn(panicFn, S[any]("will panic"))
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
	t.Helper()

	got := AsError(tc.input)
	AssertEqual(t, tc.want, got, "error")
}

func asErrorTestCases() []asErrorTestCase {
	standard := errors.New("standard error")
	custom := errors.New("custom error")
	notOK := errorWithOK{msg: "not ok error", ok: false}
	var typedNil *typedNilError

	return S(
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
	RunTestCases(t, asErrorTestCases())
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
	t.Helper()

	got := AsErrors(tc.input)
	AssertSliceEqual(t, tc.want, got, "errors")
}

func asErrorsTestCases() []asErrorsTestCase {
	errA, errB, errC := errors.New("a"), errors.New("b"), errors.New("c")
	err1, err2, err3 := errors.New("error1"), errors.New("error2"), errors.New("error3")
	fail, fail2 := errorWithOK{msg: "fail"}, errorWithOK{msg: "fail2"}
	pass := errorWithOK{msg: "pass", ok: true}

	return S(
		newAsErrorsTestCase("all errors", S[any](errA, errB, errC), S(errA, errB, errC)),
		newAsErrorsTestCase("mixed values", S[any](
			err1,
			"not an error",
			err2,
			42,
			errorWithAsError{err: err3},
			nil,
		), S(err1, err2, err3)),
		newAsErrorsTestCase("OK interface", S[any](fail, pass, fail2), S[error](fail, fail2)),
		newAsErrorsTestCase("no errors", S[any]("a", 1, true, nil), nil),
		newAsErrorsTestCase("empty slice", S[any](), nil),
		newAsErrorsTestCase("nil slice", nil, nil),
	)
}

func TestAsErrors(t *testing.T) {
	RunTestCases(t, asErrorsTestCases())
}

// TestAsWithConcreteTypes states As over source types the typed tables,
// which take any, cannot: no conversion between numeric types, and a
// pointer coming back as itself.
func TestAsWithConcreteTypes(t *testing.T) {
	i := 42

	v, ok := As[int, int64](i)
	AssertFalse(t, ok, "int to int64 ok")
	AssertEqual(t, 0, v, "int to int64 value")

	pi := &i
	p, ok := As[*int, *int](pi)
	AssertTrue(t, ok, "*int to *int ok")
	AssertSame(t, pi, p, "*int to *int value")
}

// Benchmark tests
func BenchmarkAs(b *testing.B) {
	input := "test string"
	for b.Loop() {
		_, _ = As[any, string](input)
	}
}

func BenchmarkAsFn(b *testing.B) {
	fn := func(v any) (string, bool) {
		s, ok := v.(string)
		return s, ok
	}
	var input any = "test string"

	for b.Loop() {
		_, _ = AsFn(fn, input)
	}
}

func BenchmarkSliceAs(b *testing.B) {
	input := S[any]("a", 1, "b", 2, "c", 3, "d", 4, "e", 5)

	for b.Loop() {
		_ = SliceAs[any, string](input)
	}
}

func BenchmarkAsError(b *testing.B) {
	err := errors.New("test error")

	for b.Loop() {
		_ = AsError(err)
	}
}

// typedNilError is an error whose nil pointer still satisfies error,
// for the typed-nil rows.
type typedNilError struct{}

func (*typedNilError) Error() string { return "typed-nil" }
