# Build System Documentation for Darvaza.org Projects

This document describes the shared build system architecture and usage patterns
used across all darvaza.org projects. These guidelines ensure consistent build
behaviour, quality tooling, and reliable CI/CD workflows across the entire
ecosystem.

## Overview

The darvaza.org projects use a sophisticated, shared build system designed
to work seamlessly across both individual projects and monorepos. The system
provides consistent tooling, configuration, and workflows whilst maintaining
project independence.

## Architecture

### Core Components

The build system consists of these key elements:

1. **Shared Makefile** (`/Makefile`) - Main build orchestration.
2. **Internal Build Tools** (`/internal/build/`) - Shell scripts and
   configurations.
3. **Configuration Files** - Standard files like `.editorconfig`,
   `.golangci.yml`, `renovate.json`.
4. **GitHub Workflows** (`/.github/workflows/`) - CI/CD automation.
5. **Dynamic Rule Generation** - Module discovery and Makefile rule creation.
6. **Temporary Directory** (`.tmp/`) - Generated files and build artefacts.
7. **VS Code Integration** (`.vscode/`) - Editor configuration with symlinks
   to shared settings.

### Design Philosophy

- **Monorepo-compatible**: Works with single projects or multi-module
  repositories.
- **Dynamic discovery**: Automatically finds Go modules and generates build
  rules.
- **Consistent tooling**: Same linting, formatting, and testing across all
  projects.
- **Version-aware**: Selects appropriate tool versions based on Go version.
- **Tool detection**: Graceful fallback when optional tools aren't available.

## Directory Structure

```text
project-root/
├── Makefile                    # Main build orchestration
├── .editorconfig              # Editor configuration
├── .golangci.yml              # Go linting configuration
├── .gitignore                 # Git exclusions (includes .tmp)
├── renovate.json              # Dependency update configuration
├── .github/workflows/         # CI/CD workflows
│   ├── build.yml             # Multi-version Go builds
│   ├── test.yml              # Testing (optional)
│   ├── race.yml              # Race detection testing
│   ├── codecov.yml           # Coverage reporting
│   └── renovate.yml          # Dependency updates
├── .vscode/                   # VS Code configuration
│   ├── settings.json         # Editor settings
│   └── cspell.json           # Symlink to internal/build/cspell.json
├── .tmp/                      # Generated files (gitignored)
│   ├── index                 # Module discovery index
│   ├── gen.mk                # Generated Makefile rules
│   ├── coverage/             # Coverage reports and scripts
│   └── languagetool-dict.txt # Generated dictionary
└── internal/build/           # Build system implementation
    ├── gen_index.sh          # Module discovery
    ├── gen_mk.sh             # Dynamic rule generation
    ├── get_version.sh        # Tool version selection
    ├── make_coverage.sh      # Coverage collection
    ├── make_codecov.sh       # Codecov integration
    ├── fix_whitespace.sh     # Whitespace normalisation
    ├── cspell.json           # Spell checking configuration
    ├── markdownlint.json     # Markdown linting configuration
    ├── languagetool.cfg      # Grammar checking configuration
    ├── revive.toml           # Additional Go linting rules
    └── README-coverage.md    # Coverage system documentation
```

## Key Build Scripts

### Module Discovery (`gen_index.sh`)

Automatically discovers Go modules in the repository:

- Scans for `go.mod` files recursively.
- Extracts module paths and dependencies.
- Generates an index file with module metadata.
- Handles module replacement directives.
- Supports grouping prefixes for organisation.

**Output Format**: `name:directory:module_path:dependencies`

### Dynamic Rule Generation (`gen_mk.sh`)

Creates Makefile rules for each discovered module:

- Generates file lists for each module.
- Creates per-module targets (`tidy-core`, `test-resolver`, etc.).
- Handles module dependencies and build order.
- Supports conditional rule generation based on file presence.
- Generates revive exclusions for submodules.

**Generated Commands**:

- `build` - Compile Go packages
- `coverage` - Run tests with coverage collection
- `get` - Download module dependencies
- `race` - Run tests with race detection (CGO_ENABLED=1)
- `test` - Run unit tests
- `tidy` - Format, lint, and validate code
- `up` - Update module dependencies

### Tool Version Selection (`get_version.sh`)

Selects appropriate tool versions based on Go version:

- Compares current Go version against requirements.
- Returns compatible tool versions.
- Supports version progression (newer Go = newer tools).
- Graceful fallback for unknown versions.

### Coverage System (`make_coverage.sh`)

Enhanced coverage testing for individual modules:

- Tests single module with `-covermode=atomic` for atomic coverage.
- Generates multiple output formats (.prof, .func, .html, .stdout).
- Improved error formatting with dedicated format function.
- Better failure reporting with filtered test output.
- Uses `go -C` for proper directory handling.

### Coverage Merge Utility (`merge_coverage.sh`)

Standalone utility for merging coverage profiles:

- Takes header from first file and appends data from all others.
- Proper error handling for empty or missing files.
- Follows Unix philosophy of single responsibility.
- Used by main coverage system and available for manual operations.

### Codecov Integration (`make_codecov.sh`)

Simplified Codecov integration for monorepo coverage:

- Generates only upload script (no complex codecov.yml).
- Uses separate calls per module with proper flags.
- Removes `--codecov-yml-path` dependency.
- Relies on Codecov's automatic configuration detection.
- Simplified file naming (`coverage_${name}.prof`).

## Temporary Directory (`.tmp/`)

The `.tmp/` directory contains generated files and build artefacts:

### Key Files

- **`index`**: Module discovery results from `gen_index.sh`.
- **`gen.mk`**: Generated Makefile rules included by main Makefile.
- **`languagetool-dict.txt`**: Auto-generated dictionary from cspell words.
- **`coverage/`**: Directory containing coverage reports and upload scripts.

### Gitignore Integration

The `.tmp/` directory is excluded from version control via `.gitignore`:

```gitignore
.tmp
*.test
```

This prevents build artefacts from being committed whilst allowing the build
system to cache generated files locally.

## VS Code Integration

### Configuration Sharing

The build system provides VS Code integration through `.vscode/` directories:

- **`settings.json`**: Editor-specific settings for each project.
- **`cspell.json`**: Symlinked to `internal/build/cspell.json` for consistent
  spell checking.

### Symlink Structure

The cspell configuration is shared via symbolic links:

```bash
.vscode/cspell.json -> ../internal/build/cspell.json
```

This ensures VS Code uses the same spell checking dictionary as the build
system, maintaining consistency between editor and CI environments.

## Configuration Files

### Editor Configuration (`.editorconfig`)

Standardises code formatting across editors:

```ini
[*]
charset = utf-8
end_of_line = lf
indent_style = tab
indent_size = 8
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_size = 4

[*.{json,yaml,yml,js,ts}]
indent_style = space
indent_size = 2

[*.md]
indent_style = space
indent_size = 2
trim_trailing_whitespace = false
```

### Go Linting (`.golangci.yml`)

Comprehensive Go code analysis:

- Uses golangci-lint v2.3.0+ with v2 configuration format.
- Enables 15+ linters including `fieldalignment`, `revive`, `staticcheck`.
- Configures revive with 20+ specific rules.
- Excludes generated code and common false positives.
- Integrates formatters (`gofmt`, `goimports`).

### Dependency Management (`renovate.json`)

Automated dependency updates:

- Extends Renovate's recommended configuration.
- Restricts Go version updates to supported versions.
- Runs `go mod tidy` after updates.
- JSON schema validation for configuration.

## Build Targets

### Primary Targets

- **`all`**: Complete build cycle (`get`, `generate`, `tidy`, `build`).
- **`clean`**: Remove temporary files and build artefacts.
- **`clean-coverage`**: Remove coverage files and reports.
- **`fmt`**: Format Go code and fix whitespace.
- **`tidy`**: Format, lint, and validate code.
- **`generate`**: Run `go:generate` directives.
- **`coverage`**: Run tests with coverage collection per module.
- **`codecov`**: Generate Codecov configuration and coverage data.
- **`race`**: Run tests with race detection enabled per module.

### Per-Module Targets

The system automatically generates targets for each discovered module:

- **`tidy-{module}`**: Lint and validate specific module.
- **`test-{module}`**: Run tests for specific module.
- **`build-{module}`**: Build specific module.
- **`coverage-{module}`**: Run coverage tests for specific module.
- **`race-{module}`**: Run race detection tests for specific module.
- **`get-{module}`**: Download dependencies for specific module.
- **`up-{module}`**: Update dependencies for specific module.

### Tool Integration

#### Optional Tool Detection

The build system gracefully handles missing tools:

- **markdownlint**: Markdown formatting via `pnpx`.
- **cspell**: Spell checking for docs and code.
- **languagetool**: Grammar checking for documentation.
- **shellcheck**: Shell script validation.
- **jq**: JSON processing for configuration.

Tools are auto-detected and replaced with `true` (no-op) if unavailable.

#### Required Tools

- **Go 1.23+**: Required for `go -C` directory changes.
- **golangci-lint**: Go code linting (version selected by Go version).
- **revive**: Additional Go linting rules.
- **make**: Build orchestration.
- **git**: Version control operations.

## Testing Framework

### Coverage Collection

The coverage system provides comprehensive testing:

1. **Module Discovery**: Finds all Go modules automatically.
2. **Individual Testing**: Tests each module with full coverage.
3. **Progress Reporting**: Shows real-time progress and coverage percentages.
4. **Failure Handling**: Continues testing other modules if one fails.
5. **Report Generation**: Creates merged coverage reports.

### CI/CD Integration

GitHub Actions workflows provide:

- **Build Testing**: Tests across Go 1.23 and 1.24.
- **Race Detection**: Dedicated workflow for race condition testing.
- **Coverage Reporting**: Automatic Codecov uploads.
- **Dependency Updates**: Automated Renovate PRs.
- **Branch Protection**: Ignores WIP branches.

### Test Execution Options

The `GOTEST_FLAGS` variable allows flexible test execution:

```bash
# Run with race detection via dedicated target
make race

# Run race detection with verbose output
make race GOTEST_FLAGS="-v"

# Run race detection on specific module
make race-core

# Run specific tests
make test GOTEST_FLAGS="-run TestSpecific"

# Generate coverage
make test GOTEST_FLAGS="-coverprofile=coverage.out"

# Run benchmarks
make test GOTEST_FLAGS="-bench=. -benchmem"
```

## Code Quality Standards

### Linting Rules (Revive)

Strict code quality enforcement:

- **Function length**: 40 lines maximum.
- **Function results**: 3 maximum.
- **Function arguments**: 5 maximum.
- **Cognitive complexity**: 7 maximum.
- **Cyclomatic complexity**: 10 maximum.

### Additional Quality Tools

- **Field Alignment**: Struct optimisation for memory efficiency.
- **Race Detection**: Comprehensive race condition testing with CGO enabled.
- **Spell Checking**: Documentation and code comments.
- **Grammar Checking**: Markdown documentation.
- **Whitespace Normalisation**: Consistent file formatting.

## Monorepo Features

### Multi-Module Support

The build system handles complex repository structures:

- **Automatic Discovery**: Finds modules at any depth.
- **Dependency Tracking**: Understands module relationships.
- **Isolated Building**: Each module builds independently.
- **Shared Configuration**: Common settings across modules.

### Module Dependencies

The system tracks and respects module dependencies:

- **Replacement Directives**: Handles `replace` statements in `go.mod`.
- **Build Ordering**: Ensures dependencies build first.
- **Conditional Rules**: Only generates rules when modules exist.

### Coverage Aggregation

Monorepo coverage provides:

- **Per-Module Reports**: Individual coverage for each module.
- **Unified Reporting**: Combined coverage across all modules.
- **Flag Attribution**: Proper Codecov flag assignment.
- **Path Mapping**: Accurate coverage attribution.

## Usage Examples

### Single Project

```bash
# Complete build and test
make all

# Run tests with coverage
make coverage

# Format and lint code
make tidy

# Update dependencies
make up
```

### Monorepo Operations

```bash
# Build specific module
make build-core

# Test all modules
make test

# Lint specific module
make tidy-resolver

# Update all modules
make up
```

### Development Workflow

```bash
# Before committing
make tidy

# Run tests
make test

# Run race detection
make race

# Generate coverage report with clean state
make clean-coverage coverage

# Check grammar (optional)
make check-grammar
```

## Customisation

### Environment Variables

- **`GO`**: Go command (default: `go`).
- **`GOTEST_FLAGS`**: Additional test flags.
- **`GOUP_FLAGS`**: Flags for dependency updates.
- **`COVERAGE_HTML`**: Generate HTML coverage reports.
- **`JQ`**: JSON processor command.

### Tool Overrides

Individual tools can be overridden:

- **`MARKDOWNLINT`**: Markdown linter command.
- **`CSPELL`**: Spell checker command.
- **`LANGUAGETOOL`**: Grammar checker command.
- **`SHELLCHECK`**: Shell linter command.

### Configuration Customisation

Each project can customise:

- **`internal/build/cspell.json`**: Project-specific dictionary.
- **`internal/build/revive.toml`**: Additional linting rules.
- **`internal/build/markdownlint.json`**: Markdown style rules.
- **`internal/build/languagetool.cfg`**: Grammar checking rules.

## Integration Points

### IDE Integration

The build system integrates with development environments:

- **EditorConfig**: Automatic formatting in IDEs.
- **golangci-lint**: VS Code and GoLand integration.
- **Coverage Reports**: IDE coverage display.
- **VS Code Settings**: Shared configuration via symlinks.

### CI/CD Platforms

GitHub Actions workflows provide:

- **Multi-version Testing**: Go 1.23 and 1.24.
- **Coverage Reporting**: Automatic Codecov uploads.
- **Dependency Management**: Renovate integration.
- **Branch Protection**: WIP branch exclusion.

### External Services

- **Codecov**: Coverage tracking and PR comments.
- **Renovate**: Automated dependency updates.
- **DeepSource**: Additional static analysis (where configured).

## Best Practices

### Workflow

1. **Always run `make tidy`** before committing.
2. **Use `GOTEST_FLAGS`** for flexible test execution.
3. **Generate coverage reports** for comprehensive testing.
4. **Check spell and grammar** for documentation changes.

### Project Setup

1. **Copy shared files** to new projects.
2. **Customise `internal/build/`** configurations as needed.
3. **Set up GitHub workflows** for CI/CD.
4. **Configure Codecov tokens** for coverage reporting.
5. **Create VS Code symlinks** for consistent editor experience.

### Monorepo Management

1. **Keep modules independent** where possible.
2. **Use module replacement** for local dependencies.
3. **Test modules individually** and collectively.
4. **Monitor coverage per module** for better visibility.

This build system provides a robust foundation for Go projects of any size,
from single packages to complex monorepos, whilst maintaining consistency and
quality across the darvaza.org ecosystem.
