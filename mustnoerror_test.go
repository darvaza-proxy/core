package core

import (
	"errors"
	"fmt"
	"testing"
)

var (
	errSentinel = errors.New("sentinel error")
	errOther    = errors.New("other error")
	errExtra    = errors.New("extra error")
)

// matchesSentinel and matchesOther are stable predicates for the
// MustNoErrorExceptFn rows; named (not inline closures) so the data
// table stays easy to scan.
func matchesSentinel(err error) bool { return err == errSentinel }
func matchesOther(err error) bool    { return err == errOther }

// mustNoErrorTestCase exercises MustNoError across the nil and
// non-nil err paths. The non-nil row also pins that the wrapped
// panic value chains to ErrUnreachable, so a future change in the
// panic shape fails loudly.
type mustNoErrorTestCase struct {
	err  error
	name string

	wantPanic bool
}

func newMustNoErrorTestCase(name string, err error,
	wantPanic bool) mustNoErrorTestCase {
	return mustNoErrorTestCase{name: name, err: err, wantPanic: wantPanic}
}

func (tc mustNoErrorTestCase) Name() string { return tc.name }

func (tc mustNoErrorTestCase) Test(t *testing.T) {
	t.Helper()
	fn := func() { MustNoError(tc.err) }
	if !tc.wantPanic {
		AssertNoPanic(t, fn, "MustNoError nil")
		return
	}
	AssertPanic(t, fn, ErrUnreachable,
		"panic chains to ErrUnreachable")
}

var _ TestCase = mustNoErrorTestCase{}

func mustNoErrorTestCases() []mustNoErrorTestCase {
	return []mustNoErrorTestCase{
		newMustNoErrorTestCase("nil err returns",
			nil, false),
		newMustNoErrorTestCase("non-nil err panics",
			errSentinel, true),
	}
}

// TestMustNoError exercises MustNoError across its panic and
// no-panic paths.
func TestMustNoError(t *testing.T) {
	RunTestCases(t, mustNoErrorTestCases())
}

// TestMustNoErrorPreservesOriginal verifies the panic value chains
// to the original error as well as ErrUnreachable. The two
// assertions taken together pin that the helper does not lose the
// underlying error on its way to the panic.
func TestMustNoErrorPreservesOriginal(t *testing.T) {
	fn := func() { MustNoError(errSentinel) }
	AssertPanic(t, fn, ErrUnreachable, "ErrUnreachable in chain")
	AssertPanic(t, fn, errSentinel, "original err in chain")
}

// assertUnreachablePanicShape pins that r is a *PanicError whose
// Unwrap() yields a *CompoundError holding exactly two errors,
// [ErrUnreachable, want]. Shared by the PanicShape tests for
// MustNoError, MustNoErrorExcept and MustNoErrorExceptFn so their
// structural pins stay symmetric — all three helpers route through
// the same NewUnreachableError call, so a regression in one would
// hit the others identically. r is normalised through AsRecovered
// so a nil (no-panic) input fails the first assertion with a clear
// message.
func assertUnreachablePanicShape(t T, r any, want error) {
	t.Helper()
	rec := AsRecovered(r)
	AssertMustNotNil(t, rec, "recovered value")
	pe := AssertMustTypeIs[*PanicError](t, rec,
		"recovered is *PanicError")
	ce := AssertMustTypeIs[*CompoundError](t, pe.Unwrap(),
		"payload is *CompoundError")
	AssertMustEqual(t, 2, len(ce.Errs),
		"CompoundError holds 2 errors")
	AssertMustSame(t, ErrUnreachable, ce.Errs[0],
		"first error is ErrUnreachable")
	AssertMustSame(t, want, ce.Errs[1],
		"second error is original")
}

// TestMustNoErrorPanicShape pins the structural shape of the panic
// value: a *PanicError whose Unwrap() yields a CompoundError holding
// [ErrUnreachable, original]. A future change to NewUnreachableError
// that flattens or reorders the chain will fail this test.
func TestMustNoErrorPanicShape(t *testing.T) {
	defer func() {
		assertUnreachablePanicShape(t, recover(), errSentinel)
	}()
	MustNoError(errSentinel)
}

// TestMustNoErrorExceptPreservesOriginal mirrors
// TestMustNoErrorPreservesOriginal for the no-match panic path of
// MustNoErrorExcept: when none of the allowed errors match, the
// helper must still surface the original err in the panic chain.
func TestMustNoErrorExceptPreservesOriginal(t *testing.T) {
	fn := func() {
		MustNoErrorExcept(errSentinel, errOther, errExtra)
	}
	AssertPanic(t, fn, ErrUnreachable, "ErrUnreachable in chain")
	AssertPanic(t, fn, errSentinel, "original err in chain")
}

// TestMustNoErrorExceptPanicShape mirrors TestMustNoErrorPanicShape
// for the no-match path of MustNoErrorExcept. Structural shape of
// the panic value is identical because both helpers route through
// the same NewUnreachableError call.
func TestMustNoErrorExceptPanicShape(t *testing.T) {
	defer func() {
		assertUnreachablePanicShape(t, recover(), errSentinel)
	}()
	MustNoErrorExcept(errSentinel, errOther)
}

// mustNoErrorExceptTestCase exercises MustNoErrorExcept across the
// nil/non-nil err axis and the allowed-list match outcomes. Each
// row pins both the panic outcome and the case to verify the
// helper's three guard branches behave as documented.
type mustNoErrorExceptTestCase struct {
	name string

	err     error
	allowed []error

	wantPanic bool
}

func newMustNoErrorExceptTestCase(name string, err error,
	allowed []error, wantPanic bool) mustNoErrorExceptTestCase {
	return mustNoErrorExceptTestCase{
		name:      name,
		err:       err,
		allowed:   allowed,
		wantPanic: wantPanic,
	}
}

func (tc mustNoErrorExceptTestCase) Name() string { return tc.name }

func (tc mustNoErrorExceptTestCase) Test(t *testing.T) {
	t.Helper()
	fn := func() { MustNoErrorExcept(tc.err, tc.allowed...) }
	if !tc.wantPanic {
		AssertNoPanic(t, fn, "MustNoErrorExcept allowed")
		return
	}
	AssertPanic(t, fn, ErrUnreachable,
		"panic chains to ErrUnreachable")
}

var _ TestCase = mustNoErrorExceptTestCase{}

func mustNoErrorExceptTestCases() []mustNoErrorExceptTestCase {
	return []mustNoErrorExceptTestCase{
		newMustNoErrorExceptTestCase("nil err no allowed",
			nil, nil, false),
		newMustNoErrorExceptTestCase("nil err with allowed",
			nil, []error{errSentinel}, false),
		newMustNoErrorExceptTestCase("non-nil empty allowed panics",
			errSentinel, nil, true),
		newMustNoErrorExceptTestCase("non-nil direct match",
			errSentinel, []error{errSentinel}, false),
		newMustNoErrorExceptTestCase("non-nil match second allowed",
			errSentinel, []error{errOther, errSentinel}, false),
		newMustNoErrorExceptTestCase("non-nil no match panics",
			errSentinel, []error{errOther, errExtra}, true),
		newMustNoErrorExceptTestCase("non-nil compound match",
			&CompoundError{Errs: []error{errOther, errSentinel}},
			[]error{errSentinel}, false),
		newMustNoErrorExceptTestCase("non-nil wrapped match",
			fmt.Errorf("context: %w", errSentinel),
			[]error{errSentinel}, false),
	}
}

// TestMustNoErrorExcept exercises MustNoErrorExcept across the
// nil/non-nil and allowed-list match outcomes.
func TestMustNoErrorExcept(t *testing.T) {
	RunTestCases(t, mustNoErrorExceptTestCases())
}

// mustNoErrorExceptFnTestCase exercises MustNoErrorExceptFn across the
// nil/non-nil err axis and the predicate match outcomes, including the
// nil-check degenerate path. Recursive matching is covered by the
// wrapped and compound rows.
type mustNoErrorExceptFnTestCase struct {
	err   error
	check func(error) bool
	name  string

	wantPanic bool
}

func newMustNoErrorExceptFnTestCase(name string, err error,
	check func(error) bool, wantPanic bool) mustNoErrorExceptFnTestCase {
	return mustNoErrorExceptFnTestCase{
		name:      name,
		err:       err,
		check:     check,
		wantPanic: wantPanic,
	}
}

func (tc mustNoErrorExceptFnTestCase) Name() string { return tc.name }

func (tc mustNoErrorExceptFnTestCase) Test(t *testing.T) {
	t.Helper()
	fn := func() { MustNoErrorExceptFn(tc.err, tc.check) }
	if !tc.wantPanic {
		AssertNoPanic(t, fn, "MustNoErrorExceptFn allowed")
		return
	}
	AssertPanic(t, fn, ErrUnreachable,
		"panic chains to ErrUnreachable")
}

var _ TestCase = mustNoErrorExceptFnTestCase{}

func mustNoErrorExceptFnTestCases() []mustNoErrorExceptFnTestCase {
	return []mustNoErrorExceptFnTestCase{
		newMustNoErrorExceptFnTestCase("nil err nil check",
			nil, nil, false),
		newMustNoErrorExceptFnTestCase("nil err with check",
			nil, matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("non-nil nil check panics",
			errSentinel, nil, true),
		newMustNoErrorExceptFnTestCase("non-nil direct match",
			errSentinel, matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("non-nil no match panics",
			errSentinel, matchesOther, true),
		newMustNoErrorExceptFnTestCase("non-nil compound match",
			&CompoundError{Errs: []error{errOther, errSentinel}},
			matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("non-nil wrapped match",
			fmt.Errorf("context: %w", errSentinel),
			matchesSentinel, false),
	}
}

// TestMustNoErrorExceptFn exercises MustNoErrorExceptFn across the
// nil/non-nil and predicate match outcomes.
func TestMustNoErrorExceptFn(t *testing.T) {
	RunTestCases(t, mustNoErrorExceptFnTestCases())
}

// TestMustNoErrorExceptFnPreservesOriginal mirrors
// TestMustNoErrorExceptPreservesOriginal for the no-match panic path
// of MustNoErrorExceptFn: when the predicate accepts nothing, the
// helper must still surface the original err in the panic chain.
func TestMustNoErrorExceptFnPreservesOriginal(t *testing.T) {
	fn := func() {
		MustNoErrorExceptFn(errSentinel, matchesOther)
	}
	AssertPanic(t, fn, ErrUnreachable, "ErrUnreachable in chain")
	AssertPanic(t, fn, errSentinel, "original err in chain")
}

// TestMustNoErrorExceptFnPanicShape mirrors TestMustNoErrorPanicShape
// for the no-match path of MustNoErrorExceptFn. Structural shape of
// the panic value is identical because all three helpers route through
// the same NewUnreachableError call.
func TestMustNoErrorExceptFnPanicShape(t *testing.T) {
	defer func() {
		assertUnreachablePanicShape(t, recover(), errSentinel)
	}()
	MustNoErrorExceptFn(errSentinel, matchesOther)
}

// callMustNoError is a thin wrapper around MustNoError used as a
// stable, named caller for the stack-skip verification tests. The
// captured top frame of the panic value should resolve to this
// function, not to MustNoError or NewUnreachableError.
func callMustNoError(err error) {
	MustNoError(err)
}

// callMustNoErrorExcept is the MustNoErrorExcept counterpart of
// callMustNoError.
func callMustNoErrorExcept(err error, allowed ...error) {
	MustNoErrorExcept(err, allowed...)
}

// callMustNoErrorExceptFn is the MustNoErrorExceptFn counterpart of
// callMustNoError.
func callMustNoErrorExceptFn(err error, check func(error) bool) {
	MustNoErrorExceptFn(err, check)
}

// assertTopFrameIs recovers a panic, asserts the value is a
// *PanicError, and pins:
//
//   - the top frame's FuncName is wantFunc — catches a stack that
//     lands inside the helper itself (skip too small) or skips past
//     the wrapper (skip too large);
//   - the captured stack has at least minDepth frames — guards
//     against a future change that truncates the stack and makes
//     the top-frame assertion vacuous.
//
// Both assertions are Must-* so a violation halts the test via
// Fatal rather than logging silently. The t parameter is T (not
// *testing.T) so a MockT can drive the negative meta-test rows in
// TestAssertTopFrameIs.
func assertTopFrameIs(t T, r any, wantFunc string, minDepth int) {
	t.Helper()
	AssertMustNotNil(t, r, "recovered value")
	pe := AssertMustTypeIs[*PanicError](t, r,
		"recovered value is *PanicError")
	stack := pe.CallStack()
	AssertMustTrue(t, len(stack) >= minDepth,
		"stack depth >= %d (got %d)", minDepth, len(stack))
	AssertMustEqual(t, wantFunc, stack[0].FuncName(), "top frame")
}

// TestMustNoErrorStack verifies the captured call stack of the panic
// value lands at the immediate caller (callMustNoError) and not
// inside MustNoError or NewUnreachableError. Pins the skip=1
// argument to NewUnreachableError. The minDepth of 2 ensures the
// wrapper sits above at least the test caller — a future change that
// truncates the stack to a single frame would fail here.
func TestMustNoErrorStack(t *testing.T) {
	defer func() {
		assertTopFrameIs(t, recover(), "callMustNoError", 2)
	}()
	callMustNoError(errSentinel)
}

// TestMustNoErrorExceptStack verifies MustNoErrorExcept's panic
// stack lands at the immediate caller, the same way as MustNoError.
func TestMustNoErrorExceptStack(t *testing.T) {
	defer func() {
		assertTopFrameIs(t, recover(), "callMustNoErrorExcept", 2)
	}()
	callMustNoErrorExcept(errSentinel, errOther)
}

// TestMustNoErrorExceptFnStack verifies MustNoErrorExceptFn's panic
// stack lands at the immediate caller, the same way as MustNoError.
func TestMustNoErrorExceptFnStack(t *testing.T) {
	defer func() {
		assertTopFrameIs(t, recover(), "callMustNoErrorExceptFn", 2)
	}()
	callMustNoErrorExceptFn(errSentinel, matchesOther)
}

// recoverValidPanic invokes callMustNoError and returns the
// recovered panic value. Named (not an anonymous IIFE) so its frame
// appears predictably in the captured stack and so callMustNoError
// remains the top frame after skip=1. The final return nil is
// unreachable in practice (callMustNoError always panics) but
// required by Go's flow analysis and the project's bare-return lint.
func recoverValidPanic() (r any) {
	defer func() { r = recover() }()
	callMustNoError(errSentinel)
	return nil
}

// runAssertTopFrameIs runs assertTopFrameIs against a MockT and
// swallows the panic raised by MockT.Fatal/FailNow. Returns the
// MockT so the caller can assert Failed(). Named return is required
// so the MockT survives a recovered panic in the deferred handler —
// without it the function would return a nil *MockT after the
// recover.
func runAssertTopFrameIs(r any, wantFunc string,
	minDepth int) (mock *MockT) {
	mock = &MockT{}
	defer func() { _ = recover() }()
	assertTopFrameIs(mock, r, wantFunc, minDepth)
	return mock
}

// assertTopFrameTestCase exercises assertTopFrameIs across both the
// happy path (valid input — wantFailed false) and each of the
// precondition violations (wantFailed true). The positive row pins
// that the helper does not spuriously fail; the negative rows pin
// that each Must-* check actually halts the MockT via Fatal when
// violated. Without the positive row a regression like
// "assertTopFrameIs always calls Fatal" would pass the negative rows
// silently.
type assertTopFrameTestCase struct {
	name string

	r        any
	wantFunc string
	minDepth int

	wantFailed bool
}

func newAssertTopFrameTestCase(name string, r any, wantFunc string,
	minDepth int, wantFailed bool) assertTopFrameTestCase {
	return assertTopFrameTestCase{
		name:       name,
		r:          r,
		wantFunc:   wantFunc,
		minDepth:   minDepth,
		wantFailed: wantFailed,
	}
}

func (tc assertTopFrameTestCase) Name() string { return tc.name }

func (tc assertTopFrameTestCase) Test(t *testing.T) {
	t.Helper()
	mock := runAssertTopFrameIs(tc.r, tc.wantFunc, tc.minDepth)
	AssertMustEqual(t, tc.wantFailed, mock.Failed(),
		"MockT.Failed()")
}

var _ TestCase = assertTopFrameTestCase{}

func assertTopFrameTestCases(validPanic any) []assertTopFrameTestCase {
	return []assertTopFrameTestCase{
		newAssertTopFrameTestCase("valid input passes",
			validPanic, "callMustNoError", 2, false),
		newAssertTopFrameTestCase("nil recovered",
			nil, "callMustNoError", 2, true),
		newAssertTopFrameTestCase("non-PanicError type",
			errOther, "callMustNoError", 2, true),
		newAssertTopFrameTestCase("wrong frame name",
			validPanic, "nonExistent", 2, true),
		newAssertTopFrameTestCase("depth too shallow",
			validPanic, "callMustNoError", 9999, true),
	}
}

// TestAssertTopFrameIs is a meta-test pinning the full behaviour of
// assertTopFrameIs: valid input passes silently, each precondition
// violation halts the MockT via Fatal. Without the positive row a
// regression like "assertTopFrameIs always calls Fatal" would pass
// the negative rows; without the negative rows a regression that
// drops a check would pass the positive row.
func TestAssertTopFrameIs(t *testing.T) {
	validPanic := recoverValidPanic()
	AssertMustNotNil(t, validPanic, "captured valid panic")
	RunTestCases(t, assertTopFrameTestCases(validPanic))
}
