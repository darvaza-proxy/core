package core

import "testing"

var _ TestCase = assertTopFrameTestCase{}

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

	mock := &MockT{}
	ok := AssertTopFrame(mock, tc.v, tc.want, "top frame")

	mustMock := &MockT{}
	mustOK := mustMock.Run("must", func(mt T) {
		AssertMustTopFrame(mt, tc.v, tc.want, "top frame")
		mt.Log(mustContinuationLog)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertTopFrame")
		assertMustAborted(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertTopFrame")
	assertMustContinued(t, mustMock, mustOK)
}

// assertTopFrameTestCases lists the values AssertTopFrame passes, then
// the ones it fails. NewPanicPayloadForTest builds a *PanicError whose
// stack starts in NewPanicPayloadForTest itself; a zero PanicError
// carries no stack.
func assertTopFrameTestCases() []assertTopFrameTestCase {
	return S(
		newAssertTopFrameTestCase("PanicError",
			NewPanicPayloadForTest(), "NewPanicPayloadForTest"),
		newAssertTopFrameTestCase("Catch of a plain panic",
			Catch(PanicPlainForTest), "PanicPlainForTest"),

		newAssertTopFrameTestCaseFails("other top frame",
			NewPanicPayloadForTest(), "callPanic",
			`got "NewPanicPayloadForTest"`),
		newAssertTopFrameTestCaseFails("nil",
			nil, "NewPanicPayloadForTest", "got nil"),
		newAssertTopFrameTestCaseFails("nil PanicError",
			(*PanicError)(nil), "NewPanicPayloadForTest", "got nil"),
		newAssertTopFrameTestCaseFails("not a CallStacker",
			ErrSentinelForTest, "NewPanicPayloadForTest", "expected a CallStacker"),
		newAssertTopFrameTestCaseFails("empty stack",
			&PanicError{}, "NewPanicPayloadForTest", "empty call stack"),
	)
}

func TestAssertTopFrame(t *testing.T) {
	RunTestCases(t, assertTopFrameTestCases())
}
