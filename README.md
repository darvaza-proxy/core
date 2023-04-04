# Core of helpers for darvaza.org projects

[![Go Reference](https://pkg.go.dev/badge/darvaza.org/core.svg)](https://pkg.go.dev/darvaza.org/core)

This package contains simple mechanisms used by other darvaza-proxy projects.
It's not allowed to have dependencies outside of Go' Standard Library, and if something
should be on a subdirectory, it shouldn't be here.

## Network

* GetInterfacesNames
* ParseAddr/ParseNetIP
* SplitHostPort
* AddrPort
* AddrFromNetIP
* GetIPAddresses/GetNetIPAddresses/GetStringIPAddresses

## Generics

* Coalesce/IIf
* SliceContains/SliceContainsFn
* SliceMinus/SliceMinusFn
* SliceUnique/SliceUniqueFn
* SliceUniquify/SliceUniquifyFn
* SliceReplaceFn/SliceCopyFn
* SliceRandom
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
* NewContextKey

## Errors

* Wrap/Wrapf/Unwrappable
* Errors/CompoundError
* AsRecovered/Recovered
* Catcher
* PanicError
* Panic/Panicf/PanicWrap/PanicWrapf
* WaitGroup
* Frame/Stack
* Here/StackFrame/StackTrace
* CallStacker

## See also

* [darvaza.org/slog](https://pkg.go.dev/darvaza.org/slog)
* [darvaza.org/gossipcache](https://pkg.go.dev/darvaza.org/gossipcache)
* [darvaza.org/darvaza/acme](https://pkg.go.dev/darvaza.org/darvaza/acme)
* [darvaza.org/darvaza/agent](https://pkg.go.dev/darvaza.org/darvaza/agent)
* [darvaza.org/darvaza/server](https://pkg.go.dev/darvaza.org/darvaza/server)
* [darvaza.org/darvaza/shared](https://pkg.go.dev/darvaza.org/darvaza/shared)
