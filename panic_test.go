package core

import (
	"errors"
	"testing"
)

type asRecoveredTestCase struct {
	// Large fields - interface types and strings
	expected any
	input    any
	name     string

	// Small fields (1 byte) - booleans
	isNil bool
}

var asRecoveredTestCases = []asRecoveredTestCase{
	{
		name:     "nil input",
		input:    nil,
		expected: nil,
		isNil:    true,
	},
	{
		name:     "string panic",
		input:    "test panic",
		expected: "test panic",
		isNil:    false,
	},
	{
		name:     "error panic",
		input:    errors.New("test error"),
		expected: "test error", // String comparison for error content
		isNil:    false,
	},
	{
		name:     "int panic",
		input:    42,
		expected: 42,
		isNil:    false,
	},
	{
		name:     "already recovered",
		input:    NewPanicError(1, "already wrapped"),
		expected: "already wrapped",
		isNil:    false,
	},
}

//revive:disable-next-line:cognitive-complexity
//revive:disable-next-line:cyclomatic
func (tc asRecoveredTestCase) test(t *testing.T) {
	t.Helper()
	result := AsRecovered(tc.input)

	if tc.isNil {
		if result != nil {
			t.Fatalf("expected nil, got %v", result)
		}
		return
	}

	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	// Test Recovered interface
	recovered := result.Recovered()

	// Handle different types appropriately for comparison
	switch exp := tc.expected.(type) {
	case string:
		// For string inputs, they get converted to errors
		if err, ok := recovered.(error); ok {
			AssertEqual(t, exp, err.Error(), "recovered error message mismatch")
		} else if recovered != exp {
			t.Fatalf("expected recovered value %v, got %v", exp, recovered)
		}
	case error:
		if err, ok := recovered.(error); ok {
			AssertEqual(t, exp.Error(), err.Error(), "recovered error message mismatch")
		} else {
			t.Fatalf("expected error type, got %T", recovered)
		}
	default:
		AssertEqual(t, tc.expected, recovered, "recovered value mismatch")
	}

	// Test Error method
	errorStr := result.Error()
	if errorStr == "" {
		t.Fatalf("expected non-empty error string")
	}
}

func TestAsRecovered(t *testing.T) {
	for _, tc := range asRecoveredTestCases {
		t.Run(tc.name, tc.test)
	}
}

type catcherDoTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

var catcherDoTestCases = []catcherDoTestCase{
	{
		name: "successful function",
		fn: func() error {
			return nil
		},
		expectError: false,
		expectPanic: false,
	},
	{
		name: "function returns error",
		fn: func() error {
			return errors.New("test error")
		},
		expectError: true,
		expectPanic: false,
	},
	{
		name: "function panics with string",
		fn: func() error {
			panic("test panic")
		},
		expectError: true,
		expectPanic: true,
	},
	{
		name: "function panics with error",
		fn: func() error {
			panic(errors.New("panic error"))
		},
		expectError: true,
		expectPanic: true,
	},
	{
		name: "function panics with int",
		fn: func() error {
			panic(42)
		},
		expectError: true,
		expectPanic: true,
	},
	{
		name:        "nil function",
		fn:          nil,
		expectError: false,
		expectPanic: false,
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc catcherDoTestCase) test(t *testing.T) {
	t.Helper()
	var catcher Catcher
	err := catcher.Do(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catcher.Do error")
	} else {
		AssertNoError(t, err, "Catcher.Do error")
	}

	if tc.expectError && tc.expectPanic {
		if recovered, ok := err.(Recovered); ok {
			if recovered.Recovered() == nil {
				t.Fatalf("expected recovered panic value, got nil")
			}
		} else {
			t.Fatalf("expected Recovered error, got %T", err)
		}
	}
}

func TestCatcherDo(t *testing.T) {
	for _, tc := range catcherDoTestCases {
		t.Run(tc.name, tc.test)
	}
}

type catcherTryTestCase struct {
	fn          func() error
	name        string
	expectError bool
	expectPanic bool
}

var catcherTryTestCases = []catcherTryTestCase{
	{
		name: "successful function",
		fn: func() error {
			return nil
		},
		expectError: false,
		expectPanic: false,
	},
	{
		name: "function returns error",
		fn: func() error {
			return errors.New("test error")
		},
		expectError: true,
		expectPanic: false,
	},
	{
		name: "function panics",
		fn: func() error {
			panic("test panic")
		},
		expectError: false,
		expectPanic: true,
	},
	{
		name:        "nil function",
		fn:          nil,
		expectError: false,
		expectPanic: false,
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc catcherTryTestCase) test(t *testing.T) {
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
		if recovered == nil {
			t.Fatalf("expected recovered panic, got nil")
		}
	} else if recovered != nil {
		t.Fatalf("expected no recovered panic, got %v", recovered)
	}
}

func TestCatcherTry(t *testing.T) {
	for _, tc := range catcherTryTestCases {
		t.Run(tc.name, tc.test)
	}
}

func TestCatcherRecovered(t *testing.T) {
	var catcher Catcher

	// Initially no panic
	if recovered := catcher.Recovered(); recovered != nil {
		t.Fatalf("expected nil recovered, got %v", recovered)
	}

	// After panic
	_ = catcher.Try(func() error {
		panic("test panic")
	})

	recovered := catcher.Recovered()
	if recovered == nil {
		t.Fatalf("expected recovered panic, got nil")
	}

	// String panics get converted to errors by NewPanicError
	if err, ok := recovered.Recovered().(error); ok {
		AssertEqual(t, "test panic", err.Error(), "panic error message mismatch")
	} else {
		t.Fatalf("expected error type for string panic, got %T", recovered.Recovered())
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
	if recovered == nil {
		t.Fatalf("expected recovered panic, got nil")
	}

	// Should be either "first panic" or "second panic" (converted to errors)
	panicValue := recovered.Recovered()
	if err, ok := panicValue.(error); ok {
		errorStr := err.Error()
		if errorStr != "first panic" && errorStr != "second panic" {
			t.Fatalf("unexpected panic value: %v", errorStr)
		}
	} else {
		t.Fatalf("expected error type for string panic, got %T", panicValue)
	}
}

type catchTestCase struct {
	fn          func() error
	name        string
	expectError bool
}

var catchTestCases = []catchTestCase{
	{
		name: "successful function",
		fn: func() error {
			return nil
		},
		expectError: false,
	},
	{
		name: "function returns error",
		fn: func() error {
			return errors.New("test error")
		},
		expectError: true,
	},
	{
		name: "function panics",
		fn: func() error {
			panic("test panic")
		},
		expectError: true,
	},
}

func (tc catchTestCase) test(t *testing.T) {
	t.Helper()
	err := Catch(tc.fn)

	if tc.expectError {
		AssertError(t, err, "Catch error")
	} else {
		AssertNoError(t, err, "Catch error")
	}
}

func TestCatch(t *testing.T) {
	for _, tc := range catchTestCases {
		t.Run(tc.name, tc.test)
	}
}

type catchWithPanicRecoveryTestCase struct {
	value any
	name  string
}

var catchWithPanicRecoveryTestCases = []catchWithPanicRecoveryTestCase{
	{name: "string panic", value: "string panic"},
	{name: "int panic", value: 42},
	{name: "float panic", value: 3.14},
	{name: "error panic", value: errors.New("error panic")},
	{name: "formatted error", value: errors.New("formatted error")},
	// Skip slice and map as they are not comparable
}

//revive:disable-next-line:cognitive-complexity
func (tc catchWithPanicRecoveryTestCase) test(t *testing.T) {
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
				AssertEqual(t, s, err.Error(), "panic error message mismatch")
			} else {
				t.Fatalf("expected error type for string panic, got %T", panicValue)
			}
		} else {
			AssertEqual(t, tc.value, panicValue, "panic value mismatch")
		}
	} else {
		t.Fatalf("expected Recovered error, got %T", err)
	}
}

func TestCatchWithPanicRecovery(t *testing.T) {
	for _, tc := range catchWithPanicRecoveryTestCases {
		t.Run(tc.name, tc.test)
	}
}
