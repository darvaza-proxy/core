package core

import (
	"context"
	"fmt"
	"time"
)

var (
	_ fmt.Stringer   = (*ContextKey[any])(nil)
	_ fmt.GoStringer = (*ContextKey[any])(nil)
)

// ContextKey is a type-safe key for a context.Context value
type ContextKey[T any] struct {
	name string
}

// WithValue safely attaches a value to a context.Context under this key.
func (ck *ContextKey[T]) WithValue(ctx context.Context, v T) context.Context {
	switch ctx {
	case nil, context.TODO():
		ctx = context.Background()
	}

	return context.WithValue(ctx, ck, v)
}

// Get attempts to extract a value bound to this key in a [context.Context]
// For convenience this method will safely operate over a nil receiver.
func (ck *ContextKey[T]) Get(ctx context.Context) (T, bool) {
	var zero T
	if ck == nil || ctx == nil {
		return zero, false
	}

	v, ok := ctx.Value(ck).(T)
	return v, ok
}

// String returns the name
func (ck *ContextKey[T]) String() string {
	return ck.name
}

// GoString renders this key in Go syntax for %v
func (ck *ContextKey[T]) GoString() string {
	var zero T
	return fmt.Sprintf("core.NewContextKey[%T](%q)",
		zero, ck.name)
}

// NewContextKey creates a new ContextKey bound to the
// specified type and friendly name
func NewContextKey[T any](name string) *ContextKey[T] {
	return &ContextKey[T]{name: name}
}

// WithTimeout is equivalent to [context.WithDeadline] but taking a duration
// instead of an absolute time.
//
// If the duration is zero or negative the context won't expire.
func WithTimeout(parent context.Context, tio time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	if tio > 0 {
		deadline := time.Now().Add(tio)
		return context.WithDeadline(parent, deadline)
	}

	return parent, func() {}
}

// WithTimeoutCause is equivalent to [context.WithDeadlineCause] but taking a duration
// instead of an absolute time.
//
// If the duration is zero or negative the context won't expire.
func WithTimeoutCause(parent context.Context, tio time.Duration, cause error) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	if tio > 0 {
		deadline := time.Now().Add(tio)
		return context.WithDeadlineCause(parent, deadline, cause)
	}

	return parent, func() {}
}
