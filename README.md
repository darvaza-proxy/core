# Core of helpers for darvaza.org projects

This package contains helpers shared by:

* [github.com/darvaza-proxy/slog](https://pkg.go.dev/github.com/darvaza-proxy/slog)
* [github.com/darvaza-proxy/gossipcache](https://pkg.go.dev/github.com/darvaza-proxy/gossipcache)
* [github.com/darvaza-proxy/darvaza/acme](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/acme)
* [github.com/darvaza-proxy/darvaza/agent](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/agent)
* [github.com/darvaza-proxy/darvaza/shared](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/shared)
* [github.com/darvaza-proxy/darvaza/server](https://pkg.go.dev/github.com/darvaza-proxy/darvaza/server)

## Network

* GetInterfacesNames
* ParseAddr/ParseNetIP
* AddrPort
* AddrFromNetIP
* GetIPAddresses/GetNetIPAddresses/GetStringIPAddresses

## Generics

* SliceContains/SliceContainsFn
* SliceMinus/SliceMinusFn
* ListContains/ListContainsFn
* ListForEach/ListForEachElement
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
* Recover/Recovered
* Catcher
* PanicError
* WaitGroup
* Frame/Stack
* Here/StackFrame/StackTrace
* CallStacker
