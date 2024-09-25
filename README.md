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
* Keys()/SortedKeys()
* NewContextKey

## Errors

* Wrap/QuietWrap/Unwrappable/Unwrap
* Errors/CompoundError
* CoalesceError
* AsError/AsErrors
* IsError/IsErrorFn/IsErrorFn2
* IsTemporary/CheckIsTemporary
* IsTimeout/CheckIsTimeout
* AsRecovered/Recovered
* Catcher
* PanicError
* Panic/Panicf/PanicWrap
* TemporaryError/NewTemporaryError/NewTimeoutError
* WaitGroup/ErrGroup
* Frame/Stack
* Here/StackFrame/StackTrace
* CallStacker

* ErrNotImplemented/ErrTODO
* ErrExists/ErrNotExists
* ErrInvalid/ErrUnknown

## See also

* [darvaza.org/cache](https://pkg.go.dev/darvaza.org/cache)
* [darvaza.org/resolve](https://pkg.go.dev/darvaza.org/resolve)
* [darvaza.org/slog](https://pkg.go.dev/darvaza.org/slog)
* [darvaza.org/x/config](https://pkg.go.dev/darvaza.org/x/config)
* [darvaza.org/x/fs](https://pkg.go.dev/darvaza.org/x/fs)
* [darvaza.org/x/net](https://pkg.go.dev/darvaza.org/x/net)
* [darvaza.org/x/tls](https://pkg.go.dev/darvaza.org/x/tls)
* [darvaza.org/x/web](https://pkg.go.dev/darvaza.org/x/web)
