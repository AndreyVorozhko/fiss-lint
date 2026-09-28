# System Architecture & Toolchain

## Technology Stack
- **Language:** Go (1.22+).
- **Compilation:** Pure Go without CGO (`CGO_ENABLED=0`).
- **Dependencies:** Standard library preferred. Zero external runtime dependencies.
- **Output:** Statically linked single standalone binary.

## Cross-Platform Targets
The toolchain (`Makefile`) must cross-compile binaries for:
- **Linux:** `linux/amd64`, `linux/arm64` (`bin/fiss-lint-linux-amd64`, `bin/fiss-lint-linux-arm64`)
- **macOS:** `darwin/amd64`, `darwin/arm64` (`bin/fiss-lint-darwin-amd64`, `bin/fiss-lint-darwin-arm64`)
- **Windows:** `windows/amd64`, `windows/arm64` (`bin/fiss-lint-windows-amd64.exe`, `bin/fiss-lint-windows-arm64.exe`)

## Build Configuration & Variables
Runtime configuration files (`.env`) are not required. Build metadata is injected via linker flags (`-ldflags`):
- `VERSION`: Semantic release version.
- `COMMIT`: Git commit SHA.
- `BUILD_DATE`: Build timestamp in UTC.

## Project Structure
```text
fiss-lint/
├── cmd/
│   └── fiss-lint/
│       └── main.go           # Application entry point
├── internal/
│   ├── cli/                  # CLI flags, help, version parsing
│   ├── model/                # Core domain types (Issue, Report, Severity)
│   ├── parser/               # Markdown list scanner and Read when detector
│   ├── linter/               # Rule evaluation engine
│   └── reporter/             # Text (color) and JSON output formatters
├── Makefile                  # Cross-platform build automation
├── go.mod                    # Module definition
└── FISS/                     # Intellectual space of fiss-lint
```

## CLI Interface & Behavior
- **Invocation:** `fiss-lint [flags] [target_path]`
- **Default path:** Current working directory (`.`).
- **Flags:**
  - `-h`, `--help`: Usage help and rule summary.
  - `-v`, `--version`: Version, commit, and build date.
  - `--format [text|json]`: Output format (default `text`).
  - `--strict`: Escalates warnings to errors (exit code `1`).
- **Exit Codes:**
  - `0`: No errors found (or only warnings when `--strict` is not set).
  - `1`: One or more `Error` issues detected (or warnings with `--strict`).
