# AGENT.md

This file provides guidance to AI agents when working with code in this
repository.

## Repository Overview

`darvaza.org/core` is a foundational Go utility library that provides
essential helpers and common functionality with zero external dependencies.
It serves as the base for other darvaza.org projects.

## Common Development Commands

```bash
# Full build cycle (get deps, generate, tidy, build)
make all

# Run tests
make test

# Format code and tidy dependencies (run before committing)
make tidy

# Clean build artifacts
make clean

# Update dependencies
make up

# Run go:generate directives
make generate
```

## Code Architecture

### Key Design Principles

- **Zero dependencies**: Only Go standard library and minimal golang.org/x
  packages
- **Generic programming**: Extensive use of Go 1.22+ generics for type-safe
  utilities
- **Single package**: Everything is in the `core` package, no subpackages

### Major Components

- **Error handling** (errors.go, panicerror.go, compounderror.go): Advanced
  error wrapping with stack traces, panic recovery, and compound errors
- **Context utilities** (context.go): Type-safe context keys with
  `ContextKey[T]`
- **Network utilities** (addrs.go, addrport.go, splithostport.go): Address
  parsing, interface enumeration
- **Generic collections** (slices.go, lists.go, maps.go): Functional
  programming patterns for collections
- **Synchronization** (sync.go): Advanced synchronization primitives

### Code Quality Standards

The project enforces strict linting rules via revive:

- Max function length: 40 lines
- Max function results: 3
- Max arguments: 5
- Cognitive complexity: 7
- Cyclomatic complexity: 10

Always run `make tidy` before committing to ensure proper formatting.

### Testing Patterns

- Table-driven tests are preferred
- Helper functions like `S[T]()` create test slices
- Comprehensive coverage for generic functions is expected

## Important Notes

- Go 1.22 is the minimum required version
- The Makefile dynamically generates rules for subprojects
- Tool versions (golangci-lint, revive) are selected based on Go version
- This is a utility library - no business logic, only reusable helpers
