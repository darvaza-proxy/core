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

# Generate Codecov configuration and upload scripts
make codecov

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

- Detects LanguageTool via pnpx.
- British English configuration in `internal/build/languagetool.cfg`.
- Available via `make check-grammar` target.
- Not integrated into `make tidy` due to false positives.
- Checks for missing articles, punctuation, and proper hyphenation.

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

### Codecov Integration

Enhanced coverage reporting with monorepo support:

- `make codecov` target generates Codecov configuration
- `internal/build/make_codecov.sh` creates:
  - `codecov.yml`: Dynamic configuration with per-module flags
  - `codecov.sh`: Upload script for bulk submission
- Module-specific coverage targets (80% default)
- Path mappings for accurate coverage attribution
- GitHub Actions workflow automatically uploads coverage data
- PR comments show coverage changes per module
- See [internal/build/README-coverage.md](internal/build/README-coverage.md)
  for details

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

#### Test Helper Functions

The project uses a comprehensive set of test helper functions defined in
`testutils_test.go` to reduce boilerplate and improve test consistency:

**Slice Creation:**

- `S[T](values...)` - Creates test slices concisely: `S(1, 2, 3)` instead of
  `[]int{1, 2, 3}`
- `S[T]()` - Creates empty slices: `S[string]()` instead of `[]string{}`

**Assertion Helpers:**

- `AssertEqual[T](t, expected, actual, msg...)` - Generic value comparison with
  better error messages
- `AssertSliceEqual[T](t, expected, actual, msg...)` - Slice comparison using
  `reflect.DeepEqual`
- `AssertError(t, err, expectError, msg...)` - Standardized error expectation
  checking
- `AssertBool(t, actual, expected, msg...)` - Boolean assertions with context
- `AssertPanic(t, fn, expectedPanic, msg...)` - Simplified panic testing
- `AssertNoPanic(t, fn, msg...)` - Ensure functions don't panic

**Advanced Helpers:**

- `RunConcurrentTest(t, numWorkers, workerFn)` - Concurrent testing with
  goroutines
- `RunBenchmark(b, setupFn, execFn)` - Benchmark testing with setup/execution
  phases
- `RunTestCases(t, []TestCase)` - Table-driven test runner (requires
  `TestCase` interface)

**Usage Examples:**

```go
// Before: Manual assertions
if !reflect.DeepEqual(got, expected) {
    t.Errorf("Expected %v, got %v", expected, got)
}

// After: Helper function
AssertSliceEqual(t, expected, got, "operation failed")

// Before: Manual error checking
if expectError && err == nil {
    t.Error("Expected error but got nil")
} else if !expectError && err != nil {
    t.Errorf("Expected no error but got: %v", err)
}

// After: Helper function
AssertError(t, err, expectError, "operation error expectation")
```

These helpers provide:

- Consistent error messages across all tests
- Reduced boilerplate code
- Better test maintainability
- Clearer test intent

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

# Generate coverage profile (alternative to 'make coverage')
make test GOTEST_FLAGS="-coverprofile=coverage.out"

# Run tests with timeout
make test GOTEST_FLAGS="-timeout 30s"

# Combine multiple flags
make test GOTEST_FLAGS="-v -race -coverprofile=coverage.out"

# Run benchmarks
make test GOTEST_FLAGS="-bench=. -benchmem"

# Skip long-running tests
make test GOTEST_FLAGS="-short"

# Test with specific CPU count
make test GOTEST_FLAGS="-cpu=1,2,4"
```

### Integration with Coverage

While `make coverage` provides automated coverage collection across all
modules, you can use `GOTEST_FLAGS` for more targeted coverage analysis:

```bash
# Coverage for specific package with detailed output
make test GOTEST_FLAGS="-v -coverprofile=coverage.out -covermode=atomic"

# Coverage with HTML output
make test GOTEST_FLAGS="-coverprofile=coverage.out"
go tool cover -html=coverage.out
```

### How It Works

1. The Makefile defines `GOTEST_FLAGS ?=` (empty by default).
2. The generated rules in `.tmp/gen.mk` use it in the test target:
   `$(GO) test $(GOTEST_FLAGS) ./...`.
3. Any flags passed via `GOTEST_FLAGS` are forwarded directly to `go test`.

This provides a clean interface for passing arbitrary test flags without
modifying the Makefile, making it easy to run tests with different
configurations for debugging, coverage analysis, or CI/CD pipelines.

## CI/CD and Code Analysis

### DeepSource Configuration

The project uses DeepSource for static code analysis. Configuration is in the
`.deepsource.toml` file:

- Shell analyser is configured for POSIX sh dialect.
- To ignore specific issues for certain files, use `[[issues]]` blocks with
  `paths` (not `exclude_patterns`).
- Common shell issues:
  - SH-1091: "local is undefined in POSIX sh" - excluded for all .sh files.
  - SH-2013: "Use while read for reading lines" - disable with
    ShellCheck directive comment.

### GitHub Actions

- **Codecov workflow**: Automatically runs on push/PR to generate coverage
  reports.
- **Make workflow**: Tests across Go versions 1.23 and 1.24.
- All CI checks must pass before merging PRs.

### Working with Build Tools

When LanguageTool reports issues:

- Custom dictionary is auto-generated from CSpell words in
  `.tmp/languagetool-dict.txt`.
- Technical terms should be added to `internal/build/cspell.json`.
- False positives for code-related punctuation are disabled in
  `languagetool.cfg`.

## Linting and Code Quality

### Field Alignment

The project enforces struct field alignment optimization for memory efficiency
using the `fieldalignment` tool. This helps reduce memory usage by ordering
struct fields to minimize padding.

#### Running Field Alignment Fixes

To automatically fix field alignment issues across the codebase:

```bash
# Run the fieldalignment tool with automatic fixes
GOXTOOLS="golang.org/x/tools/go/analysis/passes"
FA="$GOXTOOLS/fieldalignment/cmd/fieldalignment"
go run "$FA@latest" -fix ./...
```

This tool will:

- Analyze all struct definitions in the project.
- Reorder fields to minimize memory padding.
- Automatically update source files with optimized field ordering.

#### Field Alignment Notes

- Always run `make tidy` after field alignment fixes to ensure all linting
  passes.
- Field alignment changes may require updating struct literal initializations.
- The tool is safe to run repeatedly - it only makes changes when beneficial.
- Memory savings can be significant for frequently allocated structs.
- Run field alignment manually as needed for struct optimization.

### golangci-lint Configuration

The project uses golangci-lint for comprehensive Go code linting.
Configuration is stored in `.golangci.yml` in the project root.

#### Current Configuration Status

The project uses golangci-lint v1.64.8 (pinned to avoid v2 configuration
issues). The current `.golangci.yml` uses the v1 format that functions
properly but triggers IDE schema validation warnings:

- ✅ **Functionally works**: All linters and settings are properly applied
- ✅ **Field alignment enabled**: `govet: enable: [fieldalignment]` works
- ⚠️ **Schema validation**: IDE expects v2 format, but system uses v1.64

#### Schema Version Compatibility

**Current v1.64 format (working)**:

```yaml
linters-settings:
  govet:
    enable: [fieldalignment]
linters:
  enable:
    - govet
```

**v2 format (IDE expects, but system doesn't support)**:

```yaml
version: "2"
linters:
  enable:
    - govet
linters-settings:
  govet:
    enable: [fieldalignment]
```

#### Schema Validation Issues

Common IDE diagnostics and their meanings:

1. **Missing property "version"**: IDE schema expects `version: "2"`.
2. **Property linters-settings is not allowed**: IDE expects v2 structure.
3. **cSpell warnings**: Linter names are technical terms not in spell-check
   dictionary.

#### Resolution Status

- **Current priority**: Low - configuration works functionally.
- **IDE warnings**: Cannot be resolved without upgrading golangci-lint to v2.
- **System constraint**: Project uses v1.64.8 to avoid v2 configuration issues.
- **Workaround**: IDE warnings can be ignored as they don't affect
  functionality.

#### Field Alignment Integration

- Always run `make tidy-root` which includes golangci-lint checks
- Field alignment linter is properly configured and working
- Schema format doesn't affect linting functionality, only IDE validation
- Technical linter names are added to `internal/build/cspell.json`

### Documentation Standards

When editing Markdown files, ensure compliance with:

- **LanguageTool**: Optionally check for missing articles ("a", "an", "the"),
  punctuation, and proper hyphenation of compound modifiers via
  `make check-grammar`.
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

1. **ALWAYS run `make tidy` first** - Fix ALL issues before committing:
   - Go code formatting and whitespace clean-up
   - Markdown files checked with CSpell and markdownlint
   - Shell scripts checked with shellcheck
   - If `make tidy` fails, fix the issues and run it again until it passes
2. Verify all tests pass with `make test`.
3. Ensure no linting violations remain.
4. Update `AGENT.md` to reflect any changes in development workflow or
   standards.
5. Update `README.md` to reflect significant changes in functionality or API.

## Troubleshooting

### Common Issues and Solutions

1. **LanguageTool false positives**:
   - Add technical terms to `internal/build/cspell.json`.
   - Dictionary will auto-regenerate on next `make check-grammar`.
   - For persistent issues, consider adding rules to `languagetool.cfg`.

2. **DeepSource shell issues**:
   - Use ShellCheck disable comments for specific lines.
   - Update `.deepsource.toml` with issue-specific `paths` configurations.
   - Remember: DeepSource uses `paths`, not `exclude_patterns` in
     `[[issues]]` blocks.

3. **Coverage collection failures**:
   - Ensure `.tmp/index` exists by running `make .tmp/index`.
   - Check that all modules have test files.
   - Use `GOTEST_FLAGS` to pass additional flags to tests.

4. **Linting tool detection**:
   - Tools are auto-detected via `pnpx`.
   - If tools aren't found, they're replaced with `true` (no-op).
   - Install tools globally with `pnpm install -g <tool>` if needed.

5. **golangci-lint schema validation**:
   - IDE warnings about missing `version` field or `linters-settings` placement.
   - Configuration functions properly despite schema warnings.
   - Project uses v1.64.8 to avoid v2 configuration format issues.
   - System uses pinned version via Makefile `GOLANGCI_LINT_VERSION`.
   - cSpell warnings for linter names are added to `internal/build/cspell.json`.
