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
// A nil receiver panics with [ErrNilReceiver]: a nil key cannot hold a
// value, and [context.WithValue] would store it where [Get] cannot find
// it.
func (ck *ContextKey[T]) WithValue(ctx context.Context, v T) context.Context {
	if ck == nil {
		panic(NewPanicError(1, ErrNilReceiver))
	}

	switch ctx {
	case nil, context.TODO():
		ctx = context.Background()
	default:
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

// String returns the name. A nil receiver has none and renders as
// "<nil>", as fmt does for a nil pointer.
func (ck *ContextKey[T]) String() string {
	if ck == nil {
		return "<nil>"
	}
	return ck.name
}

// GoString renders this key in Go syntax for %#v. A nil receiver
// renders as the typed nil conversion, as fmt does for a nil pointer.
func (ck *ContextKey[T]) GoString() string {
	if ck == nil {
		return fmt.Sprintf("(*core.ContextKey[%s])(nil)", TypeName[T]())
	}
	return fmt.Sprintf("core.NewContextKey[%s](%q)",
		TypeName[T](), ck.name)
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
