package core

import "fmt"

// AssertTopFrame fails the test unless v carries a call stack whose top
// frame is the function want, as [Frame.FuncName] names it. v is taken
// as any so a recovered panic value, or the error [Catch] returns, can be
// passed as it is; it fails when v is nil, is not a [CallStacker], or
// carries an empty stack. The stack of a panic raised by [Panic], or of
// the error [Catch] returns, starts at the function that panicked.
// The name parameter can include printf-style formatting.
// Returns true if the assertion passed, false otherwise.
//
// Example usage:
//
//	defer func() {
//		AssertTopFrame(t, recover(), "callPanic", "panic stack")
//	}()
//	callPanic(err)
func AssertTopFrame(t T, v any, want, name string, args ...any) bool {
	t.Helper()
	got, reason := topFrameName(v)
	switch {
	case reason != "":
		doError(t, name, args, "%s", reason)
	case got != want:
		doError(t, name, args, "expected top frame %q, got %q", want, got)
	default:
		doLog(t, name, args, "top frame %q", got)
		return true
	}
	return false
}

// topFrameName returns the [Frame.FuncName] of the top frame of the stack
// v carries, or why there is none.
func topFrameName(v any) (funcName, reason string) {
	if IsNil(v) {
		return "", "expected a call stack, got nil"
	}

	cs, ok := v.(CallStacker)
	if !ok {
		return "", fmt.Sprintf("expected a CallStacker, got %T", v)
	}

	stack := cs.CallStack()
	if len(stack) == 0 {
		return "", "empty call stack"
	}
	return stack[0].FuncName(), ""
}

// AssertMustTopFrame calls AssertTopFrame and t.FailNow() if the assertion
// fails.
// This is a convenience function for tests that should terminate on assertion failure.
//
// Example usage:
//
//	defer func() {
//		AssertMustTopFrame(t, recover(), "callPanic", "panic stack")
//	}()
//	callPanic(err)
func AssertMustTopFrame(t T, v any, want, name string, args ...any) {
	t.Helper()
	if !AssertTopFrame(t, v, want, name, args...) {
		t.FailNow()
	}
}
