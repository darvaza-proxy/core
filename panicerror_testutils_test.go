package core_test

import "darvaza.org/core"

// assertUnreachablePanicShape pins that r is a *PanicError whose chain
// assertUnreachableChain accepts. The PanicShape tests for the
// MustNoError family share it, as their separate PanicUnreachableFrom
// calls take identical arguments, so a regression in one would hit the
// others identically. r is the raw recovered value, as assertTopFrameIs
// takes it: AsRecovered would wrap a non-Recovered panic in a
// *PanicError, leaving the type assertion to state what the line above
// it had just built.
func assertUnreachablePanicShape(t core.T, r any, want error) {
	t.Helper()
	core.AssertMustNotNil(t, r, "recovered value")
	pe := core.AssertMustTypeIs[*core.PanicError](t, r,
		"recovered value is *PanicError")
	assertUnreachableChain(t, pe, want)
}

// assertUnreachableChain pins that pe chains to both ErrUnreachable and
// want. What the payload is assembled from is NewUnreachableError's
// business; what callers depend on is that [errors.Is] finds each of the
// two, so that is what this asserts. ErrUnrelatedForTest pins the chain
// as discriminating rather than matching anything put to it.
func assertUnreachableChain(t core.T, pe *core.PanicError, want error) {
	t.Helper()
	core.AssertErrorIs(t, pe, core.ErrUnreachable, "ErrUnreachable in chain")
	core.AssertErrorIs(t, pe, want, "original error in chain")
	core.AssertNotErrorIs(t, pe, core.ErrUnrelatedForTest, "unrelated error absent")
}

// assertPlainUnreachableShape pins the no-cause panic shape:
// NewUnreachableError treats ErrUnreachable as "no distinct cause", so
// the *PanicError payload is the bare sentinel rather than a pairing of
// the sentinel with a cause. Identity is the assertion here, not
// [errors.Is]: matching the sentinel would pass for either shape, and
// the distinction between them is the whole point. Shared by the
// UnreachablePanicShape tests for MustNoError, MustNoErrorExcept and
// MustNoErrorExceptFn so their structural pins stay symmetric. r is
// the raw recovered value, for the reason assertUnreachablePanicShape
// gives.
func assertPlainUnreachableShape(t core.T, r any) {
	t.Helper()
	core.AssertMustNotNil(t, r, "recovered value")
	pe := core.AssertMustTypeIs[*core.PanicError](t, r,
		"recovered value is *PanicError")
	core.AssertMustSame(t, core.ErrUnreachable, pe.Unwrap(),
		"payload is ErrUnreachable")
}

// assertTopFrameIs takes a recovered panic value, asserts it is a
// *PanicError, returned for further checks, and pins:
//
//   - the top frame is wantFunc, through AssertMustTopFrame — catches
//     a stack that lands inside the helper itself (skip too small) or
//     skips past the wrapper (skip too large);
//   - the captured stack has at least minDepth frames — guards
//     against a future change that truncates the stack and makes
//     the top-frame assertion vacuous.
//
// Both assertions are Must-* so a violation halts the test via
// Fatal rather than logging silently. The t parameter is T (not
// *testing.T) so a MockT can drive the negative meta-test rows in
// TestAssertTopFrameIs.
func assertTopFrameIs(t core.T, r any, wantFunc string, minDepth int) *core.PanicError {
	t.Helper()
	core.AssertMustNotNil(t, r, "recovered value")
	pe := core.AssertMustTypeIs[*core.PanicError](t, r,
		"recovered value is *PanicError")
	stack := pe.CallStack()
	core.AssertMustTrue(t, len(stack) >= minDepth,
		"stack depth >= %d (got %d)", minDepth, len(stack))
	core.AssertMustTopFrame(t, pe, wantFunc, "top frame")
	return pe
}
