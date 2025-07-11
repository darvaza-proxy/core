package core

import (
	"fmt"
	"log"
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
