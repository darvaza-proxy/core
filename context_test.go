package core

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// revive:disable:cognitive-complexity
func TestNewContextKey(t *testing.T) {
	// revive:enable:cognitive-complexity
	k0 := NewContextKey[int]("k0")
	// name
	if k0.String() != "k0" {
		t.Fail()
	}
	// name and type
	s := fmt.Sprintf("core.NewContextKey[%s](%q)", "int", "k0")
	if k0.GoString() != s {
		t.Fail()
	}

	// new context
	ctx0 := k0.WithValue(context.TODO(), 123)
	v0, ok := k0.Get(ctx0)
	if !ok || v0 != 123 {
		t.Fail()
	}
	// wrong context
	_, ok = k0.Get(context.TODO())
	if ok {
		t.Fail()
	}
	// sub-context
	ctx1 := k0.WithValue(ctx0, 456)
	v1, ok := k0.Get(ctx1)
	if !ok || v1 != 456 {
		t.Fail()
	}
	// parent-context
	v, ok := k0.Get(ctx0)
	if !ok || v != v0 {
		t.Fail()
	}
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
		AssertBool(t, hasDeadline, true, "WithTimeout should set deadline for positive duration")
		return
	}

	if tc.timeout <= 0 {
		expectedParent := tc.parent
		if expectedParent == nil {
			expectedParent = context.Background()
		}
		AssertEqual(t, expectedParent, ctx, "WithTimeout should return parent for zero/negative duration")
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
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should have timed out")
	}
}

func testWithTimeoutCancellation(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background(), time.Hour)

	select {
	case <-ctx.Done():
		t.Error("Context should not be done immediately")
	default:
	}

	cancel()

	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be done after cancel")
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
		AssertBool(t, hasDeadline, true, "WithTimeoutCause should set deadline for positive duration")
		return
	}

	if tc.timeout <= 0 {
		expectedParent := tc.parent
		if expectedParent == nil {
			expectedParent = context.Background()
		}
		AssertEqual(t, expectedParent, ctx, "WithTimeoutCause should return parent for zero/negative duration")
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
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
		}
		// Check cause (if supported by Go version)
		if cause := context.Cause(ctx); cause != testErr {
			t.Errorf("Expected cause %v, got %v", testErr, cause)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should have timed out")
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
	expectedOK    bool
}

func (tc contextKeyGetTestCase) test(t *testing.T) {
	t.Helper()

	value, ok := tc.key.Get(tc.ctx)
	AssertEqual(t, tc.expectedValue, value, "Get returned wrong value")
	AssertBool(t, ok, tc.expectedOK, "Get returned wrong ok value")
}

func contextKeyGetTest(ctx context.Context, name string, key *ContextKey[int],
	expectedValue int, expectedOK bool) contextKeyGetTestCase {
	return contextKeyGetTestCase{
		ctx:           ctx,
		key:           key,
		name:          name,
		expectedValue: expectedValue,
		expectedOK:    expectedOK,
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
