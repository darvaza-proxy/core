# AGENT.md

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Repository Overview

`darvaza.org/core` is a foundational Go utility library that provides
essential helpers and common functionality with zero external dependencies.
It serves as the base for other darvaza.org projects.

## Prerequisites

Before starting development, ensure you have:

- Go 1.22 or later installed (check with `go version`).
- `make` command available (usually pre-installed on Unix systems).
- `$GOPATH` configured correctly (typically `~/go`).
- Git configured for proper line endings.

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

- **Zero dependencies**: Only the Go standard library and minimal golang.org/x
  packages.
- **Generic programming**: Extensive use of Go 1.22+ generics for type-safe
  utilities.
- **Single package**: Everything is in the `core` package, no subpackages.

### Major Components

- **Error handling** (errors.go, panicerror.go, compounderror.go): Advanced
  error wrapping with stack traces, panic recovery, and compound errors.
- **Context utilities** (context.go): Type-safe context keys with
  `ContextKey[T]`.
- **Network utilities** (addrs.go, addrport.go, splithostport.go): Address
  parsing, interface enumeration.
- **Generic collections** (slices.go, lists.go, maps.go): Functional
  programming patterns for collections.
- **Synchronization** (sync.go): Advanced synchronization primitives.

### Code Quality Standards

The project enforces strict linting rules via revive (configuration in
`internal/build/revive.toml`):

- Max function length: 40 lines.
- Max function results: 3.
- Max arguments: 5.
- Cognitive complexity: 7.
- Cyclomatic complexity: 10.

Always run `make tidy` before committing to ensure proper formatting.

### Testing Patterns

- Table-driven tests are preferred.
- Helper functions like `S[T]()` create test slices.
- Comprehensive coverage for generic functions is expected.

## Important Notes

- Go 1.22 is the minimum required version.
- The Makefile dynamically generates rules for subprojects.
- Tool versions (golangci-lint, revive) are selected based on Go version.
- This is a utility library - no business logic, only reusable helpers.
- Always use `pnpm` instead of `npm` for any JavaScript/TypeScript tooling.

## Linting and Code Quality

### Documentation Standards

When editing markdown files, ensure compliance with:

- **LanguageTool**: Check for missing articles ("a", "an", "the"), punctuation,
  and proper hyphenation of compound modifiers.
- **Markdownlint**: Follow standard Markdown formatting rules.

### Common Documentation Issues to Check

1. **Missing Articles**: Ensure proper use of "a", "an", and "the".
   - ❌ "converts value using provided function".
   - ✅ "converts value using a provided function".

2. **Missing Punctuation**: End all list items consistently.
   - ❌ "Comprehensive coverage for generic functions is expected".
   - ✅ "Comprehensive coverage for generic functions is expected.".

3. **Compound Modifiers**: Hyphenate when used as modifiers.
   - ❌ "capture specific stack frame".
   - ✅ "capture-specific stack frame".

### Writing Documentation Guidelines

When creating or editing documentation files:

1. **File Structure**:
   - Always include a link to related documentation (e.g., AGENT.md should link to README.md).
   - Add prerequisites or setup instructions before diving into commands.
   - Include paths to configuration files when mentioning tools (e.g., revive.toml).

2. **Formatting Consistency**:
   - End all bullet points with periods for consistency.
   - Capitalize proper nouns correctly (JavaScript, TypeScript, Markdown).
   - Use consistent punctuation in examples and lists.

3. **Clarity and Context**:
   - Provide context for AI agents and developers alike.
   - Include "why" explanations, not just "what" descriptions.
   - Add examples for complex concepts or common pitfalls.

4. **Maintenance**:
   - Update documentation when adding new tools or changing workflows.
   - Keep the pre-commit checklist current with project practices.
   - Review documentation changes for the issues listed above.

### Pre-commit Checklist

1. Run `make tidy` for Go code formatting.
2. Check markdown files with LanguageTool and markdownlint.
3. Verify all tests pass with `make test`.
4. Ensure no linting violations remain.
5. Update `AGENT.md` to reflect any changes in development workflow or standards.
6. Update `README.md` to reflect significant changes in functionality or API.
