# Development & Git Workflow

This document defines the conventions for Git version control, task traceability, and the task lifecycle within the `fiss-lint` project.

## Canonical Traceability
- **Canonical Task Tracker:** Self-hosted [Taiga.io](https://taiga.io/) instance is the authoritative source for user stories, tasks, and kanban state.
- **Traceability Link:** Every branch MUST reference a Taiga user story; every commit MUST reference a specific Taiga task.

## Branch Naming Conventions
Branches MUST follow the structured format:
```text
<type>/story-<id>/<slug>
```

- `<type>`: Category of the change:
  - `feature`: Implementation of new capabilities, components, or user stories.
  - `fix` or `bugfix`: Bug fixes and defect corrections.
  - `refactor`: Structural refactoring without behavioral changes.
  - `chore`: Maintenance, repository setup, or housekeeping.
- `story-<id>`: Literal prefix `story-` followed by the numeric ID (ref) of the user story in Taiga.
- `<slug>`: Short, descriptive English name in lowercase kebab-case identifying the story scope.

### Example
For Taiga User Story `#2` ("Инфраструктура проекта, CLI-каркас и кроссплатформенная сборка"):
```text
feature/story-2/infra
```

## Commit Naming Conventions
Commits MUST adhere to [Conventional Commits](https://www.conventionalcommits.org/) extended with a Taiga task reference:
```text
<type>: <summary> . T-<task_id>
```
Or optionally with scope:
```text
<type>(<scope>): <summary> . T-<task_id>
```

### Components
1. `<type>`: Standard Conventional Commit type:
   - `feat`: New feature or functionality.
   - `fix`: Bug fix.
   - `docs`: Documentation changes.
   - `style`: Formatting, missing semicolons, whitespace (no code change).
   - `refactor`: Code change that neither fixes a bug nor adds a feature.
   - `perf`: Performance improvements.
   - `test`: Adding or modifying tests.
   - `build`: Changes affecting the build system or external dependencies.
   - `ci`: CI configuration changes.
   - `chore`: Other changes that don't modify source or test files.
2. `<scope>` (optional): Subsystem or package affected (e.g., `cli`, `linter`, `model`).
3. `<summary>`: Concise explanation in Russian (using past tense indicative or concise phrasing):
   - Clear statement of what was done and why.
4. ` . T-<task_id>`: Required task reference suffix:
   - Delimiter: Space, dot, space (` . `).
   - Task tag: `T-` (Latin uppercase `T` followed by hyphen) and the numeric ID (ref) of the task in Taiga.

### Examples
- For Taiga Task `#9` ("Инициализация go.mod и базовой структуры пакетов"):
  ```text
  feat: инициализирован go.mod и базовая инфраструктура пакетов . T-9
  ```
- For Taiga Task `#10` ("Реализация CLI-парсера флагов `--help` и `--version` в `internal/cli`"):
  ```text
  feat(cli): реализован парсер флагов --help и --version . T-10
  ```

## Task Lifecycle

Every user story and its constituent tasks follow a strict seven-stage lifecycle:

1. **Branching (Ветвление):**
   - Create a feature branch from `main`: `<type>/story-<id>/<slug>` based on the Taiga User Story.
   - Example: `feature/story-2/infra`.

2. **Planning & Architecture (Планирование и архитектурный дизайн):**
   - Translate user story requirements and Taiga tasks into an implementation plan in `_currenttask/plan.md` using `writing-plans`.
   - Adhere to `FISS/overrides/planning.md`: 1 Step = 1 Taiga Task with explicit `T-<task_id>` tag and mandatory step-level commit checkpoint.
   - For module and package boundaries, apply `api-and-interface-design` and `software-architecture` (Clean Architecture, deep modules, Hyrum's Law).
   - If non-trivial architectural trade-offs arise, stress-test them with `doubt-driven-development`.

3. **Execution & Coding Standards (Исполнение и разработка):**
   - Sequentially execute plan steps using `executing-plans` (*Action → Evidence → Validation → Commit Checkpoint → User Confirmation*).
   - **Go Coding Rules:** Strictly follow `.agents/rules/go.md` and `golang-pro` (Effective Go, happy path left-aligned, error wrapping with `%w`, zero external runtime dependencies, no duplicate `package`).
   - **Markdown Standards:** When implementing or validating Markdown parsing, follow `.agents/rules/markdown.md` (CommonMark 0.31.2, fenced code block isolation).
   - **Test-Driven Development:** Apply `test-driven-development` (TDD, red-green-refactor loop) for rule evaluation and parsing logic.
   - **Systematic Debugging:** If unexpected build or test failures occur, investigate root causes using `systematic-debugging` before proposing fixes.

4. **Commit & Verification Gate (Верификация и фиксация изменений):**
   - Apply `verification-before-completion`: no claims of completion or commits without fresh, observable evidence in terminal output (`go test`, `make build-all`).
   - Invoke `git-commit` with message format: `<type>: <summary> . T-<task_id>`.
   - Never accumulate uncommitted work across multiple Taiga tasks.

5. **Agent Self-Review (Саморевью агентом):**
   - Perform comprehensive technical verification before human handover using `project-review`.
   - Use `code-review-and-quality` for multi-axis review (correctness, readability, architecture, security, performance).
   - Apply `review-and-simplify-changes` to review the git diff, eliminate unnecessary complexity, and ensure code reuse.
   - Apply `improve-codebase-architecture` to verify module depth and run deletion tests.
   - Resolve all detected issues prior to human escalation.

6. **Human Review Surface & Feedback (Человеческая приёмка):**
   - Invoke `human-review-surface` to present a concise, evidence-backed surface for human inspection.
   - When human feedback is received, apply `receiving-code-review` (rigorous technical validation and test verification, avoiding performative agreement).

7. **Task Transition & Handoff (Закрытие и шлюз перехода):**
   - Record the completed story in `FISS/state/fiss-handoff.md` using `fiss-maintain`.
   - Transition state to `fiss synchronization: synchronized`.
   - Satisfies the transition gate in `FISS/BOOTSTRAP.md`, allowing the next backlog story to begin.
