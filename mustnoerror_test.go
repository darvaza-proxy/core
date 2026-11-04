package core_test

import (
	"errors"
	"fmt"
	"testing"

	"darvaza.org/core"
)

var (
	errOther = errors.New("other error")
	errExtra = errors.New("extra error")

	// errNilPointer is a non-nil error interface holding a nil
	// pointer — the typed-nil shape MustNoError* treats as a real
	// error.
	errNilPointer error = (*nilPointerError)(nil)
)

// nilPointerError exists to build errNilPointer; Error never
// dereferences its receiver, so calling it on the nil pointer is
// safe.
type nilPointerError struct{}

func (*nilPointerError) Error() string { return "nil pointer error" }

// matchesSentinel, matchesOther and matchesNilPointer are stable
// predicates for the MustNoErrorExceptFn tests, both the data-table
// rows and the standalone no-match tests; named (not inline
// closures) so the call sites stay easy to scan.
func matchesSentinel(err error) bool   { return err == core.ErrSentinelForTest }
func matchesOther(err error) bool      { return err == errOther }
func matchesNilPointer(err error) bool { return err == errNilPointer }

// mustNoErrorTestCase exercises MustNoError across the nil and
// non-nil err paths. The non-nil rows also pin that the wrapped
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
	fn := func() { core.MustNoError(tc.err) }
	if !tc.wantPanic {
		core.AssertNoPanic(t, fn, "MustNoError nil")
		return
	}
	core.AssertPanic(t, fn, core.ErrUnreachable,
		"panic chains to ErrUnreachable")
}

var _ core.TestCase = mustNoErrorTestCase{}

func mustNoErrorTestCases() []mustNoErrorTestCase {
	return []mustNoErrorTestCase{
		newMustNoErrorTestCase("nil err returns",
			nil, false),
		newMustNoErrorTestCase("non-nil err panics",
			core.ErrSentinelForTest, true),
		newMustNoErrorTestCase("nil pointer err panics",
			errNilPointer, true),
	}
}

// TestMustNoError exercises MustNoError across its panic and
// no-panic paths.
func TestMustNoError(t *testing.T) {
	core.RunTestCases(t, mustNoErrorTestCases())
}

// TestMustNoErrorPreservesOriginal verifies the panic value chains
// to the original error as well as ErrUnreachable. The two
// assertions taken together pin that the helper does not lose the
// underlying error on its way to the panic. The typed-nil assertion
// pins that a typed nil survives into the chain unchanged too.
func TestMustNoErrorPreservesOriginal(t *testing.T) {
	fn := func() { core.MustNoError(core.ErrSentinelForTest) }
	core.AssertPanic(t, fn, core.ErrUnreachable, "ErrUnreachable in chain")
	core.AssertPanic(t, fn, core.ErrSentinelForTest, "original err in chain")

	fnNil := func() { core.MustNoError(errNilPointer) }
	core.AssertPanic(t, fnNil, errNilPointer, "typed-nil err in chain")
}

// TestMustNoErrorPanicShape pins what MustNoError panics with on its
// non-nil path: a *PanicError chaining to both ErrUnreachable and the
// original error, asserted by assertUnreachablePanicShape.
func TestMustNoErrorPanicShape(t *testing.T) {
	defer func() {
		assertUnreachablePanicShape(t, recover(), core.ErrSentinelForTest)
	}()
	core.MustNoError(core.ErrSentinelForTest)
}

// TestMustNoErrorUnreachablePanicShape pins
// MustNoError(ErrUnreachable): the sentinel as guarded err carries no
// distinct cause, so NewUnreachableError normalises it away and the
// panic takes the plain no-cause shape.
func TestMustNoErrorUnreachablePanicShape(t *testing.T) {
	defer func() {
		assertPlainUnreachableShape(t, recover())
	}()
	core.MustNoError(core.ErrUnreachable)
}

// TestMustNoErrorExceptPreservesOriginal mirrors
// TestMustNoErrorPreservesOriginal for the no-match panic path of
// MustNoErrorExcept: when none of the allowed errors match, the
// helper must still surface the original err in the panic chain.
func TestMustNoErrorExceptPreservesOriginal(t *testing.T) {
	fn := func() {
		core.MustNoErrorExcept(core.ErrSentinelForTest, errOther, errExtra)
	}
	core.AssertPanic(t, fn, core.ErrUnreachable, "ErrUnreachable in chain")
	core.AssertPanic(t, fn, core.ErrSentinelForTest, "original err in chain")
}

// TestMustNoErrorExceptPanicShape mirrors TestMustNoErrorPanicShape
// for the no-match path of MustNoErrorExcept.
func TestMustNoErrorExceptPanicShape(t *testing.T) {
	defer func() {
		assertUnreachablePanicShape(t, recover(), core.ErrSentinelForTest)
	}()
	core.MustNoErrorExcept(core.ErrSentinelForTest, errOther)
}

// TestMustNoErrorExceptUnreachablePanicShape mirrors
// TestMustNoErrorUnreachablePanicShape for the no-match path of
// MustNoErrorExcept: ErrUnreachable as the guarded err takes the
// plain no-cause shape here too.
func TestMustNoErrorExceptUnreachablePanicShape(t *testing.T) {
	defer func() {
		assertPlainUnreachableShape(t, recover())
	}()
	core.MustNoErrorExcept(core.ErrUnreachable, errOther)
}

// mustNoErrorExceptTestCase exercises MustNoErrorExcept across the
// nil/non-nil err axis and the allowed-list match outcomes. Together
// the rows verify the helper's three guard branches behave as
// documented.
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
	fn := func() { core.MustNoErrorExcept(tc.err, tc.allowed...) }
	if !tc.wantPanic {
		core.AssertNoPanic(t, fn, "MustNoErrorExcept allowed")
		return
	}
	core.AssertPanic(t, fn, core.ErrUnreachable,
		"panic chains to ErrUnreachable")
}

var _ core.TestCase = mustNoErrorExceptTestCase{}

func mustNoErrorExceptTestCases() []mustNoErrorExceptTestCase {
	return []mustNoErrorExceptTestCase{
		newMustNoErrorExceptTestCase("nil err no allowed",
			nil, nil, false),
		newMustNoErrorExceptTestCase("nil err with allowed",
			nil, []error{core.ErrSentinelForTest}, false),
		newMustNoErrorExceptTestCase("non-nil empty allowed panics",
			core.ErrSentinelForTest, nil, true),
		newMustNoErrorExceptTestCase("non-nil direct match",
			core.ErrSentinelForTest, []error{core.ErrSentinelForTest}, false),
		newMustNoErrorExceptTestCase("non-nil match second allowed",
			core.ErrSentinelForTest, []error{errOther, core.ErrSentinelForTest}, false),
		newMustNoErrorExceptTestCase("non-nil no match panics",
			core.ErrSentinelForTest, []error{errOther, errExtra}, true),
		newMustNoErrorExceptTestCase("non-nil compound match",
			&core.CompoundError{Errs: []error{errOther, core.ErrSentinelForTest}},
			[]error{core.ErrSentinelForTest}, false),
		newMustNoErrorExceptTestCase("non-nil wrapped match",
			fmt.Errorf("context: %w", core.ErrSentinelForTest),
			[]error{core.ErrSentinelForTest}, false),
		newMustNoErrorExceptTestCase("nil pointer no match panics",
			errNilPointer, []error{errOther}, true),
		newMustNoErrorExceptTestCase("nil pointer identity match",
			errNilPointer, []error{errNilPointer}, false),
		newMustNoErrorExceptTestCase("nil-only allowed panics",
			core.ErrSentinelForTest, []error{nil}, true),
		newMustNoErrorExceptTestCase("nil entry beside match",
			core.ErrSentinelForTest, []error{nil, core.ErrSentinelForTest}, false),
	}
}

// TestMustNoErrorExcept exercises MustNoErrorExcept across the
// nil/non-nil and allowed-list match outcomes.
func TestMustNoErrorExcept(t *testing.T) {
	core.RunTestCases(t, mustNoErrorExceptTestCases())
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
	fn := func() { core.MustNoErrorExceptFn(tc.err, tc.check) }
	if !tc.wantPanic {
		core.AssertNoPanic(t, fn, "MustNoErrorExceptFn allowed")
		return
	}
	core.AssertPanic(t, fn, core.ErrUnreachable,
		"panic chains to ErrUnreachable")
}

var _ core.TestCase = mustNoErrorExceptFnTestCase{}

func mustNoErrorExceptFnTestCases() []mustNoErrorExceptFnTestCase {
	return []mustNoErrorExceptFnTestCase{
		newMustNoErrorExceptFnTestCase("nil err nil check",
			nil, nil, false),
		newMustNoErrorExceptFnTestCase("nil err with check",
			nil, matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("non-nil nil check panics",
			core.ErrSentinelForTest, nil, true),
		newMustNoErrorExceptFnTestCase("non-nil direct match",
			core.ErrSentinelForTest, matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("non-nil no match panics",
			core.ErrSentinelForTest, matchesOther, true),
		newMustNoErrorExceptFnTestCase("non-nil compound match",
			&core.CompoundError{Errs: []error{errOther, core.ErrSentinelForTest}},
			matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("non-nil wrapped match",
			fmt.Errorf("context: %w", core.ErrSentinelForTest),
			matchesSentinel, false),
		newMustNoErrorExceptFnTestCase("nil pointer no match panics",
			errNilPointer, matchesOther, true),
		newMustNoErrorExceptFnTestCase("nil pointer predicate match",
			errNilPointer, matchesNilPointer, false),
	}
}

// TestMustNoErrorExceptFn exercises MustNoErrorExceptFn across the
// nil/non-nil and predicate match outcomes.
func TestMustNoErrorExceptFn(t *testing.T) {
	core.RunTestCases(t, mustNoErrorExceptFnTestCases())
}

// TestMustNoErrorExceptFnPreservesOriginal mirrors
// TestMustNoErrorExceptPreservesOriginal for the no-match panic path
// of MustNoErrorExceptFn: when the predicate accepts nothing, the
// helper must still surface the original err in the panic chain.
func TestMustNoErrorExceptFnPreservesOriginal(t *testing.T) {
	fn := func() {
		core.MustNoErrorExceptFn(core.ErrSentinelForTest, matchesOther)
	}
	core.AssertPanic(t, fn, core.ErrUnreachable, "ErrUnreachable in chain")
	core.AssertPanic(t, fn, core.ErrSentinelForTest, "original err in chain")
}

// TestMustNoErrorExceptFnPanicShape mirrors TestMustNoErrorPanicShape
// for the no-match path of MustNoErrorExceptFn.
func TestMustNoErrorExceptFnPanicShape(t *testing.T) {
	defer func() {
		assertUnreachablePanicShape(t, recover(), core.ErrSentinelForTest)
	}()
	core.MustNoErrorExceptFn(core.ErrSentinelForTest, matchesOther)
}

// TestMustNoErrorExceptFnUnreachablePanicShape mirrors
// TestMustNoErrorUnreachablePanicShape for the no-match path of
// MustNoErrorExceptFn: ErrUnreachable as the guarded err takes the
// plain no-cause shape here too.
func TestMustNoErrorExceptFnUnreachablePanicShape(t *testing.T) {
	defer func() {
		assertPlainUnreachableShape(t, recover())
	}()
	core.MustNoErrorExceptFn(core.ErrUnreachable, matchesOther)
}

// callMustNoError is a thin wrapper around MustNoError used as a
// stable, named caller for the stack-skip verification tests. The
// captured top frame of the panic value should resolve to this
// function, not to MustNoError or PanicUnreachableFrom.
func callMustNoError(err error) {
	core.MustNoError(err)
}

// callMustNoErrorExcept is the MustNoErrorExcept counterpart of
// callMustNoError.
func callMustNoErrorExcept(err error, allowed ...error) {
	core.MustNoErrorExcept(err, allowed...)
}

// callMustNoErrorExceptFn is the MustNoErrorExceptFn counterpart of
// callMustNoError.
func callMustNoErrorExceptFn(err error, check func(error) bool) {
	core.MustNoErrorExceptFn(err, check)
}

// TestMustNoErrorStack verifies the captured call stack of the panic
// value lands at the immediate caller (callMustNoError) and not
// inside MustNoError or PanicUnreachableFrom. Pins the skip=1
// argument to PanicUnreachableFrom. The minDepth of 2 ensures the
// wrapper sits above at least the test caller — a future change that
// truncates the stack to a single frame would fail here.
func TestMustNoErrorStack(t *testing.T) {
	defer func() {
		_ = assertTopFrameIs(t, recover(), "callMustNoError", 2)
	}()
	callMustNoError(core.ErrSentinelForTest)
}

// TestMustNoErrorExceptStack verifies MustNoErrorExcept's panic
// stack lands at the immediate caller, the same way as MustNoError.
func TestMustNoErrorExceptStack(t *testing.T) {
	defer func() {
		_ = assertTopFrameIs(t, recover(), "callMustNoErrorExcept", 2)
	}()
	callMustNoErrorExcept(core.ErrSentinelForTest, errOther)
}

// TestMustNoErrorExceptFnStack verifies MustNoErrorExceptFn's panic
// stack lands at the immediate caller, the same way as MustNoError.
func TestMustNoErrorExceptFnStack(t *testing.T) {
	defer func() {
		_ = assertTopFrameIs(t, recover(), "callMustNoErrorExceptFn", 2)
	}()
	callMustNoErrorExceptFn(core.ErrSentinelForTest, matchesOther)
}

// recoverValidPanic invokes callMustNoError and returns the
// recovered panic value. Named (not an anonymous IIFE) so its frame
// appears predictably in the captured stack and so callMustNoError
// remains the top frame after skip=1. The final return nil is
// unreachable in practice (callMustNoError always panics) but
// required by Go's flow analysis and the project's bare-return lint.
func recoverValidPanic() (r any) {
	defer func() { r = recover() }()
	callMustNoError(core.ErrSentinelForTest)
	return nil
}

// runAssertTopFrameIs runs assertTopFrameIs against a MockT and
// swallows the panic raised by MockT.Fatal/FailNow. Returns what the
// helper returned, nil when it halted, and the MockT so the caller can
// assert Failed(). Named returns are required so the MockT survives a
// recovered panic in the deferred handler — without them the function
// would return a nil *MockT after the recover.
func runAssertTopFrameIs(r any, wantFunc string,
	minDepth int) (pe *core.PanicError, mock *core.MockT) {
	mock = &core.MockT{}
	defer func() { _ = recover() }()
	pe = assertTopFrameIs(mock, r, wantFunc, minDepth)
	return pe, mock
}

// assertTopFrameIsTestCase exercises assertTopFrameIs across both the
// happy path (valid input — wantFailed false) and each of the
// precondition violations (wantFailed true). The positive row pins
// that the helper does not spuriously fail; the negative rows pin
// that each Must-* check actually halts the MockT via Fatal when
// violated. Without the positive row a regression like
// "assertTopFrameIs always calls Fatal" would pass the negative rows
// silently. A passing row also states that the helper returns the
// *PanicError it was given, and a failing one that it returns nothing.
type assertTopFrameIsTestCase struct {
	name string

	r        any
	wantFunc string
	minDepth int

	wantFailed bool
}

func newAssertTopFrameIsTestCase(name string, r any, wantFunc string,
	minDepth int, wantFailed bool) assertTopFrameIsTestCase {
	return assertTopFrameIsTestCase{
		name:       name,
		r:          r,
		wantFunc:   wantFunc,
		minDepth:   minDepth,
		wantFailed: wantFailed,
	}
}

func (tc assertTopFrameIsTestCase) Name() string { return tc.name }

func (tc assertTopFrameIsTestCase) Test(t *testing.T) {
	t.Helper()
	pe, mock := runAssertTopFrameIs(tc.r, tc.wantFunc, tc.minDepth)
	core.AssertMustEqual(t, tc.wantFailed, mock.Failed(),
		"MockT.Failed()")
	if tc.wantFailed {
		core.AssertNil(t, pe, "returned")
	} else {
		core.AssertSame(t, tc.r, pe, "returned")
	}
}

var _ core.TestCase = assertTopFrameIsTestCase{}

func assertTopFrameIsTestCases(validPanic any) []assertTopFrameIsTestCase {
	return []assertTopFrameIsTestCase{
		newAssertTopFrameIsTestCase("valid input passes",
			validPanic, "callMustNoError", 2, false),
		newAssertTopFrameIsTestCase("nil recovered",
			nil, "callMustNoError", 2, true),
		newAssertTopFrameIsTestCase("non-PanicError type",
			errOther, "callMustNoError", 2, true),
		newAssertTopFrameIsTestCase("wrong frame name",
			validPanic, "nonExistent", 2, true),
		newAssertTopFrameIsTestCase("depth too shallow",
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
	core.AssertMustNotNil(t, validPanic, "captured valid panic")
	core.RunTestCases(t, assertTopFrameIsTestCases(validPanic))
}
