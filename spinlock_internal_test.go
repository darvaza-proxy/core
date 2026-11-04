package core

import "testing"

func TestSpinLockNilPtr(t *testing.T) {
	var sl *SpinLock
	ptr := sl.ptr()
	AssertEqual(t, (*uint32)(nil), ptr, "nil ptr")
}
