package core

import (
	"context"
	"fmt"
	"time"
)

// pollStep is the cadence at which the Eventually assertions re-evaluate
// their predicate.
const pollStep = time.Millisecond

// AssertEventually fails the test unless the predicate holds within
// timeout, polling it every millisecond through WaitForCond. A nil
// predicate or a negative timeout fails outright.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertEventually(t, func() bool { return srv.Ready() }, time.Second, "server ready")
func AssertEventually(t T, cond func() bool, timeout time.Duration,
	name string, args ...any) bool {
	t.Helper()
	if !checkAssertEventually(t, name, args, cond, timeout) {
		return false
	}

	start := time.Now()
	if !WaitForCond(cond, timeout, pollStep) {
		doError(t, name, args, "not true within %v", timeout)
		return false
	}
	doLog(t, name, args, "true after %v", time.Since(start))
	return true
}

// AssertEventuallyContext fails the test unless the predicate holds before
// ctx ends, polling it every millisecond through WaitForCondContext. A nil
// ctx never ends, so the wait then lasts until the predicate holds. A nil
// predicate fails outright.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertEventuallyContext(t, t.Context(), func() bool { return srv.Ready() }, "server ready")
//
//revive:disable-next-line:context-as-argument
func AssertEventuallyContext(t T, ctx context.Context, cond func() bool,
	name string, args ...any) bool {
	t.Helper()
	if !checkAssertEventually(t, name, args, cond, 0) {
		return false
	}

	start := time.Now()
	if !WaitForCondContext(ctx, cond, pollStep) {
		// the wait ends only when ctx does, so ctx is not nil and its
		// error is set
		doError(t, name, args, "not true: %v", ctx.Err())
		return false
	}
	doLog(t, name, args, "true after %v", time.Since(start))
	return true
}

// checkAssertEventually reports the caller's argument that WaitForCond
// would reject without consulting the predicate, and returns false when it
// found one. A zero timeout is accepted, so a caller with no timeout to
// check passes zero.
func checkAssertEventually(t T, name string, args []any, cond func() bool, timeout time.Duration) bool {
	t.Helper()
	switch {
	case cond == nil:
		doError(t, name, args, "nil predicate never becomes true")
	case timeout < 0:
		doError(t, name, args, "negative timeout %v", timeout)
	default:
		return true
	}
	return false
}

// AssertClosed fails the test unless ch is closed within timeout. It
// receives until the close or the deadline, so a closed channel passes
// whatever it still held, and the values consumed on the way are counted
// in the report. It is meant for channels that signal by closing. A nil
// channel never closes.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertClosed(t, wg.Done(), time.Second, "workers finished")
func AssertClosed[U any](t T, ch <-chan U, timeout time.Duration,
	name string, args ...any) bool {
	t.Helper()
	start := time.Now()
	count, closed := awaitClose(ch, time.After(timeout))
	suffix := consumedSuffix(count)
	if !closed {
		doError(t, name, args, "still open after %v%s", time.Since(start), suffix)
		return false
	}
	doLog(t, name, args, "closed after %v%s", time.Since(start), suffix)
	return true
}

// AssertOpen fails the test if ch is closed within timeout. It receives
// until the close or the deadline, so an open channel passes whatever it
// held, and the values consumed on the way are counted in the report. It is
// meant for channels that signal by closing. A nil channel never closes, so
// it always passes.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	AssertOpen(t, wg.Done(), 10*time.Millisecond, "workers still running")
func AssertOpen[U any](t T, ch <-chan U, timeout time.Duration,
	name string, args ...any) bool {
	t.Helper()
	start := time.Now()
	count, closed := awaitClose(ch, time.After(timeout))
	suffix := consumedSuffix(count)
	if closed {
		doError(t, name, args, "closed after %v%s", time.Since(start), suffix)
		return false
	}
	doLog(t, name, args, "open for %v%s", time.Since(start), suffix)
	return true
}

// awaitClose receives from ch until it closes or deadline fires, counting
// the values consumed on the way. closed is false when the deadline fired
// first. Once the deadline fires it takes what ch already holds without
// waiting: the clock never outranks a subject that is already ready.
func awaitClose[U any](ch <-chan U, deadline <-chan time.Time) (count int, closed bool) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return count, true
			}
			count++
		case <-deadline:
			return drainClose(ch, count)
		}
	}
}

// drainClose continues awaitClose without waiting. It stops at the close,
// at an empty channel, or after cap(ch)+1 receives, which is all a channel
// can hold ahead of its close, so a producer still sending cannot keep it
// here.
func drainClose[U any](ch <-chan U, count int) (int, bool) {
	limit := count + cap(ch) + 1
drain:
	for count < limit {
		select {
		case _, ok := <-ch:
			if !ok {
				return count, true
			}
			count++
		default:
			break drain
		}
	}
	return count, false
}

// consumedSuffix names the values consumed on the way to a verdict, or
// returns "" when there were none.
func consumedSuffix(n int) string {
	switch n {
	case 0:
		return ""
	case 1:
		return ", 1 value consumed"
	default:
		return fmt.Sprintf(", %d values consumed", n)
	}
}

// AssertReceives fails the test unless n values arrive on ch within
// timeout. The timeout covers the whole collection rather than each receive,
// and ch closing before the n-th value counts as failure. Values already
// buffered when the timeout fires still count. A negative n fails without
// reading ch.
// The name parameter can include printf-style formatting.
// Returns the values received, in arrival order and nil when there were
// none, and true if the assertion passed.
//
// Example usage:
//
//	AssertReceives(t, ready, workers, time.Second, "workers ready")
//
//revive:disable-next-line:argument-limit
func AssertReceives[U any](t T, ch <-chan U, n int, timeout time.Duration,
	name string, args ...any) ([]U, bool) {
	t.Helper()
	if n < 0 {
		doError(t, name, args, "negative count %d", n)
		return nil, false
	}

	start := time.Now()
	got, closed := receiveN(ch, n, time.After(timeout))
	switch {
	case closed:
		doError(t, name, args, "closed after %d of %d values", len(got), n)
		return got, false
	case len(got) < n:
		doError(t, name, args, "%d of %d values within %v", len(got), n, timeout)
		return got, false
	default:
		doLog(t, name, args, "%d values after %v", n, time.Since(start))
		return got, true
	}
}

// receiveN receives up to n values from ch before deadline fires, returning
// what arrived, nil when nothing did, and whether ch closed first. Once the
// deadline fires it takes what is already buffered without waiting.
func receiveN[U any](ch <-chan U, n int, deadline <-chan time.Time) (got []U, closed bool) {
	for len(got) < n {
		select {
		case v, ok := <-ch:
			if !ok {
				return got, true
			}
			got = append(got, v)
		case <-deadline:
			return drainN(ch, got, n)
		}
	}
	return got, false
}

// drainN continues receiveN from got without waiting, taking only what ch
// already holds.
func drainN[U any](ch <-chan U, got []U, n int) ([]U, bool) {
	for len(got) < n {
		select {
		case v, ok := <-ch:
			if !ok {
				return got, true
			}
			got = append(got, v)
		default:
			return got, false
		}
	}
	return got, false
}

// AssertMustEventually calls AssertEventually and t.FailNow() if the
// assertion fails.
// This is a convenience function for tests that should terminate on assertion failure.
//
// Example usage:
//
//	AssertMustEventually(t, func() bool { return srv.Ready() }, time.Second, "server ready")
func AssertMustEventually(t T, cond func() bool, timeout time.Duration,
	name string, args ...any) {
	t.Helper()
	if !AssertEventually(t, cond, timeout, name, args...) {
		t.FailNow()
	}
}

// AssertMustEventuallyContext calls AssertEventuallyContext and t.FailNow()
// if the assertion fails.
// This is a convenience function for tests that should terminate on assertion failure.
//
// Example usage:
//
//	AssertMustEventuallyContext(t, t.Context(), func() bool { return srv.Ready() }, "server ready")
//
//revive:disable-next-line:context-as-argument
func AssertMustEventuallyContext(t T, ctx context.Context, cond func() bool,
	name string, args ...any) {
	t.Helper()
	if !AssertEventuallyContext(t, ctx, cond, name, args...) {
		t.FailNow()
	}
}

// AssertMustClosed calls AssertClosed and t.FailNow() if the assertion fails.
// This is a convenience function for tests that should terminate on assertion failure.
//
// Example usage:
//
//	AssertMustClosed(t, wg.Done(), time.Second, "workers finished")
func AssertMustClosed[U any](t T, ch <-chan U, timeout time.Duration,
	name string, args ...any) {
	t.Helper()
	if !AssertClosed(t, ch, timeout, name, args...) {
		t.FailNow()
	}
}

// AssertMustOpen calls AssertOpen and t.FailNow() if the assertion fails.
// This is a convenience function for tests that should terminate on assertion failure.
//
// Example usage:
//
//	AssertMustOpen(t, wg.Done(), 10*time.Millisecond, "workers still running")
func AssertMustOpen[U any](t T, ch <-chan U, timeout time.Duration,
	name string, args ...any) {
	t.Helper()
	if !AssertOpen(t, ch, timeout, name, args...) {
		t.FailNow()
	}
}

// AssertMustReceives calls AssertReceives and t.FailNow() if the
// assertion fails.
// This is a convenience function for tests that should terminate on assertion failure.
//
// Example usage:
//
//	AssertMustReceives(t, ready, workers, time.Second, "workers ready")
//
//revive:disable-next-line:argument-limit
func AssertMustReceives[U any](t T, ch <-chan U, n int, timeout time.Duration,
	name string, args ...any) []U {
	t.Helper()
	got, ok := AssertReceives(t, ch, n, timeout, name, args...)
	if !ok {
		t.FailNow()
	}
	return got
}
