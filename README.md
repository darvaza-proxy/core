# Core of helpers for darvaza.org projects

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]

This package contains simple mechanisms used by other darvaza-proxy projects.
It's not allowed to have dependencies outside of Go' Standard Library, and if something
should be on a subdirectory, it shouldn't be here.

[godoc]: https://pkg.go.dev/darvaza.org/core
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/core.svg
[goreport]: https://goreportcard.com/report/darvaza.org/core
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/core

## Network

* GetInterfacesNames
* ParseAddr/ParseNetIP
* SplitHostPort/SplitAddrPort
* JoinHostPort/MakeHostPort
* AddrPort
* AddrFromNetIP
* GetIPAddresses/GetNetIPAddresses/GetStringIPAddresses

## Generics

* Zero/IsZero
* Coalesce/IIf
* As/AsFn
* SliceAs/SliceAsFn
* SliceContains/SliceContainsFn
* SliceEqual/SliceEqualFn
* SliceMinus/SliceMinusFn
* SliceUnique/SliceUniqueFn
* SliceUniquify/SliceUniquifyFn
* SliceReplaceFn/SliceCopy/SliceCopyFn/SliceMap
* SliceRandom
* SliceSort/SliceSortFn/SliceSortOrdered
* SliceReverse/SliceReversed/SliceReversedFn
* ListContains/ListContainsFn
* ListForEach/ListForEachElement
* ListForEachBackward/ListForEachBackwardElement
* ListCopy/ListCopyFn
* MapContains
* MapListContains/MapListContainsFn
* MapListForEach/MapListForEachElement
* MapListInsert/MapListAppend
* MapListInsertUnique/MapListInsertUniqueFn
* MapListAppendUnique/MapListAppendUniqueFn
* MapListCopy/MapListCopyFn
* MapAllListContains/MapAllListContainsFn
* MapAllListForEach/MapAllListForEachElement
* MapValue
* Keys()/SortedKeys()
* NewContextKey

## Errors

### Wrappers

The `Unwrappable` type represents the classic `Unwrap() error` interface implemented
by `WrappedError`, while the `Errors` interface represents `Errors() []error`.

There are three factories for `Unwrappable`, the standard `"note: error description"`,
one for formatted notes, and a quiet one, not including the text of the original error
unless unwrapped first.

* `Wrap(err, note)` with a simple string,
* `Wrapf(err, format, args...)` when using a formatted note,
* and `QuietWrapf(err, format, args...)` for formatted errors not including
  the wrapped message in the text.

The `Unwrap(err error) []error` helper returns a slice of non-nil sub-errors built
from the following interfaces:
* `Unwrap() []error`
* `Errors() []error`
* `Unwrap() error`

For agreggating multiple errors and the `Unwrap() []error` or `Errors() []error` interfaces
we have the `CompoundError`.

### Panic and Recover

A `PanicError` is a special wrapper that includes a StackTrace and can wrap anything
and it's especially useful when used combined the standard `recover()` as shown below:

```go
defer func() {
  if err := core.AsRecovered(recover()); err != nil {
    // ...
  }
}()
```

This construct will return `nil` if there was a panic, pass-through the error if it implements
the `Recovered` interface, or wrap anything else in a `PanicError`.

`Catch()` is a companion of `PanicError` which will allows you to call a function and
either receive its organic `error` or a `PanicError` if it panicked, using a `Catcher`
instance internally.

To `panic()` automatically wrapping the reason in `PanicError{}` the following helpers
can be used:

* `Panic()`,
* `Panicf()`,
* and `PanicWrap`.

### Miscellaneous error related

* CoalesceError
* AsError/AsErrors
* IsError/IsErrorFn/IsErrorFn2
* IsTemporary/CheckIsTemporary
* IsTimeout/CheckIsTimeout
* TemporaryError/NewTemporaryError/NewTimeoutError
* WaitGroup/ErrGroup
* Frame/Stack
* Here/StackFrame/StackTrace
* CallStacker

* ErrNotImplemented/ErrTODO
* ErrExists/ErrNotExists
* ErrInvalid/ErrUnknown
* ErrNilReceiver

## See also

* [darvaza.org/cache](https://pkg.go.dev/darvaza.org/cache)
* [darvaza.org/resolve](https://pkg.go.dev/darvaza.org/resolve)
* [darvaza.org/slog](https://pkg.go.dev/darvaza.org/slog)
* [darvaza.org/x/config](https://pkg.go.dev/darvaza.org/x/config)
* [darvaza.org/x/fs](https://pkg.go.dev/darvaza.org/x/fs)
* [darvaza.org/x/net](https://pkg.go.dev/darvaza.org/x/net)
* [darvaza.org/x/tls](https://pkg.go.dev/darvaza.org/x/tls)
* [darvaza.org/x/web](https://pkg.go.dev/darvaza.org/x/web)
