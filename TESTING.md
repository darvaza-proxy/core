# Testing Guidelines for Darvaza.org Projects

This document describes the testing patterns and standards used across all
darvaza.org projects. These guidelines ensure consistent, maintainable, and
compliant test code that meets our strict linting requirements.

> **Note**: For core-specific testing patterns, see
> [TESTING_core.md](./TESTING_core.md)

## Core Testing Principles

1. **Clarity First**: The primary goal is clear, understandable tests.
   Patterns serve clarity, not compliance.
2. **Simplicity**: Tests should be easy to read and understand. Start
   simple and only add complexity when justified.
3. **Consistency**: All projects follow the same testing patterns when
   patterns are appropriate.
4. **Compliance**: Tests must meet linting requirements (cognitive
   complexity ≤7, cyclomatic complexity ≤10).
5. **Reliability**: Tests should be deterministic and fast.
6. **Maintainability**: Tests should be easy to modify and extend.

**⚠️ WARNING**: If you find yourself contorting your tests to fit patterns,
STOP. Write a simple test instead. Clear, straightforward tests are always
better than forced abstractions.

## Testing Utilities

All darvaza.org projects should use the testing utilities provided by
`darvaza.org/core`. Import them as:

```go
import "darvaza.org/core"
```

### Test Helper Functions

#### TestCase Interface and RunTestCases

The core testing utilities provide a `TestCase` interface and generic
`RunTestCases` function for standardised table-driven tests:

```go
// TestCase interface - implement this for your test case types
type TestCase interface {
    Name() string
    Test(t *testing.T)
}

// Generic function - works with any slice of TestCase implementations
func RunTestCases[T TestCase](t *testing.T, cases []T)
```

This eliminates boilerplate and ensures consistent test execution across all
table-driven test case types.

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
core.AssertNotEqual(t, expected, actual, "value")
core.AssertSliceEqual(t, expectedSlice, actualSlice, "slice")
core.AssertSame(t, expected, actual, "same reference/value")
core.AssertNotSame(t, expected, actual, "different reference/value")
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
core.AssertNotContain(t, text, substring, "text exclusion")
core.AssertPanic(t, func() { panic("test") }, "panic")
core.AssertNoPanic(t, func() { /* safe code */ }, "no panic")
```

**Fatal Assertions (`AssertMust*`):**

All `AssertMustFoo()` functions call the corresponding `AssertFoo()` function
and automatically call `t.FailNow()` if the assertion fails, terminating test
execution immediately:

```go
// Standard assertions - log failure, test continues
core.AssertEqual(t, expected, actual, "value")
// ... test execution continues even if assertion fails

// Fatal assertions - log failure, test terminates
core.AssertMustEqual(t, expected, actual, "critical value")
// ... test execution stops here if assertion fails
```

**Key Difference:**

- `AssertFoo()` - like `t.Error()`, allows test to continue after failure.
- `AssertMustFoo()` - like `t.Fatal()`, terminates test on failure.

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

## Testing Patterns Overview

**Two complementary testing patterns are used across all darvaza.org projects:**

1. **Table-driven tests with TestCase interface**: For testing the same
   function with multiple input/output combinations.
2. **Standard t.Run() with named functions**: For different test scenarios
   and behaviours.

## Testing Pattern Decision Framework

**CRITICAL PRINCIPLE**: Always start with the simplest approach that achieves
your testing goals. Patterns serve clarity, not compliance.

### Decision Checklist

**Use this checklist to choose the right testing pattern:**

```text
❓ Is this a single, straightforward test?
   YES → Use simple test function
   NO  → Continue checking

❓ Do I have different test scenarios or behaviours?
   YES → Use t.Run() with named functions
   NO  → Continue checking

❓ Do I have 2+ scenarios with identical test logic?
   NO  → Use simple test function
   YES → Continue checking

❓ Do scenarios differ only in input/output data?
   NO  → Use t.Run() with named functions
   YES → Consider TestCase interface

❓ Does TestCase make the test CLEARER than simple assertions?
   NO  → Use simple test function
   YES → Use TestCase interface
```

### Testing Pattern Priority

**Always start with the simplest approach:**

1. **Simple test functions** - Default choice for most tests.
2. **t.Run() with named functions** - When you have different test behaviours.
3. **TestCase interface** - ONLY when you have 2+ scenarios with
   identical Test logic.
4. **Custom TestCase implementations** - ONLY when complexity is justified.

**Golden Rule**: Start with simple tests. Only move to more complex
patterns when the simple approach becomes genuinely insufficient.

## When to Use TestCase Interface

**ONLY use TestCase for table-driven tests with multiple data cases
testing the same function logic.**

**Use TestCase when:**

- Testing a single function with multiple input/output combinations.
- All test cases share identical test logic.
- You have 2 or more data variations for the same test behaviour.

**Examples of appropriate TestCase usage:**

- URL parsing with different input formats.
- Mathematical functions with various numeric inputs.
- Validation functions with different valid/invalid inputs.
- String processing with multiple text variations.

## When NOT to Use TestCase Interface

**TestCase is inappropriate for these scenarios:**

### 1. Single or Minimal Test Cases

**Problem**: TestCase overhead for 1-2 scenarios adds complexity without
benefit.

```go
// ❌ WRONG - TestCase overkill for single test
func TestSingleCase(t *testing.T) {
    testCases := []myTestCase{
        newMyTestCase("single test", input, expected),
    }
    RunTestCases(t, testCases)
}

// ✅ RIGHT - Simple and clear
func TestSingleCase(t *testing.T) {
    result := ProcessInput(input)
    AssertEqual(t, expected, result, "result")
}
```

### 2. Different Functions or Behaviours

**Problem**: TestCase is for testing the SAME function with different
data, not different functions.

```go
// ❌ WRONG - Different behaviours forced into TestCase
func TestDifferentBehaviours(t *testing.T) {
    testCases := []TestCase{
        newGetNameTestCase("get name", user, "Alice"),
        newIsAdminTestCase("is admin", user, true),
        newLastLoginTestCase("last login", user, timestamp),
    }
}

// ✅ RIGHT - Different behaviours in separate tests or t.Run()
func TestUserProperties(t *testing.T) {
    user := &User{Name: "Alice", Role: "admin", LastLogin: timestamp}
    AssertEqual(t, "Alice", user.GetName(), "name")
    AssertTrue(t, user.IsAdmin(), "admin status")
    AssertEqual(t, timestamp, user.LastLogin(), "last login")
}
```

### 3. Unique Validation Logic Per Test

**Problem**: Each test needs different validation approaches.

```go
// ❌ WRONG - Forcing TestCase when validation differs
func TestDifferentValidations(t *testing.T) {
    testCases := []TestCase{
        newParseTestCase("parse JSON", jsonInput, complexJSONResult),
        newParseTestCase("parse XML", xmlInput, xmlResult),
    }
}

// ✅ RIGHT - Each test handles its unique validation
func TestDifferentValidations(t *testing.T) {
    t.Run("parse JSON", func(t *testing.T) {
        result := parser.Parse(jsonInput)
        AssertEqual(t, "value", result.Data["key"], "JSON key")
    })
    t.Run("parse XML", func(t *testing.T) {
        result := parser.Parse(xmlInput)
        AssertEqual(t, "namespace", result.Namespace, "XML namespace")
    })
}
```

### 4. Integration or Workflow Tests

**Problem**: TestCase patterns do not fit complex workflows.

```go
// ✅ RIGHT - Simple test models workflow naturally
func TestUserWorkflow(t *testing.T) {
    user, err := CreateUser("test@example.com")
    AssertNoError(t, err, "create user")
    defer DeleteUser(user.ID)

    token, err := Authenticate(user.Email, "password")
    AssertNoError(t, err, "authenticate")

    result, err := ProcessData(token, testData)
    AssertNoError(t, err, "process data")
    AssertEqual(t, "processed", result.Status, "process status")
}
```

### 5. Performance or Benchmark Tests

**Problem**: TestCase adds overhead to performance measurements.

```go
// ❌ WRONG - TestCase overhead affects benchmarks
func BenchmarkWithTestCase(b *testing.B) {
    testCases := []TestCase{...}
    // TestCase overhead skews benchmark results
}

// ✅ RIGHT - Direct function calls for accurate benchmarks
func BenchmarkCalculate(b *testing.B) {
    for range b.N {
        result := Calculate(inputData)
        _ = result // Prevent compiler optimisation
    }
}
```

## When to Use t.Run() with Named Functions

**Use `TestFoo() { t.Run("scenario", runTestFooScenario) }` for different
test scenarios and behaviours.**

**This is the standard and recommended pattern for:**

- Testing different functions or methods.
- Different test scenarios requiring different logic.
- Setup/tear-down that varies between tests.
- Integration tests with different configurations.
- Error handling tests with different error conditions.

**Test Function Consolidation Guidelines:**

- **Same function/type being tested** → Consolidate into `TestFoo` with
  subtests.
- **Different functions being tested** → Keep separate Test functions.
- **TestFooBar + TestFooBaz** → Consolidate to `TestFoo` with
  `t.Run("bar")` and `t.Run("baz")`.

**Example consolidation:**

```go
// ❌ AVOID - Multiple functions testing same type
func TestUserCreate(t *testing.T) { /* ... */ }
func TestUserUpdate(t *testing.T) { /* ... */ }
func TestUserDelete(t *testing.T) { /* ... */ }

// ✅ PREFERRED - Consolidated with subtests using named functions
func TestUser(t *testing.T) {
    t.Run("create", runTestUserCreate)
    t.Run("update", runTestUserUpdate)
    t.Run("delete", runTestUserDelete)
}

func runTestUserCreate(t *testing.T) {
    // Test logic here (>3 lines)
}
```

## MANDATORY Test Requirements

### General Requirements (ALL tests)

**These requirements apply to ALL test code, regardless of pattern:**

1. **Anonymous Functions**: No `t.Run("name", func(t *testing.T) { ... })`
   patterns longer than 3 lines. Use named `runTestFooBar` functions instead.

2. **Test Object Factories**: For complex test object creation, use
   `newTestObject()` factory functions to keep t.Run calls readable.

```go
// ❌ CLUTTERED - Complex setup in t.Run
func TestUser(t *testing.T) {
    t.Run("create admin", func(t *testing.T) {
        user := &User{
            Name: "Alice",
            Email: "alice@example.com",
            Settings: &UserSettings{Theme: "dark", Lang: "en"},
            Permissions: []Permission{AdminRead, AdminWrite},
        }
        // test logic...
    })
}

// ✅ READABLE - Factory function for complex setup
func TestUser(t *testing.T) {
    t.Run("create admin", func(t *testing.T) {
        user := newTestUser("Alice", "alice@example.com", "admin")
        // test logic...
    })
}

func newTestUser(name, email, role string) *User {
    return &User{
        Name: name,
        Email: email,
        Role: role,
        Settings: &UserSettings{Theme: "dark", Lang: "en"},
        Permissions: getPermissionsForRole(role),
    }
}
```

### TestCase-Specific Requirements (table-driven tests only)

**When using TestCase interface for table-driven tests, ALL files must meet
these 5 additional requirements:**

1. **TestCase Interface Validations**: `var _ TestCase = ...` declarations
   for all test case types.
2. **Factory Functions**: All TestCase types have `new{TestContext}TestCase()`
   functions (enables field alignment + logical parameters).
3. **Factory Usage**: All test case declarations use factory functions
   (no naked struct literals).
4. **RunTestCases Usage**: Test functions use `RunTestCases(t, cases)`
   instead of manual loops.
5. **Test Case List Factories**: Complex test case lists use
   `makeTestFunctionTestCases()` factory functions.

These requirements apply **ONLY** to table-driven tests using TestCase, not
to standard t.Run() patterns.

## Table-driven Test Structure Patterns (TestCase)

### Step 1: TestCase Interface Validation (MANDATORY for table-driven tests)

**ALWAYS** add interface validation declarations at the top of your test file
when using TestCase:

```go
// Compile-time verification that test case types implement TestCase interface
var _ TestCase = parseURLTestCase{}
```

### Step 2: Named TestCase Types (MANDATORY for table-driven tests)

**ALWAYS** define named types for table-driven tests that implement the
`TestCase` interface:

```go
type parseURLTestCase struct {
    // Large fields first (interfaces, strings) - 8+ bytes
    input    string
    expected *url.URL
    name     string

    // Small fields last (booleans) - 1 byte
    wantErr  bool
}

func (tc parseURLTestCase) Name() string {
    return tc.name
}

func (tc parseURLTestCase) Test(t *testing.T) {
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

### Step 3: Factory Functions (MANDATORY)

**ALWAYS** provide factory functions for ALL TestCase types.

**Critical Reason**: Factory functions decouple logical parameter order from
memory-optimised field alignment. This allows structs to be field-aligned for
performance while maintaining readable, logical function signatures.

#### Factory Naming Conventions (MANDATORY)

**Individual TestCase Factories**: Named after object type/variant, NOT Test
function

```go
func newParseURLTestCase(name, input string, expected *url.URL,
    wantErr bool) parseURLTestCase {
    return parseURLTestCase{
        name:     name,
        input:    input,
        expected: expected,
        wantErr:  wantErr,
    }
}

func newValidationTestCase(name, input string,
    wantErr bool) validationTestCase {
    return validationTestCase{
        name:    name,
        input:   input,
        wantErr: wantErr,
    }
}
```

**Context Definition Varies (Object Type/Variant):**

- **Function being tested**: `newParseURLTestCase` (creates parseURL
  test case).
- **Behaviour being tested**: `newValidationTestCase` (creates validation
  test case).
- **Component being tested**: `newHTTPClientTestCase` (creates HTTP client
  test case).
- **Type + Method**: `newUserGetNameTestCase` (creates User.GetName test case).

**Key Distinction**: `newFoo` factories are named after the TEST CASE TYPE,
not the Test function name.

### Step 4: Factory Usage (MANDATORY)

**ALWAYS** use factory functions for TestCase declarations. **NEVER** use
naked struct literals:

```go
func TestParseURL(t *testing.T) {
    testCases := []parseURLTestCase{
        newParseURLTestCase("valid URL", "https://example.com",
            &url.URL{Scheme: "https", Host: "example.com"}, false),
        newParseURLTestCase("invalid URL", "invalid-url", nil, true),
    }

    // MANDATORY: Use RunTestCases helper
    core.RunTestCases(t, testCases)
}
```

### Step 5: RunTestCases Usage (MANDATORY)

**ALWAYS** use the `RunTestCases` helper. **NEVER** use manual loops:

```go
// ✅ CORRECT - Single test group, direct call acceptable
func TestParseURL(t *testing.T) {
    core.RunTestCases(t, makeParseURLTestCases())
}

// ✅ PREFERRED - Multiple scenarios, use subtests
func TestParseURL(t *testing.T) {
    t.Run("valid", func(t *testing.T) {
        core.RunTestCases(t, makeParseURLValidTestCases())
    })
    t.Run("invalid", func(t *testing.T) {
        core.RunTestCases(t, makeParseURLInvalidTestCases())
    })
}

// ❌ FORBIDDEN - Manual loops
func TestParseURL(t *testing.T) {
    for _, tc := range testCases {
        t.Run(tc.Name(), tc.Test)
    }
}
```

### Step 6: Test Case List Factories (MANDATORY)

**Use factory functions for complex TestCase generation:**

#### Test Set Factory Naming Conventions

**Test Set Factories**: Named after `make{TestFunction}{Subtest}TestCases`,
return `[]core.TestCase` for composability

```go
// Pattern: make{TestFunction}{Subtest}TestCases() []core.TestCase
func makeHTTPClientBasicTestCases() []core.TestCase
func makeParseURLValidationTestCases() []core.TestCase
func makeValidationFieldTestCases() []core.TestCase
```

**Parameterised Test Set Factories**: Include parameter context

```go
func makeValidationTestCases(fieldName string) []core.TestCase
func makeUserTestCases(userType string) []core.TestCase
```

**Optional Composability:**

```go
// Test sets can be combined when useful
func TestCompleteAPI(t *testing.T) {
    var allTests []core.TestCase
    allTests = append(allTests, makeHTTPClientTestCases()...)
    allTests = append(allTests, makeValidationTestCases()...)
    allTests = append(allTests, makeAuthTestCases()...)
    core.RunTestCases(t, allTests)
}
```

```go
// ✅ CORRECT - Use factory function for complex lists
func makeHTTPClientTestCases() []core.TestCase {
    return []httpClientTestCase{
        newHTTPClientTestCase("GET request", "GET", "/api/users", 200, nil),
        newHTTPClientTestCase("POST request", "POST", "/api/users", 201,
            userPayload),
        newHTTPClientTestCase("PUT request", "PUT", "/api/users/1", 200,
            updatePayload),
        newHTTPClientTestCase("DELETE request", "DELETE", "/api/users/1", 204,
            nil),
        newHTTPClientTestCase("invalid endpoint", "GET", "/invalid", 404, nil),
        newHTTPClientTestCase("unauthorized", "GET", "/api/admin", 401, nil),
    }
}

func TestHTTPClient(t *testing.T) {
    core.RunTestCases(t, makeHTTPClientTestCases())
}

// ❌ FORBIDDEN - Variable declaration for complex lists
var makeHTTPClientTestCases = []httpClientTestCase{
    newHTTPClientTestCase("GET request", "GET", "/api/users", 200, nil),
    newHTTPClientTestCase("POST request", "POST", "/api/users", 201,
        userPayload),
    // ... many more cases (this becomes unwieldy)
}
```

## Anonymous Functions in t.Run

**Rule: Anonymous functions in `t.Run` are allowed ONLY if they are 3 lines
or shorter. For longer logic, use named `runTestFooBar` functions.**

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
func runTestComplexScenario(t *testing.T) {
    t.Helper()
    setup := createTestData()
    result := ProcessComplex(setup)
    validateResult(t, result)
    cleanUpTestData(setup)
}

func TestComplexFeature(t *testing.T) {
    t.Run("complex test", runTestComplexScenario)
}
```

## Managing Complexity

### Extract Helper Methods

When test methods become complex, extract helper methods:

```go
func (tc myTestCase) Test(t *testing.T) {
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

### Test Case List Factories (MANDATORY)

**Rule: Use `myFooTestCases()` factory functions for complex TestCase
generation.**

#### When to Build Test Arrays Inline

For straightforward table-driven test cases with single logic flow (regardless
of length):

```go
func TestParseURL(t *testing.T) {
    testCases := []parseURLTestCase{
        newParseURLTestCase("valid URL", "https://example.com", expectedURL,
            false),
        newParseURLTestCase("invalid URL", "invalid-url", nil, true),
        newParseURLTestCase("missing scheme", "example.com", nil, true),
        newParseURLTestCase("port included", "https://example.com:8080",
            expectedURL2, false),
        // Simple cases, even if many - keep inline
    }

    core.RunTestCases(t, testCases)
}
```

#### When to Use Test Case List Factory Functions

**MANDATORY for complex TestCase generation (computed values, conditional
logic, parameterisation, or reuse):**

```go
func makeHTTPClientTestCases() []httpClientTestCase {
    return []httpClientTestCase{
        newHTTPClientTestCase("GET request", "GET", "/api/users", 200, nil),
        newHTTPClientTestCase("POST request", "POST", "/api/users", 201,
            userPayload),
        newHTTPClientTestCase("PUT request", "PUT", "/api/users/1", 200,
            updatePayload),
        newHTTPClientTestCase("DELETE request", "DELETE", "/api/users/1", 204,
            nil),
        newHTTPClientTestCase("invalid endpoint", "GET", "/invalid", 404, nil),
        newHTTPClientTestCase("malformed JSON", "POST", "/api/users", 400,
            malformedJSON),
        newHTTPClientTestCase("unauthorized", "GET", "/api/admin", 401, nil),
        newHTTPClientTestCase("forbidden", "POST", "/api/admin", 403, nil),
    }
}

func TestHTTPClient(t *testing.T) {
    core.RunTestCases(t, makeHTTPClientTestCases())
}
```

#### Parameterised TestCase Factory Functions

Use factory functions with parameters when you need variations of the same
test suite:

```go
func validationTestCases(fieldName string, required bool) []validationTestCase {
    cases := []validationTestCase{
        newValidationTestCase("valid "+fieldName, "valid-value", false),
        newValidationTestCase("empty "+fieldName, "", required),
    }

    if fieldName == "email" {
        cases = append(cases,
            newValidationTestCase("invalid email format", "invalid-email",
                true),
            newValidationTestCase("email too long",
                strings.Repeat("a", 100)+"@example.com", true),
        )
    }

    return cases
}

func TestUserValidation(t *testing.T) {
    t.Run("name", func(t *testing.T) {
        core.RunTestCases(t, validationTestCases("name", true))
    })
    t.Run("email", func(t *testing.T) {
        core.RunTestCases(t, validationTestCases("email", true))
    })
    t.Run("phone", func(t *testing.T) {
        core.RunTestCases(t, validationTestCases("phone", false))
    })
}
```

#### Factory Functions with Computed Test Data

For TestCase requiring computation or setup:

```go
func dateParsingTestCases() []dateParseTestCase {
    now := time.Now()
    yesterday := now.AddDate(0, 0, -1)
    nextWeek := now.AddDate(0, 0, 7)

    return []dateParseTestCase{
        newDateParseTestCase("ISO format", now.Format(time.RFC3339), now,
            false),
        newDateParseTestCase("yesterday", yesterday.Format("2006-01-02"),
            yesterday, false),
        newDateParseTestCase("next week", nextWeek.Format("Jan 2, 2006"),
            nextWeek, false),
        newDateParseTestCase("invalid format", "not-a-date", time.Time{}, true),
        newDateParseTestCase("empty string", "", time.Time{}, true),
    }
}

func TestDateParsing(t *testing.T) {
    core.RunTestCases(t, dateParsingTestCases())
}
```

#### Convenience Variant Factory Functions

**Pattern: Create multiple factory functions with different argument
signatures for the same TestCase type.**

The goal is to reduce complexity and improve readability by providing
specialised factory functions that match common usage patterns rather than
forcing all callers to provide raw struct field values.

##### Type Conversion Variants

When struct fields require complex types, provide convenience variants that
accept simpler inputs:

```go
type parseAddrTestCase struct {
    want    netip.Addr  // Complex type requiring parsing
    input   string
    name    string
    wantErr bool
}

// Base factory - requires pre-parsed netip.Addr
func newParseAddrTestCase(name, input string, want netip.Addr,
    wantErr bool) parseAddrTestCase {
    return parseAddrTestCase{
        name:    name,
        input:   input,
        want:    want,
        wantErr: wantErr,
    }
}

// Convenience variant - accepts string and handles parsing
func newParseAddrTestCaseStr(name, input, wantAddr string,
    wantErr bool) parseAddrTestCase {
    var want netip.Addr
    if !wantErr && wantAddr != "" {
        want = netip.MustParseAddr(wantAddr)
    }
    return parseAddrTestCase{
        name:    name,
        input:   input,
        want:    want,
        wantErr: wantErr,
    }
}

// Usage shows the benefits
var parseAddrTestCases = []parseAddrTestCase{
    // Simple string-based test cases
    newParseAddrTestCaseStr("IPv4 address", "192.168.1.1", "192.168.1.1",
        false),
    newParseAddrTestCaseStr("IPv6 address", "2001:db8::1", "2001:db8::1",
        false),

    // Complex cases still use the base factory
    newParseAddrTestCase("IPv4 unspecified", "0", netip.IPv4Unspecified(),
        false),
    newParseAddrTestCase("IPv6 unspecified", "::", netip.IPv6Unspecified(),
        false),
}
```

##### Semantic Intent Variants

Create variants that encode common test scenarios and reduce boolean
parameter confusion:

```go
type errGroupGoTestCase struct {
    name         string
    runFunc      func(context.Context) error
    shutdownFunc func() error
    expectError  bool
    expectCancel bool
}

// Base factory - requires all flags
func newErrGroupGoTestCase(name string, runFunc func(context.Context) error,
    shutdownFunc func() error, expectError,
    expectCancel bool) errGroupGoTestCase {
    return errGroupGoTestCase{
        name:         name,
        runFunc:      runFunc,
        shutdownFunc: shutdownFunc,
        expectError:  expectError,
        expectCancel: expectCancel,
    }
}

// Semantic convenience variants - encode intent clearly
func newErrGroupGoTestCaseSuccess(name string,
    runFunc func(context.Context) error,
    shutdownFunc func() error) errGroupGoTestCase {
    return newErrGroupGoTestCase(name, runFunc, shutdownFunc, false, false)
}

func newErrGroupGoTestCaseError(name string,
    runFunc func(context.Context) error,
    shutdownFunc func() error) errGroupGoTestCase {
    return newErrGroupGoTestCase(name, runFunc, shutdownFunc, true, true)
}

func newErrGroupGoTestCaseCancel(name string,
    runFunc func(context.Context) error,
    shutdownFunc func() error) errGroupGoTestCase {
    return newErrGroupGoTestCase(name, runFunc, shutdownFunc, false, true)
}

// Usage - intent is crystal clear
var errGroupGoTestCases = []errGroupGoTestCase{
    newErrGroupGoTestCaseSuccess("successful worker", successFunc, nil),
    newErrGroupGoTestCaseError("worker with error", errorFunc, nil),
    newErrGroupGoTestCaseCancel("worker canceled", cancelFunc, nil),
}
```

##### Default Value Variants

Provide variants that supply sensible defaults for optional parameters:

```go
type compoundErrorOKTestCase struct {
    name     string
    errs     []error
    expected bool
}

// Base factory
func newCompoundErrorOKTestCase(name string, errs []error,
    expected bool) compoundErrorOKTestCase {
    return compoundErrorOKTestCase{
        name:     name,
        errs:     errs,
        expected: expected,
    }
}

// Convenience variants with semantic defaults
func newCompoundErrorOKTestCaseEmpty(name string,
    errs []error) compoundErrorOKTestCase {
    return newCompoundErrorOKTestCase(name, errs, true) // empty errors = OK
}

func newCompoundErrorOKTestCaseHasErrors(name string,
    errs []error) compoundErrorOKTestCase {
    return newCompoundErrorOKTestCase(name, errs, false) // has errors = not OK
}
```

##### Parameter Reordering for Field Alignment (CRITICAL)

**This is the fundamental reason why factory functions are MANDATORY for
every TestCase type.**

Struct fields must be ordered for memory efficiency (field alignment), but
function parameters should be ordered for logical readability. Factory
functions decouple these two concerns:

```go
type waitGroupGoTestCase struct {
    // Memory-optimised field order (largest to smallest)
    fn          func() error  // 8 bytes (function pointer)
    errorMsg    string        // 16 bytes (string header)
    name        string        // 16 bytes (string header)
    expectError bool          // 1 byte (boolean)
    // Total: 41 bytes, padded to 48 bytes
}

// Factory uses logical parameter order, NOT struct field order
func newWaitGroupGoTestCase(name string, fn func() error, expectError bool,
    errorMsg string) waitGroupGoTestCase {
    return waitGroupGoTestCase{
        // Fields assigned in memory-optimised order, regardless of
        // parameter order
        fn:          fn,          // Memory: first (largest)
        errorMsg:    errorMsg,    // Memory: second
        name:        name,        // Memory: third
        expectError: expectError, // Memory: last (smallest)
    }
}

// Without factory, callers would be forced to use memory order:
// ❌ FORBIDDEN - Forces callers to know memory layout
var badTestCases = []waitGroupGoTestCase{
    {
        fn:          func() error { return nil },  // Memory order required
        errorMsg:    "",                           // Not logical order
        name:        "test name",                  // Confusing for readers
        expectError: false,                        // Hard to understand
    },
}

// With factory, callers use logical order:
// ✅ CORRECT - Logical, readable parameter order
var goodTestCases = []waitGroupGoTestCase{
    newWaitGroupGoTestCase("test name", func() error { return nil }, false, ""),
}
```

**Why This Matters:**

1. **Memory Efficiency**: Field-aligned structs reduce memory usage and
   improve cache performance.
2. **Readability**: Logical parameter order makes test intentions clear.
3. **Maintainability**: Changes to memory layout don't affect all call
   sites.
4. **Consistency**: Same logical parameter patterns across all TestCase types.

#### Benefits of Convenience Variants

1. **Reduced Cognitive Load**: Callers don't need to understand complex type
   construction.
2. **Clear Intent**: Semantic function names make test purpose obvious.
3. **Fewer Errors**: Less chance of parameter confusion or incorrect boolean
   flags.
4. **Consistency**: Common patterns are encoded once and reused everywhere.
5. **Maintainability**: Changes to default behaviour only need to be made
   in one place.

#### Benefits of TestCase List Factories

1. **Separation of Concerns**: Test data generation is separate from test logic.
2. **Reusability**: Factory functions can be called from multiple test
   functions.
3. **Maintainability**: Complex test data logic is centralised.
4. **Readability**: Test functions focus on execution, not data setup.
5. **Parameterisation**: Easy to create variations of test suites.

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
    t.Run("creation", runTestUserCreation)
    t.Run("update", runTestUserUpdate)
    t.Run("deletion", runTestUserDeletion)
}
```

## Field Alignment in Test Structs

Order struct fields to minimise memory padding. Place larger fields first,
smaller fields last:

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

To verify optimal field ordering, create a separate sample file with just the
struct definition (no external dependencies), then use the
`fieldalignment` tool:

<!-- markdownlint-disable MD013 -->
```bash
go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest \\
  -fix sample_file.go
```
<!-- markdownlint-enable MD013 -->

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

### Testing Fatal/FailNow Scenarios

MockT supports testing functions that call Fatal/FailNow methods via the `Run()`
method, which recovers from FailNow panics:

```go
func TestAssertionFailure(t *testing.T) {
    mock := &core.MockT{}

    // Test assertion that calls Fatal internally
    ok := mock.Run("failing assertion", func(mt core.T) {
        core.AssertEqual(mt, 1, 2, "should fail")
        // This would call mt.Fatal() internally if we used:
        // if !core.AssertEqual(mt, 1, 2, "should fail") { mt.FailNow() }
    })

    core.AssertFalse(t, ok, "assertion should fail")
    core.AssertTrue(t, mock.Failed(), "mock should be marked as failed")
    core.AssertTrue(t, mock.HasErrors(), "should have error message")
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
# Generate coverage reports (.prof, .func, .html) with sub-module support
make coverage
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
- TestCase interface methods:
  - `func (tc testCaseType) Name() string` - returns test case name.
  - `func (tc testCaseType) Test(t *testing.T)` - runs the test logic.

### Documentation

```go
// parseURLTestCase tests URL parsing functionality
type parseURLTestCase struct {
    // ... fields
}

// Name returns the test case name for identification.
func (tc parseURLTestCase) Name() string {
    return tc.name
}

// Test validates URL parsing behaviour.
func (tc parseURLTestCase) Test(t *testing.T) {
    // ... implementation
}
```

### Clean Tests

- Use `t.Helper()` in all helper functions.
- Prefer table-driven tests with TestCase for multiple data variations.
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

// DON'T: Anonymous test case structs (use TestCase interface)
tests := []struct {
    name string
    // ...
}{ /* ... */ }

// DON'T: Naked struct literals (use factory functions)
testCases := []myTestCase{
    {name: "test", input: "value", expected: "result"},
}

// DON'T: Manual loops
for _, tc := range testCases {
    t.Run(tc.Name(), tc.Test)
}

// DON'T: Missing interface validations for TestCase
// (no var _ TestCase = ... declarations)

// DON'T: Missing factory functions for TestCase
// (no newTestCaseTypeName() functions)

// DON'T: Complex TestCase lists without factory functions
var complexTestCases = []myTestCase{
    newMyTestCase("test1", param1, param2, param3),
    newMyTestCase("test2", param1, param2, param3),
    newMyTestCase("test3", param1, param2, param3),
    newMyTestCase("test4", param1, param2, param3),
    newMyTestCase("test5", param1, param2, param3),
    newMyTestCase("test6", param1, param2, param3),
    // ... many more cases
}
```

### ✅ Always Use These Patterns

```go
// DO: TestCase interface validations (for table-driven tests)
var _ TestCase = myTestCase{}

// DO: Named TestCase types implementing TestCase interface
type myTestCase struct {
    name     string
    expected result
}

func (tc myTestCase) Name() string {
    return tc.name
}

func (tc myTestCase) Test(t *testing.T) {
    t.Helper()
    core.AssertEqual(t, tc.expected, actual, "result")
}

// DO: Factory functions for all TestCase types
func newMyTestCase(name string, expected result) myTestCase {
    return myTestCase{
        name:     name,
        expected: expected,
    }
}

// DO: Use factory functions for TestCase creation
testCases := []myTestCase{
    newMyTestCase("test", expectedResult),
}

// DO: Use RunTestCases for table-driven tests
func TestMyFunction(t *testing.T) {
    core.RunTestCases(t, testCases)
}

// DO: Extract complex logic to helper methods
func (tc myTestCase) Test(t *testing.T) {
    t.Helper()
    tc.runTest(t)
}

// DO: Use core assertion functions
core.AssertNoError(t, err, "operation")

// DO: Use TestCase list factories for complex table-driven tests
func makeMyFunctionTestCases() []core.TestCase {
    return []core.TestCase{
        newMyTestCase("test1", expectedResult1),
        newMyTestCase("test2", expectedResult2),
        newMyTestCase("test3", expectedResult3),
        newMyTestCase("test4", expectedResult4),
        newMyTestCase("test5", expectedResult5),
        newMyTestCase("test6", expectedResult6),
    }
}

func TestMyFunction(t *testing.T) {
    core.RunTestCases(t, makeMyFunctionTestCases())
}

// DO: Use parameterised TestCase list factories when needed
func makeValidationTestCases(fieldName string) []core.TestCase {
    return []core.TestCase{
        newValidationTestCase("valid "+fieldName, "valid-value", false),
        newValidationTestCase("empty "+fieldName, "", true),
    }
}
```

## TestCase Compliance Checklist

Before committing table-driven test code using TestCase, verify ALL 5
requirements:

- [ ] **TestCase Interface Validations**: Added `var _ TestCase = ...` for
      all TestCase types.
- [ ] **Factory Functions**: Created `new{TestContext}TestCase()` for all
      TestCase types.
- [ ] **Factory Usage**: All TestCase declarations use factory functions
      (no naked struct literals).
- [ ] **RunTestCases Usage**: All table-driven test functions use
      `RunTestCases(t, cases)`.
- [ ] **Anonymous Functions**: No `t.Run()` anonymous functions longer than
      3 lines.
- [ ] **TestCase List Factories**: Complex TestCase generation uses
      `makeTestFunctionTestCases()` factory functions.

## Summary

By following these **MANDATORY** guidelines, all darvaza.org projects will
have:

- **Consistent testing patterns** across the entire ecosystem.
- **Lint-compliant code** that meets complexity requirements.
- **Maintainable tests** that are easy to read and modify.
- **Reliable test suites** with excellent error reporting.
- **Comprehensive coverage** with meaningful assertions.

The key is to treat test code with the same care as production code, using
the excellent utilities provided by `darvaza.org/core` to maintain
consistency and quality across all projects.

**Remember**: These are not suggestions - they are **MANDATORY**
requirements. Use TestCase **ONLY** for table-driven tests with multiple
data cases. Use standard t.Run() patterns for different test scenarios.
