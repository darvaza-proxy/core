package core_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = reportedCallerTestCase{}

// callerT is a [core.MockT] that also records where the testing package
// would place each message it receives: in the first frame, outwards
// from the call to Error or Log, whose function has not called Helper.
type callerT struct {
	helpers map[string]bool
	files   []string
	core.MockT
}

// newCallerT returns a callerT that has recorded no helper and no
// message.
func newCallerT() *callerT {
	return &callerT{helpers: make(map[string]bool)}
}

// Helper marks the calling function as a helper, as testing.T does.
func (ct *callerT) Helper() {
	ct.MockT.Helper()

	var pc [1]uintptr
	runtime.Callers(2, pc[:])
	frame, _ := runtime.CallersFrames(pc[:]).Next()
	ct.helpers[frame.Function] = true
}

// Error records where the message would be placed, then passes it on.
func (ct *callerT) Error(args ...any) {
	ct.record()
	ct.MockT.Error(args...)
}

// Log records where the message would be placed, then passes it on.
func (ct *callerT) Log(args ...any) {
	ct.record()
	ct.MockT.Log(args...)
}

// record adds the file of the first frame, from the caller of Error or
// Log outwards, whose function has not called Helper, or of the last
// frame it reads when all of them have.
func (ct *callerT) record() {
	var pcs [64]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !more || !ct.helpers[frame.Function] {
			ct.files = append(ct.files, filepath.Base(frame.File))
			return
		}
	}
}

// callerTestFile returns the name of this file, where the rows call
// their assertions.
func callerTestFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Base(file)
}

// reportedCallerTestCase runs an assertion against a callerT, and checks
// that its message is placed where the row calls it, not inside the
// assertion.
type reportedCallerTestCase struct {
	assert func(core.T)
	name   string
}

func (tc reportedCallerTestCase) Name() string {
	return tc.name
}

func (tc reportedCallerTestCase) Test(t *testing.T) {
	ct := newCallerT()
	tc.assert(ct)

	core.AssertSliceEqual(t, core.S(callerTestFile()), ct.files, "reported files")
}

func newReportedCallerTestCase(name string, assert func(core.T)) reportedCallerTestCase {
	return reportedCallerTestCase{
		name:   name,
		assert: assert,
	}
}

func reportedCallerTestCases() []reportedCallerTestCase {
	return []reportedCallerTestCase{
		newReportedCallerTestCase("failing assertion", func(t core.T) {
			core.AssertEqual(t, 1, 2, "values")
		}),
		newReportedCallerTestCase("passing assertion", func(t core.T) {
			core.AssertEqual(t, 1, 1, "values")
		}),
	}
}

func TestAssertReportsCaller(t *testing.T) {
	core.RunTestCases(t, reportedCallerTestCases())
}
