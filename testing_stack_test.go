package core_test

import (
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = assertTopFrameTestCase{}

// assertTopFrameTestCase exercises AssertTopFrame and its Must form on
// the same value. A row states whether the assertion passes, and the
// essence of the failure it reports where it does not.
type assertTopFrameTestCase struct {
	v        any
	name     string
	want     string
	failure  string
	wantPass bool
}

// newAssertTopFrameTestCase declares a row the assertion passes.
func newAssertTopFrameTestCase(name string, v any,
	want string) assertTopFrameTestCase {
	return assertTopFrameTestCase{
		v:        v,
		name:     name,
		want:     want,
		wantPass: true,
	}
}

// newAssertTopFrameTestCaseFails declares a row the assertion fails,
// with the essence of the failure it reports.
func newAssertTopFrameTestCaseFails(name string, v any, want,
	failure string) assertTopFrameTestCase {
	return assertTopFrameTestCase{
		v:        v,
		name:     name,
		want:     want,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertTopFrameTestCase) Name() string {
	return tc.name
}

func (tc assertTopFrameTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &core.MockT{}
	ok := core.AssertTopFrame(mock, tc.v, tc.want, "top frame")

	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		core.AssertMustTopFrame(mt, tc.v, tc.want, "top frame")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertTopFrame")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertTopFrame")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
}

// assertTopFrameTestCases lists the values AssertTopFrame passes, then
// the ones it fails. NewPanicPayloadForTest builds a *PanicError whose
// stack starts in NewPanicPayloadForTest itself; a zero PanicError
// carries no stack.
func assertTopFrameTestCases() []assertTopFrameTestCase {
	return core.S(
		newAssertTopFrameTestCase("PanicError",
			core.NewPanicPayloadForTest(), "NewPanicPayloadForTest"),
		newAssertTopFrameTestCase("Catch of a plain panic",
			core.Catch(core.PanicPlainForTest), "PanicPlainForTest"),

		newAssertTopFrameTestCaseFails("other top frame",
			core.NewPanicPayloadForTest(), "callPanic",
			`got "NewPanicPayloadForTest"`),
		newAssertTopFrameTestCaseFails("nil",
			nil, "NewPanicPayloadForTest", "got nil"),
		newAssertTopFrameTestCaseFails("nil PanicError",
			(*core.PanicError)(nil), "NewPanicPayloadForTest", "got nil"),
		newAssertTopFrameTestCaseFails("not a CallStacker",
			core.ErrSentinelForTest, "NewPanicPayloadForTest", "expected a CallStacker"),
		newAssertTopFrameTestCaseFails("empty stack",
			&core.PanicError{}, "NewPanicPayloadForTest", "empty call stack"),
	)
}

func TestAssertTopFrame(t *testing.T) {
	core.RunTestCases(t, assertTopFrameTestCases())
}
