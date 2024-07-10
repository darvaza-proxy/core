# Core of helpers for darvaza.org projects

[![Go Reference][godoc-badge]][godoc]
[![Go Report Card][goreport-badge]][goreport]
[![Codebeat Score][codebeat-badge]][codebeat]

This package contains simple mechanisms used by other darvaza-proxy projects.
It's not allowed to have dependencies outside of Go' Standard Library, and if something
should be on a subdirectory, it shouldn't be here.

[codebeat]: https://codebeat.co/projects/github-com-darvaza-proxy-core-main
[codebeat-badge]: https://codebeat.co/badges/aaee3212-75a8-4f4d-8fe8-58bc8bcc108f
[godoc]: https://pkg.go.dev/darvaza.org/core
[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/core.svg
[goreport]: https://goreportcard.com/report/darvaza.org/core
[goreport-badge]: https://goreportcard.com/badge/darvaza.org/core

## Network

* GetInterfacesNames
* ParseAddr/ParseNetIP
* SplitHostPort/SplitAddrPort
* AddrPort
* AddrFromNetIP
* GetIPAddresses/GetNetIPAddresses/GetStringIPAddresses

## Generics

* Zero/IsZero
* Coalesce/IIf
* SliceContains/SliceContainsFn
* SliceEqual/SliceEqualFn
* SliceMinus/SliceMinusFn
* SliceUnique/SliceUniqueFn
* SliceUniquify/SliceUniquifyFn
* SliceReplaceFn/SliceCopyFn
* SliceRandom
* SliceSort/SliceSortFn
* ListContains/ListContainsFn
* ListForEach/ListForEachElement
* ListForEachBackward/ListForEachBackwardElement
* MapContains
* MapListContains/MapListContainsFn
* MapListForEach/MapListForEachElement
* MapListInsert/MapListAppend
* MapListInsertUnique/MapListInsertUniqueFn
* MapListAppendUnique/MapListAppendUniqueFn
* MapAllListContains/MapAllListContainsFn
* MapAllListForEach/MapAllListForEachElement
* Keys()/SortedKeys()
* NewContextKey

## Errors

* Wrap/QuietWrap/Unwrappable/Unwrap
* Errors/CompoundError
* CoalesceError
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

* [darvaza.org/slog](https://pkg.go.dev/darvaza.org/slog)
* [darvaza.org/gossipcache](https://pkg.go.dev/darvaza.org/gossipcache)
* [darvaza.org/darvaza/acme](https://pkg.go.dev/darvaza.org/darvaza/acme)
* [darvaza.org/darvaza/agent](https://pkg.go.dev/darvaza.org/darvaza/agent)
* [darvaza.org/darvaza/server](https://pkg.go.dev/darvaza.org/darvaza/server)
* [darvaza.org/darvaza/shared](https://pkg.go.dev/darvaza.org/darvaza/shared)
