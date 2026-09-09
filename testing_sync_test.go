package core

import (
	"context"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = assertEventuallyTestCase{}
	_ TestCase = assertEventuallyContextTestCase{}
	_ TestCase = assertChannelTestCase{}
	_ TestCase = assertReceivesTestCase{}
	_ TestCase = awaitCloseAtDeadlineTestCase{}
	_ TestCase = receiveNAtDeadlineTestCase{}
)

// atDeadlineRuns is how many times each at-deadline row runs. With both
// select arms ready the runtime picks one at random, and this many runs
// make missing the deadline arm every time as good as impossible.
const atDeadlineRuns = 16

// ms, the budgets, the predicate factories and the context setups live in
// sync_test.go.

// assertOutcome is what a row declares for the assertion under test:
// whether it passes, and the essence of its failure message when it does
// not.
type assertOutcome struct {
	failure string
	pass    bool
}

func passes() assertOutcome {
	return assertOutcome{pass: true}
}

func failsWith(failure string) assertOutcome {
	return assertOutcome{failure: failure}
}

// TestAssertOutcome pins that pass is declared rather than derived from
// failure: a failure declared without an essence stays a failure.
func TestAssertOutcome(t *testing.T) {
	AssertTrue(t, passes().pass, "passes")
	AssertEqual(t, "", passes().failure, "passes essence")
	AssertFalse(t, failsWith("closed").pass, "fails")
	AssertEqual(t, "closed", failsWith("closed").failure, "fails essence")
	AssertFalse(t, failsWith("").pass, "fails without essence")
}

// checkOutcome checks a bare assertion's result and its mock against the
// declared outcome.
func checkOutcome(t *testing.T, mock *MockT, result bool, want assertOutcome, desc string) {
	t.Helper()
	if want.pass {
		assertPassed(t, mock, result, desc)
		return
	}
	assertFailed(t, mock, result, want.failure, desc)
}

// checkMustOutcome checks a Must assertion run through MockT.Run against
// the declared outcome.
func checkMustOutcome(t *testing.T, mock *MockT, result bool, want assertOutcome) {
	t.Helper()
	if want.pass {
		assertMustContinued(t, mock, result)
		return
	}
	assertMustAborted(t, mock, result)
}

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

// assertEventuallyTestCase exercises AssertEventually and its Must form on
// the same predicate.
type assertEventuallyTestCase struct {
	cond     func() func() bool
	name     string
	want     assertOutcome
	budgetMS int
}

func newAssertEventuallyTestCase(name string, cond func() func() bool,
	budgetMS int, want assertOutcome) assertEventuallyTestCase {
	return assertEventuallyTestCase{
		name:     name,
		cond:     cond,
		budgetMS: budgetMS,
		want:     want,
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
	checkOutcome(t, mock, ok, tc.want, "AssertEventually")

	mock = &MockT{}
	ok = mock.Run("must", func(mt T) {
		AssertMustEventually(mt, tc.cond(), budget, "eventually")
		mt.Log(mustContinuationLog)
	})
	checkMustOutcome(t, mock, ok, tc.want)
}

func TestAssertEventually(t *testing.T) {
	RunTestCases(t, S(
		newAssertEventuallyTestCase("true immediately", alwaysTrue, shortBudgetMS, passes()),
		newAssertEventuallyTestCase("becomes true within budget", flagFlip, shortBudgetMS, passes()),
		newAssertEventuallyTestCase("never true within budget", alwaysFalse, tinyBudgetMS,
			failsWith("not true within")),
		newAssertEventuallyTestCase("nil predicate", noPredicate, tinyBudgetMS,
			failsWith("nil predicate")),
		newAssertEventuallyTestCase("negative timeout", alwaysTrue, -1,
			failsWith("negative timeout")),
	))
}

// assertEventuallyContextTestCase exercises AssertEventuallyContext and its
// Must form on the same context setup and predicate.
type assertEventuallyContextTestCase struct {
	ctx  func(*testing.T) context.Context
	cond func() func() bool
	name string
	want assertOutcome
}

func newAssertEventuallyContextTestCase(name string, ctx func(*testing.T) context.Context,
	cond func() func() bool, want assertOutcome) assertEventuallyContextTestCase {
	return assertEventuallyContextTestCase{
		name: name,
		ctx:  ctx,
		cond: cond,
		want: want,
	}
}

func (tc assertEventuallyContextTestCase) Name() string {
	return tc.name
}

func (tc assertEventuallyContextTestCase) Test(t *testing.T) {
	t.Helper()

	mock := &MockT{}
	ok := AssertEventuallyContext(mock, tc.ctx(t), tc.cond(), "eventually")
	checkOutcome(t, mock, ok, tc.want, "AssertEventuallyContext")

	mock = &MockT{}
	ok = mock.Run("must", func(mt T) {
		AssertMustEventuallyContext(mt, tc.ctx(t), tc.cond(), "eventually")
		mt.Log(mustContinuationLog)
	})
	checkMustOutcome(t, mock, ok, tc.want)
}

func TestAssertEventuallyContext(t *testing.T) {
	// The nil-context row carries a predicate that becomes true, since
	// nothing else ends that wait.
	RunTestCases(t, S(
		newAssertEventuallyContextTestCase("true immediately", ctxWithBudget(shortBudgetMS),
			alwaysTrue, passes()),
		newAssertEventuallyContextTestCase("becomes true before the deadline",
			ctxWithBudget(shortBudgetMS), flagFlip, passes()),
		newAssertEventuallyContextTestCase("never true before the deadline",
			ctxWithBudget(tinyBudgetMS), alwaysFalse, failsWith("deadline exceeded")),
		newAssertEventuallyContextTestCase("cancelled before the call", ctxCancelled,
			alwaysFalse, failsWith("context canceled")),
		newAssertEventuallyContextTestCase("nil context", ctxNil, flagFlip, passes()),
		newAssertEventuallyContextTestCase("nil predicate", ctxWithBudget(tinyBudgetMS),
			noPredicate, failsWith("nil predicate")),
	))
}

// channelBudgetMS is the budget every channel row runs under. A settled
// subject resolves at once whatever the budget, so the rows that wait for
// the deadline are the only ones it paces, and it keeps them quick.
const channelBudgetMS = tinyBudgetMS

// assertChannelTestCase exercises AssertClosed and AssertOpen, with their
// Must forms, on the same channel setup, so each row declares both
// outcomes. Every call gets a fresh channel, since a receive would change
// what the next call sees.
type assertChannelTestCase struct {
	setup  func() <-chan int
	name   string
	closed assertOutcome
	open   assertOutcome
}

func newAssertChannelTestCase(name string, setup func() <-chan int,
	closed, open assertOutcome) assertChannelTestCase {
	return assertChannelTestCase{
		name:   name,
		setup:  setup,
		closed: closed,
		open:   open,
	}
}

func (tc assertChannelTestCase) Name() string {
	return tc.name
}

func (tc assertChannelTestCase) Test(t *testing.T) {
	t.Helper()
	tc.testClosed(t)
	tc.testOpen(t)
}

func (tc assertChannelTestCase) testClosed(t *testing.T) {
	t.Helper()
	budget := ms(channelBudgetMS)

	mock := &MockT{}
	ok := AssertClosed(mock, tc.setup(), budget, "closed")
	checkOutcome(t, mock, ok, tc.closed, "AssertClosed")

	mock = &MockT{}
	ok = mock.Run("must", func(mt T) {
		AssertMustClosed(mt, tc.setup(), budget, "closed")
		mt.Log(mustContinuationLog)
	})
	checkMustOutcome(t, mock, ok, tc.closed)
}

func (tc assertChannelTestCase) testOpen(t *testing.T) {
	t.Helper()
	budget := ms(channelBudgetMS)

	mock := &MockT{}
	ok := AssertOpen(mock, tc.setup(), budget, "open")
	checkOutcome(t, mock, ok, tc.open, "AssertOpen")

	mock = &MockT{}
	ok = mock.Run("must", func(mt T) {
		AssertMustOpen(mt, tc.setup(), budget, "open")
		mt.Log(mustContinuationLog)
	})
	checkMustOutcome(t, mock, ok, tc.open)
}

func TestAssertClosedAndOpen(t *testing.T) {
	// Values never settle the verdict: a closed channel passes whatever it
	// held, an open one stays open whatever it held, and the report counts
	// what was consumed on the way.
	RunTestCases(t, S(
		newAssertChannelTestCase("closed", chanClosed,
			passes(), failsWith("closed after")),
		newAssertChannelTestCase("closed behind a value", chanClosedWithSends(1),
			passes(), failsWith("1 value consumed")),
		newAssertChannelTestCase("closed behind values", chanClosedWithSends(3),
			passes(), failsWith("3 values consumed")),
		newAssertChannelTestCase("value sent", chanWithSends(1),
			failsWith("1 value consumed"), passes()),
		newAssertChannelTestCase("left open", chanOpen,
			failsWith("still open after"), passes()),
		newAssertChannelTestCase("nil channel", chanNil,
			failsWith("still open after"), passes()),
	))
}

// assertReceivesTestCase exercises AssertReceives and its Must form on the
// same channel setup, declaring the outcome and the values handed back.
// Every call gets a fresh channel, since the assertion drains what it
// receives.
type assertReceivesTestCase struct {
	setup func() <-chan int
	name  string
	want  assertOutcome
	got   []int
	n     int
}

func newAssertReceivesTestCase(name string, setup func() <-chan int,
	n int, got []int, want assertOutcome) assertReceivesTestCase {
	return assertReceivesTestCase{
		name:  name,
		setup: setup,
		n:     n,
		got:   got,
		want:  want,
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
	checkOutcome(t, mock, ok, tc.want, "AssertReceives")
	AssertSliceEqual(t, tc.got, got, "AssertReceives values")

	got = nil
	mock = &MockT{}
	ok = mock.Run("must", func(mt T) {
		got = AssertMustReceives(mt, tc.setup(), tc.n, budget, "ready")
		mt.Log(mustContinuationLog)
	})
	checkMustOutcome(t, mock, ok, tc.want)
	if ok {
		AssertSliceEqual(t, tc.got, got, "AssertMustReceives values")
	}
}

func TestAssertReceives(t *testing.T) {
	// Every subject is settled before the call. A row waiting on another
	// goroutine can miss its deadline on a busy machine.
	RunTestCases(t, S(
		newAssertReceivesTestCase("immediate full", chanWithSends(4), 4,
			S(1, 2, 3, 4), passes()),
		newAssertReceivesTestCase("immediate partial", chanWithSends(3), 4,
			S(1, 2, 3), failsWith("3 of 4 values within")),
		newAssertReceivesTestCase("closed before n", chanClosed, 3,
			nil, failsWith("closed after 0 of 3")),
		newAssertReceivesTestCase("zero expected", chanWithSends(0), 0,
			nil, passes()),
		newAssertReceivesTestCase("nil channel", chanNil, 1,
			nil, failsWith("0 of 1 values within")),
		newAssertReceivesTestCase("negative expected", chanNil, -1,
			nil, failsWith("negative count -1")),
	))
}
