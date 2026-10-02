# Project Risk Register

Canonical registry of identified risks for the `fiss-lint` project.

## Technical & Architectural Risks

### [RISK-001] Drift of External FISS v1.0.0 Standard Specification

Status: Active

Risk:
Desynchronization between the linter's mechanical checks and the normative FISS v1.0.0 specification or Conformance Checklist caused by stale user story statements.

Condition / Cause:
Backlog task statements were drafted early during project inception and may contain discrepancies with the approved FISS v1.0.0 specification (e.g., assuming localization of the `Read when:` marker is allowed).

Impact:
Implementation of incorrect or non-standard checks, false positive or false negative findings for developers and AI agents, and loss of confidence in the linter.

Context / Evidence:
Audit of Story #4 revealed that Acceptance Criterion #4 and Task 101 stipulated support for a localized `Читать, когда:` marker, which is explicitly forbidden by line 155 of the FISS v1.0.0 specification.

Mitigation:
Mitigate. Mandated a pre-planning verification of task statements against the canonical specification at `https://fiss.vorozhko.ru/v1.0.0/llms.txt` before formulating implementation plans (codified in `FISS/overrides/planning.md`).

Trigger:
Discrepancy detected between the user story text and the normative standard specification.

Review Signal:
Release of updates to the FISS standard specification or the Conformance Checklist.

---

### [RISK-002] Cross-Platform Filesystem Semantic Discrepancies (Case Sensitivity & Path Separators)

Status: Active

Risk:
Divergent linter behavior across operating systems: overlooking file case violations on Windows and macOS while failing on Linux.

Condition / Cause:
NTFS (Windows) and APFS (macOS) filesystems are case-insensitive by default, whereas FISS requires strict case (`INDEX.md`, `BOOTSTRAP.md`) and forward slashes `/` in path references. Standard system call `os.Stat` on Windows matches `index.md` even when querying `INDEX.md`.

Impact:
A repository with non-conforming file casing passes validation on a developer's Windows workstation but fails in Linux CI/CD environments or for other users.

Context / Evidence:
File casing test cases in `internal/linter/root_test.go` and `internal/linter/integrity_test.go` demonstrated the necessity of checking exact file entry names by scanning directory entries (`os.ReadDir`). In Story #5, `checkPathExistsCaseSensitive` was implemented in `internal/linter/integrity.go` using component-by-component `os.ReadDir` traversal, successfully eliminating case-insensitive false negatives for all relative link targets and bare directory checks across Windows NTFS, macOS APFS, and Linux. In Story #7, `findAgentInstructionFiles` in `internal/linter/agents.go` was likewise implemented via `os.ReadDir` exact entry matching, ensuring lowercase files like `agents.md` do not inadvertently satisfy or trigger agent instruction checks on case-insensitive filesystems.

Mitigation:
Mitigate. Validate mandatory files, relative link paths, and agent instruction files against directory entry lists (`os.ReadDir`) using exact string equality, resolving symlinks via `os.Stat` only when needed, and normalizing path targets using forward slashes `/`.

Trigger:
Execution of cross-platform CI test suites on Windows/macOS runners.

Review Signal:
Closed in Story #5 for relative link targets, and in Story #7 for root agent instruction files. Retain as Active for upcoming partition rules in Story #8.

---

### [RISK-003] Symlink Loops & Filesystem Traversal Boundary Escapes

Status: Resolved

Risk:
Infinite recursion or unintended scanning outside the target repository boundary due to cyclic or external symbolic links.

Condition / Cause:
Starting from Story #6, the linter performs recursive filesystem discovery of all `INDEX.md` files and validates area reachability. Symbolic links may target parent directories or external locations (e.g., skill farm links in `.agents/skills/`).

Impact:
Infinite traversal loops, memory/stack exhaustion, or false positive diagnostics on external files outside repository scope.

Context / Evidence:
In the `fiss-lint` repository itself, `.agents/skills/` contains symlinks targeting external repository paths.

Mitigation:
Mitigate. Track visited filesystem nodes and indexes, resolve symlinks to canonical paths via `filepath.EvalSymlinks`, restrict traversal strictly within target repository root boundaries, and ignore non-FISS hidden metadata directories.

Trigger:
Development of recursive directory scanner in Story #6.

Review Signal:
Closed in Story #6 with BFS cycle and symlink loop protection (`visitedIndexes` and `visitedRealPaths` via `filepath.EvalSymlinks`).

Resolution:
In `internal/linter/topology.go`, implemented BFS navigation traversal with dual cycle protection: relative path index set `visitedIndexes` and canonical path set `visitedRealPaths` resolved via `filepath.EvalSymlinks`.

Evidence:
Full test suite passing (`TestLinter_Lint_Topology_Valid`, `TestLinter_Lint_Topology_CyclicIndexes`), and self-audit on `fiss-lint` repository (containing symlinks in `.agents/skills/`) executing in <15ms without cycles or traversal leaks.

Residual risk:
Negligible. Broken symlinks are caught safely by `filepath.EvalSymlinks` and target directory boundary checks.

---

### [RISK-004] Agent Runtime Environment Missing `fiss-lint` Binary

Status: Mitigated

Risk:
Agent execution halts or reports false failures if `fiss-lint` binary is missing from PATH in restricted or minimal CI/agent container environments.

Condition / Cause:
Agent runtime environments (e.g., Docker containers, restricted sandbox environments) may not have `fiss-lint` preinstalled in system PATH, and may lack root privileges, network access, or a Go compiler to execute `go install`.

Impact:
Skills cannot run mechanical validation, potentially leading agents into infinite retry loops or failing task handoffs.

Context / Evidence:
Identified during Story #43 architecture design for `fiss-validate` and `fiss-maintain` integration.

Mitigation:
Mitigate. Standardized a 4-step autonomous discovery ladder in both skills:
1. Check `command -v fiss-lint`;
2. Check local workspace build paths (`./bin/fiss-lint`, `/workspace/bin/fiss-lint`);
3. Attempt `go install` if Go toolchain is available;
4. Gracefully downgrade to `INSUFFICIENT_EVIDENCE` for Level 1 checks with actionable installation guidance, while proceeding with Level 2 semantic analysis without crashing.

Trigger:
Invocation of `fiss-validate` or `fiss-maintain` in an environment without `fiss-lint` in PATH.

Review Signal:
End-to-end verification in Story #43 (T-48) confirming graceful fallback and local binary detection.

Resolution:
Codified in `Establish` section of both `fiss-validate/SKILL.md` and `fiss-maintain/SKILL.md`, documented in `GOTCHAS.md` (Gotcha 10) and `DESIGN.md`.

Residual risk:
Low. If no binary can be run or built, semantic evaluation continues while mechanical status is reported as `INSUFFICIENT_EVIDENCE`.
