package core

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// CallStacker represents an object with a method CallStack()
// returning a Stack
type CallStacker interface {
	CallStack() Stack
}

const (
	// MaxDepth is the maximum depth we will go in the stack.
	MaxDepth = 32
)

// Frame represents a function call on the call Stack.
// This implementation is heavily based on
// github.com/pkg/errors.Frame but all parts are resolved
// immediately for later consumption.
type Frame struct {
	pc    uintptr
	entry uintptr
	name  string
	file  string
	line  int
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

// Name returns the name of the function,
// including package name
func (f Frame) Name() string {
	return f.name
}

// FuncName returns the name of the function,
// without the package name
func (f Frame) FuncName() string {
	_, s := f.SplitName()
	return s
}

// PkgName returns the package name
func (f Frame) PkgName() string {
	s, _ := f.SplitName()
	return s
}

// SplitName returns package name and function name
func (f Frame) SplitName() (pkgName string, funcName string) {
	i := strings.LastIndexAny(f.name, "./")
	if i < 0 {
		return "", f.name
	}
	return f.name[:i], f.name[i+1:]
}

// File returns the file name of the source code
// corresponding to this Frame
func (f Frame) File() string {
	return f.file
}

// Line returns the file number on the source code
// corresponding to this Frame, or zero if unknown.
func (f Frame) Line() int {
	return f.line
}

// FileLine returns File name and Line separated by
// a colon, or only the filename if the Line isn't known
func (f Frame) FileLine() string {
	if f.line > 0 {
		return fmt.Sprintf("%s:%v", f.file, f.line)
	}

	return f.file
}

/* Format formats the frame according to the fmt.Formatter interface.
 *
 *	%s    source file
 *	%d    source line
 *	%n    function name
 *	%v    equivalent to %s:%d
 *
 * Format accepts flags that alter the printing of some verbs, as follows:
 *
 *	%+s   function name and path of source file relative to the compile time
 *	      GOPATH separated by \n\t (<funcname>\n\t<path>)
 *	%+n   full package name followed by function name
 *  %+v   equivalent to %+s:%d
 */
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

// Stack is an snapshot of the call stack in
// the form of an array of Frames.
type Stack []Frame

// Format formats the stack of Frames following the rules
// explained in Frame.Format with the addition of the '#' flag.
//
// when '#' is passed, like for example %#+v each row
// will be prefixed by [i/n] indicating the position in the stack
// followed by the %+v representation of the Frame
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

// Here returns the Frame corresponding to where it was called,
// or nil if it wasn't possible
func Here() *Frame {
	const depth = 1
	var pcs [depth]uintptr

	if n := runtime.Callers(2, pcs[:]); n > 0 {
		f := frameForPC(pcs[0])
		return &f
	}
	return nil
}

// StackFrame returns the Frame skip levels above from where it
// was called, or nil if it wasn't possible
func StackFrame(skip int) *Frame {
	const depth = MaxDepth
	var pcs [depth]uintptr

	if n := runtime.Callers(2, pcs[:]); n > skip {
		f := frameForPC(pcs[skip])
		return &f
	}

	return nil
}

// StackTrace returns a snapshot of the call stack starting
// skip levels above from where it was called, on an empty
// array if it wasn't possible
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
