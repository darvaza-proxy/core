package core

import (
	"errors"
	"fmt"
	"testing"
)

var _ TestCase = asTestCase{}
var _ TestCase = asFnTestCase{}
var _ TestCase = sliceAsTestCase{}
var _ TestCase = sliceAsFnTestCase{}
var _ TestCase = asErrorTestCase{}
var _ TestCase = asErrorsTestCase{}

const testHello = "hello"

// asTestCase tests As function
type asTestCase struct {
	// Interface fields - input/output test data
	input any
	want  any

	// String fields - test identification
	name string

	// Boolean fields (1 byte) - expected result flags
	wantOK bool
}

// newAsTestCase creates a new asTestCase
func newAsTestCase(name string, input, want any, wantOK bool) asTestCase {
	return asTestCase{
		name:   name,
		input:  input,
		want:   want,
		wantOK: wantOK,
	}
}

func (tc asTestCase) Name() string {
	return tc.name
}

func (tc asTestCase) Test(t *testing.T) {
	t.Helper()

	switch want := tc.want.(type) {
	case string:
		tc.testStringConversion(t, want)
	case int:
		tc.testIntConversion(t, want)
	case error:
		tc.testErrorConversion(t, want)
	default:
		tc.testDefaultConversion(t)
	}
}

func (tc asTestCase) testStringConversion(t *testing.T, want string) {
	t.Helper()
	got, ok := As[any, string](tc.input)
	if ok != tc.wantOK {
		t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
	}
	if got != want {
		t.Errorf("As() got = %v, want %v", got, want)
	}
}

func (tc asTestCase) testIntConversion(t *testing.T, want int) {
	t.Helper()
	got, ok := As[any, int](tc.input)
	if ok != tc.wantOK {
		t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
	}
	if got != want {
		t.Errorf("As() got = %v, want %v", got, want)
	}
}

func (tc asTestCase) testErrorConversion(t *testing.T, want error) {
	t.Helper()
	got, ok := As[any, error](tc.input)
	if ok != tc.wantOK {
		t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
	}
	if ok && got.Error() != want.Error() {
		t.Errorf("As() got = %v, want %v", got, want)
	}
}

func (tc asTestCase) testDefaultConversion(t *testing.T) {
	t.Helper()
	// Test cases where conversion should fail
	got, ok := As[any, string](tc.input)
	if ok != tc.wantOK {
		t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
	}
	if tc.wantOK && got != "" {
		t.Errorf("As() got = %v, want zero value", got)
	}
}

// asFnTestCase tests AsFn function
type asFnTestCase struct {
	name   string
	fn     func(any) (string, bool)
	input  any
	want   string
	wantOK bool
}

// newAsFnTestCase creates a new asFnTestCase
func newAsFnTestCase(name string, fn func(any) (string, bool), input any, want string, wantOK bool) asFnTestCase {
	return asFnTestCase{
		name:   name,
		fn:     fn,
		input:  input,
		want:   want,
		wantOK: wantOK,
	}
}

func (tc asFnTestCase) Name() string {
	return tc.name
}

func (tc asFnTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok := AsFn(tc.fn, tc.input)
	if ok != tc.wantOK {
		t.Errorf("AsFn() ok = %v, want %v", ok, tc.wantOK)
	}
	if got != tc.want {
		t.Errorf("AsFn() got = %v, want %v", got, tc.want)
	}
}

// sliceAsTestCase tests SliceAs function
type sliceAsTestCase struct {
	name  string
	input []any
	want  []string
}

// newSliceAsTestCase creates a new sliceAsTestCase
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
	if len(got) != len(tc.want) {
		t.Fatalf("SliceAs() len = %v, want %v", len(got), len(tc.want))
	}
	for i, v := range got {
		if v != tc.want[i] {
			t.Errorf("SliceAs()[%d] = %v, want %v", i, v, tc.want[i])
		}
	}
}

// sliceAsFnTestCase tests SliceAsFn function
type sliceAsFnTestCase struct {
	name  string
	fn    func(any) (string, bool)
	input []any
	want  []string
}

// newSliceAsFnTestCase creates a new sliceAsFnTestCase
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
	if len(got) != len(tc.want) {
		t.Fatalf("SliceAsFn() len = %v, want %v", len(got), len(tc.want))
	}
	for i, v := range got {
		if v != tc.want[i] {
			t.Errorf("SliceAsFn()[%d] = %v, want %v", i, v, tc.want[i])
		}
	}
}

// Custom types for testing AsError
type errorWithAsError struct {
	msg string
}

func (e errorWithAsError) AsError() error {
	if e.msg == "" {
		return nil
	}
	return errors.New(e.msg)
}

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

// asErrorTestCase tests AsError function
type asErrorTestCase struct {
	// Interface fields - input test data
	input any

	// String fields - test identification and expected message
	name    string
	wantMsg string

	// Boolean fields (1 byte) - expected result flags
	wantErr bool
}

// newAsErrorTestCase creates a new asErrorTestCase
func newAsErrorTestCase(name string, input any, wantMsg string, wantErr bool) asErrorTestCase {
	return asErrorTestCase{
		name:    name,
		input:   input,
		wantMsg: wantMsg,
		wantErr: wantErr,
	}
}

func (tc asErrorTestCase) Name() string {
	return tc.name
}

func (tc asErrorTestCase) Test(t *testing.T) {
	t.Helper()

	got := AsError(tc.input)
	if (got != nil) != tc.wantErr {
		t.Errorf("AsError() error = %v, wantErr %v", got, tc.wantErr)
		return
	}
	if got != nil && got.Error() != tc.wantMsg {
		t.Errorf("AsError() error message = %v, want %v", got.Error(), tc.wantMsg)
	}
}

// asErrorsTestCase tests AsErrors function
type asErrorsTestCase struct {
	name     string
	input    []any
	wantMsgs []string
	wantLen  int
}

// newAsErrorsTestCase creates a new asErrorsTestCase
func newAsErrorsTestCase(name string, input []any, wantMsgs []string, wantLen int) asErrorsTestCase {
	return asErrorsTestCase{
		name:     name,
		input:    input,
		wantMsgs: wantMsgs,
		wantLen:  wantLen,
	}
}

func (tc asErrorsTestCase) Name() string {
	return tc.name
}

func (tc asErrorsTestCase) Test(t *testing.T) {
	t.Helper()

	got := AsErrors(tc.input)
	if len(got) != tc.wantLen {
		t.Fatalf("AsErrors() len = %v, want %v", len(got), tc.wantLen)
	}
	for i, err := range got {
		if err.Error() != tc.wantMsgs[i] {
			t.Errorf("AsErrors()[%d] = %v, want %v", i, err.Error(), tc.wantMsgs[i])
		}
	}
}

func TestAs(t *testing.T) {
	testCases := []asTestCase{
		newAsTestCase("string to string", testHello, testHello, true),
		newAsTestCase("int to int", 42, 42, true),
		newAsTestCase("error to error", errors.New("test error"), errors.New("test error"), true),
		newAsTestCase("int to string fails", 42, "", false),
		newAsTestCase("nil to string", nil, "", false),
		newAsTestCase("nil to error", nil, error(nil), false),
	}

	RunTestCases(t, testCases)
}

func TestAsFn(t *testing.T) {
	// Custom conversion function
	intToString := func(v any) (string, bool) {
		if i, ok := v.(int); ok {
			return fmt.Sprintf("%d", i), true
		}
		return "", false
	}

	testCases := []asFnTestCase{
		newAsFnTestCase("with valid conversion function", intToString, 42, "42", true),
		newAsFnTestCase("with valid conversion function but wrong type", intToString, "not an int", "", false),
		newAsFnTestCase("with nil function", nil, 42, "", false),
		newAsFnTestCase("with function returning false", func(any) (string, bool) { return "ignored", false },
			42, "ignored", false),
	}

	RunTestCases(t, testCases)
}

func TestSliceAs(t *testing.T) {
	testCases := []sliceAsTestCase{
		newSliceAsTestCase("mixed types to string", S[any](testHello, 42, "world", 3.14, "!"),
			S(testHello, "world", "!")),
		newSliceAsTestCase("all strings", S[any]("a", "b", "c"), S("a", "b", "c")),
		newSliceAsTestCase("no strings", S[any](1, 2, 3, 4.5, true), nil),
		newSliceAsTestCase("empty slice", S[any](), nil),
		newSliceAsTestCase("nil slice", nil, nil),
		newSliceAsTestCase("with nil values", S[any](testHello, nil, "world"), S(testHello, "world")),
	}

	RunTestCases(t, testCases)
}

func TestSliceAsFn(t *testing.T) {
	// Custom conversion function that adds prefix
	prefixString := func(v any) (string, bool) {
		if s, ok := v.(string); ok {
			return "prefix:" + s, true
		}
		return "", false
	}

	testCases := []sliceAsFnTestCase{
		newSliceAsFnTestCase("with custom conversion", prefixString, S[any]("a", 1, "b", 2, "c"),
			S("prefix:a", "prefix:b", "prefix:c")),
		newSliceAsFnTestCase("with nil function", nil, S[any]("a", "b", "c"), nil),
		newSliceAsFnTestCase("empty slice", prefixString, S[any](), nil),
		newSliceAsFnTestCase("nil slice", prefixString, nil, nil),
		newSliceAsFnTestCase("all filtered out", prefixString, S[any](1, 2, 3), nil),
	}

	RunTestCases(t, testCases)
}

func TestAsError(t *testing.T) {
	testCases := []asErrorTestCase{
		newAsErrorTestCase("standard error", errors.New("standard error"), "standard error", true),
		newAsErrorTestCase("nil error", error(nil), "", false),
		newAsErrorTestCase("type with AsError returning error",
			errorWithAsError{msg: "custom error"}, "custom error", true),
		newAsErrorTestCase("type with AsError returning nil",
			errorWithAsError{msg: ""}, "", false),
		newAsErrorTestCase("type with OK returning false",
			errorWithOK{msg: "not ok error", ok: false}, "not ok error", true),
		newAsErrorTestCase("type with OK returning true", errorWithOK{msg: "ok error", ok: true}, "", false),
		newAsErrorTestCase("non-error type", "not an error", "", false),
		newAsErrorTestCase("nil value", nil, "", false),
		newAsErrorTestCase("integer", 42, "", false),
	}

	RunTestCases(t, testCases)
}

func TestAsErrors(t *testing.T) {
	testCases := []asErrorsTestCase{
		newAsErrorsTestCase("mixed values with errors", S[any](
			errors.New("error1"),
			"not an error",
			errors.New("error2"),
			42,
			errorWithAsError{msg: "error3"},
			nil,
		), S("error1", "error2", "error3"), 3),
		newAsErrorsTestCase("all errors", S[any](
			errors.New("a"),
			errors.New("b"),
			errors.New("c"),
		), S("a", "b", "c"), 3),
		newAsErrorsTestCase("no errors", S[any]("a", 1, true, nil), S[string](), 0),
		newAsErrorsTestCase("empty slice", S[any](), S[string](), 0),
		newAsErrorsTestCase("nil slice", nil, S[string](), 0),
		newAsErrorsTestCase("with OK interface", S[any](
			errorWithOK{msg: "fail", ok: false},
			errorWithOK{msg: "pass", ok: true},
			errorWithOK{msg: "fail2", ok: false},
		), S("fail", "fail2"), 2),
	}

	RunTestCases(t, testCases)
}

// Benchmark tests
func BenchmarkAs(b *testing.B) {
	input := "test string"
	for i := 0; i < b.N; i++ {
		_, _ = As[any, string](input)
	}
}

func BenchmarkAsFn(b *testing.B) {
	fn := func(v any) (string, bool) {
		s, ok := v.(string)
		return s, ok
	}
	var input any = "test string"

	for i := 0; i < b.N; i++ {
		_, _ = AsFn(fn, input)
	}
}

func BenchmarkSliceAs(b *testing.B) {
	input := S[any]("a", 1, "b", 2, "c", 3, "d", 4, "e", 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceAs[any, string](input)
	}
}

func BenchmarkAsError(b *testing.B) {
	err := errors.New("test error")

	for i := 0; i < b.N; i++ {
		_ = AsError(err)
	}
}

// Test with concrete types to ensure generics work correctly
func TestAsWithConcreteTypes(t *testing.T) {
	// Test int to int64
	i := 42
	if v, ok := As[int, int64](i); ok || v != 0 {
		t.Errorf("As[int, int64](%d) = %v, %v; want 0, false", i, v, ok)
	}

	// Test *int to *int
	pi := &i
	if v, ok := As[*int, *int](pi); !ok || v != pi {
		t.Errorf("As[*int, *int](%p) = %p, %v; want %p, true", pi, v, ok, pi)
	}

	// Test interface{} to concrete type
	var value any = testHello
	if v, ok := As[any, string](value); !ok || v != testHello {
		t.Errorf("As[any, string](%v) = %v, %v; want hello, true", value, v, ok)
	}
}

// Test SliceAsFn with various function types
func TestSliceAsFnEdgeCases(t *testing.T) {
	// Function that panics
	panicFn := func(_ any) (string, bool) {
		panic("test panic")
	}

	// Verify panic propagates
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("SliceAsFn with panic function should panic")
		}
	}()

	_ = SliceAsFn(panicFn, S[any]("will panic"))
}
