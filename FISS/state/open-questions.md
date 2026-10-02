# Open Questions

Canonical registry of unresolved technical uncertainties for the `fiss-lint` project.

## Active Uncertainties

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

### [OQ-002] Machine-Readable Output Schema (--format json)

Исход:
Resolved

Закрыто через:
Реализация плоской схемы на базе `model.Report` в Story #8 (`internal/cli/reporter.go`, коммиты `cc81ea3`, `4b09fb2`). Вывод включает массив `issues` (`rule_id`, `severity`, `file`, `line`, `message`) и объект `summary` (`errors`, `warnings`, `total`). Пустой отчет сериализуется с пустым срезом `[]`.

---

### [OQ-001] Symlink Handling Strategy for External and Circular Links

Исход:
Resolved

Закрыто через:
Реализация алгоритма BFS обхода графа навигации в Story #6 (`internal/linter/topology.go`, коммиты `f655bd2`, `fd33b46`), использующая каноническое разрешение путей через `filepath.EvalSymlinks` (`visitedRealPaths`) и нормализацию относительных индексов (`visitedIndexes`). Внешние ссылки за пределы корня репозитория не включаются в граф FISS.

