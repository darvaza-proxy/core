package core

import (
	"io"
	"testing"
)

// TestCase validations
var _ TestCase = writeFormatPanicsTestCase{}

// failingWriter returns a fixed (n, err) from every Write.
type failingWriter struct {
	err error
	n   int
}

func (w failingWriter) Write(_ []byte) (int, error) { return w.n, w.err }

// writeFormatPanicsTestCase exercises the two panic paths in writeFormat:
// Write error and short write. The wantPanic substring identifies which
// branch triggered the panic.
type writeFormatPanicsTestCase struct {
	w         io.Writer
	name      string
	wantPanic string
}

func newWriteFormatPanicsTestCase(name string, w io.Writer,
	wantPanic string) writeFormatPanicsTestCase {
	return writeFormatPanicsTestCase{name: name, w: w, wantPanic: wantPanic}
}

func (tc writeFormatPanicsTestCase) Name() string { return tc.name }

func (tc writeFormatPanicsTestCase) Test(t *testing.T) {
	t.Helper()
	AssertPanic(t, func() { writeFormat(tc.w, "payload") }, tc.wantPanic, tc.name)
}

func writeFormatPanicsTestCases() []writeFormatPanicsTestCase {
	return S(
		newWriteFormatPanicsTestCase("write error",
			failingWriter{err: io.ErrShortWrite}, `Frame: failed to write "payload"`),
		// len("payload") == 7, so n=3 is a short write.
		newWriteFormatPanicsTestCase("short write",
			failingWriter{n: 3}, "Frame: incomplete write (3/7)"),
	)
}

func TestWriteFormatPanics(t *testing.T) {
	RunTestCases(t, writeFormatPanicsTestCases())
}

// Cover the unknown-PC branch of frameForPC: passing pc=0 makes
// frameForPC call runtime.FuncForPC(pc - 1), which returns nil and
// forces the "unknown" fallback.
func TestFrameForPCUnknown(t *testing.T) {
	f := frameForPC(0)
	AssertEqual(t, "unknown", f.Name(), "unknown PC name")
	AssertEqual(t, "unknown", f.File(), "unknown PC file")
}
