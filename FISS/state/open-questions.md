# Open Questions

Canonical registry of unresolved technical uncertainties for the `fiss-lint` project.

## Active Uncertainties

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

---

## Closure Notes

### [OQ-001] Symlink Handling Strategy for External and Circular Links

Исход:
Resolved

Закрыто через:
Реализация алгоритма BFS обхода графа навигации в Story #6 (`internal/linter/topology.go`, коммиты `f655bd2`, `fd33b46`), использующая каноническое разрешение путей через `filepath.EvalSymlinks` (`visitedRealPaths`) и нормализацию относительных индексов (`visitedIndexes`). Внешние ссылки за пределы корня репозитория не включаются в граф FISS.
