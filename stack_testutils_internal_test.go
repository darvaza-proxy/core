package core

// NewFrameForTest returns a Frame with the given name, file and line,
// and a zero program counter and entry point.
func NewFrameForTest(name, file string, line int) *Frame {
	return &Frame{name: name, file: file, line: line}
}
