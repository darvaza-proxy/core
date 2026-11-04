package core

// NewFrameForTest returns a Frame with the given name, file and line,
// and a zero program counter and entry point.
func NewFrameForTest(name, file string, line int) *Frame {
	return &Frame{name: name, file: file, line: line}
}

// FrameEntryForTest returns the entry point of the function f runs in,
// which a frame inlined into its caller shares with that caller's frame.
func FrameEntryForTest(f Frame) uintptr {
	return f.entry
}
