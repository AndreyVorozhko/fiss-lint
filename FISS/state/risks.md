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
File casing test cases in `internal/linter/root_test.go` and `internal/linter/integrity_test.go` demonstrated the necessity of checking exact file entry names by scanning directory entries (`os.ReadDir`). In Story #5, `checkPathExistsCaseSensitive` was implemented in `internal/linter/integrity.go` using component-by-component `os.ReadDir` traversal, successfully eliminating case-insensitive false negatives for all relative link targets and bare directory checks across Windows NTFS, macOS APFS, and Linux.

Mitigation:
Mitigate. Validate mandatory files and all relative link paths against directory entry lists (`os.ReadDir`) using exact string equality, resolving symlinks via `os.Stat` only when needed, and normalizing path targets using forward slashes `/`.

Trigger:
Execution of cross-platform CI test suites on Windows/macOS runners.

Review Signal:
Closed in Story #5 for relative link targets. Retain as Active for upcoming global discovery in Story #6.

---

### [RISK-003] Symlink Loops & Filesystem Traversal Boundary Escapes

Status: Active

Risk:
Infinite recursion or unintended scanning outside the target repository boundary due to cyclic or external symbolic links.

Condition / Cause:
Starting from Story #6, the linter performs recursive filesystem discovery of all `INDEX.md` files and validates area reachability. Symbolic links may target parent directories or external locations (e.g., skill farm links in `.agents/skills/`).

Impact:
Infinite traversal loops, memory/stack exhaustion, or false positive diagnostics on external files outside repository scope.

Context / Evidence:
In the `fiss-lint` repository itself, `.agents/skills/` contains symlinks targeting external repository paths.

Mitigation:
Mitigate. Track visited filesystem nodes (`device + inode`), restrict traversal strictly within target repository root boundaries, and ignore non-FISS hidden metadata directories.

Trigger:
Development of recursive directory scanner in Story #6.

Review Signal:
Architectural decision on directory traversal algorithm in Story #6.
