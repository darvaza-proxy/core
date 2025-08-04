package core

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = withTimeoutTestCase{}
	_ TestCase = withTimeoutCauseTestCase{}
	_ TestCase = contextKeyGetTestCase{}
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

// newWithTimeoutTestCase creates a new withTimeoutTestCase
func newWithTimeoutTestCase(parent context.Context, name string, timeout time.Duration,
	expectTimeout bool) withTimeoutTestCase {
	return withTimeoutTestCase{
		parent:        parent,
		name:          name,
		timeout:       timeout,
		expectTimeout: expectTimeout,
	}
}

func (tc withTimeoutTestCase) Name() string {
	return tc.name
}

func (tc withTimeoutTestCase) Test(t *testing.T) {
	t.Helper()

	ctx, cancel := WithTimeout(tc.parent, tc.timeout)
	defer cancel()

	if tc.expectTimeout {
		_, hasDeadline := ctx.Deadline()
		AssertTrue(t, hasDeadline, "deadline set")
		return
	}

	if tc.timeout <= 0 {
		expectedParent := tc.parent
		if expectedParent == nil {
			expectedParent = context.Background()
		}
		AssertEqual(t, expectedParent, ctx, "WithTimeout result")
	}
}

func withTimeoutTestCases() []withTimeoutTestCase {
	var nilCtx context.Context

	baseCtx := context.Background()

	return S(
		newWithTimeoutTestCase(baseCtx, "zero duration", 0, false),
		newWithTimeoutTestCase(baseCtx, "negative duration", -time.Second, false),
		newWithTimeoutTestCase(baseCtx, "positive duration", time.Millisecond, true),
		newWithTimeoutTestCase(baseCtx, "large duration", time.Hour, true),
		newWithTimeoutTestCase(nilCtx, "nil context", time.Millisecond, true),
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
	RunTestCases(t, withTimeoutTestCases())

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

// newWithTimeoutCauseTestCase creates a new withTimeoutCauseTestCase
func newWithTimeoutCauseTestCase(parent context.Context, name string,
	timeout time.Duration, cause error, expectTimeout bool) withTimeoutCauseTestCase {
	return withTimeoutCauseTestCase{
		parent:        parent,
		cause:         cause,
		name:          name,
		timeout:       timeout,
		expectTimeout: expectTimeout,
	}
}

func (tc withTimeoutCauseTestCase) Name() string {
	return tc.name
}

func (tc withTimeoutCauseTestCase) Test(t *testing.T) {
	t.Helper()

	ctx, cancel := WithTimeoutCause(tc.parent, tc.timeout, tc.cause)
	defer cancel()

	if tc.expectTimeout {
		_, hasDeadline := ctx.Deadline()
		AssertTrue(t, hasDeadline, "deadline set")
		return
	}

	if tc.timeout <= 0 {
		expectedParent := tc.parent
		if expectedParent == nil {
			expectedParent = context.Background()
		}
		AssertEqual(t, expectedParent, ctx, "WithTimeoutCause result")
	}
}

func withTimeoutCauseTestCases() []withTimeoutCauseTestCase {
	var nilCtx context.Context
	ctx := context.Background()
	return S(
		newWithTimeoutCauseTestCase(ctx, "zero duration", 0, errors.New("test cause"), false),
		newWithTimeoutCauseTestCase(ctx, "negative duration", -time.Second, errors.New("test cause"), false),
		newWithTimeoutCauseTestCase(ctx, "positive duration", time.Millisecond, errors.New("test cause"), true),
		newWithTimeoutCauseTestCase(ctx, "large duration", time.Hour, errors.New("test cause"), true),
		newWithTimeoutCauseTestCase(nilCtx, "nil context", time.Millisecond, errors.New("test cause"), true),
		newWithTimeoutCauseTestCase(ctx, "nil cause", time.Millisecond, nil, true),
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
	RunTestCases(t, withTimeoutCauseTestCases())

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

// newContextKeyGetTestCase creates a new contextKeyGetTestCase
func newContextKeyGetTestCase(ctx context.Context, name string, key *ContextKey[int],
	expectedValue int, expectedOk bool) contextKeyGetTestCase {
	return contextKeyGetTestCase{
		ctx:           ctx,
		key:           key,
		name:          name,
		expectedValue: expectedValue,
		expectedOK:    expectedOk,
	}
}

func (tc contextKeyGetTestCase) Name() string {
	return tc.name
}

func (tc contextKeyGetTestCase) Test(t *testing.T) {
	t.Helper()

	value, ok := tc.key.Get(tc.ctx)
	AssertEqual(t, tc.expectedValue, value, "Get value")
	AssertEqual(t, tc.expectedOK, ok, "Get ok result")
}

func contextKeyGetTestCases() []contextKeyGetTestCase {
	var nilCtx context.Context
	return S(
		newContextKeyGetTestCase(context.Background(), "nil receiver", nil, 0, false),
		newContextKeyGetTestCase(nilCtx, "nil context", NewContextKey[int]("test"), 0, false),
		newContextKeyGetTestCase(
			context.WithValue(context.Background(), NewContextKey[int]("test"), "string_value"),
			"wrong type", NewContextKey[int]("test"), 0, false),
	)
}

func TestContextKeyGet(t *testing.T) {
	RunTestCases(t, contextKeyGetTestCases())
}
