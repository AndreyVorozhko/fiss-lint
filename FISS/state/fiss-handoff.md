# FISS Task Handoff

task: https://taiga.vorozhko.ru/project/vorozhkoru/us/4 (User Story #4: Парсинг навигационных записей и проверка строгого формата Read when)
canonical tracker: https://taiga.vorozhko.ru/
task status: ready_for_test
fiss synchronization: pending

## Context & Baseline
- **Previous Story:** Story #3 (Root Structure MVP) successfully completed, accepted, and merged into `main` (`40ab982`).
- **Target Story:** User Story #4 (Taiga Ref: 4, ID: 51) — "Парсинг навигационных записей и проверка строгого формата Read when".
- **Branch:** `feature/story-4/read-when-format`.
- **Pre-Planning Verification:** Story #4 scope, DoD, and covered rules (`FISS-R004`, `FISS-R005`) audited against FISS v1.0.0. Obsolete localization criterion #4 removed from Taiga, Task 101 deleted.
- **Verification:** All unit and integration tests passing (`go test -v ./...`, statement coverage > 93%), `make clean && make build-all && make test` (100% PASS). CLI binary verified on 5 invalid fixtures and self repository root. Ready for human verification.
