# AGENTS.md
<!-- cspell:ignore linters -->

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Related Documentation

- [README.md](README.md) — Package overview and API reference.
- [BUILDING.md](BUILDING.md) — Shared build-system reference for all
  darvaza.org projects (make targets, tooling, linting, CI, pre-commit
  workflow, troubleshooting).
- [internal/build/README-coverage.md](internal/build/README-coverage.md) —
  Coverage system documentation.
- [TESTING.md](TESTING.md) — Testing patterns and guidelines for all
  darvaza.org projects.
- [TESTING_core.md](TESTING_core.md) — Core-specific testing patterns.

## Repository Overview

`darvaza.org/core` is a foundational Go utility library that provides
essential helpers and common functionality with zero external dependencies.
It serves as the base for other darvaza.org projects.

## Prerequisites

- Go 1.25 or later (the project's minimum).
- `make` available; `$GOPATH` configured.

See [BUILDING.md → Required Tools](BUILDING.md#required-tools) for the
full toolchain.

## Code Architecture

### Key Design Principles

- **Zero dependencies**: only the Go standard library and minimal
  `golang.org/x` packages.
- **Generic programming**: extensive use of Go generics for type-safe
  utilities.
- **Single package**: everything is in the `core` package; no subpackages.

### Major Components

- **Error handling** (`errors.go`, `panicerror.go`, `compounderror.go`):
  advanced error wrapping with stack traces, panic recovery, and compound
  errors.
- **Context utilities** (`context.go`): type-safe context keys with
  `ContextKey[T]`.
- **Network utilities** (`addrs.go`, `addrport.go`, `splithostport.go`):
  address parsing, interface enumeration.
- **Generic collections** (`slices.go`, `lists.go`, `maps.go`): functional
  programming patterns for collections.
- **Synchronisation** (`sync.go`): advanced synchronisation primitives.

## Testing

- Table-driven tests via `TestCase` apply when ≥2 rows of shared-shape
  data feed one assertion path; otherwise write plain test functions.
  See [TESTING.md](./TESTING.md) for the full decision rule.
- All testing utilities are public in `testing.go` for external use.
- Comprehensive coverage for generic functions is expected.
- Core-specific testing patterns and self-testing approaches are in
  [TESTING_core.md](./TESTING_core.md).
- Test execution flags and the generated `test` rule are documented in
  [BUILDING.md → Test Execution Options](BUILDING.md#test-execution-options).

## Build & Tooling Reference

This repo uses the shared darvaza.org build system. Most-used commands:

```bash
make all          # full build cycle (get, generate, tidy, build)
make tidy         # format, lint, validate
make test         # run tests (no cache reuse)
make coverage     # tests with coverage
```

Full reference in [BUILDING.md](BUILDING.md). Key sections:

- [Build Targets](BUILDING.md#build-targets) — primary and per-module
  targets.
- [Test Execution Options](BUILDING.md#test-execution-options) —
  `GOTEST_FLAGS` semantics; the generated `test` rule passes `-count=1`.
- [Code Quality Standards](BUILDING.md#code-quality-standards) — revive
  linting limits (function length, complexity, argument counts).
- [Field Alignment](BUILDING.md#field-alignment) — safe `fieldalignment`
  workflow via probe file.
- [golangci-lint Configuration](BUILDING.md#golangci-lint-configuration).
- [Documentation Standards](BUILDING.md#documentation-standards) —
  LanguageTool, CSpell, Markdownlint conventions and common issues.
- [Pre-commit Checklist](BUILDING.md#pre-commit-checklist).
- [CI/CD Integration](BUILDING.md#cicd-integration) including DeepSource
  configuration.
- [Troubleshooting](BUILDING.md#troubleshooting).

## Important Notes

- This is a utility library — no business logic, only reusable helpers.
- The zero-dependency discipline is load-bearing; do not introduce
  external imports beyond the standard library and minimal `golang.org/x`.
