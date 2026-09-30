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

### 4. Task Tracker (Taiga) Synchronization Protocol
To maintain seamless alignment between the agent's operational plan in `_currenttask/plan.md` and the canonical task tracker (Taiga):

1. **Plan Header to User Story Comment:**
   - Once the implementation plan is prepared and agreed, the plan overview (Goal, Selected Approach, Constraints, Non-goals, Context references, and Step-to-Task mapping) MUST be posted as a comment to the parent User Story in Taiga.
   - The canonical `description` of the User Story in Taiga MUST NOT be overwritten, preserving original acceptance criteria and business requirements intact.

2. **Step Specification to Task Description:**
   - Each fine-grained Taiga Task under the User Story MUST have its `description` populated with the technical specification of the corresponding Step from `_currenttask/plan.md` (Goal, Required reading, Modified files, Dependencies, Expected changes, Acceptance criteria, Checklist of actions).

3. **Step Completion & Evidence in Task:**
   - Upon executing a Step, obtaining verification evidence, and authoring the commit:
     - The corresponding Taiga Task MUST be transitioned to `Closed` status (`is_closed: true`).
     - A comment MUST be posted to the Task recording the completion status, commit hash, commit message, and verifiable evidence summary.

### 5. Off-Track Work & Soft Deviation Protocol
When executing a story, necessary actions may emerge that were not anticipated in the original implementation plan (e.g., standard specification updates, rule catalog additions, tooling adaptations) but meet the criteria for **Soft deviation** under the `executing-plans` skill (they preserve the core decision, contracts, and scope):

1. **Local Operational Artifact (`_currenttask/off-track.md`):**
   - All Soft deviations MUST be documented in `_currenttask/off-track.md` following the template from `executing-plans/references/off-track-template.md`:
     - Timestamp and title;
     - Affected plan step or scope;
     - Deviation description (what was actually done differently);
     - Discovered constraints and rationale;
     - Local solution and actual file modifications;
     - Verification evidence and impact assessment.

2. **Mirroring to Taiga as Off-Track Tasks:**
   - Each distinct logical package of off-track work MUST be created as a dedicated Task under the active User Story in Taiga.
   - **Task Subject:** MUST be prefixed with `[Off-track] ` followed by a concise descriptive title:
     ```text
     [Off-track] <Название выполненного действия / адаптации>
     ```
   - **Task Tag:** MUST include the tag `off-track`.
   - **Task Description:** MUST include the structured off-track specification (Plan reference, Deviation, Discovered constraints, Solution, Rationale, Modified files, Impact).
   - **Task Status & Evidence Comment:**
     - The task is transitioned to `Closed` status (`is_closed: true`).
     - A comment MUST be posted to the task containing completion status, commit hash(es), and verifiable evidence (test outputs, conformance checks).
