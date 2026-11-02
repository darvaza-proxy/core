package core

import (
	"context"
	"errors"
	"testing"
)

var _ TestCase = errGroupSetDefaultsTestCase{}

// errGroupTestKey is the key a custom parent carries its value under, read
// back through the group's context.
type errGroupTestKey struct{}

type errGroupSetDefaultsTestCase struct {
	parent         context.Context
	expectedParent context.Context
	expectedValue  any
	name           string
}

// newErrGroupSetDefaultsTestCase declares a row: the Parent the group
// starts with, the Parent SetDefaults leaves it with, and the value the
// group's context then holds under errGroupTestKey.
func newErrGroupSetDefaultsTestCase(parent, expectedParent context.Context,
	expectedValue any, name string) errGroupSetDefaultsTestCase {
	return errGroupSetDefaultsTestCase{
		name:           name,
		parent:         parent,
		expectedParent: expectedParent,
		expectedValue:  expectedValue,
	}
}

func errGroupSetDefaultsTestCases() []errGroupSetDefaultsTestCase {
	var nilCtx context.Context
	custom := context.WithValue(context.Background(), errGroupTestKey{}, "value")

	return []errGroupSetDefaultsTestCase{
		newErrGroupSetDefaultsTestCase(nilCtx, context.Background(), nil, "nil parent context"),
		newErrGroupSetDefaultsTestCase(custom, custom, "value", "custom parent context"),
	}
}

func (tc errGroupSetDefaultsTestCase) Name() string {
	return tc.name
}

func (tc errGroupSetDefaultsTestCase) Test(t *testing.T) {
	t.Helper()

	eg := &ErrGroup{Parent: tc.parent}
	eg.SetDefaults()

	AssertEqual(t, tc.expectedParent, eg.Parent, "parent")
	AssertMustNotNil(t, eg.ctx, "ctx")
	AssertMustNotNil(t, eg.cancel, "cancel")
	AssertEqual(t, tc.expectedValue, eg.ctx.Value(errGroupTestKey{}), "value through ctx")

	cause := errors.New("cause")
	eg.cancel(cause)
	AssertSame(t, cause, context.Cause(eg.ctx), "cause")
}

func TestErrGroupSetDefaults(t *testing.T) {
	RunTestCases(t, errGroupSetDefaultsTestCases())
}

// TestErrGroupSetDefaultsTwice states that a second SetDefaults keeps the
// context the first one built.
func TestErrGroupSetDefaultsTwice(t *testing.T) {
	var eg ErrGroup
	eg.SetDefaults()
	ctx := eg.ctx
	AssertMustNotNil(t, ctx, "first ctx")

	eg.SetDefaults()
	AssertSame(t, ctx, eg.ctx, "ctx")
}

func testErrGroupCatcherErrorWhenNotCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)

	AssertSame(t, testErr, result, "error instance")
}

func testErrGroupCatcherErrorWhenCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup
	eg.Cancel(errors.New("cancellation error"))

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)
	AssertSame(t, context.Canceled, result, "error instance")
}

func testErrGroupCatcherNilError(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	result := eg.defaultErrGroupCatcher(nil)
	AssertNil(t, result, "result")
}

func TestErrGroupDefaultErrGroupCatcher(t *testing.T) {
	t.Run("error when not cancelled", testErrGroupCatcherErrorWhenNotCancelled)
	t.Run("error when cancelled", testErrGroupCatcherErrorWhenCancelled)
	t.Run("nil error", testErrGroupCatcherNilError)
}
