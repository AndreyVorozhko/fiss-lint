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
- `VERSION`: Semantic release version (SemVer 2.0.0 format `vMAJOR.MINOR.PATCH`, e.g. `v1.0.0`) resolved via `git describe --tags --dirty 2>/dev/null || echo "v1.0.0-dev"`. See `versioning.md`.
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
│       ├── agents.go         # Agent instruction file discovery and FISS-R010 directive validation (AGENTS.md, CLAUDE.md, .cursorrules, QWEN.md)
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
  3. `checkIndexLinks` (`FISS-R003`, `FISS-R011`): Scans `FISS/INDEX.md` for mandatory link to `BOOTSTRAP.md` (`FISS-R003`, Error) and recommended link to the official standard website or official GitHub mirror (`FISS-R011`, Warning).
  4. `validateAllIndexes` (`FISS-R004`, `FISS-R005`, `FISS-R006`): Recursively traverses `FISS/` discovering all `INDEX.md` files via `filepath.WalkDir`.
     - **Navigation Syntax:** Uses `bufio.Scanner` to parse two-line navigation entries (`- [Title](target)\n  Read when: condition`) without regular expressions. Enforces exact two-space indentation, non-empty condition text, and the strict canonical English marker `Read when: ` without localization (`FISS-R005`, Error). In root `FISS/INDEX.md`, additionally verifies that the link targeting `BOOTSTRAP.md` is accompanied by an attached valid `Read when` condition (`FISS-R004`, Error).
     - **Link Integrity:** Calls `validateIndexLinksIntegrity` for each index (`FISS-R006`, Error). Resolves relative links relative to the containing directory of each `INDEX.md`, strips `#anchor` and `?query` fragments, excludes external URLs (`http://`, `https://`, `mailto:`, `ftp://`), enforces case-sensitive filesystem existence checks component-by-component via `os.ReadDir` (`checkPathExistsCaseSensitive`, resolving `[RISK-002]`), mandates `.md` extension on file targets, and requires `INDEX.md` within targeted directories, strictly prohibiting bare directories.
  5. `validateTopology` (`FISS-R007`, `FISS-R008`, `FISS-R009`): Executes navigation graph traversal and reachability checks via `internal/linter/topology.go`.
     - **Navigation Graph Traversal:** Starting from `FISS/INDEX.md`, traverses all child indexes using BFS. Prevents infinite recursion from cyclic links or symlink loops by tracking visited relative indexes (`visitedIndexes`) and physical directories via `filepath.EvalSymlinks` (`visitedRealPaths`, resolving `[RISK-003]`).
     - **Reachability & Orphan Detection (`FISS-R008`, Error):** Collects all physical `.md` files in `FISS/` via `filepath.WalkDir` (excluding root `FISS/INDEX.md` and `FISS/BOOTSTRAP.md` governed by `FISS-R002`) and reports any unreachable files as orphans.
     - **Composite Area Validation (`FISS-R007`, Error):** Verifies that all directories referenced as navigation targets contain their own `INDEX.md`.
     - **Overrides Validation (`FISS-R009`, Error):** If `FISS/overrides/` directory exists, verifies that it contains `INDEX.md`, root `FISS/INDEX.md` links to it, and the link has an attached read condition requiring review before skill usage (case-insensitively containing `skill`).
  6. `checkAgentInstructions` (`FISS-R010`): Scans the project root for known agent instruction files (`AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `QWEN.md`).
     - If no candidate files exist, the check passes cleanly without issues.
     - If any candidate file exists, verifies that it contains a directive or Markdown link targeting `FISS/INDEX.md` (or `FISS/`). If omitted, reports `FISS-R010` (Error).
- **Diagnostics Aggregator:** Findings are recorded in a centralized `model.Report` containing structured `model.Issue` records, with total count of errors and warnings.

## Implemented Rules (Stories #3, #4, #5, #6 & #7)
- `FISS-R001` (Error): Project root must contain `FISS/` directory.
- `FISS-R002` (Error): `FISS/` directory must contain `INDEX.md` and `BOOTSTRAP.md`.
- `FISS-R003` (Error): `FISS/INDEX.md` must contain a link targeting `BOOTSTRAP.md`.
- `FISS-R004` (Error): `FISS/INDEX.md` link to `BOOTSTRAP.md` must have an attached read condition requiring reading before project work.
- `FISS-R005` (Error): Every navigation entry in any `INDEX.md` must follow the two-line format: `- [Title](path)` followed by `  Read when: <condition>` (exact non-localized marker, 2 spaces indentation, non-empty condition text).
- `FISS-R006` (Error): Every relative link in an index must resolve to an existing physical `.md` file or `INDEX.md` of a composite area. Links to bare directories without `INDEX.md` are prohibited.
- `FISS-R007` (Error): A composite area referenced in navigation MUST contain its own `INDEX.md`.
- `FISS-R008` (Error): Every `.md` file inside `FISS/` (except root index and bootstrap) MUST be reachable from `FISS/INDEX.md` through indexes. Unreachable files (orphans) are prohibited.
- `FISS-R009` (Error): If `FISS/overrides/` exists, it MUST contain `INDEX.md`, root `FISS/INDEX.md` MUST link to it, and the link MUST require reading before skill use (`skill`).
- `FISS-R010` (Error): If `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, or `QWEN.md` exists in project root, it MUST direct agents to `FISS/INDEX.md`.
- `FISS-R011` (Warning): `FISS/INDEX.md` should contain a reference to the official standard website (`https://fiss.vorozhko.ru`) or official GitHub mirror (`https://github.com/AndreyVorozhko/fiss`).
*(Subsequent rules FISS-R012–R018 are scheduled for implementation in upcoming stories).*

## CLI Interface & Behavior
- **Invocation:** `fiss-lint [flags] [target_path]`
- **Default path:** Current working directory (`.`).
- **Supported Flags:**
  - `-h`, `--help`: Usage help and rule summary.
  - `-v`, `--version`: Version, commit, and build date.
  - `--format [text|json]`: Output format: `text` (default) or `json`.
  - `--strict`: Escalates warnings to errors (exit code `1`).
- **Reporters Architecture:**
  - `Reporter` interface (`internal/cli/reporter.go`) with `Report(report *model.Report, w io.Writer) error`.
  - `TextReporter`: Line-by-line human-readable terminal output (`[SEVERITY] [RULE_ID] path:line: message`).
  - `JSONReporter`: Indented structured JSON (`issues` array with `rule_id`, `severity`, `file`, `line`, `message`, and `summary` with `version`, `errors`, `warnings`, `total`).
- **Exit Codes:**
  - `0`: Validation passed (no errors, or only warnings when `--strict` is not set).
  - `1`: One or more `Error` issues detected (or warnings when `--strict` is set), or runtime filesystem error.
  - `2`: Invalid CLI flags or option values (e.g., unsupported format).

## Agent Skills Integration Architecture

`fiss-lint` serves as the shared deterministic validation engine for agentic skills (`ai-skills`), establishing a clean tri-partite separation of concerns:

1. **`fiss-lint` (CLI Tool)**:
   - High-speed, compiled binary executing Level 1 deterministic invariants (`FISS-R001` through `FISS-R018`).
   - Zero-token execution, zero hallucination, strict Unix exit codes (`0`, `1`, `2`).
   - Emits structured JSON diagnostic streams via `--format json`.
2. **`fiss-validate` (Agent Skill)**:
   - Read-only cognitive inspector.
   - Autonomously provisions and invokes `fiss-lint --format json .` for mechanical verification.
   - Evaluates Level 2 qualitative heuristics across the 7C principles (situational trigger clarity, absence of content leakage, domain classification, canonical drift, continuous maintenance mechanisms).
   - Combines mechanical and semantic findings into a unified report and structured diagnostic payload.
3. **`fiss-maintain` (Agent Skill)**:
   - Operational mutator and continuity orchestrator.
   - Preserves knowledge continuity across task boundaries, classifies durable outcomes (`Capture here`, `Delegate`, `No persistence`).
   - Enforces the post-mutation verification gate by running `fiss-lint --strict <workspace>` before allowing task synchronization (`fiss synchronization: synchronized`).

### Autonomous Discovery Protocol
Both skills follow a resilient discovery ladder:
1. `command -v fiss-lint` (global PATH).
2. `./bin/fiss-lint` or `/workspace/bin/fiss-lint` (local workspace build).
3. `go install github.com/AndreyVorozhko/fiss-lint/cmd/fiss-lint@latest` (when Go toolchain is available).
4. Graceful fallback to `INSUFFICIENT_EVIDENCE` with installation guidance if automated provisioning is restricted by environment permissions.

