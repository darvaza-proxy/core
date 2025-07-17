package core

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsErrorFn(t *testing.T) {
	testErr := errors.New("test error")
	differentErr := errors.New("different error")
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)

	isTestErr := func(err error) bool {
		return errors.Is(err, testErr)
	}

	for _, tc := range []struct {
		name     string
		checkFn  func(error) bool
		errs     []error
		expected bool
	}{
		{"matching error", isTestErr, S(testErr), true},
		{"non-matching error", isTestErr, S(differentErr), false},
		{"nil check function", nil, S(testErr), false},
		{"no errors", isTestErr, S[error](), false},
		{"wrapped error", isTestErr, S(wrappedErr), true},
		{"nil error in slice", isTestErr, S[error](nil, testErr), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := IsErrorFn(tc.checkFn, tc.errs...)
			if result != tc.expected {
				t.Errorf("IsErrorFn() = %v, want %v", result, tc.expected)
			}
		})
	}
}

//revive:disable:cognitive-complexity
func TestIsErrorFn2(t *testing.T) {
	//revive:enable:cognitive-complexity
	testErr := errors.New("test error")
	differentErr := errors.New("different error")
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)

	isTestErr := func(err error) (bool, bool) {
		if err == nil {
			return false, false
		}
		return errors.Is(err, testErr), true
	}

	for _, tc := range []struct {
		name          string
		checkFn       func(error) (bool, bool)
		errs          []error
		expectedIs    bool
		expectedKnown bool
	}{
		{"matching error", isTestErr, S(testErr), true, true},
		{"non-matching error", isTestErr, S(differentErr), false, true},
		{"nil check function", nil, S(testErr), false, true},
		{"no errors", isTestErr, S[error](), false, true},
		{"wrapped error", isTestErr, S(wrappedErr), true, true},
		{"unknown error type", func(_ error) (bool, bool) { return false, false }, S(testErr), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			is, known := IsErrorFn2(tc.checkFn, tc.errs...)
			if is != tc.expectedIs || known != tc.expectedKnown {
				t.Errorf("IsErrorFn2() = (%v, %v), want (%v, %v)", is, known, tc.expectedIs, tc.expectedKnown)
			}
		})
	}
}
