package core

import (
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = awaitCloseAtDeadlineTestCase{}
	_ TestCase = awaitQuietAtDeadlineTestCase{}
	_ TestCase = receiveNAtDeadlineTestCase{}
)

// atDeadlineRuns is how many times each at-deadline row runs. With both
// select arms ready the runtime picks one at random, and this many runs
// make missing the deadline arm every time as good as impossible.
const atDeadlineRuns = 16

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
	for range atDeadlineRuns {
		count, closed := awaitClose(tc.setup(), closedDeadline())
		AssertEqual(t, tc.count, count, "count")
		AssertEqual(t, tc.closed, closed, "closed")
	}
}

func TestAwaitCloseAtDeadline(t *testing.T) {
	RunTestCases(t, S(
		newAwaitCloseAtDeadlineTestCase("closed", ChanClosedForTest, 0, true),
		newAwaitCloseAtDeadlineTestCase("closed behind a value", ChanClosedWithSendsForTest(1), 1, true),
		newAwaitCloseAtDeadlineTestCase("closed behind values", ChanClosedWithSendsForTest(3), 3, true),
		newAwaitCloseAtDeadlineTestCase("value sent", ChanWithSendsForTest(1), 1, false),
		newAwaitCloseAtDeadlineTestCase("values sent", ChanWithSendsForTest(3), 3, false),
		newAwaitCloseAtDeadlineTestCase("left open", ChanOpenForTest, 0, false),
		newAwaitCloseAtDeadlineTestCase("nil channel", ChanNilForTest, 0, false),
	))
}

// awaitQuietAtDeadlineTestCase exercises awaitQuiet with a deadline that
// has already fired, so both select arms are ready: an event ch already
// holds still settles the verdict, and only the first one is read.
type awaitQuietAtDeadlineTestCase struct {
	setup  func() <-chan int
	name   string
	got    int
	quiet  bool
	closed bool
}

// newAwaitQuietAtDeadlineTestCaseQuiet declares a channel with nothing to
// read.
func newAwaitQuietAtDeadlineTestCaseQuiet(name string,
	setup func() <-chan int) awaitQuietAtDeadlineTestCase {
	return awaitQuietAtDeadlineTestCase{
		name:   name,
		setup:  setup,
		got:    0,
		quiet:  true,
		closed: false,
	}
}

// newAwaitQuietAtDeadlineTestCaseClosed declares a channel whose close is
// the first thing to read.
func newAwaitQuietAtDeadlineTestCaseClosed(name string,
	setup func() <-chan int) awaitQuietAtDeadlineTestCase {
	return awaitQuietAtDeadlineTestCase{
		name:   name,
		setup:  setup,
		got:    0,
		quiet:  false,
		closed: true,
	}
}

// newAwaitQuietAtDeadlineTestCaseValue declares a channel whose first
// event is the value got, whatever follows it.
func newAwaitQuietAtDeadlineTestCaseValue(name string,
	setup func() <-chan int, got int) awaitQuietAtDeadlineTestCase {
	return awaitQuietAtDeadlineTestCase{
		name:   name,
		setup:  setup,
		got:    got,
		quiet:  false,
		closed: false,
	}
}

func (tc awaitQuietAtDeadlineTestCase) Name() string {
	return tc.name
}

func (tc awaitQuietAtDeadlineTestCase) Test(t *testing.T) {
	for range atDeadlineRuns {
		got, quiet, closed := awaitQuiet(tc.setup(), closedDeadline())
		AssertEqual(t, tc.got, got, "got")
		AssertEqual(t, tc.quiet, quiet, "quiet")
		AssertEqual(t, tc.closed, closed, "closed")
	}
}

// awaitQuietAtDeadlineTestCases lists the channels by their first event:
// nothing, the close, then a value. A value ahead of the close is what the
// verdict rests on, so a channel closed behind values reports the value.
func awaitQuietAtDeadlineTestCases() []awaitQuietAtDeadlineTestCase {
	return S(
		newAwaitQuietAtDeadlineTestCaseQuiet("left open", ChanOpenForTest),
		newAwaitQuietAtDeadlineTestCaseQuiet("nil channel", ChanNilForTest),

		newAwaitQuietAtDeadlineTestCaseClosed("closed", ChanClosedForTest),

		newAwaitQuietAtDeadlineTestCaseValue("value sent", ChanWithSendsForTest(1), 1),
		newAwaitQuietAtDeadlineTestCaseValue("values sent", ChanWithSendsForTest(3), 1),
		newAwaitQuietAtDeadlineTestCaseValue("values sent, first taken", ChanFirstTakenForTest, 2),
		newAwaitQuietAtDeadlineTestCaseValue("closed behind a value",
			ChanClosedWithSendsForTest(1), 1),
		newAwaitQuietAtDeadlineTestCaseValue("closed behind values",
			ChanClosedWithSendsForTest(3), 1),
	)
}

func TestAwaitQuietAtDeadline(t *testing.T) {
	RunTestCases(t, awaitQuietAtDeadlineTestCases())
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
	for range atDeadlineRuns {
		got, closed := receiveN(tc.setup(), tc.n, closedDeadline())
		AssertSliceEqual(t, tc.got, got, "got")
		AssertEqual(t, tc.closed, closed, "closed")
	}
}

func TestReceiveNAtDeadline(t *testing.T) {
	RunTestCases(t, S(
		newReceiveNAtDeadlineTestCase("all buffered", ChanWithSendsForTest(3), 3, S(1, 2, 3), false),
		newReceiveNAtDeadlineTestCase("some buffered", ChanWithSendsForTest(2), 4, S(1, 2), false),
		newReceiveNAtDeadlineTestCase("closed", ChanClosedForTest, 1, nil, true),
		newReceiveNAtDeadlineTestCase("closed behind enough", ChanClosedWithSendsForTest(2), 2, S(1, 2), false),
		newReceiveNAtDeadlineTestCase("closed behind too few", ChanClosedWithSendsForTest(3), 4, S(1, 2, 3), true),
		newReceiveNAtDeadlineTestCase("nil channel", ChanNilForTest, 2, nil, false),
	))
}
