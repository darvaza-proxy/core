# Testing Guidelines for Darvaza.org Projects

This document describes the testing patterns and standards used across all
darvaza.org projects. These guidelines ensure consistent, maintainable, and
compliant test code that meets our strict linting requirements.

> **Note**: For core-specific testing patterns, see
> [TESTING_core.md](./TESTING_core.md)

## Core Testing Principles

1. **Consistency**: All projects follow the same testing patterns.
2. **Simplicity**: Tests should be easy to read and understand.
3. **Compliance**: Tests must meet linting requirements (cognitive
   complexity ≤7, cyclomatic complexity ≤10).
4. **Reliability**: Tests should be deterministic and fast.
5. **Maintainability**: Tests should be easy to modify and extend.

## Testing Utilities

All darvaza.org projects should use the testing utilities provided by
`darvaza.org/core`. Import them as:

```go
import "darvaza.org/core"
```

### Test Helper Functions

#### Slice Creation

Use `S[T]()` for concise test slice creation:

```go
// Instead of []int{1, 2, 3}
testData := core.S(1, 2, 3)

// Instead of []string{}
emptyStrings := core.S[string]()
```

#### Assertion Functions

All assertions work with both `*testing.T` and `core.MockT`, log successful
assertions, and return boolean results. The final parameter is a **prefix**
for log messages, not a complete message:

**Basic Assertions:**

```go
core.AssertEqual(t, expected, actual, "value")
core.AssertSliceEqual(t, expectedSlice, actualSlice, "slice")
core.AssertTrue(t, condition, "condition")
core.AssertFalse(t, condition, "negation")
core.AssertNil(t, value, "nil check")
core.AssertNotNil(t, value, "non-nil check")
```

**Error Assertions:**

```go
core.AssertError(t, err, "error")
core.AssertNoError(t, err, "success")
core.AssertErrorIs(t, err, target, "error chain")
```

**Advanced Assertions:**

```go
result, ok := core.AssertTypeIs[MyType](t, value, "type cast")
core.AssertContains(t, text, substring, "text content")
core.AssertPanic(t, func() { panic("test") }, "panic")
core.AssertNoPanic(t, func() { /* safe code */ }, "no panic")
```

### Assertion Function Hierarchy Guidelines

When creating custom assertion functions, follow these hierarchy principles.
For detailed information about the core assertion hierarchy, see
[TESTING_core.md](./TESTING_core.md#assertion-function-hierarchy).

**Base Functions**: Independent implementations that don't call other
assertions.
**Derived Functions**: Call base functions for consistency and code reuse.

#### Designing New Assertion Functions

When implementing custom assertions in your projects, follow this pattern:

```go
// Base assertion - independent implementation
func AssertCustomCondition(t core.T, value SomeType, name string,
    args ...any) bool {
    t.Helper()
    ok := customLogic(value)
    if !ok {
        doError(t, name, args, "expected custom condition, got %v",
            value)
    } else {
        doLog(t, name, args, "custom: %v", value)
    }
    return ok
}

// Derived assertion - calls base function for consistency
func AssertCustomTrue(t core.T, condition bool, name string,
    args ...any) bool {
    t.Helper()
    return AssertCustomCondition(t, condition, name, args...)
}
```

**Key Principles:**

- **Avoid circular dependencies**: Don't test assertions using themselves.
- **Maintain consistency**: Derived functions should use base functions for
  uniform error formatting and logging behaviour.
- **Use helper pattern**: Always call `t.Helper()` in assertion functions.
- **Follow naming**: Use `Assert` prefix and descriptive suffixes.

**Understanding Assertion Prefixes:**

The assertion functions use the prefix parameter to create meaningful log
messages:

```go
// ❌ WRONG - Complete sentences as prefixes
core.AssertEqual(t, "expected", actual, "should return expected value")
core.AssertTrue(t, found, "should find the key in the map")

// ✅ CORRECT - Short prefixes that work with formatting
core.AssertEqual(t, "expected", actual, "value")
// logs: "value: expected"
core.AssertTrue(t, found, "key %q found", k)
// logs: "key \"myKey\" found: true"
core.AssertTrue(t, SliceContains(got, k), "key %q present", k)
// logs: "key \"myKey\" present: true"
```

**Prefix Guidelines:**

- Use **short, descriptive prefixes** (1-3 words).
- The prefix will be combined with the actual value: `"prefix: value"`.
- For formatted messages, use printf-style formatting: `"contains %v", key`.
- Avoid complete sentences - they become redundant with the logged value.

## Test Structure Patterns

### Named Test Types with Test Methods

**Always** define named types for test cases with a `test` method:

```go
type parseURLTestCase struct {
 // Large fields first (interfaces, strings) - 8+ bytes
 input    string
 expected *url.URL

 // Small fields last (booleans) - 1 byte
 wantErr  bool
}

func (tc parseURLTestCase) test(t *testing.T) {
 t.Helper()

 result, err := ParseURL(tc.input)

 if tc.wantErr {
  core.AssertError(t, err, "parse error")
  return
 }

 core.AssertNoError(t, err, "parse")
 core.AssertEqual(t, tc.expected.String(), result.String(), "URL")
}
```

### Table-Driven Tests

Use table-driven patterns with the `test` method:

```go
func TestParseURL(t *testing.T) {
 testCases := []parseURLTestCase{
  {
   input:    "https://example.com",
   expected: &url.URL{Scheme: "https", Host: "example.com"},
   wantErr:  false,
  },
  {
   input:   "invalid-url",
   wantErr: true,
  },
 }

 for _, tc := range testCases {
  t.Run(tc.input, tc.test)
 }
}
```

### Test Case Constructor Functions

For complex test cases, provide constructor functions:

```go
func newParseURLTest(name, input string, expected *url.URL,
 wantErr bool) parseURLTestCase {
 return parseURLTestCase{
  input:    input,
  expected: expected,
  wantErr:  wantErr,
 }
}
```

## Anonymous Functions in t.Run

**Rule: Anonymous functions in `t.Run` are allowed ONLY if they are 3 lines
or shorter.**

```go
// ✅ ALLOWED - Short anonymous function (≤3 lines)
t.Run("nil input", func(t *testing.T) {
 result := ProcessInput(nil)
 core.AssertNil(t, result, "result")
})

// ❌ NEVER DO THIS - Long anonymous function (>3 lines)
t.Run("complex test", func(t *testing.T) {
 setup := createTestData()
 result := ProcessComplex(setup)
 validateResult(t, result)
 cleanUpTestData(setup)
})

// ✅ CORRECT - Extract to named function
func testComplexScenario(t *testing.T) {
 t.Helper()
 setup := createTestData()
 result := ProcessComplex(setup)
 validateResult(t, result)
 cleanUpTestData(setup)
}

t.Run("complex test", testComplexScenario)
```

## Managing Complexity

### Extract Helper Methods

When test methods become complex, extract helper methods:

```go
func (tc myTestCase) test(t *testing.T) {
 t.Helper()

 tc.setupTest(t)
 result := tc.runTest(t)
 tc.validateResult(t, result)
}

func (tc myTestCase) setupTest(t *testing.T) {
 t.Helper()
 // Setup logic
}

func (tc myTestCase) runTest(t *testing.T) ResultType {
 t.Helper()
 // Test execution
 return result
}

func (tc myTestCase) validateResult(t *testing.T, result ResultType) {
 t.Helper()
 // Validation logic
}
```

### Extract Test Case Generation

For complex test case creation:

```go
func httpClientTestCases() []httpClientTestCase {
 return []httpClientTestCase{
  newHTTPTest("GET request", "GET", "/api/users", 200, nil),
  newHTTPTest("POST request", "POST", "/api/users", 201, userPayload),
  newHTTPTest("invalid endpoint", "GET", "/invalid", 404, nil),
 }
}

func TestHTTPClient(t *testing.T) {
 for _, tc := range httpClientTestCases() {
  t.Run(tc.name, tc.test)
 }
}
```

### Split Complex Tests

If a single test function exceeds complexity limits, split it:

```go
// Instead of TestEverything
func TestUserCreation(t *testing.T) { /* ... */ }
func TestUserUpdate(t *testing.T) { /* ... */ }
func TestUserDeletion(t *testing.T) { /* ... */ }
func TestUserValidation(t *testing.T) { /* ... */ }

// Group related tests
func TestUserCRUD(t *testing.T) {
 t.Run("creation", TestUserCreation)
 t.Run("update", TestUserUpdate)
 t.Run("deletion", TestUserDeletion)
}
```

## Field Alignment in Test Structs

Order struct fields to minimize memory padding:

```go
type testCase struct {
 // 8-byte fields first (pointers, interfaces, strings on 64-bit)
 input    interface{}
 expected interface{}
 name     string

 // 4-byte fields (int32, float32)
 timeout  int32

 // 1-byte fields last (bool, int8)
 wantErr  bool
 wantOK   bool
}
```

Use the field alignment tool to verify:

```bash
go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/\
 fieldalignment@latest -fix ./...
```

## Concurrent Testing

Use the built-in concurrent testing utility:

```go
func TestConcurrentAccess(t *testing.T) {
 err := core.RunConcurrentTest(t, 10, func(workerID int) error {
  // Test concurrent access with worker ID
  return validateConcurrentOperation(workerID)
 })
 core.AssertNoError(t, err, "concurrent")
}
```

## Benchmark Testing

Use the benchmark utility for consistent setup:

```go
func BenchmarkProcessing(b *testing.B) {
 err := core.RunBenchmark(b,
  func() interface{} {
   // Setup phase - not timed
   return createLargeDataset()
  },
  func(data interface{}) {
   // Execution phase - timed
   ProcessData(data.(*Dataset))
  },
 )
 if err != nil {
  b.Fatal(err)
 }
}
```

## MockT for Testing Test Code

When testing assertion functions themselves. For detailed core-specific
patterns, see
[TESTING_core.md](./TESTING_core.md#testing-the-testing-utilities):

```go
func TestMyAssertion(t *testing.T) {
 mock := &core.MockT{}

 // Test successful assertion
 core.AssertEqual(mock, 42, 42, "equality")
 core.AssertTrue(t, mock.HasLogs(), "has logs")

 lastLog, ok := mock.LastLog()
 core.AssertTrue(t, ok, "has log")
 core.AssertContains(t, lastLog, "test equality: 42", "log content")

 // Reset for next test
 mock.Reset()
}
```

## Error Testing Patterns

### Expected Errors

```go
func TestValidation(t *testing.T) {
 err := ValidateInput("")
 core.AssertError(t, err, "validation error")
 core.AssertErrorIs(t, err, ErrInvalidInput, "error type")
}
```

### Error Types

```go
func TestErrorTypes(t *testing.T) {
 err := ProcessRequest(invalidData)

 validationErr, ok := core.AssertTypeIs[*ValidationError](t, err, "cast")
 if ok {
  core.AssertEqual(t, "invalid field: name", validationErr.Message,
    "message")
 }
}
```

## Integration with CI/CD

### Coverage Requirements

Tests should maintain high coverage. Use with CI:

```bash
# Generate coverage
make test GOTEST_FLAGS="-coverprofile=coverage.out"

# Check coverage threshold
go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | \
  sed 's/%//' | awk '{if($1<80) exit 1}'
```

### Test Tags

Use build tags for different test categories:

```go
//go:build integration
// +build integration

func TestDatabaseIntegration(t *testing.T) {
 // Integration test code
}
```

Run with: `go test -tags=integration`

## Best Practices

### Test Naming

- Test functions: `TestFunctionName`.
- Test types: `functionNameTestCase`.
- Test methods: `func (tc testCaseType) test(t *testing.T)`.

### Documentation

```go
// parseURLTestCase tests URL parsing functionality
type parseURLTestCase struct {
 // ... fields
}

// test validates URL parsing behaviour
func (tc parseURLTestCase) test(t *testing.T) {
 // ... implementation
}
```

### Clean Tests

- Use `t.Helper()` in all helper functions.
- Prefer table-driven tests over individual test functions.
- Keep setup and clean-up minimal.
- Use meaningful assertion descriptions.
- Test both success and failure paths.

### Test Data

```go
// Use meaningful test data
var validUserData = User{
 Name:  "John Doe",
 Email: "john@example.com",
 Age:   30,
}

// Use edge cases
var testUsers = []User{
 {}, // empty user
 {Name: "A"}, // minimal data
 validUserData, // normal case
 {Name: strings.Repeat("A", 1000)}, // large data
}
```

## Forbidden Patterns

### ❌ Never Use These Patterns

```go
// DON'T: Anonymous functions >3 lines
t.Run("test", func(t *testing.T) {
 setup()
 result := execute()
 validate(result)
 cleanUp()
})

// DON'T: Suppress complexity with comments
//revive:disable-next-line:cognitive-complexity
func complexTest(t *testing.T) { /* ... */ }

// DON'T: Direct testing.T methods when assertions exist
if result != expected {
 t.Errorf("got %v, want %v", result, expected)
}

// DON'T: Anonymous test case structs
tests := []struct {
 name string
 // ...
}{ /* ... */ }
```

### ✅ Always Use These Patterns

```go
// DO: Named test types with test methods
type myTestCase struct {
 name     string
 expected result
}

func (tc myTestCase) test(t *testing.T) {
 t.Helper()
 core.AssertEqual(t, tc.expected, actual, tc.name)
}

// DO: Extract complex logic to helper methods
func (tc myTestCase) test(t *testing.T) {
 t.Helper()
 tc.runTest(t)
}

// DO: Use core assertion functions
core.AssertNoError(t, err, "operation")
```

## Summary

By following these guidelines, all darvaza.org projects will have:

- **Consistent testing patterns** across the entire ecosystem.
- **Lint-compliant code** that meets complexity requirements.
- **Maintainable tests** that are easy to read and modify.
- **Reliable test suites** with excellent error reporting.
- **Comprehensive coverage** with meaningful assertions.

The key is to treat test code with the same care as production code, using
the excellent utilities provided by `darvaza.org/core` to maintain consistency
and quality across all projects.
