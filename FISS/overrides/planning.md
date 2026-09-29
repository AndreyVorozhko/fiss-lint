# Task Planning & Commit Traceability Overrides

This document defines project-specific overrides adapting the behavior of the `writing-plans` skill for the `fiss-lint` project.

## Context & Rationale
In `fiss-lint`, canonical tasks in Taiga are fine-grained (atomic work packages). To prevent confusion between task numbers and to ensure clean, bisectable Git history, implementation plans must tightly couple each Taiga task to a plan step and mandate an explicit commit checkpoint.

## Override Rules for Implementation Planning

### 1. Step-to-Task 1:1 Mapping
- When planning work decomposed into fine-grained Taiga tasks, each **Шаг (Step)** in `_currenttask/plan.md` MUST map to exactly one Taiga task.
- The step header MUST explicitly include the Taiga task identifier in the format:
  ```markdown
  ### Шаг N (T-<task_id>): <Название задачи из Taiga>
  ```
  *Example:* `### Шаг 1 (T-9): Инициализация go.mod и базовой структуры пакетов`
- If a Taiga task requires multiple steps due to complexity, subsequent steps MUST be numbered hierarchically (e.g., `Шаг 1.1 (T-9)`, `Шаг 1.2 (T-9)`) to preserve unambiguous task identity.

### 2. Mandatory Commit Checkpoint at Step Level
- Commits are placed at the **уровень Шага (Step level)** as the concluding action of that step.
- Every Step MUST define its final action as an explicit Git commit checkpoint:
  ```markdown
  - **Действие N (Коммит):** Зафиксировать изменения через `git-commit` с сообщением: `<type>: <краткое описание> . T-<task_id>`
  ```
- The commit message in the plan MUST strictly copy the `T-<task_id>` from the step header.

### 3. Precondition: Evidence Before Commit
- The commit action MUST NOT be executed until the Step's **Свидетельство (Evidence)** has been obtained and verified (e.g., successful build, passing unit tests, verified CLI output).
- Committing incomplete, unverified, or broken intermediate states across multiple actions within a step is prohibited.
