# Core helpers for darvaza.org projects

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]
[![codecov][codecov-badge]][codecov]

This package contains simple mechanisms used by other darvaza.org
projects. It's not allowed to have dependencies outside of Go's Standard
Library, and if something should be on a subdirectory, it shouldn't be here.

[godoc]: https://pkg.go.dev/darvaza.org/core
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/core.svg
[goreport]: https://goreportcard.com/report/darvaza.org/core
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/core
[codecov]: https://codecov.io/gh/darvaza-proxy/core
[codecov-badge]: https://codecov.io/gh/darvaza-proxy/core/graph/badge.svg

## Type Constraints

Generic type constraints for use with Go generics:

* `Signed` - signed integer types.
* `Unsigned` - unsigned integer types.
* `Integer` - all integer types (signed and unsigned).
* `Float` - floating-point types.
* `Complex` - complex number types.
* `Bool` - boolean type.
* `String` - string type.
* `Ordered` - types that support ordering operations.

## Context

### Context Keys

* `NewContextKey[T](name)` - creates a new type-safe key bound to specified
  type and friendly name.
* `ContextKey[T].WithValue(ctx, value)` - safely attach value to context,
  comparable to standard `context.WithValue()`.
* `ContextKey[T].Get(ctx)` - extract value bound to this key in context,
  returns (value, found) with nil receiver safety.

### Timeout Utilities

* `WithTimeout(parent, duration)` - equivalent to `context.WithDeadline()`
  but takes duration instead of absolute time. Returns parent context and
  no-op cancel for zero/negative durations.
* `WithTimeoutCause(parent, duration, cause)` - equivalent to
  `context.WithDeadlineCause()` but takes duration instead of absolute time.
  Attaches custom cause error to timeout context.

## Network Utilities

### IP Address Functions

* `GetIPAddresses()` - get IP addresses as `netip.Addr`.
* `GetNetIPAddresses()` - get IP addresses as `net.IP`.
* `GetStringIPAddresses()` - get IP addresses as strings.
* `AddrFromNetIP(ip)` - convert `net.IP` to `netip.Addr`.
* `ParseAddr(s)` - parse string to `netip.Addr`.
* `ParseNetIP(s)` - parse string to `net.IP`.

### Host/Port Functions

#### Parsing and Splitting

* `SplitHostPort(hostport)` - enhanced version of `net.SplitHostPort` that
  accepts portless strings and validates both host and port. Supports IPv6
  addresses, international domain names, and descriptive error messages.
* `SplitAddrPort(addrport)` - splits IP address and optional port into
  `netip.Addr` and `uint16`. Validates address format and port range
  (1-65535), returns zero values for portless addresses.

#### Joining and Construction

* `JoinHostPort(host, port)` - enhanced version of `net.JoinHostPort` that
  validates inputs and returns portless host when port is empty. Properly
  handles IPv6 bracketing and international domain names.
* `MakeHostPort(hostport, defaultPort)` - constructs validated host:port
  string from input with optional default port. Rejects port 0 in input,
  supports portless output when default is 0.
* `AddrPort(addr, port)` - creates `netip.AddrPort` from components.

### Interface Functions

* `GetInterfacesNames()` - get network interface names.

## Generic Utilities

### Zero Value Utilities

#### Zero Value Creation

* `Zero[T]()` - returns the zero value for type T using reflection when
  needed. Supports all Go types including complex generics, interfaces, and
  custom types.

#### Zero Value Detection

* `IsZero(v)` - reports whether a value is in an uninitialized state and ready
  to be set. Answers the question: "Is this value uninitialized and ready to
  be set?"

Key semantic distinctions:

* **Nil vs Empty**: `[]int(nil)` returns `true` (needs initialization),
  `[]int{}` returns `false` (already initialized).
* **Pointer States**: `(*int)(nil)` returns `true` (can be assigned),
  `new(int)` returns `false` (already points to memory).
* **Interface Support**: Types implementing `IsZero() bool` are handled
  via their method, enabling custom zero semantics.

#### Nil Value Detection

* `IsNil(v)` - reports whether a value is nil (typed or untyped). Answers
  the question: "Is this value nil?"

Key distinctions from `IsZero`:

* **Scope**: Only checks for nil state, not zero state.
* **Basic Types**: `IsNil(0)` returns `false` (integers cannot be nil),
  `IsZero(0)` returns `true` (zero integer is uninitialized).
* **Collections**: `IsNil([]int{})` returns `false` (empty slice is not nil),
  `IsZero([]int{})` returns `false` (empty slice is initialized).
* **Structs**: `IsNil(struct{}{})` returns `false` (structs cannot be nil),
  `IsZero(struct{}{})` returns `true` (zero struct is uninitialized).

#### Same Value Detection

* `IsSame(a, b)` - reports whether two values are the same. For reference
  types (slices, maps, pointers), compares by pointer equality. For value
  types (numbers, strings, booleans), compares by equal values. Two nils
  of the same type are considered the same.

#### Other Utilities

* `Coalesce[T](values...)` returns the first non-zero value.
* `IIf[T](condition, ifTrue, ifFalse)` conditional expression.

### Type Conversion

* `As[T,V](v)` attempts to convert value to target type
* `AsFn[T,V](v, fn)` converts value using a provided function
* `AsError[T](v)` attempts to convert value to error
* `AsErrors[T](v)` attempts to convert value to error slice

### Slice Operations

#### Search and Comparison

* `SliceContains[T](slice, value)` - check if slice contains value.
* `SliceContainsFn[T](slice, value, eq)` - check containment with custom
  equality.
* `SliceEqual[T](a, b)` - compare two slices for equality.
* `SliceEqualFn[T](a, b, eq)` - compare slices with custom equality function.

#### Transformation

* `SliceAs[T,V]` / `SliceAsFn[T,V]` - convert slice elements to different type.
* `SliceMap[T1,T2](slice, fn)` - transform each element with cumulative
  function.
* `SliceReplaceFn[T](slice, fn)` - replace/filter elements in-place.
* `SliceCopy[T](slice)` - create shallow copy of slice.
* `SliceCopyFn[T](slice, fn)` - create filtered/transformed copy.

#### Set Operations

* `SliceMinus[T](a, b)` - elements in `a` but not in `b`.
* `SliceMinusFn[T](a, b, eq)` - set difference with custom equality.
* `SliceUnique[T](slice)` - return slice with unique elements only.
* `SliceUniqueFn[T](slice, eq)` - unique elements with custom equality.
* `SliceUniquify[T](ptr)` - remove duplicates in-place, modify original.
* `SliceUniquifyFn[T](ptr, eq)` - remove duplicates with custom equality.

#### Sorting and Ordering

* `SliceSort[T](slice, cmp)` - sort using comparison function (returns int).
* `SliceSortFn[T](slice, less)` - sort using less function (returns bool).
* `SliceSortOrdered[T](slice)` - sort ordered types (int, string, float64).
* `SliceReverse[T](slice)` - reverse slice in-place.
* `SliceReversed[T](slice)` - return reversed copy.
* `SliceReversedFn[T](slice, fn)` - return transformed and reversed copy.

#### Utilities

* `SliceRandom[T](slice)` - select random element, returns (value, found).

### List Operations (container/list)

#### Search and Membership

* `ListContains[T](list, value)` - check if list contains element with default
  equality.
* `ListContainsFn[T](list, value, eq)` - check if list contains element with
  custom equality function.

#### Iteration

* `ListForEach[T](list, fn)` - iterate forward over list values until fn
  returns true.
* `ListForEachElement(list, fn)` - iterate forward over list elements until fn
  returns true.
* `ListForEachBackward[T](list, fn)` - iterate backward over list values until
  fn returns true.
* `ListForEachBackwardElement(list, fn)` - iterate backward over list elements
  until fn returns true.

#### Copying and Transformation

* `ListCopy[T](list)` - create shallow copy of list.
* `ListCopyFn[T](list, fn)` - create filtered/transformed copy with helper
  function.

### Map Operations

#### Basic Map Functions

* `MapContains[K]()` checks if a map contains a key.
* `MapValue[K,V]()` returns the value for a key, or a fallback value.
* `Keys[K,T]()` returns a slice of the keys in the map.
* `SortedKeys[K,T]()` returns a sorted slice of the keys.
* `SortedValues[K,T]()` returns values sorted by key.
* `SortedValuesCond[K,T]()` returns filtered values sorted by key.
* `SortedValuesUnlikelyCond[K,T]()` like `SortedValuesCond` but more efficient.

#### Map List Operations

* `MapListContains[K,T]` / `MapListContainsFn[K,T]`
* `MapListForEach[K,T]` / `MapListForEachElement[K]`
* `MapListInsert[K,T]` / `MapListAppend[K,T]`
* `MapListInsertUnique[K,T]` / `MapListInsertUniqueFn[K,T]`
* `MapListAppendUnique[K,T]` / `MapListAppendUniqueFn[K,T]`
* `MapListCopy[T]` / `MapListCopyFn[K,V]`

#### Map All List Operations

* `MapAllListContains[K,T]` / `MapAllListContainsFn[K,T]`
* `MapAllListForEach[K,T]` / `MapAllListForEachElement[K]`

## Error Handling

### Standard Error Variables

Predefined error values for common conditions:

* `ErrNotImplemented` - functionality not yet implemented.
* `ErrTODO` - placeholder for future implementation.
* `ErrExists` - resource already exists.
* `ErrNotExists` - resource does not exist.
* `ErrInvalid` - invalid input or state.
* `ErrUnknown` - unknown or unspecified error.
* `ErrNilReceiver` - method called on nil receiver.
* `ErrUnreachable` - indicates impossible condition.

### Error Wrapping

The `Unwrappable` interface represents the classic `Unwrap() error` pattern,
implemented by `WrappedError`. The `Errors` interface represents multi-error
containers with `Errors() []error`.

Error wrapping functions:

* `Wrap(err, note)` - wrap with simple string note.
* `Wrapf(err, format, args...)` - wrap with formatted note.
* `QuietWrap(err, note)` - wrap without including original error text.
* `Unwrap(err) []error` - extract all sub-errors from wrapped errors.

### Compound Errors

The `CompoundError` type aggregates multiple errors:

* Implements both `Unwrap() []error` and `Errors() []error` interfaces.
* `.AppendError(err)` / `.Append(errs...)` - add errors.
* `.AsError()` - convert to single error or nil.
* `.OK()` - check if no errors.

### Panic Handling

The `PanicError` type wraps panic values with stack traces:

* `NewPanicError()` / `NewPanicErrorf()` - create panic errors.
* `NewPanicWrap()` / `NewPanicWrapf()` - wrap existing errors as panics.
* `Panic()` / `Panicf()` / `PanicWrap()` / `PanicWrapf()` - panic with
  `PanicError`.

Panic recovery utilities:

* `Recovered` interface - marks errors from recovered panics.
* `AsRecovered(v)` - convert `recover()` result to error.
* `Catcher` type - safely call functions that might panic.
* `Catch(fn)` - execute function, returning error if panic occurs.

```go
defer func() {
  if err := core.AsRecovered(recover()); err != nil {
    // handle panic as error
  }
}()
```

### Must/Maybe Utilities

Convenience functions for common error-handling patterns:

* `Must[T](value T, err error) T` - returns value or panics with `PanicError` if
  err is not nil. Follows the common Go pattern of Must* functions for cases
  where errors should never occur.
* `Maybe[T](value T, err error) T` - always returns the value, ignoring any
  error. When an error occurs, the returned value may be the zero value of type
  T depending on the function that produced it.
* `MustOK[T](value T, ok bool) T` - returns value or panics with `PanicError` if
  ok is false. Useful for operations that should always succeed, such as map
  access or type assertions that are guaranteed to be valid.
* `MaybeOK[T](value T, ok bool) T` - always returns the value, ignoring the ok
  flag. When ok is false, the returned value is typically the zero value of
  type T from the failed operation.
* `MustT[T](value any) T` - equivalent to `MustOK(As[any, T](value))`. Returns
  converted value or panics with `PanicError` if type conversion fails.
* `MaybeT[T](value any) T` - equivalent to `MaybeOK(As[any, T](value))`. Returns
  converted value on success, or the zero value of type T if conversion fails.

All these utilities work together with the `As[any, T]` function, which returns
`(value, ok)` for type conversion, allowing custom handling of conversion
failures.

```go
// Must - panic on error (use in tests, config loading, etc.)
config := Must(loadConfig("config.json"))  // panics if loadConfig fails
conn := Must(net.Dial("tcp", "localhost:8080"))  // panics if dial fails

// Maybe - ignore errors, proceed with values (may be zero values)
content := Maybe(os.ReadFile("optional.txt"))  // []byte{} if file missing
count := Maybe(strconv.Atoi(userInput))       // 0 if parsing fails

// MustOK - panic on failure (use when operation should always succeed)
value := MustOK(MapValue(m, "key", 0))  // panics if key doesn't exist
str := MustOK(As[any, string](v))       // panics if v is not a string

// MaybeOK - ignore ok flag, proceed with values (zero values on failure)
value := MaybeOK(MapValue(m, "key", 0))  // 0 if key doesn't exist
str := MaybeOK(As[any, string](v))       // "" if type assertion fails

// MustT - equivalent to MustOK(As[any, T](value))
str := MustT[string](value)  // panics if value is not a string
num := MustT[int](value)     // panics if value is not an int

// MaybeT - equivalent to MaybeOK(As[any, T](value))
str := MaybeT[string](value)  // "" (zero value) if value is not a string
num := MaybeT[int](value)     // 0 (zero value) if value is not an int

// As - the building block for type conversion with custom handling
if str, ok := As[any, string](value); ok {
    // use converted string
} else {
    str = "custom default"  // your own fallback logic
}
```

### Unreachable Conditions

For indicating impossible code paths:

* `NewUnreachableError()` - create unreachable error.
* `NewUnreachableErrorf(format, args...)` - create formatted unreachable error.

These create `PanicError` instances with stack traces.

### Temporary and Timeout Errors

Special error types for network-style temporary and timeout conditions:

* `TemporaryError` type - implements `Temporary() bool` and `IsTemporary() bool`
  interfaces for marking recoverable errors.
* `NewTemporaryError(err)` - wrap error as temporary condition.
* `NewTimeoutError(err)` - wrap error as timeout condition with both temporary
  and timeout properties.
* `IsTemporary(err)` - recursively test if error chain contains temporary
  condition via `Temporary()` or `IsTemporary()` methods.
* `CheckIsTemporary(err)` - test single error for temporary condition without
  unwrapping chain, returns (is, known) tuple.
* `IsTimeout(err)` - recursively test if error chain contains timeout
  condition via `Timeout()` or `IsTimeout()` methods.
* `CheckIsTimeout(err)` - test single error for timeout condition without
  unwrapping chain, returns (is, known) tuple.

### Error Testing and Utilities

* `IsError[T](err)` / `IsErrorFn[T](err, fn)` / `IsErrorFn2[T](err, fn)` -
  type-safe error testing with generic constraints and custom checker functions.
* `CoalesceError(errs...)` - return first non-nil error from argument list.

## Stack Tracing

Stack tracing utilities for debugging, error reporting, and call context:

### Core Types

* `Frame` - represents a single function call frame with source location.
* `Stack` - slice of frames representing a complete call stack.
* `MaxDepth` - maximum stack capture depth (32 frames).
* `CallStacker` interface - types that can provide their call stack.

### Stack Capture Functions

#### Frame Capture

* `Here()` - capture the current stack frame where called. Returns nil if
  capture fails. Useful for immediate calling context.
* `StackFrame(skip)` - capture a specific frame in the call stack, skipping
  the specified number of levels. Returns nil if insufficient frames.

#### Complete Stack Capture

* `StackTrace(skip)` - capture complete call stack starting from skip level.
  Returns empty Stack on failure. Maximum depth limited by MaxDepth.

### Frame Information Methods

#### Function Names

* `Frame.Name()` - full qualified function name including package path
  (e.g., "darvaza.org/core.TestFunction").
* `Frame.FuncName()` - function name only without package qualification
  (e.g., "TestFunction").
* `Frame.PkgName()` - package path portion only (e.g., "darvaza.org/core").
* `Frame.SplitName()` - split full name into (package, function) components.
  Handles generic functions by ignoring "[...]" suffixes.

#### Source Location

* `Frame.File()` - full path to source file containing the function.
* `Frame.Line()` - line number within source file (0 if unavailable).
* `Frame.FileLine()` - formatted "file:line" string for display.

#### Formatting

* `Frame.Format(fmt.State, rune)` - implements fmt.Formatter interface with
  support for multiple format verbs:
  * `%s` - source file basename.
  * `%d` - line number.
  * `%n` - function name (short form).
  * `%v` - equivalent to `%s:%d`.
  * `%+s` - function name + full file path (newline separated).
  * `%+n` - full qualified function name.
  * `%+v` - equivalent to `%+s:%d`.

* `Stack.Format(fmt.State, rune)` - format entire stack with same verbs as
  Frame plus '#' flag support:
  * `%#s`, `%#n`, `%#v` - each frame on new line.
  * `%#+s`, `%#+n`, `%#+v` - numbered frames with [index/total] prefix.

### Usage Examples

```go
// Capture current location
frame := Here()
if frame != nil {
    fmt.Printf("Called from %s at %s", frame.FuncName(), frame.FileLine())
}

// Capture complete stack for error reporting
stack := StackTrace(1) // skip current function
fmt.Printf("Stack trace:%+v", stack)

// Numbered stack output
fmt.Printf("Debug stack:%#+v", stack)
```

## Testing Utilities

This package provides comprehensive public testing utilities for both internal
library tests and external library users.

### T Interface and MockT

* `T` interface - abstracts testing functionality for both `*testing.T` and
  mock implementations.
* `MockT` - thread-safe mock testing.T implementation with error/log
  collection, helper tracking, state inspection (`HasErrors()`, `HasLogs()`,
  `LastError()`, `LastLog()`), reset capabilities, and full Fatal/FailNow
  support with panic recovery via the `Run()` method.

### Test Helpers

* `S[T](values...)` - concise slice creation: `S(1, 2, 3)` instead of
  `[]int{1, 2, 3}`.
* `S[T]()` - empty slice creation: `S[string]()` instead of `[]string{}`.

### Assertion Functions

All assertions return boolean results, log successful cases, and work with
both `*testing.T` and `MockT`:

#### Basic Assertions

* `AssertEqual[T](t, expected, actual, msg...)` - generic value comparison.
* `AssertNotEqual[T](t, expected, actual, msg...)` - generic inequality
  comparison.
* `AssertSliceEqual[T](t, expected, actual, msg...)` - slice comparison using
  `reflect.DeepEqual`.
* `AssertSame(t, expected, actual, msg...)` - same value/reference comparison
  using pointer equality for reference types, value equality for basic types.
* `AssertNotSame(t, expected, actual, msg...)` - different value/reference
  comparison.
* `AssertTrue(t, condition, msg...)` / `AssertFalse(t, condition, msg...)` -
  boolean assertions.
* `AssertNil(t, value, msg...)` / `AssertNotNil(t, value, msg...)` - nil
  checking.
* `AssertContains(t, text, substring, msg...)` - string containment.

#### Error and Type Assertions

* `AssertError(t, err, msg...)` / `AssertNoError(t, err, msg...)` - error
  presence/absence.
* `AssertErrorIs(t, err, target, msg...)` - error chain checking with
  `errors.Is`.
* `AssertTypeIs[T](t, value, msg...)` - type assertion with casting, returns
  (value, ok).
* `AssertPanic(t, fn, expectedPanic, msg...)` / `AssertNoPanic(t, fn, msg...)` -
  panic testing.

#### Fatal Assertions

All `AssertMustFoo()` functions call the corresponding `AssertFoo()` function
and automatically call `t.FailNow()` if the assertion fails, terminating test
execution immediately. These follow Go testing conventions where "Fatal"
methods terminate execution, similar to `t.Error()` vs `t.Fatal()`.

**Key Differences:**

* `AssertFoo()` - like `t.Error()`, logs failure and allows test to continue.
* `AssertMustFoo()` - like `t.Fatal()`, logs failure and terminates test
  execution.

**Fatal Assertion Functions:**

* `AssertMustEqual[T](t, expected, actual, msg...)` - terminate on inequality.
* `AssertMustNotEqual[T](t, expected, actual, msg...)` - terminate on equality.
* `AssertMustSliceEqual[T](t, expected, actual, msg...)` - terminate on slice
  inequality.
* `AssertMustTrue(t, condition, msg...)` /
  `AssertMustFalse(t, condition, msg...)` - terminate on boolean mismatch.
* `AssertMustNil(t, value, msg...)` / `AssertMustNotNil(t, value, msg...)` -
  terminate on nil check failure.
* `AssertMustContains(t, text, substring, msg...)` - terminate if substring not
  found.
* `AssertMustError(t, err, msg...)` / `AssertMustNoError(t, err, msg...)` -
  terminate on error expectation mismatch.
* `AssertMustErrorIs(t, err, target, msg...)` - terminate on error chain
  mismatch.
* `AssertMustTypeIs[T](t, value, msg...) T` - terminate on type assertion
  failure, returns cast value.
* `AssertMustPanic(t, fn, expectedPanic, msg...)` /
  `AssertMustNoPanic(t, fn, msg...)` - terminate on panic expectation mismatch.

**Usage Examples:**

```go
// Error pattern - logs failure, test continues
AssertEqual(t, expected, actual, "value check")
// execution continues even if assertion fails

// Fatal pattern - logs failure, test terminates
AssertMustEqual(t, expected, actual, "critical value check")
// execution stops here if assertion fails

// Traditional early abort pattern (equivalent to AssertMust*)
if !AssertEqual(t, expected, actual, "critical value") {
    t.FailNow()
}
```

### Advanced Testing Utilities

* `TestCase` interface - standardised interface for table-driven tests with
  `Name()` and `Test(t)` methods.
* `RunTestCases[T TestCase](t, cases)` - table-driven test runner for
  TestCase implementations.
* `RunConcurrentTest(t, numWorkers, workerFn)` - concurrent testing with
  goroutines.
* `RunBenchmark(b, setupFn, execFn)` - benchmark testing with
  setup/execution phases.

### Documentation

For detailed testing patterns and guidelines:

* [TESTING.md](./TESTING.md) - General testing patterns for all darvaza.org
  projects
* [TESTING_core.md](./TESTING_core.md) - Core-specific testing patterns and
  self-testing approaches

## Synchronization

### WaitGroup

Enhanced wait group with error handling:

* `WaitGroup` - wait group that collects errors.
* `.OnError(fn)` - set error handler.
* `.Go(fn)` / `.GoCatch(fn)` - run functions in `goroutines`.
* `.Wait()` - wait for completion.
* `.Err()` - get first error.

### ErrGroup

Context-aware error group with cancellation:

* `ErrGroup` - context-based error group.
* `.SetDefaults()` - configure with defaults.
* `.OnError(fn)` - set error handler.
* `.Cancel()` / `.Context()` - cancellation control.
* `.Go(fn)` / `.GoCatch(fn)` - run functions with context.
* `.Wait()` - wait and return first error.
* `.IsCancelled()` / `.Cancelled()` - check cancellation state.

### Deprecated

* ~~SpinLock~~ Deprecated in favour of
  [darvaza.org/x/sync/spinlock][x-sync-spinlock]

## Development

**Requirements:** Go 1.23 or later

For detailed development setup, build commands, and AI agent guidance:

* [AGENT.md](./AGENT.md) - Development guidelines, build system, and testing
  patterns.

### Quick Start

```bash
make all    # Full build cycle (get deps, generate, tidy, build)
make test   # Run tests
make tidy   # Format and tidy (run before committing)
```

## See also

* [darvaza.org/cache][cache]
* [darvaza.org/resolver][resolver]
* [darvaza.org/slog][slog]
* [darvaza.org/x/cmp][x-cmp]
* [darvaza.org/x/config][x-config]
* [darvaza.org/x/container][x-container]
* [darvaza.org/x/fs][x-fs]
* [darvaza.org/x/net][x-net]
* [darvaza.org/x/sync][x-sync]
* [darvaza.org/x/tls][x-tls]
* [darvaza.org/x/web][x-web]

[cache]: https://pkg.go.dev/darvaza.org/cache
[resolver]: https://pkg.go.dev/darvaza.org/resolver
[slog]: https://pkg.go.dev/darvaza.org/slog
[x-cmp]: https://pkg.go.dev/darvaza.org/x/cmp
[x-config]: https://pkg.go.dev/darvaza.org/x/config
[x-container]: https://pkg.go.dev/darvaza.org/x/container
[x-fs]: https://pkg.go.dev/darvaza.org/x/fs
[x-net]: https://pkg.go.dev/darvaza.org/x/net
[x-sync]: https://pkg.go.dev/darvaza.org/x/sync
[x-sync-spinlock]: https://pkg.go.dev/darvaza.org/x/sync/spinlock
[x-tls]: https://pkg.go.dev/darvaza.org/x/tls
[x-web]: https://pkg.go.dev/darvaza.org/x/web
