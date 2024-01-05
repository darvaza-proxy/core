package core

import (
	"context"
	"fmt"
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
	if ck == nil {
		var zero T
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
