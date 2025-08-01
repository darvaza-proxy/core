package core

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewContextKey(t *testing.T) {
	k0 := NewContextKey[int]("k0")
	// name
	AssertEqual(t, "k0", k0.String(), "key name")
	// name and type
	s := fmt.Sprintf("core.NewContextKey[%s](%q)", "int", "k0")
	AssertEqual(t, s, k0.GoString(), "GoString")

	// new context
	ctx0 := k0.WithValue(context.TODO(), 123)
	v0, ok := k0.Get(ctx0)
	AssertTrue(t, ok, "context value found")
	AssertEqual(t, 123, v0, "value")
	// wrong context
	_, ok = k0.Get(context.TODO())
	AssertFalse(t, ok, "wrong context")
	// sub-context
	ctx1 := k0.WithValue(ctx0, 456)
	v1, ok := k0.Get(ctx1)
	AssertTrue(t, ok, "sub-context value found")
	AssertEqual(t, 456, v1, "sub-context")
	// parent-context
	v, ok := k0.Get(ctx0)
	AssertTrue(t, ok, "parent context value found")
	AssertEqual(t, v0, v, "parent value")
}

// Test cases for WithTimeout function
type withTimeoutTestCase struct {
	parent        context.Context
	name          string
	timeout       time.Duration
	expectTimeout bool
}

func (tc withTimeoutTestCase) test(t *testing.T) {
	t.Helper()

	ctx, cancel := WithTimeout(tc.parent, tc.timeout)
	defer cancel()

	if tc.expectTimeout {
		_, hasDeadline := ctx.Deadline()
		AssertTrue(t, hasDeadline, "has deadline")
		return
	}

	if tc.timeout <= 0 {
		expectedParent := tc.parent
		if expectedParent == nil {
			expectedParent = context.Background()
		}
		AssertEqual(t, expectedParent, ctx, "parent")
	}
}

func timeoutTestCase(parent context.Context, name string, timeout time.Duration,
	expectTimeout bool) withTimeoutTestCase {
	return withTimeoutTestCase{
		parent:        parent,
		name:          name,
		timeout:       timeout,
		expectTimeout: expectTimeout,
	}
}

func withTimeoutTestCases() []withTimeoutTestCase {
	var nilCtx context.Context
	return S(
		timeoutTestCase(context.Background(), "zero duration", 0, false),
		timeoutTestCase(context.Background(), "negative duration", -time.Second, false),
		timeoutTestCase(context.Background(), "positive duration", time.Millisecond, true),
		timeoutTestCase(context.Background(), "large duration", time.Hour, true),
		timeoutTestCase(nilCtx, "nil context", time.Millisecond, true),
	)
}

func testWithTimeoutExpiration(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		// Should timeout
		AssertEqual(t, context.DeadlineExceeded, ctx.Err(), "timeout error")
	case <-time.After(100 * time.Millisecond):
		t.Error("context should have timed out")
	}
}

func testWithTimeoutCancellation(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background(), time.Hour)

	select {
	case <-ctx.Done():
		t.Error("context should not be done immediately")
	default:
	}

	cancel()

	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("context should be done after cancel")
	}
}

func TestWithTimeout(t *testing.T) {
	for _, tc := range withTimeoutTestCases() {
		t.Run(tc.name, tc.test)
	}

	// Test timeout expiration
	t.Run("timeout expiration", testWithTimeoutExpiration)

	// Test cancellation
	t.Run("cancellation", testWithTimeoutCancellation)
}

// Test cases for WithTimeoutCause function
type withTimeoutCauseTestCase struct {
	parent        context.Context
	cause         error
	name          string
	timeout       time.Duration
	expectTimeout bool
}

func (tc withTimeoutCauseTestCase) test(t *testing.T) {
	t.Helper()

	ctx, cancel := WithTimeoutCause(tc.parent, tc.timeout, tc.cause)
	defer cancel()

	if tc.expectTimeout {
		_, hasDeadline := ctx.Deadline()
		AssertTrue(t, hasDeadline, "has deadline")
		return
	}

	if tc.timeout <= 0 {
		expectedParent := tc.parent
		if expectedParent == nil {
			expectedParent = context.Background()
		}
		AssertEqual(t, expectedParent, ctx, "parent")
	}
}

func timeoutCauseTestCase(parent context.Context, name string, timeout time.Duration,
	cause error, expectTimeout bool) withTimeoutCauseTestCase {
	return withTimeoutCauseTestCase{
		parent:        parent,
		cause:         cause,
		name:          name,
		timeout:       timeout,
		expectTimeout: expectTimeout,
	}
}

func withTimeoutCauseTestCases() []withTimeoutCauseTestCase {
	var nilCtx context.Context
	ctx := context.Background()
	return S(
		timeoutCauseTestCase(ctx, "zero duration", 0, errors.New("test cause"), false),
		timeoutCauseTestCase(ctx, "negative duration", -time.Second, errors.New("test cause"), false),
		timeoutCauseTestCase(ctx, "positive duration", time.Millisecond, errors.New("test cause"), true),
		timeoutCauseTestCase(ctx, "large duration", time.Hour, errors.New("test cause"), true),
		timeoutCauseTestCase(nilCtx, "nil context", time.Millisecond, errors.New("test cause"), true),
		timeoutCauseTestCase(ctx, "nil cause", time.Millisecond, nil, true),
	)
}

func testWithTimeoutCauseExpiration(t *testing.T) {
	testErr := errors.New("custom timeout cause")
	ctx, cancel := WithTimeoutCause(context.Background(), 10*time.Millisecond, testErr)
	defer cancel()

	select {
	case <-ctx.Done():
		// Should timeout
		AssertEqual(t, context.DeadlineExceeded, ctx.Err(), "timeout error")
		// Check cause (if supported by Go version)
		if cause := context.Cause(ctx); cause != testErr {
			AssertEqual(t, testErr, cause, "cause")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("context should have timed out")
	}
}

func TestWithTimeoutCause(t *testing.T) {
	for _, tc := range withTimeoutCauseTestCases() {
		t.Run(tc.name, tc.test)
	}

	// Test timeout expiration with cause
	t.Run("timeout expiration with cause", testWithTimeoutCauseExpiration)
}

// Test cases for ContextKey Get function
type contextKeyGetTestCase struct {
	ctx           context.Context
	key           *ContextKey[int]
	name          string
	expectedValue int
	expectedOk    bool
}

func (tc contextKeyGetTestCase) test(t *testing.T) {
	t.Helper()

	value, ok := tc.key.Get(tc.ctx)
	AssertEqual(t, tc.expectedValue, value, "value")
	AssertEqual(t, tc.expectedOk, ok, "ok")
}

func contextKeyGetTest(ctx context.Context, name string, key *ContextKey[int],
	expectedValue int, expectedOk bool) contextKeyGetTestCase {
	return contextKeyGetTestCase{
		ctx:           ctx,
		key:           key,
		name:          name,
		expectedValue: expectedValue,
		expectedOk:    expectedOk,
	}
}

func contextKeyGetTestCases() []contextKeyGetTestCase {
	var nilCtx context.Context
	return S(
		contextKeyGetTest(context.Background(), "nil receiver", nil, 0, false),
		contextKeyGetTest(nilCtx, "nil context", NewContextKey[int]("test"), 0, false),
		contextKeyGetTest(
			context.WithValue(context.Background(), NewContextKey[int]("test"), "string_value"),
			"wrong type", NewContextKey[int]("test"), 0, false),
	)
}

func TestContextKeyGet(t *testing.T) {
	for _, tc := range contextKeyGetTestCases() {
		t.Run(tc.name, tc.test)
	}
}
