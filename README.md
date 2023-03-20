# Core of helpers for darvaza.org projects

[![Go Reference](https://pkg.go.dev/badge/github.com/darvaza-proxy/core.svg)](https://pkg.go.dev/github.com/darvaza-proxy/core)

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
* SliceUnique
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

## Errors

* Wrap/Wrapf/Unwrappable
* AsRecovered/Recovered
* Catcher
* PanicError
* Panic/Panicf/PanicWrap/PanicWrapf
* WaitGroup
* Frame/Stack
* Here/StackFrame/StackTrace
* CallStacker

## See also

* [github.com/darvaza-proxy/slog](https://pkg.go.dev/github.com/darvaza-proxy/slog)
* [github.com/darvaza-proxy/gossipcache](https://pkg.go.dev/github.com/darvaza-proxy/gossipcache)
* [github.com/darvaza-proxy/darvaza/acme](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/acme)
* [github.com/darvaza-proxy/darvaza/agent](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/agent)
* [github.com/darvaza-proxy/darvaza/server](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/server)
* [github.com/darvaza-proxy/darvaza/shared](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/shared)
