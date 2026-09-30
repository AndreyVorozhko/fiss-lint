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
│   ├── cli/                  # CLI flags, help, version parsing and terminal reporting
│   ├── model/                # Core domain types (Issue, Report, Severity, BuildInfo)
│   └── linter/               # Rule evaluation engine and validation pipeline
├── testdata/                 # Test suites and fixtures for CLI and rules
├── Makefile                  # Cross-platform build automation
├── go.mod                    # Module definition
└── FISS/                     # Intellectual space of fiss-lint
```

## Linter Engine & Pipeline Architecture
The validation engine (`internal/linter`) executes a sequential, deterministic pipeline against the target directory:
- **Pipeline Stages:**
  1. `checkRootDir` (`FISS-R001`): Verifies existence and directory status of `FISS/`.
     - *Short-circuit Invariant:* If `FISS/` is missing, the linter reports `FISS-R001` (Error) and immediately halts the pipeline, suppressing cascading errors from downstream file checks.
  2. `checkMandatoryFiles` (`FISS-R002`): Verifies presence and regular file/symlink status of `FISS/INDEX.md` and `FISS/BOOTSTRAP.md` (strictly case-sensitive).
  3. `checkIndexLinks` (`FISS-R003`, `FISS-R011`): Scans `FISS/INDEX.md` for mandatory link to `BOOTSTRAP.md` (`FISS-R003`, Error) and recommended link to the official standard website (`FISS-R011`, Warning).
- **Diagnostics Aggregator:** Findings are recorded in a centralized `model.Report` containing structured `model.Issue` records, with total count of errors and warnings.

## Implemented Rules (Root Structure MVP)
- `FISS-R001` (Error): Project root must contain `FISS/` directory.
- `FISS-R002` (Error): `FISS/` directory must contain `INDEX.md` and `BOOTSTRAP.md`.
- `FISS-R003` (Error): `FISS/INDEX.md` must contain a link targeting `BOOTSTRAP.md`.
- `FISS-R011` (Warning): `FISS/INDEX.md` should contain a reference to `https://fiss.vorozhko.ru`.
*(Subsequent rules FISS-R004–R010, R012–R018 are scheduled for implementation in upcoming stories).*

## CLI Interface & Behavior
- **Invocation:** `fiss-lint [flags] [target_path]`
- **Default path:** Current working directory (`.`).
- **Supported Flags:**
  - `-h`, `--help`: Usage help and rule summary.
  - `-v`, `--version`: Version, commit, and build date.
- **Planned Flags:**
  - `--format [text|json]`: Output format (default `text`).
  - `--strict`: Escalates warnings to errors (exit code `1`).
- **Exit Codes:**
  - `0`: No errors found (or only warnings when `--strict` is not set).
  - `1`: One or more `Error` issues detected (or warnings with `--strict`).
