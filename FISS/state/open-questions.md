# Open Questions

Canonical registry of unresolved technical uncertainties for the `fiss-lint` project.

## Active Uncertainties

### [OQ-001] Symlink Handling Strategy for External and Circular Links

Status:
Active

Unresolved:
How the linter should handle symbolic links during recursive directory traversal (starting with Story #6):
1. Should symlinks pointing outside the project root be ignored?
2. Should external links be treated as link integrity violations?
3. Which cycle detection mechanism should be adopted (canonical path resolution via `filepath.EvalSymlinks` or device/inode tracking via `os.SameFile`)?

Why it matters:
In repositories utilizing symlink farms (e.g., `.agents/skills/` or monorepos), improper symlink traversal will cause infinite loops or erroneous findings on third-party files.

Origin:
Filesystem traversal architecture discussion and risk analysis during Story #3 / Story #6.

Revisit if:
Design and implementation of recursive area validation begins in Story #6 (`FISS-R007`, `FISS-R008`).

---

### [OQ-002] Machine-Readable Output Schema (--format json)

Status:
Active

Unresolved:
Which JSON output schema should be supported by `--format json`:
1. Proprietary flat schema based on `model.Report` (`issues: [...]`, `summary: {...}`);
2. Industry-standard SARIF (Static Analysis Results Interchange Format) for native integration with GitHub Code Scanning and GitLab SAST;
3. Two-tier support (`--format json` and `--format sarif`).

Why it matters:
Defines the integration contract for `fiss-lint` in CI/CD pipelines and external tooling consumption.

Origin:
System architecture document `FISS/knowledge/project/architecture.md` (planned flags section).

Revisit if:
Specification and implementation of structured output reporters (Story #9).

---

### [OQ-003] Diagnostic Error Message Localization

Status:
Active

Unresolved:
Whether terminal diagnostic messages should remain strictly in English (following the uniform format `[ERROR] [FISS-R001] FISS/ directory missing`) or support localization (e.g., Russian) via a `--lang` flag or `LANG` environment variable.

Why it matters:
Affects deterministic output parsing in automation scripts, developer readability across regions, and error catalog package architecture.

Origin:
User experience evaluation and CLI output format requirements.

Revisit if:
Specification and implementation of CLI text reporter and error message catalog.
