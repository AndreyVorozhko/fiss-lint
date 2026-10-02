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
│   └── linter/               # Rule evaluation engine, file checks, and nav parsing
│       ├── integrity.go      # Link integrity, case-sensitive FS checks, bare directories (FISS-R006)
│       ├── linter.go         # Core engine and sequential pipeline orchestration
│       ├── links.go          # Root index link checks (FISS-R003, FISS-R011)
│       ├── nav.go            # Nav parsing & Read when format checks (FISS-R004, FISS-R005)
│       └── topology.go       # Navigation graph traversal, orphan reachability, composite areas & overrides (FISS-R007, FISS-R008, FISS-R009)
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
  4. `validateAllIndexes` (`FISS-R004`, `FISS-R005`, `FISS-R006`): Recursively traverses `FISS/` discovering all `INDEX.md` files via `filepath.WalkDir`.
     - **Navigation Syntax:** Uses `bufio.Scanner` to parse two-line navigation entries (`- [Title](target)\n  Read when: condition`) without regular expressions. Enforces exact two-space indentation, non-empty condition text, and the strict canonical English marker `Read when: ` without localization (`FISS-R005`, Error). In root `FISS/INDEX.md`, additionally verifies that the link targeting `BOOTSTRAP.md` is accompanied by an attached valid `Read when` condition (`FISS-R004`, Error).
     - **Link Integrity:** Calls `validateIndexLinksIntegrity` for each index (`FISS-R006`, Error). Resolves relative links relative to the containing directory of each `INDEX.md`, strips `#anchor` and `?query` fragments, excludes external URLs (`http://`, `https://`, `mailto:`, `ftp://`), enforces case-sensitive filesystem existence checks component-by-component via `os.ReadDir` (`checkPathExistsCaseSensitive`, resolving `[RISK-002]`), mandates `.md` extension on file targets, and requires `INDEX.md` within targeted directories, strictly prohibiting bare directories.
  5. `validateTopology` (`FISS-R007`, `FISS-R008`, `FISS-R009`): Executes navigation graph traversal and reachability checks via `internal/linter/topology.go`.
     - **Navigation Graph Traversal:** Starting from `FISS/INDEX.md`, traverses all child indexes using BFS. Prevents infinite recursion from cyclic links or symlink loops by tracking visited relative indexes (`visitedIndexes`) and physical directories via `filepath.EvalSymlinks` (`visitedRealPaths`, resolving `[RISK-003]`).
     - **Reachability & Orphan Detection (`FISS-R008`, Error):** Collects all physical `.md` files in `FISS/` via `filepath.WalkDir` (excluding root `FISS/INDEX.md` and `FISS/BOOTSTRAP.md` governed by `FISS-R002`) and reports any unreachable files as orphans.
     - **Composite Area Validation (`FISS-R007`, Error):** Verifies that all directories referenced as navigation targets contain their own `INDEX.md`.
     - **Overrides Validation (`FISS-R009`, Error):** If `FISS/overrides/` directory exists, verifies that it contains `INDEX.md`, root `FISS/INDEX.md` links to it, and the link has an attached read condition requiring review before skill usage (case-insensitively containing `skill`).
- **Diagnostics Aggregator:** Findings are recorded in a centralized `model.Report` containing structured `model.Issue` records, with total count of errors and warnings.

## Implemented Rules (Stories #3, #4, #5 & #6)
- `FISS-R001` (Error): Project root must contain `FISS/` directory.
- `FISS-R002` (Error): `FISS/` directory must contain `INDEX.md` and `BOOTSTRAP.md`.
- `FISS-R003` (Error): `FISS/INDEX.md` must contain a link targeting `BOOTSTRAP.md`.
- `FISS-R004` (Error): `FISS/INDEX.md` link to `BOOTSTRAP.md` must have an attached read condition requiring reading before project work.
- `FISS-R005` (Error): Every navigation entry in any `INDEX.md` must follow the two-line format: `- [Title](path)` followed by `  Read when: <condition>` (exact non-localized marker, 2 spaces indentation, non-empty condition text).
- `FISS-R006` (Error): Every relative link in an index must resolve to an existing physical `.md` file or `INDEX.md` of a composite area. Links to bare directories without `INDEX.md` are prohibited.
- `FISS-R007` (Error): A composite area referenced in navigation MUST contain its own `INDEX.md`.
- `FISS-R008` (Error): Every `.md` file inside `FISS/` (except root index and bootstrap) MUST be reachable from `FISS/INDEX.md` through indexes. Unreachable files (orphans) are prohibited.
- `FISS-R009` (Error): If `FISS/overrides/` exists, it MUST contain `INDEX.md`, root `FISS/INDEX.md` MUST link to it, and the link MUST require reading before skill use (`skill`).
- `FISS-R011` (Warning): `FISS/INDEX.md` should contain a reference to `https://fiss.vorozhko.ru`.
*(Subsequent rules FISS-R010, R012–R018 are scheduled for implementation in upcoming stories).*

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
