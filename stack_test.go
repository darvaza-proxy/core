package core

import (
	"fmt"
	"log"
	"strings"
	"testing"
)

// TestCase validations
var _ TestCase = frameSplitNameTestCase{}
var _ TestCase = frameNameTestCase{}
var _ TestCase = framePkgNameTestCase{}
var _ TestCase = framePkgFileTestCase{}
var _ TestCase = frameFileLineTestCase{}
var _ TestCase = frameFormatTestCase{}
var _ TestCase = stackFormatTestCase{}
var _ TestCase = formatLineTestCase{}
var _ TestCase = writeFormatTestCase{}

const (
	MaxTestDepth = 16
	MaxTestSpace = 6
)

func TestHere(t *testing.T) {
	if s := fmt.Sprintf("%n", Here()); s != "TestHere" {
		t.FailNow()
	}
	for i := 0; i < MaxDepth; i++ {
		f := deepHere(i)
		if s := fmt.Sprintf("%n", f); s != "hereHere" {
			t.FailNow()
		}
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
	if s := fmt.Sprintf("%n", StackFrame(0)); s != "TestStackFrame" {
		log.Print(s)
		t.FailNow()
	}

	for i := 0; i < MaxTestDepth; i++ {
		for j := 0; j < MaxTestSpace; j++ {
			f := deepStackFrame(i, j)
			if s := fmt.Sprintf("%n", f); s != "hereStackFrame" {
				log.Print(s)
				t.FailNow()
			}
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
	if len(stack) < 2 || fmt.Sprintf("%n", stack[0]) != "TestStackTrace" {
		t.Fatalf("StackTrace(%v): %s", 0, fmt.Sprintln(stack))
	}

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

	if !validateStackTrace(stack, expectation) {
		t.Fatalf("StackTrace(%v, %v): %v: %s",
			depth, space, len(stack), fmt.Sprintf("%n", stack))
	}
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

func validateStackTrace(stack Stack, expectation stackTraceExpectation) bool {
	analysis := analyzeStackTrace(stack, expectation)
	return validateStackAnalysis(analysis, expectation)
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
	}
}

func validateStackAnalysis(analysis stackAnalysis, expectation stackTraceExpectation) bool {
	return analysis.bottomPos == 0 &&
		analysis.topPos != -1 &&
		analysis.recurringCount == expectation.expectedCount
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
	assertFrameNames(t, tc.expectedPkgName, tc.expectedFuncName, pkgName, funcName)
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

func assertFrameNames(t *testing.T, expectedPkg, expectedFunc, actualPkg, actualFunc string) {
	t.Helper()
	if actualPkg != expectedPkg {
		t.Errorf("Expected package name '%s', got '%s'", expectedPkg, actualPkg)
	}
	if actualFunc != expectedFunc {
		t.Errorf("Expected function name '%s', got '%s'", expectedFunc, actualFunc)
	}
}

// Test case for Frame.Name method
type frameNameTestCase struct {
	name     string
	frame    *Frame
	expected string
}

func (tc frameNameTestCase) Name() string {
	return tc.name
}

func (tc frameNameTestCase) Test(t *testing.T) {
	t.Helper()
	if tc.frame == nil {
		// Test nil Frame - create empty frame to test Name method
		var frame Frame
		AssertEqual(t, "", frame.Name(), "empty frame name")
	} else {
		AssertEqual(t, tc.expected, tc.frame.Name(), "frame name mismatch")
	}
}

func newFrameNameTestCase(name string, frame *Frame, expected string) frameNameTestCase {
	return frameNameTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func frameNameTestCases() []frameNameTestCase {
	return []frameNameTestCase{
		newFrameNameTestCase("current function", Here(), "darvaza.org/core.frameNameTestCases"),
		newFrameNameTestCase("nil frame", nil, ""),
	}
}

// Test Frame.Name method (0% coverage)
func TestFrameName(t *testing.T) {
	RunTestCases(t, frameNameTestCases())
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
	AssertEqual(t, tc.expected, tc.frame.PkgName(), "package name mismatch")
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
	if !strings.Contains(file, "stack_test.go") {
		t.Errorf("Expected file to contain 'stack_test.go', got '%s'", file)
	}

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
	AssertEqual(t, tc.expected, tc.frame.PkgFile(), "PkgFile output mismatch")
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
	if line <= 0 {
		t.Errorf("Expected positive line number, got %d", line)
	}

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
	AssertEqual(t, tc.expected, tc.frame.FileLine(), "FileLine output mismatch")
}

func newFrameFileLineTestCase(name string, frame Frame, expected string) frameFileLineTestCase {
	return frameFileLineTestCase{
		name:     name,
		frame:    frame,
		expected: expected,
	}
}

func frameFileLineTestCases() []frameFileLineTestCase {
	return S(
		newFrameFileLineTestCase("frame with line", Frame{file: "test.go", line: 42}, "test.go:42"),
		newFrameFileLineTestCase("frame without line", Frame{file: "test.go", line: 0}, "test.go"),
		newFrameFileLineTestCase("empty frame", Frame{file: "", line: 0}, ""),
		newFrameFileLineTestCase("frame with negative line", Frame{file: "test.go", line: -1}, "test.go"),
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

	// String() should be equivalent to %v format
	expected := fmt.Sprintf("%v", frame)
	actual := frame.String()

	AssertEqual(t, expected, actual, "String format")

	// Test with empty frame
	emptyFrame := &Frame{}
	expectedEmpty := fmt.Sprintf("%v", emptyFrame)
	actualEmpty := emptyFrame.String()

	AssertEqual(t, expectedEmpty, actualEmpty, "empty frame String")
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
	AssertEqual(t, tc.expected, result, "format output mismatch")
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
	)
}

// Test Frame.Format method and helper functions (0% coverage)
func TestFrameFormat(t *testing.T) {
	RunTestCases(t, frameFormatTestCases())
}

// Test case for Stack.Format method
type stackFormatTestCase struct {
	name     string
	stack    Stack
	format   string
	contains []string
}

func (tc stackFormatTestCase) Name() string {
	return tc.name
}

func (tc stackFormatTestCase) Test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf(tc.format, tc.stack)

	// Special case for empty stack - check exact match
	if len(tc.stack) == 0 && len(tc.contains) == 1 && tc.contains[0] == "" {
		AssertEqual(t, "", result, "empty stack output")
		return
	}

	for _, expected := range tc.contains {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got '%s'", expected, result)
		}
	}
}

func newStackFormatTestCase(name string, stack Stack, format string, contains []string) stackFormatTestCase {
	return stackFormatTestCase{
		name:     name,
		stack:    stack,
		format:   format,
		contains: contains,
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
		newStackFormatTestCase("empty stack", emptyStack, "%+v", S("")),
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

	// String() should be equivalent to %v format
	expected := fmt.Sprintf("%v", stack)
	actual := stack.String()

	AssertEqual(t, expected, actual, "String format")

	// Test with empty stack
	emptyStack := Stack{}
	expectedEmpty := fmt.Sprintf("%v", emptyStack)
	actualEmpty := emptyStack.String()

	AssertEqual(t, expectedEmpty, actualEmpty, "empty stack String")
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

// Test case for writeFormat edge cases
type writeFormatTestCase struct {
	name   string
	format string
}

func (tc writeFormatTestCase) Name() string {
	return tc.name
}

func (tc writeFormatTestCase) Test(t *testing.T) {
	t.Helper()
	frame := &Frame{name: "test", file: "test.go", line: 10}
	result := fmt.Sprintf(tc.format, frame)
	// Just ensure it doesn't panic and produces some output
	if result == "" {
		t.Errorf("Expected non-empty result for format %s", tc.format)
	}
}

func newWriteFormatTestCase(name, format string) writeFormatTestCase {
	return writeFormatTestCase{
		name:   name,
		format: format,
	}
}

func writeFormatTestCases() []writeFormatTestCase {
	return S(
		newWriteFormatTestCase("basic file", "%s"),
		newWriteFormatTestCase("basic line", "%d"),
		newWriteFormatTestCase("basic name", "%n"),
		newWriteFormatTestCase("plus file", "%+s"),
		newWriteFormatTestCase("plus name", "%+n"),
		newWriteFormatTestCase("file:line", "%v"),
		newWriteFormatTestCase("plus file:line", "%+v"),
	)
}

// Test writeFormat error conditions for better coverage (60% -> higher)
func TestWriteFormatEdgeCases(t *testing.T) {
	RunTestCases(t, writeFormatTestCases())
}
