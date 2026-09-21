package core

import (
	"context"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = awaitCloseAtDeadlineTestCase{}
	_ TestCase = receiveNAtDeadlineTestCase{}
	_ TestCase = assertEventuallyTestCase{}
	_ TestCase = assertEventuallyContextTestCase{}
	_ TestCase = assertClosedTestCase{}
	_ TestCase = assertOpenTestCase{}
	_ TestCase = assertReceivesTestCase{}
)

// atDeadlineRuns is how many times each at-deadline row runs. With both
// select arms ready the runtime picks one at random, and this many runs
// make missing the deadline arm every time as good as impossible.
const atDeadlineRuns = 16

// ms, the budgets, the predicate factories and the context setups live in
// sync_test.go.

// The channel setups carry int values, numbered from 1 in send order, so a
// row can declare what an assertion hands back.

// chanOpen returns a channel nothing will close or send to.
func chanOpen() <-chan int {
	return make(chan int)
}

// chanClosed returns an already closed channel.
func chanClosed() <-chan int {
	ch := make(chan int)
	close(ch)
	return ch
}

// chanNil returns a nil channel, which never closes.
func chanNil() <-chan int {
	return nil
}

// chanWithSends returns a setup that fills a buffered channel with n values
// before handing it over.
func chanWithSends(n int) func() <-chan int {
	return func() <-chan int {
		ch := make(chan int, n)
		for i := range n {
			ch <- i + 1
		}
		return ch
	}
}

// chanClosedWithSends returns a setup that fills a buffered channel with n
// values and closes it, so the close sits behind the values.
func chanClosedWithSends(n int) func() <-chan int {
	return func() <-chan int {
		ch := make(chan int, n)
		for i := range n {
			ch <- i + 1
		}
		close(ch)
		return ch
	}
}

// closedDeadline returns a deadline that has already fired.
func closedDeadline() <-chan time.Time {
	ch := make(chan time.Time)
	close(ch)
	return ch
}

// awaitCloseAtDeadlineTestCase exercises awaitClose with a deadline that
// has already fired, so both select arms are ready: whatever ch holds is
// consumed, and a close behind it still counts.
type awaitCloseAtDeadlineTestCase struct {
	setup  func() <-chan int
	name   string
	count  int
	closed bool
}

func newAwaitCloseAtDeadlineTestCase(name string, setup func() <-chan int,
	count int, closed bool) awaitCloseAtDeadlineTestCase {
	return awaitCloseAtDeadlineTestCase{
		name:   name,
		setup:  setup,
		count:  count,
		closed: closed,
	}
}

func (tc awaitCloseAtDeadlineTestCase) Name() string {
	return tc.name
}

func (tc awaitCloseAtDeadlineTestCase) Test(t *testing.T) {
	t.Helper()
	for range atDeadlineRuns {
		count, closed := awaitClose(tc.setup(), closedDeadline())
		AssertEqual(t, tc.count, count, "count")
		AssertEqual(t, tc.closed, closed, "closed")
	}
}

func TestAwaitCloseAtDeadline(t *testing.T) {
	RunTestCases(t, S(
		newAwaitCloseAtDeadlineTestCase("closed", chanClosed, 0, true),
		newAwaitCloseAtDeadlineTestCase("closed behind a value", chanClosedWithSends(1), 1, true),
		newAwaitCloseAtDeadlineTestCase("closed behind values", chanClosedWithSends(3), 3, true),
		newAwaitCloseAtDeadlineTestCase("value sent", chanWithSends(1), 1, false),
		newAwaitCloseAtDeadlineTestCase("values sent", chanWithSends(3), 3, false),
		newAwaitCloseAtDeadlineTestCase("left open", chanOpen, 0, false),
		newAwaitCloseAtDeadlineTestCase("nil channel", chanNil, 0, false),
	))
}

// receiveNAtDeadlineTestCase exercises receiveN with a deadline that has
// already fired: whatever ch holds is taken, and no more.
type receiveNAtDeadlineTestCase struct {
	setup  func() <-chan int
	name   string
	got    []int
	n      int
	closed bool
}

func newReceiveNAtDeadlineTestCase(name string, setup func() <-chan int,
	n int, got []int, closed bool) receiveNAtDeadlineTestCase {
	return receiveNAtDeadlineTestCase{
		name:   name,
		setup:  setup,
		n:      n,
		got:    got,
		closed: closed,
	}
}

func (tc receiveNAtDeadlineTestCase) Name() string {
	return tc.name
}

func (tc receiveNAtDeadlineTestCase) Test(t *testing.T) {
	t.Helper()
	for range atDeadlineRuns {
		got, closed := receiveN(tc.setup(), tc.n, closedDeadline())
		AssertSliceEqual(t, tc.got, got, "got")
		AssertEqual(t, tc.closed, closed, "closed")
	}
}

func TestReceiveNAtDeadline(t *testing.T) {
	RunTestCases(t, S(
		newReceiveNAtDeadlineTestCase("all buffered", chanWithSends(3), 3, S(1, 2, 3), false),
		newReceiveNAtDeadlineTestCase("some buffered", chanWithSends(2), 4, S(1, 2), false),
		newReceiveNAtDeadlineTestCase("closed", chanClosed, 1, nil, true),
		newReceiveNAtDeadlineTestCase("closed behind enough", chanClosedWithSends(2), 2, S(1, 2), false),
		newReceiveNAtDeadlineTestCase("closed behind too few", chanClosedWithSends(3), 4, S(1, 2, 3), true),
		newReceiveNAtDeadlineTestCase("nil channel", chanNil, 2, nil, false),
	))
}

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
	t.Helper()
	budget := ms(tc.budgetMS)

	mock := &MockT{}
	ok := AssertEventually(mock, tc.cond(), budget, "eventually")

	mustMock := &MockT{}
	mustOK := mustMock.Run("must", func(mt T) {
		AssertMustEventually(mt, tc.cond(), budget, "eventually")
		mt.Log(mustContinuationLog)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertEventually")
		assertMustAborted(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertEventually")
	assertMustContinued(t, mustMock, mustOK)
}

// assertEventuallyTestCases lists the predicates AssertEventually passes,
// then the calls it fails. A nil predicate and a negative timeout are
// call-site mistakes, named before the wait.
func assertEventuallyTestCases() []assertEventuallyTestCase {
	return S(
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
	RunTestCases(t, assertEventuallyTestCases())
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
	t.Helper()

	mock := &MockT{}
	ok := AssertEventuallyContext(mock, tc.ctx(t), tc.cond(), "eventually")

	mustMock := &MockT{}
	mustOK := mustMock.Run("must", func(mt T) {
		AssertMustEventuallyContext(mt, tc.ctx(t), tc.cond(), "eventually")
		mt.Log(mustContinuationLog)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertEventuallyContext")
		assertMustAborted(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertEventuallyContext")
	assertMustContinued(t, mustMock, mustOK)
}

// assertEventuallyContextTestCases lists the calls AssertEventuallyContext
// passes, then the ones it fails. The nil-context row carries a predicate
// that becomes true, since nothing else ends that wait.
func assertEventuallyContextTestCases() []assertEventuallyContextTestCase {
	return S(
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
	RunTestCases(t, assertEventuallyContextTestCases())
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
	t.Helper()
	budget := ms(channelBudgetMS)

	mock := &MockT{}
	ok := AssertClosed(mock, tc.setup(), budget, "closed")

	mustMock := &MockT{}
	mustOK := mustMock.Run("must", func(mt T) {
		AssertMustClosed(mt, tc.setup(), budget, "closed")
		mt.Log(mustContinuationLog)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertClosed")
		assertMustAborted(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertClosed")
	assertMustContinued(t, mustMock, mustOK)
}

// assertClosedTestCases lists the channels AssertClosed passes, then the
// ones it fails. Values never settle the verdict: a closed channel passes
// whatever it held, and the report counts what was consumed on the way.
func assertClosedTestCases() []assertClosedTestCase {
	return S(
		newAssertClosedTestCase("closed", chanClosed),
		newAssertClosedTestCase("closed behind a value", chanClosedWithSends(1)),
		newAssertClosedTestCase("closed behind values", chanClosedWithSends(3)),

		newAssertClosedTestCaseFails("value sent", chanWithSends(1), "1 value consumed"),
		newAssertClosedTestCaseFails("left open", chanOpen, "still open after"),
		newAssertClosedTestCaseFails("nil channel", chanNil, "still open after"),
	)
}

func TestAssertClosed(t *testing.T) {
	RunTestCases(t, assertClosedTestCases())
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
	t.Helper()
	budget := ms(channelBudgetMS)

	mock := &MockT{}
	ok := AssertOpen(mock, tc.setup(), budget, "open")

	mustMock := &MockT{}
	mustOK := mustMock.Run("must", func(mt T) {
		AssertMustOpen(mt, tc.setup(), budget, "open")
		mt.Log(mustContinuationLog)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertOpen")
		assertMustAborted(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertOpen")
	assertMustContinued(t, mustMock, mustOK)
}

// assertOpenTestCases lists the channels AssertOpen passes, then the ones
// it fails. An open channel stays open whatever it held, and a closed one
// fails with a count of what was consumed on the way.
func assertOpenTestCases() []assertOpenTestCase {
	return S(
		newAssertOpenTestCase("value sent", chanWithSends(1)),
		newAssertOpenTestCase("left open", chanOpen),
		newAssertOpenTestCase("nil channel", chanNil),

		newAssertOpenTestCaseFails("closed", chanClosed, "closed after"),
		newAssertOpenTestCaseFails("closed behind a value", chanClosedWithSends(1),
			"1 value consumed"),
		newAssertOpenTestCaseFails("closed behind values", chanClosedWithSends(3),
			"3 values consumed"),
	)
}

func TestAssertOpen(t *testing.T) {
	RunTestCases(t, assertOpenTestCases())
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
	t.Helper()
	budget := ms(channelBudgetMS)

	mock := &MockT{}
	got, ok := AssertReceives(mock, tc.setup(), tc.n, budget, "ready")
	AssertSliceEqual(t, tc.got, got, "AssertReceives values")

	var mustGot []int
	mustMock := &MockT{}
	mustOK := mustMock.Run("must", func(mt T) {
		mustGot = AssertMustReceives(mt, tc.setup(), tc.n, budget, "ready")
		mt.Log(mustContinuationLog)
	})

	if !tc.wantPass {
		assertFailed(t, mock, ok, tc.failure, "AssertReceives")
		assertMustAborted(t, mustMock, mustOK)
		return
	}

	assertPassed(t, mock, ok, "AssertReceives")
	assertMustContinued(t, mustMock, mustOK)
	AssertSliceEqual(t, tc.got, mustGot, "AssertMustReceives values")
}

// assertReceivesTestCases lists the channels AssertReceives passes, then
// the ones it fails, each with the values it hands back. Every subject is
// settled before the call: a row waiting on another goroutine can miss its
// deadline on a busy machine.
func assertReceivesTestCases() []assertReceivesTestCase {
	return S(
		newAssertReceivesTestCase("immediate full", chanWithSends(4), 4, S(1, 2, 3, 4)),
		newAssertReceivesTestCase("zero expected", chanWithSends(0), 0, nil),

		newAssertReceivesTestCaseFails("immediate partial", chanWithSends(3), 4,
			S(1, 2, 3), "3 of 4 values within"),
		newAssertReceivesTestCaseFails("closed before n", chanClosed, 3,
			nil, "closed after 0 of 3"),
		newAssertReceivesTestCaseFails("nil channel", chanNil, 1,
			nil, "0 of 1 values within"),
		newAssertReceivesTestCaseFails("negative expected", chanNil, -1,
			nil, "negative count -1"),
	)
}

func TestAssertReceives(t *testing.T) {
	RunTestCases(t, assertReceivesTestCases())
}
