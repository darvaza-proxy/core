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

# Run tests with coverage
make test GOTEST_FLAGS="-cover"

# Run tests with verbose output and coverage
make test GOTEST_FLAGS="-v -cover"

# Build test binaries without running (useful for debugging)
make test GOTEST_FLAGS="-c"

# Generate coverage reports
make coverage

# Format code and tidy dependencies (run before committing)
make tidy

# Clean build artifacts
make clean

# Update dependencies
make up

# Run go:generate directives
make generate
```

## Build System Features

### Whitespace and EOF Handling

The `internal/build/fix_whitespace.sh` script automatically:

- Removes trailing whitespace from all text files
- Ensures files end with a newline
- Excludes binary files and version control directories
- Integrates with `make fmt` for non-Go files
- Supports both directory scanning and explicit file arguments

### Markdownlint Integration

The build system includes automatic Markdown linting:

- Detects markdownlint-cli via pnpx
- Configuration in `internal/build/markdownlint.json`
- 80-character line limits and strict formatting rules
- Selective HTML allowlist (comments, br, kbd, etc.)
- Runs automatically with `make fmt` when available

### LanguageTool Integration

Grammar and style checking for Markdown files:

- Detects LanguageTool via pnpx
- British English configuration in `internal/build/languagetool.cfg`
- New `check-grammar` target
- Integrated into `make tidy`
- Checks for missing articles, punctuation, and proper hyphenation

### CSpell Integration

Spell checking for both Markdown and Go source files:

- Detects cspell via pnpx
- British English configuration in `internal/build/cspell.json`
- New `check-spelling` target
- Integrated into `make tidy`
- Custom word list for project-specific terminology
- Checks both documentation and code comments

### Coverage Collection

The build system includes automated coverage report generation:

- `make coverage` target runs tests with coverage flags
- `internal/build/make_coverage.sh` handles test execution
- Generates coverage reports in multiple formats (text, HTML)
- Coverage artifacts stored in `.tmp/coverage/` directory
- Integrated with CI/CD workflows for automated reporting

## Code Architecture

### Key Design Principles

- **Zero dependencies**: Only the Go standard library and minimal golang.org/x
  packages.
- **Generic programming**: Extensive use of Go 1.23+ generics for type-safe
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

- Go 1.23 is the minimum required version.
- The Makefile dynamically generates rules for subprojects.
- Tool versions (golangci-lint, revive) are selected based on Go version.
- This is a utility library - no business logic, only reusable helpers.
- Always use `pnpm` instead of `npm` for any JavaScript/TypeScript tooling.

## Testing with GOTEST_FLAGS

The `GOTEST_FLAGS` environment variable allows flexible test execution by
passing additional flags to `go test`. This variable is defined in the
Makefile (line 10) with an empty default value and is used when running tests
through the generated rules in `.tmp/gen.mk:39`.

### Common Usage Examples

```bash
# Run tests with race detection
make test GOTEST_FLAGS="-race"

# Run specific tests by pattern
make test GOTEST_FLAGS="-run TestSpecific"

# Generate coverage profile
make test GOTEST_FLAGS="-coverprofile=coverage.out"

# Run tests with timeout
make test GOTEST_FLAGS="-timeout 30s"

# Combine multiple flags
make test GOTEST_FLAGS="-v -race -coverprofile=coverage.out"
```

### How It Works

1. The Makefile defines `GOTEST_FLAGS ?=` (empty by default).
2. The generated rules in `.tmp/gen.mk` use it in the test target:
   `$(GO) test $(GOTEST_FLAGS) ./...`.
3. Any flags passed via `GOTEST_FLAGS` are forwarded directly to `go test`.

This provides a clean interface for passing arbitrary test flags without
modifying the Makefile, making it easy to run tests with different
configurations for debugging, coverage analysis, or CI/CD pipelines.

## Linting and Code Quality

### Documentation Standards

When editing Markdown files, ensure compliance with:

- **LanguageTool**: Check for missing articles ("a", "an", "the"), punctuation,
  and proper hyphenation of compound modifiers.
- **CSpell**: Check spelling in both documentation and code comments.
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
   - Always include a link to related documentation (e.g., AGENT.md should
     link to README.md).
   - Add prerequisites or setup instructions before diving into commands.
   - Include paths to configuration files when mentioning tools (e.g.,
     revive.toml).

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

1. Run `make tidy` for Go code formatting and whitespace clean-up.
2. Check Markdown files with CSpell, LanguageTool and markdownlint.
3. Verify all tests pass with `make test`.
4. Ensure no linting violations remain.
5. Update `AGENT.md` to reflect any changes in development workflow or
   standards.
6. Update `README.md` to reflect significant changes in functionality or API.
