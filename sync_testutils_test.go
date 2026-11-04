package core_test

import (
	"context"
	"testing"
	"time"
)

const (
	// shortBudgetMS caps a path that should succeed well inside it.
	shortBudgetMS = 50
	// tinyBudgetMS caps a path that should time out, keeping the run short.
	tinyBudgetMS = 10
)

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
