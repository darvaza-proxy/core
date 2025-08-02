package core

import (
	"errors"
	"fmt"
	"testing"
)

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

func (tc asTestCase) test(t *testing.T) {
	//revive:disable:cyclomatic,cognitive-complexity
	t.Helper()

	switch want := tc.want.(type) {
	case string:
		got, ok := As[any, string](tc.input)
		if ok != tc.wantOK {
			t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
		}
		if got != want {
			t.Errorf("As() got = %v, want %v", got, want)
		}
	case int:
		got, ok := As[any, int](tc.input)
		if ok != tc.wantOK {
			t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
		}
		if got != want {
			t.Errorf("As() got = %v, want %v", got, want)
		}
	case error:
		got, ok := As[any, error](tc.input)
		if ok != tc.wantOK {
			t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
		}
		if ok && got.Error() != want.Error() {
			t.Errorf("As() got = %v, want %v", got, want)
		}
	default:
		// Test cases where conversion should fail
		got, ok := As[any, string](tc.input)
		if ok != tc.wantOK {
			t.Errorf("As() ok = %v, want %v", ok, tc.wantOK)
		}
		if tc.wantOK && got != "" {
			t.Errorf("As() got = %v, want zero value", got)
		}
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

func (tc asFnTestCase) test(t *testing.T) {
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

func (tc sliceAsTestCase) test(t *testing.T) {
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

func (tc sliceAsFnTestCase) test(t *testing.T) {
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

func (tc asErrorTestCase) test(t *testing.T) {
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

func (tc asErrorsTestCase) test(t *testing.T) {
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
		{
			name:   "string to string",
			input:  "hello",
			want:   "hello",
			wantOK: true,
		},
		{
			name:   "int to int",
			input:  42,
			want:   42,
			wantOK: true,
		},
		{
			name:   "error to error",
			input:  errors.New("test error"),
			want:   errors.New("test error"),
			wantOK: true,
		},
		{
			name:   "int to string fails",
			input:  42,
			want:   "",
			wantOK: false,
		},
		{
			name:   "nil to string",
			input:  nil,
			want:   "",
			wantOK: false,
		},
		{
			name:   "nil to error",
			input:  nil,
			want:   error(nil),
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
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
		{
			name:   "with valid conversion function",
			fn:     intToString,
			input:  42,
			want:   "42",
			wantOK: true,
		},
		{
			name:   "with valid conversion function but wrong type",
			fn:     intToString,
			input:  "not an int",
			want:   "",
			wantOK: false,
		},
		{
			name:   "with nil function",
			fn:     nil,
			input:  42,
			want:   "",
			wantOK: false,
		},
		{
			name:   "with function returning false",
			fn:     func(any) (string, bool) { return "ignored", false },
			input:  42,
			want:   "ignored",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestSliceAs(t *testing.T) {
	testCases := []sliceAsTestCase{
		{
			name:  "mixed types to string",
			input: S[any]("hello", 42, "world", 3.14, "!"),
			want:  S("hello", "world", "!"),
		},
		{
			name:  "all strings",
			input: S[any]("a", "b", "c"),
			want:  S("a", "b", "c"),
		},
		{
			name:  "no strings",
			input: S[any](1, 2, 3, 4.5, true),
			want:  nil,
		},
		{
			name:  "empty slice",
			input: S[any](),
			want:  nil,
		},
		{
			name:  "nil slice",
			input: nil,
			want:  nil,
		},
		{
			name:  "with nil values",
			input: S[any]("hello", nil, "world"),
			want:  S("hello", "world"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
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
		{
			name:  "with custom conversion",
			fn:    prefixString,
			input: S[any]("a", 1, "b", 2, "c"),
			want:  S("prefix:a", "prefix:b", "prefix:c"),
		},
		{
			name:  "with nil function",
			fn:    nil,
			input: S[any]("a", "b", "c"),
			want:  nil,
		},
		{
			name:  "empty slice",
			fn:    prefixString,
			input: S[any](),
			want:  nil,
		},
		{
			name:  "nil slice",
			fn:    prefixString,
			input: nil,
			want:  nil,
		},
		{
			name:  "all filtered out",
			fn:    prefixString,
			input: S[any](1, 2, 3),
			want:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestAsError(t *testing.T) {
	testCases := []asErrorTestCase{
		{
			name:    "standard error",
			input:   errors.New("standard error"),
			wantErr: true,
			wantMsg: "standard error",
		},
		{
			name:    "nil error",
			input:   error(nil),
			wantErr: false,
		},
		{
			name:    "type with AsError returning error",
			input:   errorWithAsError{msg: "custom error"},
			wantErr: true,
			wantMsg: "custom error",
		},
		{
			name:    "type with AsError returning nil",
			input:   errorWithAsError{msg: ""},
			wantErr: false,
		},
		{
			name:    "type with OK returning false",
			input:   errorWithOK{msg: "not ok error", ok: false},
			wantErr: true,
			wantMsg: "not ok error",
		},
		{
			name:    "type with OK returning true",
			input:   errorWithOK{msg: "ok error", ok: true},
			wantErr: false,
		},
		{
			name:    "non-error type",
			input:   "not an error",
			wantErr: false,
		},
		{
			name:    "nil value",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "integer",
			input:   42,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

func TestAsErrors(t *testing.T) {
	testCases := []asErrorsTestCase{
		{
			name: "mixed values with errors",
			input: S[any](
				errors.New("error1"),
				"not an error",
				errors.New("error2"),
				42,
				errorWithAsError{msg: "error3"},
				nil,
			),
			wantLen:  3,
			wantMsgs: S("error1", "error2", "error3"),
		},
		{
			name: "all errors",
			input: S[any](
				errors.New("a"),
				errors.New("b"),
				errors.New("c"),
			),
			wantLen:  3,
			wantMsgs: S("a", "b", "c"),
		},
		{
			name:     "no errors",
			input:    S[any]("a", 1, true, nil),
			wantLen:  0,
			wantMsgs: S[string](),
		},
		{
			name:     "empty slice",
			input:    S[any](),
			wantLen:  0,
			wantMsgs: S[string](),
		},
		{
			name:     "nil slice",
			input:    nil,
			wantLen:  0,
			wantMsgs: S[string](),
		},
		{
			name: "with OK interface",
			input: S[any](
				errorWithOK{msg: "fail", ok: false},
				errorWithOK{msg: "pass", ok: true},
				errorWithOK{msg: "fail2", ok: false},
			),
			wantLen:  2,
			wantMsgs: S("fail", "fail2"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
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
	var value any = "hello"
	if v, ok := As[any, string](value); !ok || v != "hello" {
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
