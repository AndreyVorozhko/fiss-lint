# Semantic Versioning & Release Policy

This document defines the semantic versioning schema, increment criteria, and release lifecycle for the `fiss-lint` utility.

## 1. Version Format

`fiss-lint` strictly adheres to [Semantic Versioning 2.0.0 (SemVer)](https://semver.org/).

The release version string format:
```text
vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

- `v`: Literal version prefix applied to Git tags and CLI version displays.
- `MAJOR`: Incompatible API or CLI changes, or breaking rule semantics.
- `MINOR`: Backward-compatible feature additions, new FISS rules, or CLI capabilities.
- `PATCH`: Backward-compatible bug fixes, diagnostic refinements, and internal maintenance.
- `PRERELEASE`: Optional identifier for development or pre-release builds (e.g., `v1.0.0-dev`, `v1.1.0-rc.1`).
- `BUILD`: Optional build metadata (e.g., commit hash, build date).

## 2. Increment Rules

### MAJOR Version (X.0.0)
Incremented when changes break backward compatibility for users, automated tooling, or existing conforming FISS spaces:
- **CLI Contract Breaking Changes:** Removal or incompatible alteration of existing flags (e.g. `--strict`, `--format`).
- **Machine Output Breaking Changes:** Incompatible modification of the JSON report schema (`issues` array or `summary` object structure).
- **Rule Breaking Changes:** Modifying the validation behavior of existing rules (`FISS-R001` through `FISS-R018`) such that previously compliant FISS spaces produce new errors or failures without changes in standard specification.
- **Minimum Environment Requirements:** Incompatible elevation of minimum supported Go runtime or OS platforms.

### MINOR Version (X.Y.0)
Incremented when new functionality is introduced in a backward-compatible manner:
- **New Validation Rules:** Introduction of new deterministic rules (e.g., `FISS-R019+`).
- **CLI Enhancements:** Addition of new optional flags or non-breaking reporting modes.
- **Relaxation or Mirror Support:** Broadening rule acceptance without breaking existing conforming spaces (e.g., supporting official GitHub mirrors in `FISS-R011`).
- **Report Extensions:** Adding non-breaking fields to JSON reports (e.g., `summary.version` with `omitempty`).

### PATCH Version (X.Y.Z)
Incremented for backward-compatible bug fixes, optimizations, and documentation updates:
- **Linter Bug Fixes:** Correcting false positives or false negatives in existing rule implementations.
- **Diagnostic Clarifications:** Refining error or warning messages for clearer human and agent remediation.
- **Internal Maintenance:** Code refactoring, test coverage improvements, and performance enhancements.
- **Documentation Updates:** Updating guides, README, or internal FISS knowledge files.

## 3. Toolchain & Build Integration

The project build automation (`Makefile`) automatically extracts and injects the SemVer version during compilation:

```makefile
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || echo "v1.0.0-dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
```

- When built within a Git repository with release tags, `git describe --tags --dirty` yields the exact tag (e.g., `v1.0.0`) or tag-based development description (`v1.0.0-2-g05a99dd-dirty`).
- When built outside a Git clone (such as from a source tarball or in restricted CI), the fallback default is `v1.0.0-dev`.
- Both human-readable CLI outputs (`fiss-lint --version`) and machine-readable JSON reports (`summary.version` under `--format json`) display the resolved SemVer version.

## 4. Git Tagging & Release Procedure

1. **Tag Format:** Release tags MUST be annotated Git tags with a descriptive release message:
   ```bash
   git tag -a v1.0.0 <commit-sha> -m "Release v1.0.0: Initial stable release of fiss-lint"
   ```
2. **Release Commit:** The release tag is created on the canonical commit on the `main` branch after review and acceptance.
3. **Distribution Artifacts:** Every release is accompanied by statically linked standalone binaries built via `make build-all`:
   - `fiss-lint-linux-amd64`, `fiss-lint-linux-arm64`
   - `fiss-lint-darwin-amd64`, `fiss-lint-darwin-arm64`
   - `fiss-lint-windows-amd64.exe`, `fiss-lint-windows-arm64.exe`
4. **Traceability:** Releases correspond to milestones and user stories in the canonical task tracker (Taiga).
