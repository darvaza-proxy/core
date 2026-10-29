package core

import (
	"context"
	"errors"
	"testing"
)

var _ TestCase = errGroupSetDefaultsTestCase{}

type errGroupSetDefaultsTestCase struct {
	expectedParent context.Context
	setup          func() *ErrGroup
	name           string
}

// newErrGroupSetDefaultsTestCase creates a new errGroupSetDefaultsTestCase
func newErrGroupSetDefaultsTestCase(expectedParent context.Context, name string,
	setup func() *ErrGroup) errGroupSetDefaultsTestCase {
	return errGroupSetDefaultsTestCase{
		name:           name,
		setup:          setup,
		expectedParent: expectedParent,
	}
}

func errGroupSetDefaultsTestCases() []errGroupSetDefaultsTestCase {
	var nilCtx context.Context

	return []errGroupSetDefaultsTestCase{
		newErrGroupSetDefaultsTestCase(context.Background(), "nil parent context", func() *ErrGroup {
			return &ErrGroup{}
		}),
		newErrGroupSetDefaultsTestCase(nilCtx, "custom parent context", func() *ErrGroup {
			type testKey string
			ctx := context.WithValue(context.Background(), testKey("test"), "value")
			return &ErrGroup{Parent: ctx}
		}), // Will be set by test
	}
}

func (tc errGroupSetDefaultsTestCase) Name() string {
	return tc.name
}

func (tc errGroupSetDefaultsTestCase) Test(t *testing.T) {
	t.Helper()

	eg := tc.setup()
	expectedParent := tc.expectedParent
	if tc.name == "custom parent context" {
		expectedParent = eg.Parent
	}

	eg.SetDefaults()

	if eg.Parent != expectedParent {
		t.Errorf("Expected Parent %v, got %v", expectedParent, eg.Parent)
	}

	AssertNotNil(t, eg.ctx, "ctx")
	AssertNotNil(t, eg.cancel, "cancel")
}

func TestErrGroupSetDefaults(t *testing.T) {
	RunTestCases(t, errGroupSetDefaultsTestCases())
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
