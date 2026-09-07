package core

import (
	"context"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = waitForCondTestCase{}
	_ TestCase = waitForCondContextTestCase{}
)

const (
	// shortBudgetMS caps a path that should succeed well inside it.
	shortBudgetMS = 50
	// tinyBudgetMS caps a path that should time out, keeping the run short.
	tinyBudgetMS = 10
)

func ms(n int) time.Duration {
	return time.Duration(n) * time.Millisecond
}

// alwaysTrue returns a predicate that is true from the first call.
func alwaysTrue() func() bool {
	return func() bool { return true }
}

// alwaysFalse returns a predicate that never becomes true.
func alwaysFalse() func() bool {
	return func() bool { return false }
}

// noPredicate returns no predicate at all.
func noPredicate() func() bool {
	return nil
}

// flagFlip returns a predicate that becomes true two milliseconds after it
// was made. Time-based, so it needs no goroutine and no shared state.
func flagFlip() func() bool {
	target := time.Now().Add(ms(2))
	return func() bool {
		return !time.Now().Before(target)
	}
}

// trueOnSecondCall returns a predicate that is false once and true from
// then on. Call-counted rather than time-based, so a row built on it does
// not depend on the scheduler.
func trueOnSecondCall() func() bool {
	calls := 0
	return func() bool {
		calls++
		return calls > 1
	}
}

// waitForCondTestCase exercises WaitForCond, polling every millisecond.
type waitForCondTestCase struct {
	cond     func() func() bool
	name     string
	step     time.Duration
	budgetMS int
	want     bool
}

// newWaitForCondTestCase polls every millisecond.
func newWaitForCondTestCase(name string, cond func() func() bool,
	budgetMS int, want bool) waitForCondTestCase {
	return waitForCondTestCase{
		name:     name,
		cond:     cond,
		step:     ms(1),
		budgetMS: budgetMS,
		want:     want,
	}
}

// newWaitForCondTestCaseStep polls at stepMS instead.
func newWaitForCondTestCaseStep(name string, cond func() func() bool,
	budgetMS, stepMS int, want bool) waitForCondTestCase {
	tc := newWaitForCondTestCase(name, cond, budgetMS, want)
	tc.step = ms(stepMS)
	return tc
}

func (tc waitForCondTestCase) Name() string {
	return tc.name
}

func (tc waitForCondTestCase) Test(t *testing.T) {
	t.Helper()
	got := WaitForCond(tc.cond(), ms(tc.budgetMS), tc.step)
	AssertEqual(t, tc.want, got, "result")
}

func TestWaitForCond(t *testing.T) {
	// The "zero timeout but predicate true" row pins the order of the
	// checks: the predicate is consulted before the deadline. The row after
	// it pins the final evaluation: with the deadline already passed, the
	// predicate is asked once more before the wait gives up, and nothing
	// about timing decides whether that second call happens. The two
	// guard rows carry a predicate that is already true, so a false result
	// can only come from the argument guard.
	RunTestCases(t, S(
		newWaitForCondTestCase("true immediately", alwaysTrue, shortBudgetMS, true),
		newWaitForCondTestCase("becomes true within budget", flagFlip, shortBudgetMS, true),
		newWaitForCondTestCase("never true within budget", alwaysFalse, tinyBudgetMS, false),
		newWaitForCondTestCase("zero timeout", alwaysFalse, 0, false),
		newWaitForCondTestCase("zero timeout but predicate true", alwaysTrue, 0, true),
		newWaitForCondTestCase("zero timeout, true on the second call", trueOnSecondCall, 0, true),
		newWaitForCondTestCase("negative timeout", alwaysTrue, -1, false),
		newWaitForCondTestCaseStep("zero step", alwaysTrue, shortBudgetMS, 0, false),
		newWaitForCondTestCase("nil predicate", noPredicate, tinyBudgetMS, false),
	))
}

// ctxWithBudget returns a context setup whose context ends budgetMS after
// it runs.
func ctxWithBudget(budgetMS int) func(*testing.T) context.Context {
	return func(t *testing.T) context.Context {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), ms(budgetMS))
		t.Cleanup(cancel)
		return ctx
	}
}

// ctxCancelled returns a context that has already ended.
func ctxCancelled(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	return ctx
}

// ctxNil returns no context at all.
func ctxNil(*testing.T) context.Context {
	return nil
}

// waitForCondContextTestCase exercises WaitForCondContext, polling every
// millisecond.
type waitForCondContextTestCase struct {
	ctx  func(*testing.T) context.Context
	cond func() func() bool
	name string
	step time.Duration
	want bool
}

// newWaitForCondContextTestCase polls every millisecond.
func newWaitForCondContextTestCase(name string, ctx func(*testing.T) context.Context,
	cond func() func() bool, want bool) waitForCondContextTestCase {
	return waitForCondContextTestCase{
		name: name,
		ctx:  ctx,
		cond: cond,
		step: ms(1),
		want: want,
	}
}

// newWaitForCondContextTestCaseStep polls at stepMS instead.
func newWaitForCondContextTestCaseStep(name string, ctx func(*testing.T) context.Context,
	cond func() func() bool, stepMS int, want bool) waitForCondContextTestCase {
	tc := newWaitForCondContextTestCase(name, ctx, cond, want)
	tc.step = ms(stepMS)
	return tc
}

func (tc waitForCondContextTestCase) Name() string {
	return tc.name
}

func (tc waitForCondContextTestCase) Test(t *testing.T) {
	t.Helper()
	got := WaitForCondContext(tc.ctx(t), tc.cond(), tc.step)
	AssertEqual(t, tc.want, got, "result")
}

func TestWaitForCondContext(t *testing.T) {
	// The "cancelled but predicate true" row pins the order of the checks:
	// the predicate is consulted before the context, and the row after it
	// pins the final evaluation once the context has ended. The nil-context
	// row carries a predicate that becomes true, since nothing else ends
	// that wait, and the guard rows one that is already true.
	RunTestCases(t, S(
		newWaitForCondContextTestCase("true immediately", ctxWithBudget(shortBudgetMS),
			alwaysTrue, true),
		newWaitForCondContextTestCase("becomes true before the deadline",
			ctxWithBudget(shortBudgetMS), flagFlip, true),
		newWaitForCondContextTestCase("never true before the deadline",
			ctxWithBudget(tinyBudgetMS), alwaysFalse, false),
		newWaitForCondContextTestCase("cancelled before the call", ctxCancelled,
			alwaysFalse, false),
		newWaitForCondContextTestCase("cancelled but predicate true", ctxCancelled,
			alwaysTrue, true),
		newWaitForCondContextTestCase("cancelled, true on the second call", ctxCancelled,
			trueOnSecondCall, true),
		newWaitForCondContextTestCase("nil context", ctxNil, flagFlip, true),
		newWaitForCondContextTestCaseStep("zero step", ctxWithBudget(shortBudgetMS),
			alwaysTrue, 0, false),
		newWaitForCondContextTestCase("nil predicate", ctxWithBudget(tinyBudgetMS),
			noPredicate, false),
	))
}
