package core

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"testing"
)

// TestCase validations
var _ TestCase = frameSplitNameTestCase{}
var _ TestCase = framePkgNameTestCase{}
var _ TestCase = framePkgFileTestCase{}
var _ TestCase = frameFileLineTestCase{}
var _ TestCase = frameFormatTestCase{}
var _ TestCase = capturedFrameFormatTestCase{}
var _ TestCase = stackFormatTestCase{}
var _ TestCase = formatLineTestCase{}
var _ TestCase = writeFormatPanicsTestCase{}

const (
	MaxTestDepth = 16
	MaxTestSpace = 6
)

func TestHere(t *testing.T) {
	AssertMustEqual(t, "TestHere", fmt.Sprintf("%n", Here()), "Here")
	for i := range MaxDepth {
		f := deepHere(i)
		AssertMustEqual(t, "hereHere", fmt.Sprintf("%n", f), "depth %v", i)
	}
}

func hereHere() *Frame {
	return Here()
}

func deepHere(depth int) *Frame {
	if depth > 0 {
		return deepHere(depth - 1)
	}
	return hereHere()
}

func TestStackFrame(t *testing.T) {
	AssertMustEqual(t, "TestStackFrame", fmt.Sprintf("%n", StackFrame(0)), "StackFrame(0)")

	for i := range MaxTestDepth {
		for j := range MaxTestSpace {
			f := deepStackFrame(i, j)
			AssertMustEqual(t, "hereStackFrame", fmt.Sprintf("%n", f),
				"depth %v space %v", i, j)
		}
	}
}

func deeperStackFrame(depth, skip int) *Frame {
	if depth > 0 {
		return deeperStackFrame(depth-1, skip)
	}
	return StackFrame(skip + 1)
}

func hereStackFrame(depth int) *Frame {
	return deeperStackFrame(depth, depth)
}

func deepStackFrame(depth, space int) *Frame {
	if depth > 0 {
		return deepStackFrame(depth-1, space)
	}
	return hereStackFrame(space)
}

func TestStackTrace(t *testing.T) {
	stack := StackTrace(0)
	AssertMustTrue(t, len(stack) >= 2, "StackTrace(0) depth %v", len(stack))
	AssertMustEqual(t, "TestStackTrace", fmt.Sprintf("%n", stack[0]), "StackTrace(0) bottom frame")

	for i := range MaxTestDepth {
		for j := range MaxTestSpace {
			testDeepStackTrace(t, i, j)
		}
	}
}

type stackTraceExpectation struct {
	bottomFunc    string // Function at position 0
	recurringFunc string // Function that appears multiple times
	topFunc       string // Function we expect to find somewhere in the stack
	expectedCount int    // Expected count of recurringFunc
}

func testDeepStackTrace(t *testing.T, depth, space int) {
	t.Helper()
	stack := deepStackTrace(depth, space)

	expectation := stackTraceExpectation{
		bottomFunc:    "deeperStackTrace",
		recurringFunc: "deepStackTrace",
		topFunc:       "TestStackTrace",
		expectedCount: depth + 1, // deepStackTrace calls itself 'depth' times
	}

	analysis := analyzeStackTrace(stack, expectation)

	AssertMustEqual(t, 0, analysis.bottomPos, "StackTrace(%v, %v) %s position",
		depth, space, expectation.bottomFunc)
	AssertMustNotEqual(t, -1, analysis.topPos, "StackTrace(%v, %v) %s position",
		depth, space, expectation.topFunc)
	AssertMustEqual(t, expectation.expectedCount, analysis.recurringCount,
		"StackTrace(%v, %v) %s count", depth, space, expectation.recurringFunc)
}

func deeperStackTrace(depth, skip int) Stack {
	if depth > 0 {
		return deeperStackTrace(depth-1, skip)
	}
	return StackTrace(skip)
}

func deepStackTrace(depth, space int) Stack {
	if depth > 0 {
		return deepStackTrace(depth-1, space)
	}
	return deeperStackTrace(space, space)
}

type stackAnalysis struct {
	bottomPos      int
	topPos         int
	recurringCount int
}

func analyzeStackTrace(stack Stack, expectation stackTraceExpectation) stackAnalysis {
	analysis := stackAnalysis{bottomPos: -1, topPos: -1}

	for i, frame := range stack {
		analyzeFrame(&analysis, frame, i, expectation)
	}

	return analysis
}

func analyzeFrame(analysis *stackAnalysis, frame Frame, position int, expectation stackTraceExpectation) {
	funcName := frame.FuncName()
	switch funcName {
	case expectation.bottomFunc:
		if analysis.bottomPos == -1 {
			analysis.bottomPos = position
		}
	case expectation.topFunc:
		if analysis.topPos == -1 {
			analysis.topPos = position
		}
	case expectation.recurringFunc:
		analysis.recurringCount++
	default:
	}
}

// Test case for Frame.SplitName method
type frameSplitNameTestCase struct {
	name             string
	frame            *Frame
	expectedPkgName  string
	expectedFuncName string
}

func (tc frameSplitNameTestCase) Name() string {
	return tc.name
}

func (tc frameSplitNameTestCase) Test(t *testing.T) {
	t.Helper()
	pkgName, funcName := tc.frame.SplitName()
	AssertEqual(t, tc.expectedPkgName, pkgName, "package name")
	AssertEqual(t, tc.expectedFuncName, funcName, "function name")
}

func newFrameSplitNameTestCase(name string, frame *Frame, expectedPkgName,
	expectedFuncName string) frameSplitNameTestCase {
	return frameSplitNameTestCase{
		name:             name,
		frame:            frame,
		expectedPkgName:  expectedPkgName,
		expectedFuncName: expectedFuncName,
	}
}

func frameSplitNameTestCases() []frameSplitNameTestCase {
	return []frameSplitNameTestCase{
		newFrameSplitNameTestCase("current function", Here(), "darvaza.org/core", "frameSplitNameTestCases"),
	}
}

func TestFrameSplitName(t *testing.T) {
	RunTestCases(t, frameSplitNameTestCases())
}

// Test Frame.Name for a populated frame and for the zero value.
func TestFrameName(t *testing.T) {
	AssertEqual(t, "darvaza.org/core.TestFrameName", Here().Name(), "frame name")

	var empty Frame
	AssertEqual(t, "", empty.Name(), "empty frame name")
}

// Test case for Frame.PkgName method
type framePkgNameTestCase struct {
	name     string
	frame    *Frame
	expected string
}

func (tc framePkgNameTestCase) Name() string {
	return tc.name
}

func (tc framePkgNameTestCase) Test(t *testing.T) {
	t.Helper()
	AssertEqual(t, tc.expected, tc.frame.PkgName(), "package name")
}

func newFramePkgNameTestCase(name string, frame *Frame, expected string) framePkgNameTestCase {
	return framePkgNameTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func framePkgNameTestCases() []framePkgNameTestCase {
	return []framePkgNameTestCase{
		newFramePkgNameTestCase("current function", Here(), "darvaza.org/core"),
		newFramePkgNameTestCase("empty frame", &Frame{name: ""}, ""),
		newFramePkgNameTestCase("no package frame", &Frame{name: "func"}, ""),
		newFramePkgNameTestCase("dot separator", &Frame{name: "pkg.func"}, "pkg"),
		newFramePkgNameTestCase("slash separator", &Frame{name: "pkg/module.func"}, "pkg/module"),
	}
}

// Test Frame.PkgName method (0% coverage)
func TestFramePkgName(t *testing.T) {
	RunTestCases(t, framePkgNameTestCases())
}

// Test Frame.File method (0% coverage)
func TestFrameFile(t *testing.T) {
	frame := Here()
	file := frame.File()

	// Should contain the test file name
	AssertContains(t, file, "stack_test.go", "frame file")

	// Test empty frame
	emptyFrame := &Frame{file: ""}
	AssertEqual(t, "", emptyFrame.File(), "empty frame file")

	// Test frame with file
	testFrame := &Frame{file: "/path/to/file.go"}
	AssertEqual(t, "/path/to/file.go", testFrame.File(), "frame file path")
}

// Test case for Frame.PkgFile method
type framePkgFileTestCase struct {
	name     string
	expected string
	frame    Frame
}

func (tc framePkgFileTestCase) Name() string {
	return tc.name
}

func (tc framePkgFileTestCase) Test(t *testing.T) {
	t.Helper()
	AssertEqual(t, tc.expected, tc.frame.PkgFile(), "PkgFile")
}

func newFramePkgFileTestCase(name string, frame Frame, expected string) framePkgFileTestCase {
	return framePkgFileTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

// Test Frame.PkgFile method
func TestFramePkgFile(t *testing.T) {
	tests := []framePkgFileTestCase{
		newFramePkgFileTestCase(
			"empty frame",
			Frame{},
			"",
		),
		newFramePkgFileTestCase(
			"frame with file but no name",
			Frame{file: "/path/to/file.go"},
			"file.go",
		),
		newFramePkgFileTestCase(
			"frame with absolute path and no package",
			Frame{file: "/absolute/path/to/source.go", name: "main"},
			"source.go",
		),
		newFramePkgFileTestCase(
			"frame with package and file",
			Frame{file: "/go/src/github.com/user/repo/file.go", name: "github.com/user/repo.FuncName"},
			"github.com/user/repo/file.go",
		),
		newFramePkgFileTestCase(
			"frame with nested package",
			Frame{file: "/workspace/project/internal/utils/helper.go", name: "internal/utils.Helper"},
			"internal/utils/helper.go",
		),
		newFramePkgFileTestCase(
			"frame with standard library package",
			Frame{file: "/usr/local/go/src/fmt/print.go", name: "fmt.Printf"},
			"fmt/print.go",
		),
		newFramePkgFileTestCase(
			"frame with generic function",
			Frame{file: "/src/generics.go", name: "example.com/pkg.GenericFunc[...]"},
			"example.com/pkg/generics.go",
		),
		newFramePkgFileTestCase(
			"current test frame",
			*Here(),
			"darvaza.org/core/stack_test.go",
		),
	}

	RunTestCases(t, tests)
}

// Test Frame.Line method (0% coverage)
func TestFrameLine(t *testing.T) {
	frame := Here()
	line := frame.Line()

	// Should have a valid line number (greater than 0)
	AssertTrue(t, line > 0, "line %v positive", line)

	// Test empty frame
	emptyFrame := &Frame{line: 0}
	AssertEqual(t, 0, emptyFrame.Line(), "empty frame line")

	// Test frame with line
	testFrame := &Frame{line: 42}
	AssertEqual(t, 42, testFrame.Line(), "frame line number")
}

// Test case for Frame.FileLine method
type frameFileLineTestCase struct {
	name     string
	expected string
	frame    Frame
}

func (tc frameFileLineTestCase) Name() string {
	return tc.name
}

func (tc frameFileLineTestCase) Test(t *testing.T) {
	t.Helper()
	AssertEqual(t, tc.expected, tc.frame.FileLine(), "FileLine")
}

func newFrameFileLineTestCase(name, file string, line int, expected string) frameFileLineTestCase {
	return frameFileLineTestCase{
		name:     name,
		frame:    Frame{file: file, line: line},
		expected: expected,
	}
}

func frameFileLineTestCases() []frameFileLineTestCase {
	return S(
		newFrameFileLineTestCase("frame with line", "test.go", 42, "test.go:42"),
		newFrameFileLineTestCase("frame without line", "test.go", 0, "test.go"),
		newFrameFileLineTestCase("empty frame", "", 0, ""),
		newFrameFileLineTestCase("frame with negative line", "test.go", -1, "test.go"),
	)
}

// Test Frame.FileLine method (0% coverage)
func TestFrameFileLine(t *testing.T) {
	RunTestCases(t, frameFileLineTestCases())
}

// Test Frame.String method (implements fmt.Stringer)
func TestFrameString(t *testing.T) {
	frame := &Frame{
		name: "darvaza.org/core.TestFunction",
		file: "/path/to/test.go",
		line: 42,
	}

	AssertEqual(t, "test.go:42", frame.String(), "String")

	var empty Frame
	AssertEqual(t, ":0", empty.String(), "empty frame String")
}

// Test case for Frame.Format method
type frameFormatTestCase struct {
	name     string
	frame    *Frame
	format   string
	expected string
}

func (tc frameFormatTestCase) Name() string {
	return tc.name
}

func (tc frameFormatTestCase) Test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf(tc.format, tc.frame)
	AssertEqual(t, tc.expected, result, "formatted output")
}

func newFrameFormatTestCase(name string, frame *Frame, format, expected string) frameFormatTestCase {
	return frameFormatTestCase{
		name:     name,
		frame:    frame,
		format:   format,
		expected: expected,
	}
}

func frameFormatTestCases() []frameFormatTestCase {
	frame := &Frame{
		name: "darvaza.org/core.TestFunction",
		file: "/path/to/test.go",
		line: 42,
	}
	emptyFrame := &Frame{}

	return S(
		newFrameFormatTestCase("file format %s", frame, "%s", "test.go"),
		newFrameFormatTestCase("file format with + flag %+s", frame, "%+s",
			"darvaza.org/core.TestFunction\n\t/path/to/test.go"),
		newFrameFormatTestCase("file format with # flag %#s", frame, "%#s", "darvaza.org/core/test.go"),
		newFrameFormatTestCase("line format %d", frame, "%d", "42"),
		newFrameFormatTestCase("name format %n", frame, "%n", "TestFunction"),
		newFrameFormatTestCase("name format with + flag %+n", frame, "%+n", "darvaza.org/core.TestFunction"),
		newFrameFormatTestCase("file:line format %v", frame, "%v", "test.go:42"),
		newFrameFormatTestCase("file:line format with + flag %+v", frame, "%+v",
			"darvaza.org/core.TestFunction\n\t/path/to/test.go:42"),
		newFrameFormatTestCase("file:line format with # flag %#v", frame, "%#v", "darvaza.org/core/test.go:42"),
		newFrameFormatTestCase("empty file %s", emptyFrame, "%s", ""),
		newFrameFormatTestCase("empty file with # flag %#s", emptyFrame, "%#s", ""),
		newFrameFormatTestCase("empty line %d", emptyFrame, "%d", "0"),
		newFrameFormatTestCase("empty name %n", emptyFrame, "%n", ""),
		newFrameFormatTestCase("empty file:line %v", emptyFrame, "%v", ":0"),
		newFrameFormatTestCase("empty file:line with # flag %#v", emptyFrame, "%#v", ":0"),
		// Unknown verbs fall through to the empty default branch.
		newFrameFormatTestCase("unknown verb %x", frame, "%x", ""),
	)
}

// Test Frame.Format method and helper functions (0% coverage)
func TestFrameFormat(t *testing.T) {
	RunTestCases(t, frameFormatTestCases())
}

// capturedFrameFormatTestCase formats a frame captured from a real call,
// where frameFormatTestCases builds its frames from literals.
type capturedFrameFormatTestCase struct {
	name     string
	frame    *Frame
	format   string
	expected string
}

func (tc capturedFrameFormatTestCase) Name() string {
	return tc.name
}

func (tc capturedFrameFormatTestCase) Test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf(tc.format, tc.frame)
	AssertEqual(t, tc.expected, result, "formatted output")
}

func newCapturedFrameFormatTestCase(name string, frame *Frame,
	format, expected string) capturedFrameFormatTestCase {
	return capturedFrameFormatTestCase{
		name:     name,
		frame:    frame,
		format:   format,
		expected: expected,
	}
}

// capturedFrame returns the frame StackFrame reports for the call site of
// capturedFrame's caller, alongside the file and line runtime.Caller
// reports for the same site. The two reach the runtime by different
// routes, so the caller can state one against the other.
func capturedFrame(t *testing.T) (*Frame, string, int) {
	t.Helper()

	frame := StackFrame(1)
	_, file, line, ok := runtime.Caller(1)
	AssertMustTrue(t, ok, "runtime.Caller")

	return frame, file, line
}

func capturedFrameFormatTestCases(frame *Frame, file string,
	line int) []capturedFrameFormatTestCase {
	const funcName = "TestCapturedFrameFormat"
	const fullName = "darvaza.org/core." + funcName
	const baseFile = "stack_test.go"
	const pkgFile = "darvaza.org/core/" + baseFile

	lineText := strconv.Itoa(line)

	return S(
		newCapturedFrameFormatTestCase("file %s", frame, "%s", baseFile),
		newCapturedFrameFormatTestCase("line %d", frame, "%d", lineText),
		newCapturedFrameFormatTestCase("name %n", frame, "%n", funcName),
		newCapturedFrameFormatTestCase("file:line %v", frame, "%v",
			baseFile+":"+lineText),
		newCapturedFrameFormatTestCase("file %+s", frame, "%+s",
			fullName+"\n\t"+file),
		newCapturedFrameFormatTestCase("name %+n", frame, "%+n", fullName),
		newCapturedFrameFormatTestCase("file:line %+v", frame, "%+v",
			fullName+"\n\t"+file+":"+lineText),
		newCapturedFrameFormatTestCase("file %#s", frame, "%#s", pkgFile),
		newCapturedFrameFormatTestCase("file:line %#v", frame, "%#v",
			pkgFile+":"+lineText),
	)
}

// Test Frame.Format against a captured frame. The expectations are built
// from what runtime.Caller reports for the same call site, so the
// preconditions state that frameForPC agrees with it on file and line —
// the line being resolved there from the raw pc, where the name comes
// from pc - 1.
func TestCapturedFrameFormat(t *testing.T) {
	frame, file, line := capturedFrame(t)

	AssertMustNotNil(t, frame, "captured frame")
	AssertMustEqual(t, file, frame.File(), "captured file")
	AssertMustEqual(t, line, frame.Line(), "captured line")

	RunTestCases(t, capturedFrameFormatTestCases(frame, file, line))
}

// Test case for Stack.Format method
type stackFormatTestCase struct {
	name      string
	stack     Stack
	format    string
	contains  []string
	wantEmpty bool
}

func (tc stackFormatTestCase) Name() string {
	return tc.name
}

func (tc stackFormatTestCase) Test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf(tc.format, tc.stack)

	if tc.wantEmpty {
		AssertEqual(t, "", result, "output")
		return
	}

	for _, expected := range tc.contains {
		AssertContains(t, result, expected, "output")
	}
}

// newStackFormatTestCase declares a row by the substrings the output must
// contain, and every row must name at least one: an empty list would leave
// the Test body asserting nothing at all.
func newStackFormatTestCase(name string, stack Stack, format string, contains []string) stackFormatTestCase {
	if len(contains) == 0 {
		panic("stackFormatTestCase: every row must state a substring")
	}

	return stackFormatTestCase{
		name:     name,
		stack:    stack,
		format:   format,
		contains: contains,
	}
}

// newStackFormatTestCaseEmpty expects the format to produce no output at all,
// which no substring row can state: every string contains "".
func newStackFormatTestCaseEmpty(name string, stack Stack, format string) stackFormatTestCase {
	return stackFormatTestCase{
		name:      name,
		stack:     stack,
		format:    format,
		wantEmpty: true,
	}
}

func stackFormatTestCases() []stackFormatTestCase {
	stack := Stack{
		{name: "darvaza.org/core.func1", file: "/path/to/file1.go", line: 10},
		{name: "darvaza.org/core.func2", file: "/path/to/file2.go", line: 20},
	}
	emptyStack := Stack{}

	return S(
		newStackFormatTestCase("basic format %s", stack, "%s", S("\nfile1.go", "\nfile2.go")),
		newStackFormatTestCase("verbose format %+v", stack, "%+v",
			S("\ndarvaza.org/core.func1\n\t/path/to/file1.go:10",
				"\ndarvaza.org/core.func2\n\t/path/to/file2.go:20")),
		newStackFormatTestCase("numbered format %#+v", stack, "%#+v",
			S("\n[0/2] darvaza.org/core.func1\n\t/path/to/file1.go:10",
				"\n[1/2] darvaza.org/core.func2\n\t/path/to/file2.go:20")),
		newStackFormatTestCase("numbered name format %#+n", stack, "%#+n",
			S("\n[0/2] darvaza.org/core.func1", "\n[1/2] darvaza.org/core.func2")),
		newStackFormatTestCaseEmpty("empty stack", emptyStack, "%+v"),
	)
}

// Test Stack.Format method (0% coverage)
func TestStackFormat(t *testing.T) {
	RunTestCases(t, stackFormatTestCases())
}

// Test Stack.String method (implements fmt.Stringer)
func TestStackString(t *testing.T) {
	stack := Stack{
		{name: "darvaza.org/core.func1", file: "/path/to/file1.go", line: 10},
		{name: "darvaza.org/core.func2", file: "/path/to/file2.go", line: 20},
	}

	AssertEqual(t, "\nfile1.go:10\nfile2.go:20", stack.String(), "String")

	AssertEqual(t, "", Stack{}.String(), "empty stack String")
}

// Test case for formatLine function
type formatLineTestCase struct {
	name     string
	frame    *Frame
	expected string
}

func (tc formatLineTestCase) Name() string {
	return tc.name
}

func (tc formatLineTestCase) Test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf("%d", tc.frame)
	AssertEqual(t, tc.expected, result, "formatted line")
}

func newFormatLineTestCase(name string, frame *Frame, expected string) formatLineTestCase {
	return formatLineTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func formatLineTestCases() []formatLineTestCase {
	return S(
		newFormatLineTestCase("positive line", &Frame{line: 123}, "123"),
		newFormatLineTestCase("zero line", &Frame{line: 0}, "0"),
	)
}

// Test formatLine function directly (0% coverage)
func TestFormatLineMethod(t *testing.T) {
	RunTestCases(t, formatLineTestCases())
}

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

// Cover the early-return branch of StackFrame when skip exceeds depth.
func TestStackFrameSkipExceeds(t *testing.T) {
	AssertNil(t, StackFrame(10000), "StackFrame returns nil when skip > depth")
}
