package core

import (
	"context"
	"fmt"
	"testing"
)

// revive:disable:cognitive-complexity
func TestNewContextKey(t *testing.T) {
	// revive:enable:cognitive-complexity
	k0 := NewContextKey[int]("k0")
	// name
	if k0.String() != "k0" {
		t.Fail()
	}
	// name and type
	s := fmt.Sprintf("core.NewContextKey[%s](%q)", "int", "k0")
	if k0.GoString() != s {
		t.Fail()
	}

	// new context
	ctx0 := k0.WithValue(context.TODO(), 123)
	v0, ok := k0.Get(ctx0)
	if !ok || v0 != 123 {
		t.Fail()
	}
	// wrong context
	_, ok = k0.Get(context.TODO())
	if ok {
		t.Fail()
	}
	// sub-context
	ctx1 := k0.WithValue(ctx0, 456)
	v1, ok := k0.Get(ctx1)
	if !ok || v1 != 456 {
		t.Fail()
	}
	// parent-context
	v, ok := k0.Get(ctx0)
	if !ok || v != v0 {
		t.Fail()
	}
}
