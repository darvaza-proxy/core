package core_test

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"darvaza.org/core"
)

// TestCase validations
var (
	_ core.TestCase = frameSplitNameTestCase{}
	_ core.TestCase = frameNameTestCase{}
	_ core.TestCase = frameStringTestCase{}
	_ core.TestCase = framePkgNameTestCase{}
	_ core.TestCase = framePkgFileTestCase{}
	_ core.TestCase = frameFileLineTestCase{}
	_ core.TestCase = frameFormatTestCase{}
	_ core.TestCase = capturedFrameFormatTestCase{}
	_ core.TestCase = capturedLineTestCase{}
	_ core.TestCase = stackStringTestCase{}
	_ core.TestCase = stackFormatTestCase{}
	_ core.TestCase = formatLineTestCase{}
	_ core.TestCase = stackTraceMaxDepthTestCase{}
)

const (
	MaxTestDepth = 16
	MaxTestSpace = 6
)

func TestHere(t *testing.T) {
	core.AssertMustEqual(t, "TestHere", fmt.Sprintf("%n", core.Here()), "Here")
	for i := range core.MaxDepth {
		f := deepHere(i)
		core.AssertMustEqual(t, "hereHere", fmt.Sprintf("%n", f), "depth %v", i)
	}
}

// Test that Here allocates no more than the frame it returns, and
// resolves its one program counter without allocating anything else.
func TestHereAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() { core.Here() })
	core.AssertTrue(t, allocs <= 1, "at most one allocation, got %v", allocs)
}

func hereHere() *core.Frame {
	return core.Here()
}

func deepHere(depth int) *core.Frame {
	if depth > 0 {
		return deepHere(depth - 1)
	}
	return hereHere()
}

func TestStackFrame(t *testing.T) {
	core.AssertMustEqual(t, "TestStackFrame", fmt.Sprintf("%n", core.StackFrame(0)), "StackFrame(0)")

	for i := range MaxTestDepth {
		for j := range MaxTestSpace {
			f := deepStackFrame(i, j)
			core.AssertMustEqual(t, "hereStackFrame", fmt.Sprintf("%n", f),
				"depth %v space %v", i, j)
		}
	}
}

func deeperStackFrame(depth, skip int) *core.Frame {
	if depth > 0 {
		return deeperStackFrame(depth-1, skip)
	}
	return core.StackFrame(skip + 1)
}

func hereStackFrame(depth int) *core.Frame {
	return deeperStackFrame(depth, depth)
}

func deepStackFrame(depth, space int) *core.Frame {
	if depth > 0 {
		return deepStackFrame(depth-1, space)
	}
	return hereStackFrame(space)
}

func TestStackTrace(t *testing.T) {
	stack := core.StackTrace(0)
	core.AssertMustTrue(t, len(stack) >= 2, "StackTrace(0) depth %v", len(stack))
	core.AssertMustEqual(t, "TestStackTrace", fmt.Sprintf("%n", stack[0]), "StackTrace(0) bottom frame")

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

	core.AssertMustEqual(t, 0, analysis.bottomPos, "StackTrace(%v, %v) %s position",
		depth, space, expectation.bottomFunc)
	core.AssertMustNotEqual(t, -1, analysis.topPos, "StackTrace(%v, %v) %s position",
		depth, space, expectation.topFunc)
	core.AssertMustEqual(t, expectation.expectedCount, analysis.recurringCount,
		"StackTrace(%v, %v) %s count", depth, space, expectation.recurringFunc)
}

func deeperStackTrace(depth, skip int) core.Stack {
	if depth > 0 {
		return deeperStackTrace(depth-1, skip)
	}
	return core.StackTrace(skip)
}

func deepStackTrace(depth, space int) core.Stack {
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

func analyzeStackTrace(stack core.Stack, expectation stackTraceExpectation) stackAnalysis {
	analysis := stackAnalysis{bottomPos: -1, topPos: -1}

	for i, frame := range stack {
		analyzeFrame(&analysis, frame, i, expectation)
	}

	return analysis
}

func analyzeFrame(analysis *stackAnalysis, frame core.Frame, position int, expectation stackTraceExpectation) {
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
	frame            *core.Frame
	expectedPkgName  string
	expectedFuncName string
}

func (tc frameSplitNameTestCase) Name() string {
	return tc.name
}

func (tc frameSplitNameTestCase) Test(t *testing.T) {
	pkgName, funcName := tc.frame.SplitName()
	core.AssertEqual(t, tc.expectedPkgName, pkgName, "package name")
	core.AssertEqual(t, tc.expectedFuncName, funcName, "function name")
}

func newFrameSplitNameTestCase(name string, frame *core.Frame, expectedPkgName,
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
		newFrameSplitNameTestCase("current function", core.Here(), "darvaza.org/core_test", "frameSplitNameTestCases"),
	}
}

func TestFrameSplitName(t *testing.T) {
	core.RunTestCases(t, frameSplitNameTestCases())
}

// Test case for Frame.Name method
type frameNameTestCase struct {
	name     string
	frame    *core.Frame
	expected string
}

func (tc frameNameTestCase) Name() string {
	return tc.name
}

func (tc frameNameTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.expected, tc.frame.Name(), "frame name")
}

func newFrameNameTestCase(name string, frame *core.Frame, expected string) frameNameTestCase {
	return frameNameTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func frameNameTestCases() []frameNameTestCase {
	return []frameNameTestCase{
		newFrameNameTestCase("current function", core.Here(), "darvaza.org/core_test.frameNameTestCases"),
		newFrameNameTestCase("zero value", &core.Frame{}, ""),
	}
}

func TestFrameName(t *testing.T) {
	core.RunTestCases(t, frameNameTestCases())
}

// Test case for Frame.PkgName method
type framePkgNameTestCase struct {
	name     string
	frame    *core.Frame
	expected string
}

func (tc framePkgNameTestCase) Name() string {
	return tc.name
}

func (tc framePkgNameTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.expected, tc.frame.PkgName(), "package name")
}

func newFramePkgNameTestCase(name string, frame *core.Frame, expected string) framePkgNameTestCase {
	return framePkgNameTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func framePkgNameTestCases() []framePkgNameTestCase {
	return []framePkgNameTestCase{
		newFramePkgNameTestCase("current function", core.Here(), "darvaza.org/core_test"),
		newFramePkgNameTestCase("empty frame", &core.Frame{}, ""),
		newFramePkgNameTestCase("no package frame", core.NewFrameForTest("func", "", 0), ""),
		newFramePkgNameTestCase("dot separator", core.NewFrameForTest("pkg.func", "", 0), "pkg"),
		newFramePkgNameTestCase("slash separator", core.NewFrameForTest("pkg/module.func", "", 0), "pkg/module"),
	}
}

// Test Frame.PkgName method (0% coverage)
func TestFramePkgName(t *testing.T) {
	core.RunTestCases(t, framePkgNameTestCases())
}

// Test Frame.File method (0% coverage)
func TestFrameFile(t *testing.T) {
	frame := core.Here()
	file := frame.File()

	// Should contain the test file name
	core.AssertContains(t, file, "stack_test.go", "frame file")

	// Test empty frame
	emptyFrame := &core.Frame{}
	core.AssertEqual(t, "", emptyFrame.File(), "empty frame file")

	// Test frame with file
	testFrame := core.NewFrameForTest("", "/path/to/file.go", 0)
	core.AssertEqual(t, "/path/to/file.go", testFrame.File(), "frame file path")
}

// Test case for Frame.PkgFile method
type framePkgFileTestCase struct {
	name     string
	expected string
	frame    core.Frame
}

func (tc framePkgFileTestCase) Name() string {
	return tc.name
}

func (tc framePkgFileTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.expected, tc.frame.PkgFile(), "PkgFile")
}

func newFramePkgFileTestCase(name string, frame core.Frame, expected string) framePkgFileTestCase {
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
			core.Frame{},
			"",
		),
		newFramePkgFileTestCase(
			"frame with file but no name",
			*core.NewFrameForTest("", "/path/to/file.go", 0),
			"file.go",
		),
		newFramePkgFileTestCase(
			"frame with absolute path and no package",
			*core.NewFrameForTest("main", "/absolute/path/to/source.go", 0),
			"source.go",
		),
		newFramePkgFileTestCase(
			"frame with package and file",
			*core.NewFrameForTest("github.com/user/repo.FuncName", "/go/src/github.com/user/repo/file.go", 0),
			"github.com/user/repo/file.go",
		),
		newFramePkgFileTestCase(
			"frame with nested package",
			*core.NewFrameForTest("internal/utils.Helper", "/workspace/project/internal/utils/helper.go", 0),
			"internal/utils/helper.go",
		),
		newFramePkgFileTestCase(
			"frame with standard library package",
			*core.NewFrameForTest("fmt.Printf", "/usr/local/go/src/fmt/print.go", 0),
			"fmt/print.go",
		),
		newFramePkgFileTestCase(
			"frame with generic function",
			*core.NewFrameForTest("example.com/pkg.GenericFunc[...]", "/src/generics.go", 0),
			"example.com/pkg/generics.go",
		),
		newFramePkgFileTestCase(
			"current test frame",
			*core.Here(),
			"darvaza.org/core_test/stack_test.go",
		),
	}

	core.RunTestCases(t, tests)
}

// Test Frame.Line method (0% coverage)
func TestFrameLine(t *testing.T) {
	frame := core.Here()
	line := frame.Line()

	// Should have a valid line number (greater than 0)
	core.AssertTrue(t, line > 0, "line %v positive", line)

	// Test empty frame
	emptyFrame := &core.Frame{}
	core.AssertEqual(t, 0, emptyFrame.Line(), "empty frame line")

	// Test frame with line
	testFrame := core.NewFrameForTest("", "", 42)
	core.AssertEqual(t, 42, testFrame.Line(), "frame line number")
}

// Test case for Frame.FileLine method
type frameFileLineTestCase struct {
	name     string
	expected string
	frame    core.Frame
}

func (tc frameFileLineTestCase) Name() string {
	return tc.name
}

func (tc frameFileLineTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.expected, tc.frame.FileLine(), "FileLine")
}

func newFrameFileLineTestCase(name, file string, line int, expected string) frameFileLineTestCase {
	return frameFileLineTestCase{
		name:     name,
		frame:    *core.NewFrameForTest("", file, line),
		expected: expected,
	}
}

func frameFileLineTestCases() []frameFileLineTestCase {
	return core.S(
		newFrameFileLineTestCase("frame with line", "test.go", 42, "test.go:42"),
		newFrameFileLineTestCase("frame without line", "test.go", 0, "test.go"),
		newFrameFileLineTestCase("empty frame", "", 0, ""),
		newFrameFileLineTestCase("frame with negative line", "test.go", -1, "test.go"),
	)
}

// Test Frame.FileLine method (0% coverage)
func TestFrameFileLine(t *testing.T) {
	core.RunTestCases(t, frameFileLineTestCases())
}

// Test case for Frame.String method (implements fmt.Stringer)
type frameStringTestCase struct {
	name     string
	frame    *core.Frame
	expected string
}

func (tc frameStringTestCase) Name() string {
	return tc.name
}

func (tc frameStringTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.expected, tc.frame.String(), "String")
}

func newFrameStringTestCase(name string, frame *core.Frame, expected string) frameStringTestCase {
	return frameStringTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func frameStringTestCases() []frameStringTestCase {
	populated := core.NewFrameForTest("darvaza.org/core.TestFunction", "/path/to/test.go", 42)

	return []frameStringTestCase{
		newFrameStringTestCase("populated frame", populated, "test.go:42"),
		newFrameStringTestCase("zero value", &core.Frame{}, ":0"),
	}
}

func TestFrameString(t *testing.T) {
	core.RunTestCases(t, frameStringTestCases())
}

// Test case for Frame.Format method
type frameFormatTestCase struct {
	name     string
	frame    *core.Frame
	format   string
	expected string
}

func (tc frameFormatTestCase) Name() string {
	return tc.name
}

func (tc frameFormatTestCase) Test(t *testing.T) {
	result := fmt.Sprintf(tc.format, tc.frame)
	core.AssertEqual(t, tc.expected, result, "formatted output")
}

func newFrameFormatTestCase(name string, frame *core.Frame, format, expected string) frameFormatTestCase {
	return frameFormatTestCase{
		name:     name,
		frame:    frame,
		format:   format,
		expected: expected,
	}
}

func frameFormatTestCases() []frameFormatTestCase {
	frame := core.NewFrameForTest("darvaza.org/core.TestFunction", "/path/to/test.go", 42)
	emptyFrame := &core.Frame{}

	return core.S(
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
	core.RunTestCases(t, frameFormatTestCases())
}

// capturedFrameFormatTestCase formats a frame captured from a real call,
// where frameFormatTestCases builds its frames from chosen fields.
type capturedFrameFormatTestCase struct {
	name     string
	frame    *core.Frame
	format   string
	expected string
}

func (tc capturedFrameFormatTestCase) Name() string {
	return tc.name
}

func (tc capturedFrameFormatTestCase) Test(t *testing.T) {
	result := fmt.Sprintf(tc.format, tc.frame)
	core.AssertEqual(t, tc.expected, result, "formatted output")
}

func newCapturedFrameFormatTestCase(name string, frame *core.Frame,
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
func capturedFrame(t *testing.T) (*core.Frame, string, int) {
	t.Helper()

	frame := core.StackFrame(1)
	_, file, line, ok := runtime.Caller(1)
	core.AssertMustTrue(t, ok, "runtime.Caller")

	return frame, file, line
}

func capturedFrameFormatTestCases(frame *core.Frame, file string,
	line int) []capturedFrameFormatTestCase {
	const funcName = "TestCapturedFrameFormat"
	const fullName = "darvaza.org/core_test." + funcName
	const baseFile = "stack_test.go"
	const pkgFile = "darvaza.org/core_test/" + baseFile

	lineText := strconv.Itoa(line)

	return core.S(
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
// preconditions state that StackFrame agrees with it on file and line.
func TestCapturedFrameFormat(t *testing.T) {
	frame, file, line := capturedFrame(t)

	core.AssertMustNotNil(t, frame, "captured frame")
	core.AssertMustEqual(t, file, frame.File(), "captured file")
	core.AssertMustEqual(t, line, frame.Line(), "captured line")

	core.RunTestCases(t, capturedFrameFormatTestCases(frame, file, line))
}

// capturedLineTestCase states the function and line that one frame of a
// captured stack resolves to. Where callerOfInlined is set, the row
// states the frame of a caller whose callee, the frame below it, was
// inlined into it, and it skips in a build where that did not happen.
type capturedLineTestCase struct {
	name            string
	wantFunc        string
	stack           core.Stack
	index           int
	wantLine        int
	callerOfInlined bool
}

func newCapturedLineTestCase(name string, stack core.Stack, index int,
	wantFunc string, wantLine int) capturedLineTestCase {
	return capturedLineTestCase{
		stack:    stack,
		name:     name,
		wantFunc: wantFunc,
		index:    index,
		wantLine: wantLine,
	}
}

func newCapturedLineInlinedTestCase(name string, stack core.Stack, index int,
	wantFunc string, wantLine int) capturedLineTestCase {
	tc := newCapturedLineTestCase(name, stack, index, wantFunc, wantLine)
	tc.callerOfInlined = true
	return tc
}

func (tc capturedLineTestCase) Name() string {
	return tc.name
}

func (tc capturedLineTestCase) Test(t *testing.T) {
	core.AssertMustTrue(t, tc.index < len(tc.stack), "frame %v captured", tc.index)
	if tc.callerOfInlined {
		core.AssertMustTrue(t, tc.index > 0, "callee frame captured")
		callee := core.FrameEntryForTest(tc.stack[tc.index-1])
		if callee != core.FrameEntryForTest(tc.stack[tc.index]) {
			t.Skip("the callee is not inlined in this build")
		}
	}

	frame := tc.stack[tc.index]
	core.AssertEqual(t, tc.wantFunc, frame.FuncName(), "function")
	core.AssertEqual(t, tc.wantLine, frame.Line(), "line")
}

// stackFromLastCall captures its stack through a call that ends the
// function, so the address the call returns to falls past the line of
// the call. It reports, through above, the line just above the call.
//
//go:noinline
func stackFromLastCall(st *core.Stack, above *int) {
	_, _, *above, _ = runtime.Caller(0)
	captureStack(st)
}

// captureStack stores the stack above it in st.
//
//go:noinline
func captureStack(st *core.Stack) {
	*st = core.StackTrace(1)
}

// stackThroughInlined captures its stack through stackInlined, which is
// small enough to inline into it, and reports the line just above that
// call.
//
//go:noinline
func stackThroughInlined() (st core.Stack, above int) {
	_, _, above, _ = runtime.Caller(0)
	st = stackInlined()
	return st, above
}

func stackInlined() core.Stack {
	return core.StackTrace(0)
}

func capturedLineTestCases() []capturedLineTestCase {
	var lastCall core.Stack
	var lastCallAbove int
	stackFromLastCall(&lastCall, &lastCallAbove)

	inlined, inlinedAbove := stackThroughInlined()

	return core.S(
		newCapturedLineTestCase("call ending its function", lastCall, 0,
			"stackFromLastCall", lastCallAbove+1),
		newCapturedLineInlinedTestCase("caller of an inlined function", inlined, 1,
			"stackThroughInlined", inlinedAbove+1),
	)
}

// Test that a captured frame reports the line of its call, where the
// address the call returns to sits on a later line, and where the
// function it calls is inlined into it.
func TestCapturedLine(t *testing.T) {
	core.RunTestCases(t, capturedLineTestCases())
}

// stackFormatTestCase tests Stack.Format by the substrings its output
// contains, or, where wantEmpty is set, by the output being empty, which
// no substring can state: every string contains "".
type stackFormatTestCase struct {
	name      string
	stack     core.Stack
	format    string
	contains  []string
	wantEmpty bool
}

func (tc stackFormatTestCase) Name() string {
	return tc.name
}

func (tc stackFormatTestCase) Test(t *testing.T) {
	result := fmt.Sprintf(tc.format, tc.stack)
	if tc.wantEmpty {
		core.AssertEqual(t, "", result, "output")
		return
	}

	for _, expected := range tc.contains {
		core.AssertContains(t, result, expected, "output")
	}
}

// newStackFormatTestCase declares a row by the substrings the output must
// contain, and every row must name at least one: an empty list would leave
// the Test body asserting nothing at all.
func newStackFormatTestCase(name string, stack core.Stack, format string, contains []string) stackFormatTestCase {
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

func newStackFormatTestCaseEmpty(name string, stack core.Stack, format string) stackFormatTestCase {
	return stackFormatTestCase{
		name:      name,
		stack:     stack,
		format:    format,
		wantEmpty: true,
	}
}

func stackFormatTestCases() []stackFormatTestCase {
	stack := core.Stack{
		*core.NewFrameForTest("darvaza.org/core.func1", "/path/to/file1.go", 10),
		*core.NewFrameForTest("darvaza.org/core.func2", "/path/to/file2.go", 20),
	}
	emptyStack := core.Stack{}

	return core.S(
		newStackFormatTestCase("basic format %s", stack, "%s", core.S("\nfile1.go", "\nfile2.go")),
		newStackFormatTestCase("verbose format %+v", stack, "%+v",
			core.S("\ndarvaza.org/core.func1\n\t/path/to/file1.go:10",
				"\ndarvaza.org/core.func2\n\t/path/to/file2.go:20")),
		newStackFormatTestCase("numbered format %#+v", stack, "%#+v",
			core.S("\n[0/2] darvaza.org/core.func1\n\t/path/to/file1.go:10",
				"\n[1/2] darvaza.org/core.func2\n\t/path/to/file2.go:20")),
		newStackFormatTestCase("numbered name format %#+n", stack, "%#+n",
			core.S("\n[0/2] darvaza.org/core.func1", "\n[1/2] darvaza.org/core.func2")),
		newStackFormatTestCaseEmpty("empty stack", emptyStack, "%+v"),
	)
}

// Test Stack.Format method (0% coverage)
func TestStackFormat(t *testing.T) {
	core.RunTestCases(t, stackFormatTestCases())
}

// Test case for Stack.String method (implements fmt.Stringer)
type stackStringTestCase struct {
	name     string
	expected string
	stack    core.Stack
}

func (tc stackStringTestCase) Name() string {
	return tc.name
}

func (tc stackStringTestCase) Test(t *testing.T) {
	core.AssertEqual(t, tc.expected, tc.stack.String(), "String")
}

func newStackStringTestCase(name string, stack core.Stack, expected string) stackStringTestCase {
	return stackStringTestCase{
		name:     name,
		stack:    stack,
		expected: expected,
	}
}

func stackStringTestCases() []stackStringTestCase {
	stack := core.Stack{
		*core.NewFrameForTest("darvaza.org/core.func1", "/path/to/file1.go", 10),
		*core.NewFrameForTest("darvaza.org/core.func2", "/path/to/file2.go", 20),
	}

	return []stackStringTestCase{
		newStackStringTestCase("two frames", stack, "\nfile1.go:10\nfile2.go:20"),
		newStackStringTestCase("empty stack", core.Stack{}, ""),
	}
}

func TestStackString(t *testing.T) {
	core.RunTestCases(t, stackStringTestCases())
}

// Test case for formatLine function
type formatLineTestCase struct {
	name     string
	frame    *core.Frame
	expected string
}

func (tc formatLineTestCase) Name() string {
	return tc.name
}

func (tc formatLineTestCase) Test(t *testing.T) {
	result := fmt.Sprintf("%d", tc.frame)
	core.AssertEqual(t, tc.expected, result, "formatted line")
}

func newFormatLineTestCase(name string, frame *core.Frame, expected string) formatLineTestCase {
	return formatLineTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func formatLineTestCases() []formatLineTestCase {
	return core.S(
		newFormatLineTestCase("positive line", core.NewFrameForTest("", "", 123), "123"),
		newFormatLineTestCase("zero line", &core.Frame{}, "0"),
	)
}

// Test formatLine function directly (0% coverage)
func TestFormatLineMethod(t *testing.T) {
	core.RunTestCases(t, formatLineTestCases())
}

// Cover the early-return branch of StackFrame when skip exceeds depth.
func TestStackFrameSkipExceeds(t *testing.T) {
	core.AssertNil(t, core.StackFrame(10000), "StackFrame returns nil when skip > depth")
}

// Cover the same branch through StackTrace, which yields an empty Stack
// rather than a nil Frame. Statement coverage cannot show this one: the
// arm it takes holds no statements.
func TestStackTraceSkipExceeds(t *testing.T) {
	core.AssertEqual(t, 0, len(core.StackTrace(10000)), "stack length")
}

// Cover the negative-skip guard, which rejects rather than clamping to zero.
// Without it the buffer is sliced from a negative index, so the assertion
// that matters is that the call returns at all.
func TestStackFrameNegativeSkip(t *testing.T) {
	var frame *core.Frame

	core.AssertMustNoPanic(t, func() { frame = core.StackFrame(-1) }, "StackFrame(-1)")
	core.AssertNil(t, frame, "frame")
}

func TestStackTraceNegativeSkip(t *testing.T) {
	var stack core.Stack

	core.AssertMustNoPanic(t, func() { stack = core.StackTrace(-1) }, "StackTrace(-1)")
	core.AssertEqual(t, 0, len(stack), "stack length")
}

// StackTrace documents a capture limited to MaxDepth frames, counted after
// the skip. Recurse past it so the buffer, rather than the stack, decides
// the length.
type stackTraceMaxDepthTestCase struct {
	name string
	skip int
}

func newStackTraceMaxDepthTestCase(name string, skip int) stackTraceMaxDepthTestCase {
	return stackTraceMaxDepthTestCase{name: name, skip: skip}
}

func (tc stackTraceMaxDepthTestCase) Name() string { return tc.name }

func (tc stackTraceMaxDepthTestCase) Test(t *testing.T) {
	stack := deepestStackTrace(core.MaxDepth, tc.skip)
	core.AssertEqual(t, core.MaxDepth, len(stack), "stack length")
}

func TestStackTraceMaxDepth(t *testing.T) {
	core.RunTestCases(t, core.S(
		newStackTraceMaxDepthTestCase("no skip", 0),
		newStackTraceMaxDepthTestCase("skip one", 1),
	))
}

func deepestStackTrace(depth, skip int) core.Stack {
	if depth > 0 {
		return deepestStackTrace(depth-1, skip)
	}
	return core.StackTrace(skip)
}

// StackFrame reaches a caller MaxDepth levels up when the stack is deep
// enough: the skip is counted before the capture, not out of it.
func TestStackFrameMaxDepth(t *testing.T) {
	frame := deepestStackFrame(core.MaxDepth+1, core.MaxDepth)
	core.AssertMustNotNil(t, frame, "frame")
	core.AssertEqual(t, "deepestStackFrame", fmt.Sprintf("%n", frame), "frame name")
}

func deepestStackFrame(depth, skip int) *core.Frame {
	if depth > 0 {
		return deepestStackFrame(depth-1, skip)
	}
	return core.StackFrame(skip)
}
