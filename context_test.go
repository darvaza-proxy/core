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
	_ TestCase = withTimeoutNewTestCase[withTimeoutArgs]{}
	_ TestCase = withTimeoutClampedTestCase[withTimeoutArgs]{}
	_ TestCase = withTimeoutParentTestCase[withTimeoutArgs]{}
	_ TestCase = withTimeoutNewTestCase[withTimeoutCauseArgs]{}
	_ TestCase = withTimeoutClampedTestCase[withTimeoutCauseArgs]{}
	_ TestCase = withTimeoutParentTestCase[withTimeoutCauseArgs]{}
	_ TestCase = contextKeyGetTestCase{}
	_ TestCase = contextKeyStringTestCase[int]{}
)

// contextKeyStringTestCase states how one key renders: String is the
// name, and GoString names the value type as well, an interface
// included when the zero value has no dynamic type to print. A nil
// key has no name, and renders as the typed nil conversion.
type contextKeyStringTestCase[T any] struct {
	key    *ContextKey[T]
	want   string
	wantGo string
	name   string
}

func newContextKeyStringTestCase[T any](name string, key *ContextKey[T],
	want, wantGo string) TestCase {
	return contextKeyStringTestCase[T]{
		key:    key,
		want:   want,
		wantGo: wantGo,
		name:   name,
	}
}

func (tc contextKeyStringTestCase[T]) Name() string {
	return tc.name
}

func (tc contextKeyStringTestCase[T]) Test(t *testing.T) {
	t.Helper()
	AssertEqual(t, tc.want, tc.key.String(), "String")
	AssertEqual(t, tc.wantGo, tc.key.GoString(), "GoString")
}

func contextKeyStringTestCases() []TestCase {
	return S(
		newContextKeyStringTestCase("concrete type", NewContextKey[int]("k0"),
			"k0", `core.NewContextKey[int]("k0")`),
		newContextKeyStringTestCase("interface type",
			NewContextKey[fmt.Stringer]("k1"),
			"k1", `core.NewContextKey[fmt.Stringer]("k1")`),
		newContextKeyStringTestCase("nil key", (*ContextKey[int])(nil),
			"<nil>", "(*core.ContextKey[int])(nil)"),
	)
}

func TestContextKeyString(t *testing.T) {
	RunTestCases(t, contextKeyStringTestCases())
}

// TestContextKeyWithValueNilReceiver states that a nil key refuses
// the value rather than storing it where Get cannot find it.
func TestContextKeyWithValueNilReceiver(t *testing.T) {
	AssertPanic(t, func() {
		(*ContextKey[int])(nil).WithValue(context.Background(), 1)
	}, ErrNilReceiver, "nil key")
}

func TestNewContextKey(t *testing.T) {
	k0 := NewContextKey[int]("k0")

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

// withTimeoutArgs holds what a row passes to WithTimeout.
type withTimeoutArgs struct {
	parent  context.Context
	name    string
	timeout time.Duration
}

func (a withTimeoutArgs) Name() string {
	return a.name
}

func (a withTimeoutArgs) args() withTimeoutArgs {
	return a
}

// run calls WithTimeout, returning the context and the instants either
// side of the call.
func (a withTimeoutArgs) run(t *testing.T) (ctx context.Context, before, after time.Time) {
	t.Helper()

	var cancel context.CancelFunc
	before = time.Now()
	ctx, cancel = WithTimeout(a.parent, a.timeout)
	after = time.Now()
	t.Cleanup(cancel)
	return ctx, before, after
}

func newWithTimeoutArgs(parent context.Context, name string, timeout time.Duration) withTimeoutArgs {
	return withTimeoutArgs{
		parent:  parent,
		name:    name,
		timeout: timeout,
	}
}

// withTimeoutCauseArgs holds what a row passes to WithTimeoutCause.
type withTimeoutCauseArgs struct {
	cause error
	withTimeoutArgs
}

// run calls WithTimeoutCause, returning the context and the instants
// either side of the call.
func (a withTimeoutCauseArgs) run(t *testing.T) (ctx context.Context, before, after time.Time) {
	t.Helper()

	var cancel context.CancelFunc
	before = time.Now()
	ctx, cancel = WithTimeoutCause(a.parent, a.timeout, a.cause)
	after = time.Now()
	t.Cleanup(cancel)
	return ctx, before, after
}

func newWithTimeoutCauseArgs(parent context.Context, name string, timeout time.Duration,
	cause error) withTimeoutCauseArgs {
	return withTimeoutCauseArgs{
		cause:           cause,
		withTimeoutArgs: newWithTimeoutArgs(parent, name, timeout),
	}
}

// withTimeoutCall is a row's call to WithTimeout or WithTimeoutCause.
type withTimeoutCall interface {
	Name() string
	args() withTimeoutArgs
	run(t *testing.T) (ctx context.Context, before, after time.Time)
}

// withTimeoutNewTestCase tests a call returning a new context whose
// deadline is the timeout from when the call ran.
type withTimeoutNewTestCase[C withTimeoutCall] struct {
	call C
}

func (tc withTimeoutNewTestCase[C]) Name() string {
	return tc.call.Name()
}

func (tc withTimeoutNewTestCase[C]) Test(t *testing.T) {
	t.Helper()
	ctx, before, after := tc.call.run(t)
	a := tc.call.args()

	AssertNotEqual(t, a.parent, ctx, "new context")
	deadline, ok := ctx.Deadline()
	AssertMustTrue(t, ok, "deadline set")
	AssertFalse(t, deadline.Before(before.Add(a.timeout)), "deadline not before timeout")
	AssertFalse(t, deadline.After(after.Add(a.timeout)), "deadline not after timeout")
}

// withTimeoutClampedTestCase tests a call returning a new context that
// kept its parent's deadline, that being sooner than the timeout.
type withTimeoutClampedTestCase[C withTimeoutCall] struct {
	call C
}

func (tc withTimeoutClampedTestCase[C]) Name() string {
	return tc.call.Name()
}

func (tc withTimeoutClampedTestCase[C]) Test(t *testing.T) {
	t.Helper()
	ctx, _, _ := tc.call.run(t)
	a := tc.call.args()

	AssertNotEqual(t, a.parent, ctx, "new context")
	want, ok := a.parent.Deadline()
	AssertMustTrue(t, ok, "parent deadline")
	deadline, ok := ctx.Deadline()
	AssertMustTrue(t, ok, "deadline set")
	AssertEqual(t, want, deadline, "parent's deadline kept")
}

// withTimeoutParentTestCase tests a call on a timeout that is not
// positive, returning the parent itself, or Background for a nil one.
type withTimeoutParentTestCase[C withTimeoutCall] struct {
	want context.Context
	call C
}

func (tc withTimeoutParentTestCase[C]) Name() string {
	return tc.call.Name()
}

func (tc withTimeoutParentTestCase[C]) Test(t *testing.T) {
	t.Helper()
	ctx, _, _ := tc.call.run(t)

	// AssertSame never holds for context.Background(), an empty
	// struct value; AssertEqual compares it by value, and the
	// pointer contexts these rows use by identity.
	AssertEqual(t, tc.want, ctx, "parent returned")
}

func newWithTimeoutTestCase(parent context.Context, name string, timeout time.Duration) TestCase {
	return withTimeoutNewTestCase[withTimeoutArgs]{
		call: newWithTimeoutArgs(parent, name, timeout),
	}
}

func newWithTimeoutTestCaseClamped(parent context.Context, name string, timeout time.Duration) TestCase {
	return withTimeoutClampedTestCase[withTimeoutArgs]{
		call: newWithTimeoutArgs(parent, name, timeout),
	}
}

func newWithTimeoutTestCaseParent(parent context.Context, name string, timeout time.Duration) TestCase {
	return withTimeoutParentTestCase[withTimeoutArgs]{
		want: parent,
		call: newWithTimeoutArgs(parent, name, timeout),
	}
}

func newWithTimeoutTestCaseBackground(name string, timeout time.Duration) TestCase {
	var nilCtx context.Context
	return withTimeoutParentTestCase[withTimeoutArgs]{
		want: context.Background(),
		call: newWithTimeoutArgs(nilCtx, name, timeout),
	}
}

// expiringContext returns a context whose deadline is a minute away,
// cancelled when the test ends.
func expiringContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func withTimeoutTestCases(t *testing.T) []TestCase {
	t.Helper()

	var nilCtx context.Context
	ctx := context.Background()
	expiring := expiringContext(t)

	return S[TestCase](
		// A new context
		newWithTimeoutTestCase(ctx, "positive duration", time.Millisecond),
		newWithTimeoutTestCase(ctx, "large duration", time.Hour),
		newWithTimeoutTestCase(expiring, "parent deadline later", time.Millisecond),
		newWithTimeoutTestCaseClamped(expiring, "parent deadline sooner", time.Hour),
		newWithTimeoutTestCase(nilCtx, "nil context", time.Millisecond),
		// The parent back
		newWithTimeoutTestCaseParent(ctx, "zero duration", 0),
		newWithTimeoutTestCaseParent(ctx, "negative duration", -time.Second),
		newWithTimeoutTestCaseParent(expiring, "zero duration, parent deadline", 0),
		newWithTimeoutTestCaseBackground("zero duration, nil context", 0),
	)
}

func testWithTimeoutExpiration(t *testing.T) {
	t.Helper()

	ctx, cancel := WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	AssertMustClosed(t, ctx.Done(), 100*time.Millisecond, "done")
	AssertEqual(t, context.DeadlineExceeded, ctx.Err(), "timeout error")
}

func testWithTimeoutCancellation(t *testing.T) {
	t.Helper()

	ctx, cancel := WithTimeout(context.Background(), time.Hour)

	AssertOpen(t, ctx.Done(), 0, "done before cancel")
	cancel()
	AssertClosed(t, ctx.Done(), 0, "done after cancel")
}

func TestWithTimeout(t *testing.T) {
	RunTestCases(t, withTimeoutTestCases(t))

	// Test timeout expiration
	t.Run("timeout expiration", testWithTimeoutExpiration)

	// Test cancellation
	t.Run("cancellation", testWithTimeoutCancellation)
}

func newWithTimeoutCauseTestCase(parent context.Context, name string,
	timeout time.Duration, cause error) TestCase {
	return withTimeoutNewTestCase[withTimeoutCauseArgs]{
		call: newWithTimeoutCauseArgs(parent, name, timeout, cause),
	}
}

func newWithTimeoutCauseTestCaseClamped(parent context.Context, name string,
	timeout time.Duration, cause error) TestCase {
	return withTimeoutClampedTestCase[withTimeoutCauseArgs]{
		call: newWithTimeoutCauseArgs(parent, name, timeout, cause),
	}
}

func newWithTimeoutCauseTestCaseParent(parent context.Context, name string,
	timeout time.Duration, cause error) TestCase {
	return withTimeoutParentTestCase[withTimeoutCauseArgs]{
		want: parent,
		call: newWithTimeoutCauseArgs(parent, name, timeout, cause),
	}
}

func newWithTimeoutCauseTestCaseBackground(name string, timeout time.Duration,
	cause error) TestCase {
	var nilCtx context.Context
	return withTimeoutParentTestCase[withTimeoutCauseArgs]{
		want: context.Background(),
		call: newWithTimeoutCauseArgs(nilCtx, name, timeout, cause),
	}
}

func withTimeoutCauseTestCases(t *testing.T) []TestCase {
	t.Helper()

	var nilCtx context.Context
	ctx := context.Background()
	expiring := expiringContext(t)
	cause := errors.New("test cause")

	return S[TestCase](
		// A new context
		newWithTimeoutCauseTestCase(ctx, "positive duration", time.Millisecond, cause),
		newWithTimeoutCauseTestCase(ctx, "large duration", time.Hour, cause),
		newWithTimeoutCauseTestCase(expiring, "parent deadline later", time.Millisecond, cause),
		newWithTimeoutCauseTestCaseClamped(expiring, "parent deadline sooner", time.Hour, cause),
		newWithTimeoutCauseTestCase(nilCtx, "nil context", time.Millisecond, cause),
		newWithTimeoutCauseTestCase(ctx, "nil cause", time.Millisecond, nil),
		// The parent back
		newWithTimeoutCauseTestCaseParent(ctx, "zero duration", 0, cause),
		newWithTimeoutCauseTestCaseParent(ctx, "negative duration", -time.Second, cause),
		newWithTimeoutCauseTestCaseParent(expiring, "zero duration, parent deadline", 0, cause),
		newWithTimeoutCauseTestCaseBackground("zero duration, nil context", 0, cause),
	)
}

func testWithTimeoutCauseExpiration(t *testing.T) {
	t.Helper()

	testErr := errors.New("custom timeout cause")
	ctx, cancel := WithTimeoutCause(context.Background(), 10*time.Millisecond, testErr)
	defer cancel()

	AssertMustClosed(t, ctx.Done(), 100*time.Millisecond, "done")
	AssertEqual(t, context.DeadlineExceeded, ctx.Err(), "timeout error")
	AssertSame(t, testErr, context.Cause(ctx), "cause")
}

func TestWithTimeoutCause(t *testing.T) {
	RunTestCases(t, withTimeoutCauseTestCases(t))

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
