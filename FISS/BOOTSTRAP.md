# fiss-lint — Baseline Context

## Project Purpose
`fiss-lint` is an autonomous, zero-dependency command-line linter written in Go for the [File-based Intellectual Space Standard (FISS v1.0.0)](https://fiss.vorozhko.ru/v1.0.0/llms.txt).

The core mission of `fiss-lint`:
> `fiss-lint` performs strict mechanical verification of those FISS requirements for which an unambiguous, deterministic verification criterion exists.

It serves as a fast, deterministic gate for developers (CLI), CI/CD pipelines, pre-commit hooks, and AI agents.

## Core Invariants & Division of Responsibility
- **Mechanical vs. Semantic:** `fiss-lint` verifies structural, syntactic, and link-integrity invariants deterministically. Natural language comprehension, qualitative assessment, and semantic evaluation are out of scope and handled exclusively by the AI agent skill `fiss-validate`.
- **Zero Runtime Dependencies:** Compiles into a single standalone binary without CGO and without external runtime dependencies for Linux (`amd64`, `arm64`), macOS (`amd64`, `arm64`), and Windows (`amd64`, `arm64`).
- **Read-Only Audit:** `fiss-lint` performs read-only checks with standard exit codes (`0` = clean, `1` = errors found). Modes like `--fix` and `init` are intentionally excluded.

## Canonical Sources
- **Task & Issue Tracking:** Canonical task management is hosted in self-hosted [Taiga.io](https://taiga.io/) (Kanban mode). FISS does not duplicate the task tracker.
- **Operational Task Artifacts:** Temporary operational artifacts for the active task reside in `_currenttask/` in the project root.

## Task Transition Gate
Work on any subsequent task MAY begin only when the preceding FISS synchronization state in `FISS/state/fiss-handoff.md` is `synchronized`. The states `pending` and `unresolved` strictly block the transition to the next task.
