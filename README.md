# Core helpers for darvaza.org projects

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]

This package contains simple mechanisms used by other darvaza-proxy
projects. It's not allowed to have dependencies outside of Go's Standard
Library, and if something should be on a subdirectory, it shouldn't be here.

[godoc]: https://pkg.go.dev/darvaza.org/core
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/core.svg
[goreport]: https://goreportcard.com/report/darvaza.org/core
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/core

## Type Constraints

Generic type constraints for use with Go generics:

* `Signed` - signed integer types
* `Unsigned` - unsigned integer types
* `Integer` - all integer types (signed and unsigned)
* `Float` - floating-point types
* `Complex` - complex number types
* `Bool` - boolean type
* `String` - string type
* `Ordered` - types that support ordering operations

## Context

* `ContextKey[T]` - type-safe context key type
* `NewContextKey[T]()` creates a ContextKey adding type-safety and ease of use
  to the standard `context.WithValue()`
* `WithTimeout()` and `WithTimeoutCause()` are equivalent to
  `context.WithDeadline()` and `context.WithDeadlineCause()` but receiving
  a duration instead of an absolute time

## Network Utilities

### IP Address Functions

* `GetIPAddresses()` - get IP addresses as `netip.Addr`
* `GetNetIPAddresses()` - get IP addresses as `net.IP`
* `GetStringIPAddresses()` - get IP addresses as strings
* `AddrFromNetIP(ip)` - convert `net.IP` to `netip.Addr`
* `ParseAddr(s)` - parse string to `netip.Addr`
* `ParseNetIP(s)` - parse string to `net.IP`

### Host/Port Functions

* `SplitHostPort(hostport)` - split host:port string
* `SplitAddrPort(addrport)` - split address:port string
* `JoinHostPort(host, port)` - join host and port
* `MakeHostPort(host, port)` - create host:port string
* `AddrPort(addr, port)` - create `netip.AddrPort`

### Interface Functions

* `GetInterfacesNames()` - get network interface names

## Generic Utilities

### Basic Utilities

* `Zero[T]()` returns the zero value for type T
* `IsZero[T](v)` checks if a value is the zero value for its type
* `Coalesce[T](values...)` returns the first non-zero value
* `IIf[T](condition, ifTrue, ifFalse)` conditional expression

### Type Conversion

* `As[T,V](v)` attempts to convert value to target type
* `AsFn[T,V](v, fn)` converts value using a provided function
* `AsError[T](v)` attempts to convert value to error
* `AsErrors[T](v)` attempts to convert value to error slice

### Slice Operations

#### Search and Comparison

* `SliceContains[T]` / `SliceContainsFn[T]`
* `SliceEqual[T]` / `SliceEqualFn[T]`

#### Transformation

* `SliceAs[T,V]` / `SliceAsFn[T,V]`
* `SliceMap[T1,T2]` - maps slice elements to new type
* `SliceReplaceFn[T]` - replaces elements matching condition
* `SliceCopy[T]` / `SliceCopyFn[T]`

#### Set Operations

* `SliceMinus[T]` / `SliceMinusFn[T]` - set difference
* `SliceUnique[T]` / `SliceUniqueFn[T]` - unique elements
* `SliceUniquify[T]` / `SliceUniquifyFn[T]` - remove duplicates in-place

#### Sorting and Ordering

* `SliceSort[T]` / `SliceSortFn[T]` / `SliceSortOrdered[T]`
* `SliceReverse[T]` / `SliceReversed[T]` / `SliceReversedFn[T]`

#### Utilities

* `SliceRandom[T]` - random element selection

### List Operations (container/list)

* `ListContains[T]` / `ListContainsFn[T]`
* `ListForEach[T]` / `ListForEachElement`
* `ListForEachBackward[T]` / `ListForEachBackwardElement`
* `ListCopy[T]` / `ListCopyFn[T]`

### Map Operations

#### Basic Map Functions

* `MapContains[K]()` checks if a map contains a key
* `MapValue[K,V]()` returns the value for a key, or a fallback value
* `Keys[K,T]()` returns a slice of the keys in the map
* `SortedKeys[K,T]()` returns a sorted slice of the keys
* `SortedValues[K,T]()` returns values sorted by key
* `SortedValuesCond[K,T]()` returns filtered values sorted by key
* `SortedValuesUnlikelyCond[K,T]()` like `SortedValuesCond` but more efficient

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

* `ErrNotImplemented` - functionality not yet implemented
* `ErrTODO` - placeholder for future implementation
* `ErrExists` - resource already exists
* `ErrNotExists` - resource does not exist
* `ErrInvalid` - invalid input or state
* `ErrUnknown` - unknown or unspecified error
* `ErrNilReceiver` - method called on nil receiver
* `ErrUnreachable` - indicates impossible condition

### Error Wrapping

The `Unwrappable` interface represents the classic `Unwrap() error` pattern,
implemented by `WrappedError`. The `Errors` interface represents multi-error
containers with `Errors() []error`.

Error wrapping functions:

* `Wrap(err, note)` - wrap with simple string note
* `Wrapf(err, format, args...)` - wrap with formatted note
* `QuietWrap(err, note)` - wrap without including original error text
* `Unwrap(err) []error` - extract all sub-errors from wrapped errors

### Compound Errors

The `CompoundError` type aggregates multiple errors:

* Implements both `Unwrap() []error` and `Errors() []error` interfaces
* `.AppendError(err)` / `.Append(errs...)` - add errors
* `.AsError()` - convert to single error or nil
* `.Ok()` - check if no errors

### Panic Handling

The `PanicError` type wraps panic values with stack traces:

* `NewPanicError()` / `NewPanicErrorf()` - create panic errors
* `NewPanicWrap()` / `NewPanicWrapf()` - wrap existing errors as panics
* `Panic()` / `Panicf()` / `PanicWrap()` / `PanicWrapf()` - panic with
  `PanicError`

Panic recovery utilities:

* `Recovered` interface - marks errors from recovered panics
* `AsRecovered(v)` - convert `recover()` result to error
* `Catcher` type - safely call functions that might panic
* `Catch(fn)` - execute function, returning error if panic occurs

```go
defer func() {
  if err := core.AsRecovered(recover()); err != nil {
    // handle panic as error
  }
}()
```

### Unreachable Conditions

For indicating impossible code paths:

* `NewUnreachableError()` - create unreachable error
* `NewUnreachableErrorf(format, args...)` - create formatted unreachable error

These create `PanicError` instances with stack traces.

### Temporary and Timeout Errors

Special error types for network-style temporary and timeout conditions:

* `TemporaryError` type - implements `Temporary() bool`
* `NewTemporaryError(err)` - wrap error as temporary
* `NewTimeoutError(err)` - wrap error as timeout
* `IsTemporary(err)` / `CheckIsTemporary(err)` - test if error is temporary
* `IsTimeout(err)` / `CheckIsTimeout(err)` - test if error is timeout

### Error Testing and Utilities

* `IsError[T](err)` / `IsErrorFn[T](err, fn)` / `IsErrorFn2[T](err, fn)` -
  type-safe error testing
* `CoalesceError(errs...)` - return first non-nil error

## Stack Tracing

Utilities for capturing and working with call stacks:

* `Frame` - represents a single stack frame
* `Stack` - represents a call stack
* `MaxDepth` - maximum stack depth constant
* `CallStacker` interface - types that provide call stacks

### Stack Capture Functions

* `Here()` - capture current stack frame
* `StackFrame(skip)` - capture-specific stack frame
* `StackTrace(skip, depth)` - capture call stack

### Frame Methods

* `.Name()` / `.FuncName()` / `.PkgName()` - function/package names
* `.SplitName()` - split full name into package and function
* `.File()` / `.Line()` / `.FileLine()` - source location
* `.Format()` - formatted representation

## Synchronization

### WaitGroup

Enhanced wait group with error handling:

* `WaitGroup` - wait group that collects errors
* `.OnError(fn)` - set error handler
* `.Go(fn)` / `.GoCatch(fn)` - run functions in `goroutines`
* `.Wait()` - wait for completion
* `.Err()` - get first error

### ErrGroup

Context-aware error group with cancellation:

* `ErrGroup` - context-based error group
* `.SetDefaults()` - configure with defaults
* `.OnError(fn)` - set error handler
* `.Cancel()` / `.Context()` - cancellation control
* `.Go(fn)` / `.GoCatch(fn)` - run functions with context
* `.Wait()` - wait and return first error
* `.IsCancelled()` / `.Cancelled()` - check cancellation state

### Deprecated

* ~~SpinLock~~ Deprecated in favour of
  [darvaza.org/x/sync/spinlock][x-sync-spinlock]

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
