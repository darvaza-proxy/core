package core_test

import (
	"context"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = assertEventuallyTestCase{}
	_ core.TestCase = assertEventuallyContextTestCase{}
	_ core.TestCase = assertClosedTestCase{}
	_ core.TestCase = assertOpenTestCase{}
	_ core.TestCase = assertQuietTestCase{}
	_ core.TestCase = assertReceivesTestCase{}
)

// ms lives in testutils_test.go, and the budgets, the predicate
// factories and the context setups in sync_testutils_test.go.

// assertEventuallyTestCase exercises AssertEventually and its Must form
// on the same predicate. A row states whether the assertion passes, and
// the essence of the failure it reports where it does not.
type assertEventuallyTestCase struct {
	cond     func() func() bool
	name     string
	failure  string
	budgetMS int
	wantPass bool
}

// newAssertEventuallyTestCase declares a row the assertion passes.
func newAssertEventuallyTestCase(name string, cond func() func() bool,
	budgetMS int) assertEventuallyTestCase {
	return assertEventuallyTestCase{
		name:     name,
		cond:     cond,
		budgetMS: budgetMS,
		wantPass: true,
	}
}

// newAssertEventuallyTestCaseFails declares a row the assertion fails,
// with the essence of the failure it reports.
func newAssertEventuallyTestCaseFails(name string, cond func() func() bool,
	budgetMS int, failure string) assertEventuallyTestCase {
	return assertEventuallyTestCase{
		name:     name,
		cond:     cond,
		budgetMS: budgetMS,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertEventuallyTestCase) Name() string {
	return tc.name
}

func (tc assertEventuallyTestCase) Test(t *testing.T) {
	budget := ms(tc.budgetMS)

	mock := &core.MockT{}
	ok := core.AssertEventually(mock, tc.cond(), budget, "eventually")

	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		core.AssertMustEventually(mt, tc.cond(), budget, "eventually")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertEventually")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertEventually")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
}

// assertEventuallyTestCases lists the predicates AssertEventually passes,
// then the calls it fails. A nil predicate and a negative timeout are
// call-site mistakes, named before the wait.
func assertEventuallyTestCases() []assertEventuallyTestCase {
	return core.S(
		newAssertEventuallyTestCase("true immediately", alwaysTrue, shortBudgetMS),
		newAssertEventuallyTestCase("becomes true within budget", flagFlip, shortBudgetMS),

		newAssertEventuallyTestCaseFails("never true within budget", alwaysFalse,
			tinyBudgetMS, "not true within"),
		newAssertEventuallyTestCaseFails("nil predicate", noPredicate,
			tinyBudgetMS, "nil predicate"),
		newAssertEventuallyTestCaseFails("negative timeout", alwaysTrue,
			-1, "negative timeout"),
	)
}

func TestAssertEventually(t *testing.T) {
	core.RunTestCases(t, assertEventuallyTestCases())
}

// assertEventuallyContextTestCase exercises AssertEventuallyContext and
// its Must form on the same context setup and predicate. A row states
// whether the assertion passes, and the essence of the failure it reports
// where it does not.
type assertEventuallyContextTestCase struct {
	ctx      func(*testing.T) context.Context
	cond     func() func() bool
	name     string
	failure  string
	wantPass bool
}

// newAssertEventuallyContextTestCase declares a row the assertion passes.
func newAssertEventuallyContextTestCase(name string, ctx func(*testing.T) context.Context,
	cond func() func() bool) assertEventuallyContextTestCase {
	return assertEventuallyContextTestCase{
		name:     name,
		ctx:      ctx,
		cond:     cond,
		wantPass: true,
	}
}

// newAssertEventuallyContextTestCaseFails declares a row the assertion
// fails, with the essence of the failure it reports.
func newAssertEventuallyContextTestCaseFails(name string, ctx func(*testing.T) context.Context,
	cond func() func() bool, failure string) assertEventuallyContextTestCase {
	return assertEventuallyContextTestCase{
		name:     name,
		ctx:      ctx,
		cond:     cond,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertEventuallyContextTestCase) Name() string {
	return tc.name
}

func (tc assertEventuallyContextTestCase) Test(t *testing.T) {
	mock := &core.MockT{}
	ok := core.AssertEventuallyContext(mock, tc.ctx(t), tc.cond(), "eventually")

	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		core.AssertMustEventuallyContext(mt, tc.ctx(t), tc.cond(), "eventually")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertEventuallyContext")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertEventuallyContext")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
}

// assertEventuallyContextTestCases lists the calls AssertEventuallyContext
// passes, then the ones it fails. The nil-context row carries a predicate
// that becomes true, since nothing else ends that wait.
func assertEventuallyContextTestCases() []assertEventuallyContextTestCase {
	return core.S(
		newAssertEventuallyContextTestCase("true immediately",
			ctxWithBudget(shortBudgetMS), alwaysTrue),
		newAssertEventuallyContextTestCase("becomes true before the deadline",
			ctxWithBudget(shortBudgetMS), flagFlip),
		newAssertEventuallyContextTestCase("nil context",
			ctxNil, flagFlip),

		newAssertEventuallyContextTestCaseFails("never true before the deadline",
			ctxWithBudget(tinyBudgetMS), alwaysFalse, "deadline exceeded"),
		newAssertEventuallyContextTestCaseFails("cancelled before the call",
			ctxCancelled, alwaysFalse, "context canceled"),
		newAssertEventuallyContextTestCaseFails("nil predicate",
			ctxWithBudget(tinyBudgetMS), noPredicate, "nil predicate"),
	)
}

func TestAssertEventuallyContext(t *testing.T) {
	core.RunTestCases(t, assertEventuallyContextTestCases())
}

// channelBudgetMS is the budget every channel row runs under. A settled
// subject resolves at once whatever the budget, so the rows that wait for
// the deadline are the only ones it paces, and it keeps them quick.
const channelBudgetMS = tinyBudgetMS

// assertClosedTestCase exercises AssertClosed and its Must form on the
// same channel setup. A row states whether the assertion passes, and the
// essence of the failure it reports where it does not. Every call gets a
// fresh channel, since a receive would change what the next call sees.
type assertClosedTestCase struct {
	setup    func() <-chan int
	name     string
	failure  string
	wantPass bool
}

// newAssertClosedTestCase declares a row the assertion passes.
func newAssertClosedTestCase(name string, setup func() <-chan int) assertClosedTestCase {
	return assertClosedTestCase{
		name:     name,
		setup:    setup,
		wantPass: true,
	}
}

// newAssertClosedTestCaseFails declares a row the assertion fails, with
// the essence of the failure it reports.
func newAssertClosedTestCaseFails(name string, setup func() <-chan int,
	failure string) assertClosedTestCase {
	return assertClosedTestCase{
		name:     name,
		setup:    setup,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertClosedTestCase) Name() string {
	return tc.name
}

func (tc assertClosedTestCase) Test(t *testing.T) {
	budget := ms(channelBudgetMS)

	mock := &core.MockT{}
	ok := core.AssertClosed(mock, tc.setup(), budget, "closed")

	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		core.AssertMustClosed(mt, tc.setup(), budget, "closed")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertClosed")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertClosed")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
}

// assertClosedTestCases lists the channels AssertClosed passes, then the
// ones it fails. Values never settle the verdict: a closed channel passes
// whatever it held, and the report counts what was consumed on the way.
func assertClosedTestCases() []assertClosedTestCase {
	return core.S(
		newAssertClosedTestCase("closed", core.ChanClosedForTest),
		newAssertClosedTestCase("closed behind a value", core.ChanClosedWithSendsForTest(1)),
		newAssertClosedTestCase("closed behind values", core.ChanClosedWithSendsForTest(3)),

		newAssertClosedTestCaseFails("value sent", core.ChanWithSendsForTest(1), "1 value consumed"),
		newAssertClosedTestCaseFails("left open", core.ChanOpenForTest, "still open after"),
		newAssertClosedTestCaseFails("nil channel", core.ChanNilForTest, "still open after"),
	)
}

func TestAssertClosed(t *testing.T) {
	core.RunTestCases(t, assertClosedTestCases())
}

// assertOpenTestCase exercises AssertOpen and its Must form on the same
// channel setup. A row states whether the assertion passes, and the
// essence of the failure it reports where it does not.
type assertOpenTestCase struct {
	setup    func() <-chan int
	name     string
	failure  string
	wantPass bool
}

// newAssertOpenTestCase declares a row the assertion passes.
func newAssertOpenTestCase(name string, setup func() <-chan int) assertOpenTestCase {
	return assertOpenTestCase{
		name:     name,
		setup:    setup,
		wantPass: true,
	}
}

// newAssertOpenTestCaseFails declares a row the assertion fails, with the
// essence of the failure it reports.
func newAssertOpenTestCaseFails(name string, setup func() <-chan int,
	failure string) assertOpenTestCase {
	return assertOpenTestCase{
		name:     name,
		setup:    setup,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertOpenTestCase) Name() string {
	return tc.name
}

func (tc assertOpenTestCase) Test(t *testing.T) {
	budget := ms(channelBudgetMS)

	mock := &core.MockT{}
	ok := core.AssertOpen(mock, tc.setup(), budget, "open")

	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		core.AssertMustOpen(mt, tc.setup(), budget, "open")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertOpen")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertOpen")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
}

// assertOpenTestCases lists the channels AssertOpen passes, then the ones
// it fails. An open channel stays open whatever it held, and a closed one
// fails with a count of what was consumed on the way.
func assertOpenTestCases() []assertOpenTestCase {
	return core.S(
		newAssertOpenTestCase("value sent", core.ChanWithSendsForTest(1)),
		newAssertOpenTestCase("left open", core.ChanOpenForTest),
		newAssertOpenTestCase("nil channel", core.ChanNilForTest),

		newAssertOpenTestCaseFails("closed", core.ChanClosedForTest, "closed after"),
		newAssertOpenTestCaseFails("closed behind a value", core.ChanClosedWithSendsForTest(1),
			"1 value consumed"),
		newAssertOpenTestCaseFails("closed behind values", core.ChanClosedWithSendsForTest(3),
			"3 values consumed"),
	)
}

func TestAssertOpen(t *testing.T) {
	core.RunTestCases(t, assertOpenTestCases())
}

// assertQuietTestCase exercises AssertQuiet and its Must form on the same
// channel setup. A row states whether the assertion passes, and the
// essence of the failure it reports where it does not.
type assertQuietTestCase struct {
	setup    func() <-chan int
	name     string
	failure  string
	wantPass bool
}

// newAssertQuietTestCase declares a row the assertion passes.
func newAssertQuietTestCase(name string, setup func() <-chan int) assertQuietTestCase {
	return assertQuietTestCase{
		name:     name,
		setup:    setup,
		wantPass: true,
	}
}

// newAssertQuietTestCaseFails declares a row the assertion fails, with the
// essence of the failure it reports.
func newAssertQuietTestCaseFails(name string, setup func() <-chan int,
	failure string) assertQuietTestCase {
	return assertQuietTestCase{
		name:     name,
		setup:    setup,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertQuietTestCase) Name() string {
	return tc.name
}

func (tc assertQuietTestCase) Test(t *testing.T) {
	budget := ms(channelBudgetMS)

	mock := &core.MockT{}
	ok := core.AssertQuiet(mock, tc.setup(), budget, "quiet")

	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		core.AssertMustQuiet(mt, tc.setup(), budget, "quiet")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertQuiet")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertQuiet")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
}

// assertQuietTestCases lists the channels AssertQuiet passes, then the ones
// it fails. The first event of either kind is the failure, and a value is
// named rather than counted: where several wait, the report names the
// first and leaves the rest unread.
func assertQuietTestCases() []assertQuietTestCase {
	return core.S(
		newAssertQuietTestCase("left open", core.ChanOpenForTest),
		newAssertQuietTestCase("nil channel", core.ChanNilForTest),

		newAssertQuietTestCaseFails("closed", core.ChanClosedForTest, "closed after"),
		newAssertQuietTestCaseFails("value sent", core.ChanWithSendsForTest(1), "received 1 after"),
		newAssertQuietTestCaseFails("values sent, first taken", core.ChanFirstTakenForTest,
			"received 2 after"),
		newAssertQuietTestCaseFails("closed behind a value", core.ChanClosedWithSendsForTest(1),
			"received 1 after"),
		newAssertQuietTestCaseFails("closed behind values", core.ChanClosedWithSendsForTest(3),
			"received 1 after"),
	)
}

func TestAssertQuiet(t *testing.T) {
	core.RunTestCases(t, assertQuietTestCases())
}

// AssertQuiet reads the event it fails on and nothing past it, which is
// where it parts from AssertOpen.
func TestAssertQuietStopsAtFirstEvent(t *testing.T) {
	ch := core.ChanWithSendsForTest(3)()

	mock := &core.MockT{}
	ok := core.AssertQuiet(mock, ch, ms(channelBudgetMS), "quiet")

	assertFailed(t, mock, ok, "received 1 after", "AssertQuiet")
	core.AssertEqual(t, 2, len(ch), "left unread")
}

// assertReceivesTestCase exercises AssertReceives and its Must form on
// the same channel setup. A row states the values handed back, whether
// the assertion passes, and the essence of the failure it reports where
// it does not. Every call gets a fresh channel, since the assertion
// drains what it receives.
type assertReceivesTestCase struct {
	setup    func() <-chan int
	name     string
	failure  string
	got      []int
	n        int
	wantPass bool
}

// newAssertReceivesTestCase declares a row the assertion passes.
func newAssertReceivesTestCase(name string, setup func() <-chan int,
	n int, got []int) assertReceivesTestCase {
	return assertReceivesTestCase{
		name:     name,
		setup:    setup,
		n:        n,
		got:      got,
		wantPass: true,
	}
}

// newAssertReceivesTestCaseFails declares a row the assertion fails, with
// what it collected on the way and the essence of the failure it reports.
func newAssertReceivesTestCaseFails(name string, setup func() <-chan int,
	n int, got []int, failure string) assertReceivesTestCase {
	return assertReceivesTestCase{
		name:     name,
		setup:    setup,
		n:        n,
		got:      got,
		failure:  failure,
		wantPass: false,
	}
}

func (tc assertReceivesTestCase) Name() string {
	return tc.name
}

func (tc assertReceivesTestCase) Test(t *testing.T) {
	budget := ms(channelBudgetMS)

	mock := &core.MockT{}
	got, ok := core.AssertReceives(mock, tc.setup(), tc.n, budget, "ready")
	core.AssertSliceEqual(t, tc.got, got, "AssertReceives values")

	var mustGot []int
	mustMock := &core.MockT{}
	mustOK := mustMock.Run("must", func(mt core.T) {
		mustGot = core.AssertMustReceives(mt, tc.setup(), tc.n, budget, "ready")
		mt.Log(core.MustContinuationLogForTest)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertReceives")
		core.AssertMustAbortedForTest(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertReceives")
	core.AssertMustContinuedForTest(t, mustMock, mustOK)
	core.AssertSliceEqual(t, tc.got, mustGot, "AssertMustReceives values")
}

// assertReceivesTestCases lists the channels AssertReceives passes, then
// the ones it fails, each with the values it hands back. Every subject is
// settled before the call: a row waiting on another goroutine can miss its
// deadline on a busy machine.
func assertReceivesTestCases() []assertReceivesTestCase {
	return core.S(
		newAssertReceivesTestCase("immediate full", core.ChanWithSendsForTest(4), 4, core.S(1, 2, 3, 4)),
		newAssertReceivesTestCase("zero expected", core.ChanWithSendsForTest(0), 0, nil),

		newAssertReceivesTestCaseFails("immediate partial", core.ChanWithSendsForTest(3), 4,
			core.S(1, 2, 3), "3 of 4 values within"),
		newAssertReceivesTestCaseFails("closed before n", core.ChanClosedForTest, 3,
			nil, "closed after 0 of 3"),
		newAssertReceivesTestCaseFails("nil channel", core.ChanNilForTest, 1,
			nil, "0 of 1 values within"),
		newAssertReceivesTestCaseFails("negative expected", core.ChanNilForTest, -1,
			nil, "negative count -1"),
	)
}

func TestAssertReceives(t *testing.T) {
	core.RunTestCases(t, assertReceivesTestCases())
}
