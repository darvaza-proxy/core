package core

import (
	"fmt"
	"log"
	"strings"
	"testing"
)

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

// revive:disable:cognitive-complexity
func TestStackTrace(t *testing.T) {
	// revive:enable:cognitive-complexity
	stack := StackTrace(0)
	if len(stack) < 2 || fmt.Sprintf("%n", stack[0]) != "TestStackTrace" {
		t.Fatalf("StackTrace(%v): %s", 0, fmt.Sprintln(stack))
	}

	for i := 0; i < MaxTestDepth; i++ {
		for j := 0; j < MaxTestSpace; j++ {
			stack := deepStackTrace(i, j)
			if !checkDeepStackTrace(stack, i) {
				t.Fatalf("StackTrace(%v, %v): %v: %s",
					i, j, len(stack), fmt.Sprintf("%n", stack))
			}
		}
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

func checkStackFrameName(stack Stack, offset int, name string) bool {
	if len(stack) > offset {
		s := fmt.Sprintf("%n", stack[offset])
		return s == name
	}
	return false
}

func checkDeepStackTrace(stack Stack, depth int) bool {
	t0 := 2 + depth
	if !checkStackFrameName(stack, t0, "TestStackTrace") {
		return false
	}
	if !checkStackFrameName(stack, 0, "deeperStackTrace") {
		return false
	}
	for i := 1; i < t0; i++ {
		if !checkStackFrameName(stack, i, "deepStackTrace") {
			return false
		}
	}
	return true
}

func TestFrameSplitName(t *testing.T) {
	for _, tc := range []struct {
		name             string
		frame            *Frame
		expectedPkgName  string
		expectedFuncName string
	}{
		{"current function", Here(), "darvaza.org/core", "TestFrameSplitName"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pkgName, funcName := tc.frame.SplitName()
			assertFrameNames(t, tc.expectedPkgName, tc.expectedFuncName, pkgName, funcName)
		})
	}
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

// Test Frame.Name method (0% coverage)
func TestFrameName(t *testing.T) {
	testCases := []struct {
		name     string
		frame    *Frame
		expected string
	}{
		{"current function", Here(), "darvaza.org/core.TestFrameName"},
		{"nil frame", nil, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.frame == nil {
				// Test nil Frame - create empty frame to test Name method
				var frame Frame
				AssertEqual(t, "", frame.Name(), "empty frame should return empty name")
			} else {
				AssertEqual(t, tc.expected, tc.frame.Name(), "frame name mismatch")
			}
		})
	}
}

// Test Frame.PkgName method (0% coverage)
func TestFramePkgName(t *testing.T) {
	testCases := []struct {
		name     string
		frame    *Frame
		expected string
	}{
		{"current function", Here(), "darvaza.org/core"},
		{"empty frame", &Frame{name: ""}, ""},
		{"no package frame", &Frame{name: "func"}, ""},
		{"dot separator", &Frame{name: "pkg.func"}, "pkg"},
		{"slash separator", &Frame{name: "pkg/module.func"}, "pkg/module"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			AssertEqual(t, tc.expected, tc.frame.PkgName(), "package name mismatch")
		})
	}
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
	AssertEqual(t, "", emptyFrame.File(), "empty frame should return empty file")

	// Test frame with file
	testFrame := &Frame{file: "/path/to/file.go"}
	AssertEqual(t, "/path/to/file.go", testFrame.File(), "frame should return file path")
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
	AssertEqual(t, 0, emptyFrame.Line(), "empty frame should return 0 line")

	// Test frame with line
	testFrame := &Frame{line: 42}
	AssertEqual(t, 42, testFrame.Line(), "frame should return line number")
}

// Test case for Frame.FileLine method
type frameFileLineTestCase struct {
	name     string
	expected string
	frame    Frame
}

func (tc frameFileLineTestCase) test(t *testing.T) {
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
	for _, tc := range frameFileLineTestCases() {
		t.Run(tc.name, tc.test)
	}
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

	AssertEqual(t, expected, actual, "String() should be equivalent to %v format")

	// Test with empty frame
	emptyFrame := &Frame{}
	expectedEmpty := fmt.Sprintf("%v", emptyFrame)
	actualEmpty := emptyFrame.String()

	AssertEqual(t, expectedEmpty, actualEmpty, "String() should work for empty frame")
}

// Test case for Frame.Format method
type frameFormatTestCase struct {
	name     string
	frame    *Frame
	format   string
	expected string
}

func (tc frameFormatTestCase) test(t *testing.T) {
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
		newFrameFormatTestCase("line format %d", frame, "%d", "42"),
		newFrameFormatTestCase("name format %n", frame, "%n", "TestFunction"),
		newFrameFormatTestCase("name format with + flag %+n", frame, "%+n", "darvaza.org/core.TestFunction"),
		newFrameFormatTestCase("file:line format %v", frame, "%v", "test.go:42"),
		newFrameFormatTestCase("file:line format with + flag %+v", frame, "%+v",
			"darvaza.org/core.TestFunction\n\t/path/to/test.go:42"),
		newFrameFormatTestCase("empty file %s", emptyFrame, "%s", ""),
		newFrameFormatTestCase("empty line %d", emptyFrame, "%d", "0"),
		newFrameFormatTestCase("empty name %n", emptyFrame, "%n", ""),
		newFrameFormatTestCase("empty file:line %v", emptyFrame, "%v", ":0"),
	)
}

// Test Frame.Format method and helper functions (0% coverage)
func TestFrameFormat(t *testing.T) {
	for _, tc := range frameFormatTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// Test case for Stack.Format method
type stackFormatTestCase struct {
	name     string
	stack    Stack
	format   string
	contains []string
}

func (tc stackFormatTestCase) test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf(tc.format, tc.stack)

	// Special case for empty stack - check exact match
	if len(tc.stack) == 0 && len(tc.contains) == 1 && tc.contains[0] == "" {
		AssertEqual(t, "", result, "empty stack should produce empty output")
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
	for _, tc := range stackFormatTestCases() {
		t.Run(tc.name, tc.test)
	}
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

	AssertEqual(t, expected, actual, "String() should be equivalent to %v format")

	// Test with empty stack
	emptyStack := Stack{}
	expectedEmpty := fmt.Sprintf("%v", emptyStack)
	actualEmpty := emptyStack.String()

	AssertEqual(t, expectedEmpty, actualEmpty, "String() should work for empty stack")
}

// Test case for formatLine function
type formatLineTestCase struct {
	name     string
	frame    *Frame
	expected string
}

func (tc formatLineTestCase) test(t *testing.T) {
	t.Helper()
	result := fmt.Sprintf("%d", tc.frame)
	AssertEqual(t, tc.expected, result, "formatLine should format line number")
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
	for _, tc := range formatLineTestCases() {
		t.Run(tc.name, tc.test)
	}
}

// Test case for writeFormat edge cases
type writeFormatTestCase struct {
	name   string
	format string
}

func (tc writeFormatTestCase) test(t *testing.T) {
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
	for _, tc := range writeFormatTestCases() {
		t.Run(tc.name, tc.test)
	}
}
