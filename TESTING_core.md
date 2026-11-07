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
Trusted Ground (verified with plain checks on the real t):
├── comparableEqual[U]      (== with the panic recovered as unknown)
└── assertComparable[U]     (reports what comparableEqual decides)

Independent Base Functions:
├── AssertTrue              (uses assertComparable)
├── AssertFalse             (uses assertComparable)
├── AssertEqual[T]          (uses AreEqual)
├── AssertNotEqual[T]       (uses AreEqual)
├── AssertSliceEqual[T]     (uses AreEqual)
├── AssertDeepEqual[T]      (uses reflect.DeepEqual)
├── AssertNotDeepEqual[T]   (uses reflect.DeepEqual)
├── AssertContains          (uses strings.Contains)
├── AssertNotContain        (uses strings.Contains)
├── AssertNil               (uses IsNil utility)
├── AssertNotNil            (uses IsNil utility)
├── AssertError             (standalone implementation)
├── AssertNoError           (standalone implementation)
├── AssertErrorIs           (uses errors.Is)
├── AssertNotErrorIs        (uses errors.Is)
├── AssertErrorIsFn         (uses IsErrorFn)
├── AssertErrorAs[E]        (uses errors.As)
├── AssertTypeIs[T]         (uses type assertion)
├── AssertPanic             (uses recover mechanism)
├── AssertNoPanic           (uses recover mechanism)
├── AssertTopFrame          (uses the CallStacker interface)
├── AssertEventually        (uses WaitForCond)
├── AssertEventuallyContext (uses WaitForCondContext)
├── AssertClosed[U]         (uses select with a timeout)
├── AssertOpen[U]           (uses select with a timeout)
├── AssertQuiet[U]          (uses select with a timeout)
└── AssertReceives[U]       (uses select with a shared deadline)
```

`comparableEqual` is `==` and a recover, and `AreEqual` attempts its
comparisons through it as well. Its tests read their results with plain
checks, so everything above it can rely on it as it relies on the `!=` of
a plain check. `assertComparable` is unexported and serves `AssertTrue`
and `AssertFalse`. Tests read through it as `AssertComparableForTest`,
which `testutils_internal_test.go` exports to package `core_test` as well.

### Testing Implications

**✅ Safe Testing Patterns:**

```go
// Read a subject's results through something it does not use
func TestAssertEqual(t *testing.T) {
    mock := &core.MockT{}
    result := core.AssertEqual(mock, 42, 42, "test")
    // AssertTrue does not reach AssertEqual or AreEqual
    core.AssertTrue(t, result, "result")
}

// The same for a utility: AssertComparableForTest does not reach IsNil
func TestIsNil(t *testing.T) {
    var p *int
    core.AssertComparableForTest(t, true, core.IsNil(p), "typed nil")
    core.AssertComparableForTest(t, false, core.IsNil(42), "non-nil")
}
```

```go
// Where the subject is the trusted ground, use plain checks; in package
// core, as assertComparable is unexported
func TestAssertComparable(t *testing.T) {
    mock := &MockT{}
    result := assertComparable(mock, 42, 42, "test")
    plainEqual(t, true, result, "result")
}
```

The tests of the trusted ground make their plain checks through `plainEqual`
and `plainContains`, defined in `testutils_internal_test.go`: they compare
and report without an assertion.

**❌ Circular Testing Anti-Patterns:**

```go
// DON'T: Test AreEqual using AssertEqual (AssertEqual calls AreEqual)
func TestAreEqual(t *testing.T) {
    is, _ := core.AreEqual(42, 42)
    core.AssertEqual(t, true, is, "is") // CIRCULAR!
}

// DON'T: Test AssertTrue using AssertTrue
func TestAssertTrue(t *testing.T) {
    mock := &core.MockT{}
    result := core.AssertTrue(mock, true, "test")
    core.AssertTrue(t, result, "result") // CIRCULAR!
}

// DON'T: Test IsNil using AssertNil (AssertNil calls IsNil internally)
func TestIsNil(t *testing.T) {
    var p *int
    core.AssertNil(t, p, "typed nil") // CIRCULAR!
}
```

### Architecture Rationale

The hierarchy exists for consistency and code reuse:

1. **Trusted Ground**: `comparableEqual` and `assertComparable` are small
   enough to verify with plain checks, which gives the tests of everything
   else a reader that shares only verified code with their subject.
   `AssertTrue` and `AssertFalse` report with the wording of `AssertEqual`
   without reaching `AreEqual`.

2. **Independent Base Functions**: Provide fundamental comparison logic
   without dependencies on other assertion functions. Each base function
   is responsible for its own domain-specific validation and messaging.

This design allows for:

- Consistent error message formatting
- Shared logging behaviour
- Reduced code duplication
- Clear separation of concerns

### Testing Guidelines for Hierarchy

1. **Test the trusted ground** with plain checks on the real `t`
2. **Test everything else** through readers it does not use
3. **Avoid circular testing**: a subject is never checked through itself
   or through anything it uses internally, apart from the trusted ground
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
- **Helper call tracking**: NumHelperCalls() counts how many times
  Helper() was called
- **Failed state management**: Failed() reports test failure state; Error()
  and Errorf() automatically mark as failed
- **Formatted logging**: Errorf() and Logf() provide printf-style formatting
- **Complete state inspection**: HasErrors(), HasLogs(), LastError(),
  LastLog(), NumErrors(), NumLogs(), ErrorAt(i) and LogAt(i) for detailed
  testing. The At accessors count a negative index from the end, so
  ErrorAt(-1) is LastError(), and return ("", false) out of range. Read
  through them rather than the Errors, Logs and HelperCalled fields: the
  accessors take the lock, a bare field read does not
- **State reset**: Reset() clears all collected data and resets counters
- **Fatal/FailNow support**: Full implementation of Fatal(), Fatalf(), and
  FailNow() methods, which stop the test
- **Skip/SkipNow support**: Skip(), Skipf() and SkipNow() stop the test
  the same way and mark it skipped without failing it; Skipped() reports
  it
- **Stopping a test**: Run() method executes test functions and returns on
  FailNow and SkipNow, enabling testing of fatal assertion patterns
  and of code that skips

#### MockT Usage Examples

```go
func TestAssertContains(t *testing.T) {
 mock := &core.MockT{}

 // Test successful assertion
 result := core.AssertContains(mock, "hello world", "world", "text")
 core.AssertTrue(t, result, "returns true")
 core.AssertTrue(t, mock.HasLogs(), "has logs")

 lastLog, ok := mock.LastLog()
 core.AssertTrue(t, ok, "has log")
 core.AssertEqual(t, `text: contains "world"`, lastLog, "log content")

 // Test failed assertion
 mock.Reset()
 result = core.AssertContains(mock, "hello world", "moon", "text")
 core.AssertFalse(t, result, "returns false")
 core.AssertTrue(t, mock.HasErrors(), "has errors")
}
```

#### Testing Fatal/FailNow Scenarios

MockT's Run() method enables testing functions that call Fatal/FailNow methods,
including the `AssertMust*` family of functions which automatically call
`t.FailNow()` on assertion failure:

```go
func TestFatalAssertion(t *testing.T) {
    mock := &core.MockT{}

    // AssertMustEqual reports, then calls FailNow; Run returns there
    ok := mock.Run("fatal assertion test", func(mt core.T) {
        core.AssertMustEqual(mt, 1, 2, "value")
        mt.Log("continued") // not reached
    })

    core.AssertFalse(t, ok, "passed")
    core.AssertTrue(t, mock.Failed(), "failed")
    core.AssertEqual(t, 1, mock.NumErrors(), "errors")
    core.AssertEqual(t, 0, mock.NumLogs(), "logs")

    // The same through AssertEqual and an explicit FailNow
    mock.Reset()
    ok = mock.Run("manual fatal test", func(mt core.T) {
        if !core.AssertEqual(mt, 1, 2, "value") {
            mt.FailNow() // Run returns here
        }
    })
    core.AssertFalse(t, ok, "passed")
}
```

#### Comprehensive MockT Example

```go
// Example MockT usage for testing assertion functions
func TestMyAssertion(t *testing.T) {
    mock := &core.MockT{}

    // Test successful assertion
    MyAssert(mock, true, "pass")
    core.AssertFalse(t, mock.HasErrors(), "errors")
    core.AssertTrue(t, mock.HasLogs(), "logs")

    // Test failed assertion
    mock.Reset()
    MyAssert(mock, false, "fail")
    core.AssertTrue(t, mock.HasErrors(), "errors")
    core.AssertTrue(t, mock.Failed(), "failed")

    lastErr, ok := mock.LastError()
    core.AssertTrue(t, ok, "error recorded")
    core.AssertContains(t, lastErr, "fail", "error message content")
}
```

### Generic Function Testing

Core provides many generic functions that need comprehensive type testing:

```go
func TestSliceContainsGeneric(t *testing.T) {
 // Test with different types
 t.Run("int", func(t *testing.T) {
  slice := core.S(1, 2, 3, 4, 5)
  core.AssertTrue(t, core.SliceContains(slice, 3), "slice has int")
  core.AssertFalse(t, core.SliceContains(slice, 6), "slice missing int")
 })

 t.Run("string", func(t *testing.T) {
  slice := core.S("a", "b", "c")
  core.AssertTrue(t, core.SliceContains(slice, "b"), "slice has string")
  core.AssertFalse(t, core.SliceContains(slice, "d"), "slice missing string")
 })

 t.Run("custom type", func(t *testing.T) {
  type Custom struct{ ID int }
  slice := core.S(Custom{1}, Custom{2})
  core.AssertTrue(t, core.SliceContains(slice, Custom{1}), "slice has custom")
  core.AssertFalse(t, core.SliceContains(slice, Custom{3}),
    "should not find missing custom")
 })
}
```

### Context Key Testing

Core provides type-safe context keys that require special testing:

```go
func TestContextKey(t *testing.T) {
 key := core.NewContextKey[string]("test-key")

 // Test string representation
 core.AssertEqual(t, "test-key", key.String(), "key name")
 core.AssertContains(t, key.GoString(), "core.NewContextKey[string]",
   "GoString format")

 // Test context operations
 ctx := key.WithValue(context.Background(), "test-value")
 value, ok := key.Get(ctx)
 core.AssertTrue(t, ok, "found")
 core.AssertEqual(t, "test-value", value, "retrieved value")

 // Test wrong context
 _, ok = key.Get(context.Background())
 core.AssertFalse(t, ok, "found in empty context")
}
```

### Error Handling Testing

Core provides sophisticated error handling that needs thorough testing:

```go
func TestPanicError(t *testing.T) {
 // Build a PanicError from a payload
 panicErr := core.NewPanicError(0, 42)

 // Test interfaces
 core.AssertEqual(t, "panic: 42", panicErr.Error(), "error message")
 core.AssertEqual(t, 42, panicErr.Recovered(), "recovered value")

 // Test stack trace
 stack := panicErr.CallStack()
 core.AssertTrue(t, len(stack) > 0, "stack")

 // Test unwrapping
 cause := errors.New("root cause")
 wrappedErr := core.NewPanicError(0, cause)
 core.AssertErrorIs(t, wrappedErr, cause, "cause")
}
```

### Synchronisation Primitive Testing

Core provides advanced sync primitives:

```go
func TestSpinLock(t *testing.T) {
 var lock core.SpinLock
 var counter int

 err := core.RunConcurrentTest(t, 10, func(workerID int) error {
  for i := 0; i < 100; i++ {
   lock.Lock()
   counter++
   lock.Unlock()
  }
  return nil
 })

 core.AssertNoError(t, err, "concurrent test")
 core.AssertEqual(t, 1000, counter, "counter")
}
```

### Network Address Testing

Core provides network utilities that need system-dependent testing:

```go
func TestGetIPAddresses(t *testing.T) {
 // Test with system interfaces (may vary by system)
 addrs, err := core.GetIPAddresses()
 if err != nil {
  // Some CI systems may not have network interfaces
  t.Logf("No interfaces available: %v", err)
  return
 }

 // Validate all addresses
 for i, addr := range addrs {
  core.AssertTrue(t, addr.IsValid(), "address[%d] valid", i)
 }
}
```

## Core-Specific Test Utilities

### Custom Assertion Testing

Test custom assertions with MockT:

```go
func TestAssertSliceEqual(t *testing.T) {
 mock := &core.MockT{}

 // Test equal slices
 a := core.S(1, 2, 3)
 b := core.S(1, 2, 3)
 result := core.AssertSliceEqual(mock, a, b, "equal slices")
 core.AssertTrue(t, result, "result")
 core.AssertTrue(t, mock.HasLogs(), "logs")

 // Test different slices
 mock.Reset()
 c := core.S(1, 2, 4)
 result = core.AssertSliceEqual(mock, a, c, "different slices")
 core.AssertFalse(t, result, "result")
 core.AssertTrue(t, mock.HasErrors(), "errors")
}
```

### Benchmark Utility Testing

```go
func TestRunBenchmark(t *testing.T) {
 execCount := 0

 core.RunBenchmark(&testing.B{N: 10},
  func() any {
   return "test data"
  },
  func(data any) {
   core.AssertEqual(t, "test data", data, "benchmark data")
   execCount++
  },
 )

 core.AssertEqual(t, 10, execCount, "execution count")
}
```

## Testing Zero Dependencies

Core's tests bring in no testing or mocking frameworks. They use:

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

// ❌ FORBIDDEN - testing and mocking frameworks
import (
 "github.com/stretchr/testify/assert" // cspell:disable-line
 "github.com/golang/mock/gomock" // cspell:disable-line
)
```

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
  _ = core.SliceContains(slice, 500)
 }
}

func BenchmarkAssertEqual(b *testing.B) {
 mock := &core.MockT{}

 b.ResetTimer()
 for i := 0; i < b.N; i++ {
  core.AssertEqual(mock, 42, 42, "benchmark test")
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
 core.AssertTrue(t, hasTestMethod(tc), "test method")
}

func hasTestMethod(tc any) bool {
 // Use reflection to verify test method exists
 // Implementation details...
 return true
}
```

## Documentation Testing

Core's testing utilities should be well-documented:

```go
func ExampleAssertEqual() {
 mock := &core.MockT{}

 // Test equality
 result := core.AssertEqual(mock, 42, 42, "numbers match")
 fmt.Printf("Equal values: %t\n", result)
 fmt.Printf("Has logs: %t\n", mock.HasLogs())

 // Output:
 // Equal values: true
 // Has logs: true
}

func ExampleS() {
 // Create test slices concisely
 numbers := core.S(1, 2, 3)
 strings := core.S("a", "b", "c")
 empty := core.S[int]()

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
 mock := &core.MockT{}

 // Test nil vs nil
 result := core.AssertEqual[any](mock, nil, nil, "both nil")
 core.AssertTrue(t, result, "result")

 // Test nil vs non-nil
 mock.Reset()
 result = core.AssertEqual[any](mock, nil, 42, "nil vs non-nil")
 core.AssertFalse(t, result, "result")

 // Test different nil types
 mock.Reset()
 var nilSlice []int
 var nilMap map[string]int
 result = core.AssertEqual[any](mock, nilSlice, nilMap, "different nil types")
 core.AssertFalse(t, result, "result")
}
```

## Summary

Testing `darvaza.org/core` requires:

1. **Self-testing** using MockT for assertion functions.
2. **Generic testing** across multiple types.
3. **Zero dependencies** - no testing frameworks beyond core's own utilities.
4. **Performance awareness** - benchmarks for utilities.
5. **Edge case coverage** - nil handling, empty inputs.
6. **Documentation** - examples for public APIs.
7. **Consistency** - following our own guidelines.

The core package sets the standard for all other darvaza.org projects, so its
tests must exemplify best practices.
