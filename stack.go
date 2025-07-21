package core

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// CallStacker represents an object that can provide its call stack.
// Types implementing this interface can be used for stack trace collection
// and debugging purposes.
type CallStacker interface {
	CallStack() Stack
}

const (
	// MaxDepth is the maximum depth we will go in the stack.
	MaxDepth = 32
)

// Frame represents a single function call frame in a call stack.
// It captures function name, source file, line number, and program counter
// information at the time of stack capture.
//
// This implementation is heavily based on github.com/pkg/errors.Frame
// but resolves all information immediately for efficient later consumption.
// Each Frame contains:
//   - Function name (package.function)
//   - Source file path
//   - Line number
//   - Program counter and entry point
type Frame struct {
	// string fields - alphabetically ordered
	file string
	name string

	// uintptr fields - alphabetically ordered
	entry uintptr
	pc    uintptr

	// int fields - single field
	line int
}

func frameForPC(pc uintptr) Frame {
	var entry uintptr
	var name string
	var file string
	var line int

	if fp := runtime.FuncForPC(pc - 1); fp != nil {
		entry = fp.Entry()
		name = fp.Name()
		file, line = fp.FileLine(pc)
	} else {
		name = "unknown"
		file = "unknown"
	}

	return Frame{
		pc:    pc,
		entry: entry,
		name:  name,
		file:  file,
		line:  line,
	}
}

// Name returns the full qualified function name including package path.
// For example: "darvaza.org/core.TestFunction" or "main.main".
// Returns empty string for zero-valued frames.
func (f Frame) Name() string {
	return f.name
}

// FuncName returns only the function name without package qualification.
// For example: "TestFunction" or "main".
// Uses SplitName internally to separate package from function name.
func (f Frame) FuncName() string {
	_, s := f.SplitName()
	return s
}

// PkgName returns the package path portion of the function name.
// For example: "darvaza.org/core" or "main".
// Returns empty string if no package separator is found.
func (f Frame) PkgName() string {
	s, _ := f.SplitName()
	return s
}

// SplitName splits the full function name into package and function
// components. Handles generic function names by ignoring trailing "[...]"
// suffixes. Searches for the last "." or "/" separator to determine the
// split point.
//
// Returns:
//   - pkgName: package path ("darvaza.org/core")
//   - funcName: function name ("TestFunction")
//
// If no separator is found, returns an empty package name and the full
// name as the function name.
func (f Frame) SplitName() (pkgName, funcName string) {
	// ignore trailing[...] from generic functions
	name, _ := strings.CutSuffix(f.name, "[...]")
	i := strings.LastIndexAny(name, "./")
	if i < 0 {
		return "", f.name
	}
	return f.name[:i], f.name[i+1:]
}

// File returns the full path to the source file containing this frame's function.
// Returns empty string for zero-valued frames or when file information is unavailable.
func (f Frame) File() string {
	return f.file
}

// PkgFile returns the package name (if present) followed by '/' and the
// file name. For example: "darvaza.org/core/stack.go".
// If no package name exists, returns just the file name.
// Returns empty string for zero-valued frames or when file information
// is unavailable.
func (f Frame) PkgFile() string {
	if f.file == "" {
		return ""
	}

	pkgName := f.PkgName()
	fileName := path.Base(f.file)
	if pkgName == "" {
		return fileName
	}

	return pkgName + "/" + fileName
}

// Line returns the line number within the source file for this frame.
// Returns 0 for zero-valued frames or when line information is unavailable.
func (f Frame) Line() int {
	return f.line
}

// FileLine returns a formatted "file:line" string for this frame.
// If line number is available and positive, returns "filename:123".
// If line is zero or negative, returns just the filename.
// Useful for displaying source location in error messages and logs.
func (f Frame) FileLine() string {
	if f.line > 0 {
		return fmt.Sprintf("%s:%v", f.file, f.line)
	}

	return f.file
}

// String returns a string representation of the frame equivalent to %v format.
// This implements the fmt.Stringer interface and returns "file:line" format
// using the same logic as FileLine but with basename only for the file.
func (f Frame) String() string {
	return fmt.Sprintf("%v", f)
}

// Format formats the frame according to the fmt.Formatter interface.
//
// The following verbs are supported:
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcName>\n\t<path>)
//	%+n   full package name followed by function name
//	%+v   equivalent to %+s:%d
//	%#s   package/file format using PkgFile()
//	%#v   equivalent to %#s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		f.formatFile(s)
	case 'd':
		f.formatLine(s)
	case 'n':
		f.formatName(s)
	case 'v':
		f.formatFile(s)
		writeFormat(s, ":")
		f.formatLine(s)
	}
}

func (f Frame) formatFile(s fmt.State) {
	switch {
	case s.Flag('+'):
		writeFormat(s, f.name)
		writeFormat(s, "\n\t")
		writeFormat(s, f.file)
	case s.Flag('#'):
		writeFormat(s, f.PkgFile())
	case f.file == "":
		writeFormat(s, "")
	default:
		writeFormat(s, path.Base(f.file))
	}
}

func (f Frame) formatLine(s fmt.State) {
	writeFormat(s, strconv.Itoa(f.line))
}

func (f Frame) formatName(s fmt.State) {
	var name string
	if s.Flag('+') {
		name = f.Name()
	} else {
		name = f.FuncName()
	}
	writeFormat(s, name)
}

func writeFormat(s io.Writer, str string) {
	n, err := io.WriteString(s, str)
	if err != nil {
		panic(fmt.Errorf("Frame: failed to write %q to buffer", str))
	} else if l := len(str); n < l {
		panic(fmt.Errorf("Frame: incomplete write (%v/%v)", n, l))
	}
}

// Stack represents a captured call stack as a slice of Frame objects.
// Each Frame in the stack corresponds to a function call, ordered from
// the most recent call (index 0) to the oldest call.
//
// Stack implements custom formatting via the Format method, supporting
// various output formats for debugging and logging purposes.
type Stack []Frame

// Format formats the entire stack using the same verbs as Frame.Format
// with additional support for the '#' flag for numbered output.
//
// Supported format verbs (same as Frame.Format):
//
//	%s, %d, %n, %v    - basic formatting
//	%+s, %+n, %+v     - verbose formatting with full paths/names
//
// Additional '#' flag support:
//
//	%#s, %#n, %#v     - each frame on new line
//	%#+s, %#+n, %#+v  - numbered frames with [index/total] prefix
//
// Example outputs:
//
//	fmt.Printf("%+v", stack)   - multi-line stack with full info
//	fmt.Printf("%#+v", stack)  - numbered multi-line stack
//	fmt.Printf("%#n", stack)   - function names only, numbered
func (st Stack) Format(s fmt.State, verb rune) {
	if s.Flag('#') {
		l := len(st)
		for i, f := range st {
			writeFormat(s, fmt.Sprintf("\n[%v/%v] ", i, l))
			f.Format(s, verb)
		}
	} else {
		for _, f := range st {
			writeFormat(s, "\n")
			f.Format(s, verb)
		}
	}
}

// String returns a string representation of the stack equivalent to %v format.
// This implements the fmt.Stringer interface and returns a multi-line representation
// with each frame formatted as "file:line" on separate lines.
func (st Stack) String() string {
	return fmt.Sprintf("%v", st)
}

// Here captures and returns the current stack frame where it was called.
// This is useful for capturing the immediate calling context for debugging.
//
// Returns nil if stack capture fails or if called in an environment where
// runtime stack information is unavailable.
//
// Example:
//
//	frame := Here()
//	if frame != nil {
//	    fmt.Printf("Called from %s at %s", frame.FuncName(), frame.FileLine())
//	}
func Here() *Frame {
	const depth = 1
	var pcs [depth]uintptr

	if n := runtime.Callers(2, pcs[:]); n > 0 {
		f := frameForPC(pcs[0])
		return &f
	}
	return nil
}

// StackFrame captures a specific frame in the call stack, skipping the
// specified number of stack levels above the caller.
//
// Parameters:
//   - skip: number of stack frames to skip (0 = caller, 1 = caller's caller, etc.)
//
// Returns nil if the stack doesn't have enough frames or if capture fails.
// Useful for creating stack-aware error reporting utilities.
//
// Example:
//
//	// Get the frame 2 levels up the stack
//	frame := StackFrame(2)
//	if frame != nil {
//	    log.Printf("Error originated from %s", frame.Name())
//	}
func StackFrame(skip int) *Frame {
	const depth = MaxDepth
	var pcs [depth]uintptr

	if n := runtime.Callers(2, pcs[:]); n > skip {
		f := frameForPC(pcs[skip])
		return &f
	}

	return nil
}

// StackTrace captures a complete call stack starting from the specified
// skip level. The resulting Stack contains frames ordered from most recent
// to oldest call.
//
// Parameters:
//   - skip: number of initial stack frames to skip before capture begins
//
// Returns an empty Stack if capture fails or if there are insufficient frames.
// The maximum capture depth is limited by MaxDepth (32 frames).
//
// This function is commonly used for error reporting, debugging, and logging
// where complete call context is needed.
//
// Example:
//
//	// Capture stack excluding this function and its caller
//	stack := StackTrace(2)
//	for i, frame := range stack {
//	    fmt.Printf("[%d] %s at %s", i, frame.Name(), frame.FileLine())
//	}
func StackTrace(skip int) Stack {
	const depth = MaxDepth
	var pcs [depth]uintptr
	var st Stack

	if n := runtime.Callers(2, pcs[:]); n > skip {
		var frames []Frame

		frames = make([]Frame, 0, n-skip)

		for _, pc := range pcs[skip:n] {
			frames = append(frames, frameForPC(pc))
		}

		st = Stack(frames)
	}

	return st
}
