# Task Planning & Commit Traceability Overrides

This document defines project-specific overrides adapting the behavior of the `writing-plans` skill for the `fiss-lint` project.

## Context & Rationale
In `fiss-lint`, canonical tasks in Taiga are fine-grained (atomic work packages). To prevent confusion between task numbers and to ensure clean, bisectable Git history, implementation plans must tightly couple each Taiga task to a plan step and mandate an explicit commit checkpoint.

## Override Rules for Implementation Planning

### 1. Pre-Planning: Verification Against Current FISS Standard
- Before decomposing work or formulating the implementation plan, the agent MUST verify the User Story description, scope, Definition of Done, and referenced validation rules against:
  1. The normative FISS v1.0.0 specification (`https://fiss.vorozhko.ru/v1.0.0/llms.txt`);
  2. The official FISS v1.0.0 Conformance Checklist (`https://fiss.vorozhko.ru/v1.0.0/en/conformance.html`);
  3. The deterministic rules registry in `FISS/knowledge/subject/rules.md`.
- **Purpose:** Ensure that the task statement, rule IDs, severities, and acceptance criteria have not become outdated or drifted due to standard evolution since the user story was originally drafted.
- If discrepancies, obsolete requirements, or missing invariants are identified, the agent MUST explicitly highlight them during pre-analysis and align them with the human before committing to the implementation plan.

### 2. Step-to-Task 1:1 Mapping
- When planning work decomposed into fine-grained Taiga tasks, each **Шаг (Step)** of the implementation plan MUST map to exactly one Taiga task.
- The step header MUST explicitly include the Taiga task identifier in the format:
  ```markdown
  ### Шаг N (T-<task_id>): <Название задачи из Taiga>
  ```
  *Example:* `### Шаг 1 (T-9): Инициализация go.mod и базовой структуры пакетов`
- If a Taiga task requires multiple steps due to complexity, subsequent steps MUST be numbered hierarchically (e.g., `Шаг 1.1 (T-9)`, `Шаг 1.2 (T-9)`) to preserve unambiguous task identity.

### 3. Mandatory Commit Checkpoint at Step Level
- Commits are placed at the **уровень Шага (Step level)** as the concluding action of that step.
- Every Step MUST define its final action as an explicit Git commit checkpoint:
  ```markdown
  - **Действие N (Коммит):** Зафиксировать изменения через `git-commit` с сообщением: `<type>: <краткое описание> . T-<task_id>`
  ```
- The commit message in the plan MUST strictly copy the `T-<task_id>` from the step header.

### 4. Precondition: Evidence Before Commit
- The commit action MUST NOT be executed until the Step's **Свидетельство (Evidence)** has been obtained and verified (e.g., successful build, passing unit tests, verified CLI output).
- Committing incomplete, unverified, or broken intermediate states across multiple actions within a step is prohibited.

### 5. Task Tracker (Taiga) Synchronization Protocol
To maintain seamless alignment between the agent's operational plan and the canonical task tracker (Taiga):

1. **Plan Header to User Story Comment:**
   - Once the implementation plan is prepared and agreed, the plan overview (Goal, Selected Approach, Constraints, Non-goals, Context references, and Step-to-Task mapping) MUST be posted as a comment to the parent User Story in Taiga.
   - The canonical `description` of the User Story in Taiga MUST NOT be overwritten, preserving original acceptance criteria and business requirements intact.

2. **Step Specification to Task Description:**
   - Each fine-grained Taiga Task under the User Story MUST have its `description` populated with the technical specification of the corresponding Step of the implementation plan (Goal, Required reading, Modified files, Dependencies, Expected changes, Acceptance criteria, Checklist of actions).

3. **Step Completion & Evidence in Task:**
   - Upon executing a Step, obtaining verification evidence, and authoring the commit:
     - The corresponding Taiga Task MUST be transitioned to `Closed` status (`is_closed: true`).
     - A comment MUST be posted to the Task recording the completion status, commit hash, commit message, and verifiable evidence summary.

4. **User Story Kanban State Transitions:**
   - **Active Development (Agent):** On taking the story into work (Stages 1–2), transition the User Story to `In progress` (`status: 23`) and set the assignee.
   - **Review Gate (Agent):** Upon completing all constituent tasks, passing verification suite, and completing agent self-review (Stage 5), transition the User Story to `Ready for test` (`status: 24`).
   - **Rework Loop (Agent):** If human feedback during Stage 6 requests revisions, transition the User Story back to `In progress`, perform fixes, and return to `Ready for test` upon obtaining fresh evidence.
   - **Story Acceptance & Archival (Human Only):** The agent MUST NOT transition stories to `Done` or `Archived`. The transition to `Done` (`status: 25`, `is_closed: true`) and subsequent archival is executed exclusively by the human after independent verification and merging the feature branch into `main`.

5. **AI Model Attribution:**
   - In all completion comments posted to Taiga Tasks and User Stories, the agent MUST explicitly record the AI model used in the format:
     ```markdown
     **Модель:** <model_identifier>
     ```
     *(Example: `**Модель:** Gemini 3.8 Flash`)*.
   - Each Task and User Story MUST be tagged with the model tag (e.g., `gemini-3.8-flash`) to enable kanban board filtering and model performance tracking.
   - When the `AI Model` custom attribute is defined in the Taiga project settings, the agent MUST also populate this attribute via the API.

### 6. Off-Track Work & Soft Deviation Protocol
When executing a story, necessary actions may emerge that were not anticipated in the original implementation plan (e.g., standard specification updates, rule catalog additions, tooling adaptations) but meet the criteria for **Soft deviation** under the `executing-plans` skill (they preserve the core decision, contracts, and scope):

1. **Documentation of Off-Track Deviations:**
   - All Soft deviations MUST be documented following the template from `executing-plans/references/off-track-template.md`:
     - Timestamp and title;
     - Affected plan step or scope;
     - Deviation description (what was actually done differently);
     - Discovered constraints and rationale;
     - Local solution and actual file modifications;
     - Verification evidence and impact assessment.
   - The deviation specification MUST be saved to Taiga (as an off-track task or comment) and MUST NOT be linked using ephemeral scratch directories.

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

### 7. Preservation of Operational Artifacts in Taiga
- If an operational artifact exists or is created during work (e.g., `intent-map`, `statement`, `pre-analysis-report`, `specification`, `plan`, `off-track`, `review-rules`), it MUST be preserved in Taiga in a convenient form — as a comment or attachment to the corresponding User Story or Task.
- It is strictly prohibited to reference `_currenttask/` anywhere in project documentation, FISS spaces, handoff artifacts, or commit messages, as `_currenttask/` contains only transient working files.

### 8. Language Policy for Taiga and Operational Artifacts
- **Operational Artifacts:** All operational task artifacts (`intent-map`, `statement`, `pre-analysis-report`, `specification`, `plan`, `off-track`, `review-rules`) MUST be authored in Russian.
- **Taiga Management:** All records and communication in Taiga (User Story subjects, descriptions, tags, Task subjects, descriptions, completion comments, and discussions) MUST be maintained exclusively in Russian.

