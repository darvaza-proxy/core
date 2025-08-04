# Core Testing Guidelines

This document provides specific testing guidance for the `darvaza.org/core`
package, which provides the testing utilities used across all darvaza.org
projects.

For general testing patterns applicable to all darvaza.org projects, see
[TESTING.md](./TESTING.md).

## Assertion Function Hierarchy

Understanding the internal hierarchy of assertion functions is crucial for
proper testing, especially when testing the assertion functions themselves.

### Function Dependencies

```text
Independent Base Functions:
├── AssertEqual[T]         (standalone implementation)
├── AssertNotEqual[T]      (standalone implementation)
├── AssertSliceEqual[T]    (uses reflect.DeepEqual)
├── AssertContains         (uses strings.Contains)
├── AssertNil              (uses IsNil utility)
├── AssertNotNil           (uses IsNil utility)
├── AssertErrorIs          (uses errors.Is)
├── AssertTypeIs[T]        (uses type assertion)
├── AssertPanic            (uses recover mechanism)
└── AssertNoPanic          (uses recover mechanism)

Derived Functions (depend on base functions):
├── AssertTrue             → calls AssertEqual(t, true, value, ...)
├── AssertFalse            → calls AssertEqual(t, false, value, ...)
├── AssertError            → calls AssertNotNil(t, err, ...)
└── AssertNoError          → calls AssertNil(t, err, ...)
```

### Testing Implications

**✅ Safe Testing Patterns:**

```go
// Test base functions with MockT
func TestAssertEqual(t *testing.T) {
    mock := &MockT{}
    result := AssertEqual(mock, 42, 42, "test")
    // Use standard testing methods, not AssertTrue (which calls AssertEqual)
    if !result {
        t.Error("AssertEqual should return true")
    }
}

// Test derived functions knowing their dependencies
func TestAssertTrue(t *testing.T) {
    mock := &MockT{}
    result := AssertTrue(mock, true, "test")
    // This is safe because we're testing AssertTrue, not AssertEqual
    if !result {
        t.Error("AssertTrue should return true")
    }
}
```

**❌ Circular Testing Anti-Patterns:**

```go
// DON'T: Test AssertEqual using AssertTrue (AssertTrue calls AssertEqual)
func TestAssertEqual(t *testing.T) {
    mock := &MockT{}
    result := AssertEqual(mock, 42, 42, "test")
    AssertTrue(t, result, "should be true") // CIRCULAR!
}

// DON'T: Test AssertTrue using AssertTrue
func TestAssertTrue(t *testing.T) {
    mock := &MockT{}
    result := AssertTrue(mock, true, "test")
    AssertTrue(t, result, "should be true") // CIRCULAR!
}
```

### Architecture Rationale

The hierarchy exists for consistency and code reuse:

1. **AssertTrue/AssertFalse → AssertEqual**: Ensures consistent formatting
   and logging behaviour for boolean assertions.

2. **AssertError/AssertNoError → AssertNil/AssertNotNil**: Treats errors as
   special cases of nil checking with appropriate naming.

3. **Independent Base Functions**: Provide fundamental comparison logic
   without dependencies on other assertion functions.

This design allows for:

- Consistent error message formatting
- Shared logging behaviour
- Reduced code duplication
- Clear separation of concerns

### Testing Guidelines for Hierarchy

1. **Test base functions first** using standard testing methods
2. **Test derived functions** knowing their dependencies
3. **Avoid circular testing** where functions test themselves
4. **Use MockT appropriately** to capture assertion behaviour
5. **Document the hierarchy** in tests for maintainability

## Core-Specific Testing Patterns

### Testing the Testing Utilities

Since `core` provides testing utilities, we need to test the test code
itself using `MockT`.

#### MockT Features

MockT provides a complete mock implementation of the testing.T interface with
enhanced capabilities:

- **Thread-safe operations**: All methods are protected by sync.RWMutex for
  concurrent use
- **Helper call tracking**: HelperCalled field counts how many times
  Helper() was called
- **Failed state management**: Failed() reports test failure state; Error()
  and Errorf() automatically mark as failed
- **Formatted logging**: Errorf() and Logf() provide printf-style formatting
- **Complete state inspection**: HasErrors(), HasLogs(), LastError(),
  LastLog() for detailed testing
- **State reset**: Reset() clears all collected data and resets counters

#### MockT Usage Examples

```go
func TestAssertEqual(t *testing.T) {
 mock := &MockT{}

 // Test successful assertion
 result := AssertEqual(mock, 42, 42, "equality")
 AssertTrue(t, result, "returns true")
 AssertTrue(t, mock.HasLogs(), "has logs")

 lastLog, ok := mock.LastLog()
 AssertTrue(t, ok, "has log")
 AssertContains(t, lastLog, "equality test: 42", "log content")

 // Test failed assertion
 mock.Reset()
 result = AssertEqual(mock, 42, 24, "inequality")
 AssertFalse(t, result, "returns false")
 AssertTrue(t, mock.HasErrors(), "has errors")
}
```

#### Comprehensive MockT Example

```go
// Example MockT usage for testing assertion functions
func TestMyAssertion(t *testing.T) {
    mock := &MockT{}

    // Test successful assertion
    MyAssert(mock, true, "should pass")
    AssertFalse(t, mock.HasErrors(), "no errors expected")
    AssertTrue(t, mock.HasLogs(), "success should be logged")

    // Test failed assertion
    mock.Reset()
    MyAssert(mock, false, "should fail")
    AssertTrue(t, mock.HasErrors(), "error expected")
    AssertTrue(t, mock.Failed(), "test should be marked as failed")

    lastErr, ok := mock.LastError()
    AssertTrue(t, ok, "should have error message")
    AssertContains(t, lastErr, "should fail", "error message content")
}
```

### Generic Function Testing

Core provides many generic functions that need comprehensive type testing:

```go
func TestSliceContainsGeneric(t *testing.T) {
 // Test with different types
 t.Run("int", func(t *testing.T) {
  slice := S(1, 2, 3, 4, 5)
  AssertTrue(t, SliceContains(slice, 3), "slice has int")
  AssertFalse(t, SliceContains(slice, 6), "slice missing int")
 })

 t.Run("string", func(t *testing.T) {
  slice := S("a", "b", "c")
  AssertTrue(t, SliceContains(slice, "b"), "slice has string")
  AssertFalse(t, SliceContains(slice, "d"), "slice missing string")
 })

 t.Run("custom type", func(t *testing.T) {
  type Custom struct{ ID int }
  slice := S(Custom{1}, Custom{2})
  AssertTrue(t, SliceContains(slice, Custom{1}), "slice has custom")
  AssertFalse(t, SliceContains(slice, Custom{3}),
    "should not find missing custom")
 })
}
```

### Context Key Testing

Core provides type-safe context keys that require special testing:

```go
func TestContextKey(t *testing.T) {
 key := NewContextKey[string]("test-key")

 // Test string representation
 AssertEqual(t, "test-key", key.String(), "key name")
 AssertContains(t, key.GoString(), "core.NewContextKey[string]",
   "GoString format")

 // Test context operations
 ctx := key.WithValue(context.Background(), "test-value")
 value, ok := key.Get(ctx)
 AssertTrue(t, ok, "should retrieve value")
 AssertEqual(t, "test-value", value, "retrieved value")

 // Test wrong context
 _, ok = key.Get(context.Background())
 AssertFalse(t, ok, "should not find in empty context")
}
```

### Error Handling Testing

Core provides sophisticated error handling that needs thorough testing:

```go
func TestPanicError(t *testing.T) {
 // Test panic recovery
 panicErr := NewPanicError(42, "test panic")

 // Test interfaces
 AssertEqual(t, "test panic", panicErr.Error(), "error message")
 AssertEqual(t, 42, panicErr.Recovered(), "recovered value")

 // Test stack trace
 frames := panicErr.Frames()
 AssertTrue(t, len(frames) > 0, "should have stack frames")

 // Test unwrapping
 cause := errors.New("root cause")
 wrappedErr := NewPanicError(cause, "wrapped panic")
 AssertErrorIs(t, wrappedErr, cause, "should unwrap to cause")
}
```

### Synchronization Primitive Testing

Core provides advanced sync primitives:

```go
func TestSpinLock(t *testing.T) {
 var lock SpinLock
 var counter int

 err := RunConcurrentTest(t, 10, func(workerID int) error {
  for i := 0; i < 100; i++ {
   lock.Lock()
   counter++
   lock.Unlock()
  }
  return nil
 })

 AssertNoError(t, err, "concurrent test")
 AssertEqual(t, 1000, counter, "counter should be exactly 1000")
}
```

### Network Address Testing

Core provides network utilities that need system-dependent testing:

```go
func TestGetIPAddresses(t *testing.T) {
 // Test with system interfaces (may vary by system)
 addrs, err := GetIPAddresses()
 if err != nil {
  // Some CI systems may not have network interfaces
  t.Logf("No interfaces available: %v", err)
  return
 }

 AssertTrue(t, len(addrs) >= 0, "should return slice")

 // Validate all addresses
 for i, addr := range addrs {
  AssertTrue(t, addr.IsValid(), "address[%d] should be valid", i)
 }
}
```

## Core-Specific Test Utilities

### Custom Assertion Testing

Test custom assertions with MockT:

```go
func TestAssertSliceEqual(t *testing.T) {
 mock := &MockT{}

 // Test equal slices
 a := S(1, 2, 3)
 b := S(1, 2, 3)
 result := AssertSliceEqual(mock, a, b, "equal slices")
 AssertTrue(t, result, "should return true")
 AssertTrue(t, mock.HasLogs(), "should log success")

 // Test different slices
 mock.Reset()
 c := S(1, 2, 4)
 result = AssertSliceEqual(mock, a, c, "different slices")
 AssertFalse(t, result, "should return false")
 AssertTrue(t, mock.HasErrors(), "should log error")
}
```

### Benchmark Utility Testing

```go
func TestRunBenchmark(t *testing.T) {
 called := false
 err := RunBenchmark(&testing.B{},
  func() interface{} {
   return "test data"
  },
  func(data interface{}) {
   called = true
   AssertEqual(t, "test data", data.(string), "benchmark data")
  },
 )

 AssertNoError(t, err, "benchmark should run")
 AssertTrue(t, called, "benchmark function should be called")
}
```

## Testing Zero Dependencies

Core has zero external dependencies, so all tests must use only:

- Go standard library
- Core's own utilities (for self-testing)

```go
// ✅ ALLOWED
import (
 "context"
 "errors"
 "fmt"
 "testing"
 "time"
)

// ❌ FORBIDDEN - no external dependencies
import (
 "github.com/stretchr/testify/assert" // cspell:disable-line
 "github.com/golang/mock/gomock" // cspell:disable-line
)
```

### Exception: testing_test.go

The `testing_test.go` file is a special case where we test the testing
utilities themselves. In this file, we use the standard `testing` package
directly rather than our own Assert functions to avoid circular dependencies:

```go
// testing_test.go - uses standard testing package
func TestAssertEqual(t *testing.T) {
 mock := &MockT{}

 // Use standard testing methods to test our Assert functions
 result := AssertEqual(mock, 42, 42, "test")
 if !result {
  t.Error("AssertEqual should return true for equal values")
 }

 if !mock.HasLogs() {
  t.Error("AssertEqual should log success")
 }
}
```

This is the **only** exception to using our Assert functions - when testing
the Assert functions themselves.

**Additional Exception**: Avoid using Assert helpers that internally use the
functions being tested (circular dependencies):

```go
// ❌ DON'T: Test IsNil using AssertNil (AssertNil calls IsNil internally)
func TestIsNil(t *testing.T) {
    AssertNil(t, IsNil(nil), "IsNil with nil") // Circular dependency!
}

// ❌ DON'T: Test SliceContains using AssertTrue with SliceContains
func TestSliceContains(t *testing.T) {
    slice := S(1, 2, 3)
    AssertTrue(t, SliceContains(slice, 2), "contains 2")
    // Uses the function being tested!
}

// ✅ DO: Use standard testing methods when testing core utilities
func TestIsNil(t *testing.T) {
    if !IsNil(nil) {
        t.Error("IsNil should return true for nil")
    }
    if IsNil(42) {
        t.Error("IsNil should return false for non-nil")
    }
}

func TestSliceContains(t *testing.T) {
    slice := S(1, 2, 3)
    if !SliceContains(slice, 2) {
        t.Error("SliceContains should find existing element")
    }
    if SliceContains(slice, 4) {
        t.Error("SliceContains should not find missing element")
    }
}
```

**Rule**: When testing a function, don't use Assert helpers that call that same
function internally. This ensures the test is actually validating the function
rather than creating circular logic.

## Performance Testing

Core utilities should be efficient:

```go
func BenchmarkSliceContains(b *testing.B) {
 slice := make([]int, 1000)
 for i := range slice {
  slice[i] = i
 }

 b.ResetTimer()
 for i := 0; i < b.N; i++ {
  _ = SliceContains(slice, 500)
 }
}

func BenchmarkAssertEqual(b *testing.B) {
 mock := &MockT{}

 b.ResetTimer()
 for i := 0; i < b.N; i++ {
  AssertEqual(mock, 42, 42, "benchmark test")
  mock.Reset()
 }
}
```

## Self-Consistency Testing

Core must test its own patterns:

```go
// Test that our test utilities follow our guidelines
func TestTestCaseCompliance(t *testing.T) {
 // Verify field alignment
 tc := parseAddrTestCase{
  input:   "192.168.1.1",
  want:    netip.MustParseAddr("192.168.1.1"),
  name:    "test",
  wantErr: false,
 }

 // Test case should have test method
 AssertTrue(t, hasTestMethod(tc), "test case should have test method")
}

func hasTestMethod(tc interface{}) bool {
 // Use reflection to verify test method exists
 // Implementation details...
 return true
}
```

## Documentation Testing

Core's testing utilities should be well-documented:

```go
func ExampleAssertEqual() {
 mock := &MockT{}

 // Test equality
 result := AssertEqual(mock, 42, 42, "numbers match")
 fmt.Printf("Equal values: %t\n", result)
 fmt.Printf("Has logs: %t\n", mock.HasLogs())

 // Output:
 // Equal values: true
 // Has logs: true
}

func ExampleS() {
 // Create test slices concisely
 numbers := S(1, 2, 3)
 strings := S("a", "b", "c")
 empty := S[int]()

 fmt.Printf("Numbers: %v\n", numbers)
 fmt.Printf("Strings: %v\n", strings)
 fmt.Printf("Empty length: %d\n", len(empty))

 // Output:
 // Numbers: [1 2 3]
 // Strings: [a b c]
 // Empty length: 0
}
```

## Edge Cases and Error Conditions

Core utilities must handle edge cases gracefully:

```go
func TestAssertEqualNilHandling(t *testing.T) {
 mock := &MockT{}

 // Test nil vs nil
 result := AssertEqual(mock, nil, nil, "both nil")
 AssertTrue(t, result, "nil should equal nil")

 // Test nil vs non-nil
 mock.Reset()
 result = AssertEqual(mock, nil, 42, "nil vs non-nil")
 AssertFalse(t, result, "nil should not equal non-nil")

 // Test different nil types
 mock.Reset()
 var nilSlice []int
 var nilMap map[string]int
 result = AssertEqual(mock, nilSlice, nilMap, "different nil types")
 AssertFalse(t, result, "different nil types should not be equal")
}
```

## Summary

Testing `darvaza.org/core` requires:

1. **Self-testing** using MockT for assertion functions.
2. **Generic testing** across multiple types.
3. **Zero dependencies** - only standard library.
4. **Performance awareness** - benchmarks for utilities.
5. **Edge case coverage** - nil handling, empty inputs.
6. **Documentation** - examples for public APIs.
7. **Consistency** - following our own guidelines.

The core package sets the standard for all other darvaza.org projects, so its
tests must exemplify best practices.
