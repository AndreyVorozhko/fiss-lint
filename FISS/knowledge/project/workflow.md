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

2. **Planning (Планирование):**
   - Translate user story requirements and Taiga tasks into an implementation plan in `_currenttask/plan.md` using `writing-plans`.
   - Each fine-grained Taiga task corresponds to a dedicated Step in the plan with an explicit `T-<task_id>` tag (per project overrides).
   - Every Step defines actions, expected evidence, and an explicit Git commit checkpoint.

3. **Execution (Исполнение):**
   - Sequentially execute plan steps using `executing-plans`.
   - Follow the loop: *Action → Evidence → Validation → Commit Checkpoint → User Confirmation*.
   - Never skip steps or accumulate uncommitted work across multiple Taiga tasks.

4. **Commit & Traceability (Фиксация изменений):**
   - Once step evidence is verified, invoke `git-commit`.
   - Message format: `<type>: <summary> . T-<task_id>`.
   - Verifies worktree cleanliness and precise task attribution.

5. **Agent Self-Review (Саморевью агентом):**
   - Before handing over work to the human reviewer, perform comprehensive technical verification using the `project-review` skill.
   - Audits the entire branch/diff against correctness (`Logic Correctness`), project architecture and invariants (`Architecture & Invariants`), code quality, and build/dependency integrity (`Supply Chain & Build Integrity`).
   - Ensures all identified defects or coverage gaps are resolved before escalating to human attention.

6. **Human Review Surface (Человеческая приёмка):**
   - After passing all automated checks and the agent self-review, invoke `human-review-surface`.
   - Isolates non-automated properties and architectural choices requiring human judgment from already-proven mechanics.

7. **Task Transition & Handoff (Закрытие и шлюз перехода):**
   - Use `fiss-maintain` to record the completed story in `FISS/state/fiss-handoff.md`.
   - Transition state to `fiss synchronization: synchronized`.
   - Fulfills the task transition gate in `FISS/BOOTSTRAP.md` allowing work on the next user story to begin.
