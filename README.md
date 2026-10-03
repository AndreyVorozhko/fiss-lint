# fiss-lint

[Русский](README_RU.md) | **English**

`fiss-lint` is an autonomous, zero-dependency command-line static analyzer and linter for projects following the [File-based Intellectual Space Standard (FISS v1.0.0)](https://fiss.vorozhko.ru/v1.0.0/llms.txt) (or its [official GitHub mirror](https://github.com/AndreyVorozhko/fiss)).

The utility verifies the physical structure of the `FISS/` directory, the integrity and reachability of the navigation graph, the correctness of situational read conditions (`Read when:`), derivation syntax (`Derived from:`), area partitioning, and agent entry points (`AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `QWEN.md`).

---

## Features

- **Deterministic FISS v1.0.0 Rule Coverage:** 18 mechanical validation rules (`FISS-R001` through `FISS-R018`) spanning root structure, navigation, overrides, link integrity, domain classification, and agent entry points.
- **Machine-Readable Diagnostics:** Structured JSON output (`--format json`) with injected SemVer version in the `summary` block for CI/CD pipelines and automated tooling.
- **Strict Verification Mode (`--strict`):** Escalates any warning (`Warning`) to a non-zero exit code `1`.
- **Zero Runtime Dependencies:** Pure Go compiled without CGO (`CGO_ENABLED=0`).
- **Cross-Platform Support:** Precompiled standalone binaries for Linux (`amd64`, `arm64`), macOS (`amd64`, `arm64`), and Windows (`amd64`, `arm64`).

---

## Installation

### 1. Quick One-Line Installer Script (Linux, macOS, Windows)

Install `fiss-lint` with a single command that downloads the precompiled release binary matching your operating system and CPU architecture without requiring Go:

**Linux and macOS (POSIX sh/bash via `curl`):**
```bash
curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
```

Install a specific version (e.g., `v1.0.0`):
```bash
curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash -s -- v1.0.0
```

By default, the binary is installed into `~/.local/bin` (or `/usr/local/bin` if running with root privileges). The installation directory can be customized via the `INSTALL_DIR` environment variable:
```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.ps1 | iex
```

Install a specific version:
```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.ps1))) -Version v1.0.0
```

The script downloads `fiss-lint.exe` into `%LOCALAPPDATA%\Programs\fiss-lint` (or `~/bin`) and automatically registers the directory in your `User PATH`.

### 2. Install via Go Toolchain

If Go 1.21 or later is installed on your system:

```bash
go install github.com/AndreyVorozhko/fiss-lint/cmd/fiss-lint@latest
```

### 3. Precompiled Release Binaries

Directly download precompiled standalone binaries from [GitHub Releases](https://github.com/AndreyVorozhko/fiss-lint/releases):
- Linux: `fiss-lint-linux-amd64`, `fiss-lint-linux-arm64`
- macOS: `fiss-lint-darwin-amd64`, `fiss-lint-darwin-arm64`
- Windows: `fiss-lint-windows-amd64.exe`, `fiss-lint-windows-arm64.exe`

Make the binary executable (`chmod +x fiss-lint-*`) and place it in your `$PATH`.

### 4. Build from Source

```bash
git clone https://github.com/AndreyVorozhko/fiss-lint.git
cd fiss-lint
make build
```

The compiled binary will be placed at `bin/fiss-lint`.

To cross-compile binaries for all supported platforms:
```bash
make build-all
```

---

## CLI Usage

```bash
fiss-lint [flags] [path]
```

If `[path]` is omitted, `fiss-lint` checks the current working directory (`.`).

### Command-Line Flags

| Flag | Description | Default |
|---|---|---|
| `-h`, `--help` | Show CLI help and rules summary | `false` |
| `-v`, `--version` | Display version, commit hash, build date, and Built with FISS badge | `false` |
| `--format string` | Output format: `text` (default terminal lines) or `json` | `text` |
| `--strict` | Fail with exit code `1` if any warnings (`Warning`) are detected | `false` |

### Exit Codes

| Exit Code | Description |
|---|---|
| `0` | Verification passed (no errors; warnings permitted without `--strict`). |
| `1` | Validation failed (errors detected, warnings under `--strict`, or filesystem error). |
| `2` | Command-line argument error (e.g., unsupported `--format`). |

---

## Output Formats

### Text Format (`--format text`)

By default, output is formatted as concise diagnostic lines:

```text
[ERROR] [FISS-R001] FISS: directory FISS/ not found
[WARNING] [FISS-R011] FISS/INDEX.md: recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found
```

When all checks pass, `fiss-lint` produces zero stdout output and exits with code `0`.

### Structured JSON (`--format json`)

Passing `--format json` outputs a formatted JSON document:

```json
{
  "issues": [
    {
      "rule_id": "FISS-R011",
      "severity": "WARNING",
      "file": "FISS/INDEX.md",
      "line": 0,
      "message": "recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found"
    }
  ],
  "summary": {
    "version": "v1.0.0",
    "errors": 0,
    "warnings": 1,
    "total": 1
  }
}
```

When no violations are detected, the `issues` array is empty and summary counters are `0`:

```json
{
  "issues": [],
  "summary": {
    "version": "v1.0.0",
    "errors": 0,
    "warnings": 0,
    "total": 0
  }
}
```

---

## Deterministic Rules Catalog (FISS-R001 — FISS-R018)

`fiss-lint` enforces 18 deterministic rules defined in the FISS v1.0.0 specification:

| Rule ID | Severity | Category | Normative Requirement |
|---|---|---|---|
| `FISS-R001` | Error | Root Structure | The project root MUST contain the `FISS/` directory. |
| `FISS-R002` | Error | Root Structure | The `FISS/` directory MUST contain `INDEX.md` and `BOOTSTRAP.md`. |
| `FISS-R003` | Error | Navigation | `FISS/INDEX.md` MUST contain a link targeting `BOOTSTRAP.md`. |
| `FISS-R004` | Error | Navigation | The link to `BOOTSTRAP.md` MUST have an attached read condition requiring reading before project work. |
| `FISS-R005` | Error | Syntax | Every navigation entry in any `INDEX.md` MUST follow the two-line format (`- [Title](target.md)` + `  Read when: condition`). The marker MUST NOT be localized or overridden. |
| `FISS-R006` | Error | Link Integrity | Every relative link in an index MUST resolve to an existing physical `.md` file or `INDEX.md` of a composite area. Links to bare directories without `INDEX.md` are prohibited. |
| `FISS-R007` | Error | Topology | A composite area (a directory representing an area referenced in navigation) MUST contain its own `INDEX.md`. |
| `FISS-R008` | Error | Topology | Every used area MUST be reachable from `FISS/INDEX.md` through indexes. Isolated (orphan) files are flagged. |
| `FISS-R009` | Error | Overrides Entry | If `FISS/overrides/` exists, it MUST contain `INDEX.md`, and `FISS/INDEX.md` MUST link to `FISS/overrides/INDEX.md` with a read condition requiring it before skill use. |
| `FISS-R010` | Error | Agent Entry | If `AGENTS.md` (or root agent instruction file) exists in the project root, it MUST direct agents to `FISS/INDEX.md` without duplicating space content. |
| `FISS-R011` | Warning | Standard Reference | `FISS/INDEX.md` SHOULD contain a reference to the official standard website (`https://fiss.vorozhko.ru`) or official GitHub mirror (`https://github.com/AndreyVorozhko/fiss`). |
| `FISS-R012` | Error | Knowledge Partitioning | If `FISS/knowledge/` exists, it MUST function strictly as a container grouping `subject` and/or `project` areas. Loose files directly in `FISS/knowledge/` are prohibited. |
| `FISS-R013` | Error | Human Partitioning | If `FISS/human/` exists, it MUST function strictly as a container grouping `knowledge` and/or `hmm` areas. Loose files directly in `FISS/human/` are prohibited. |
| `FISS-R014` | Error | State Registry | Registries of state items in `FISS/state/` (such as ADRs, risks, open questions) MUST NOT be scattered loosely; they MUST be structured as composite areas with `INDEX.md` or consolidated single files. |
| `FISS-R015` | Error | Derivation Syntax | Derived knowledge MUST use the exact non-localized marker `Derived from:` followed by Markdown links to source materials. Local links MUST resolve to existing physical files. |
| `FISS-R016` | Error | HMM Derivation | Every content material in `FISS/human/hmm/` (and `FISS/human/hmm.md`) MUST be explicitly marked as derived (`Derived from:`) and link to source materials. |
| `FISS-R017` | Error | Override Routing | Override navigation in `FISS/overrides/` MUST be organized by the subject of a rule, NOT by tool or skill name. |
| `FISS-R018` | Error | Handoff Validity | The resolved FISS handoff record MUST declare a valid `fiss synchronization` state (`synchronized`, `pending`, or `unresolved`) and identify the current work item. |

---

## Git Hooks Integration

### `pre-push` Hook (Recommended)

A `pre-push` hook guarantees that broken context or non-conforming intellectual spaces cannot be pushed to remote repositories. Create `.git/hooks/pre-push`:

```bash
#!/bin/sh
set -e

echo "Running fiss-lint before push..."
fiss-lint --strict .
```

Make the hook executable:
```bash
chmod +x .git/hooks/pre-push
```

### `pre-commit` Hook

To validate locally before every commit, create `.git/hooks/pre-commit`:

```bash
#!/bin/sh
set -e

fiss-lint --strict .
```

```bash
chmod +x .git/hooks/pre-commit
```

### Pre-commit Framework

To integrate with the [pre-commit framework](https://pre-commit.com), add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: fiss-lint
        name: fiss-lint
        entry: fiss-lint --strict .
        language: system
        pass_filenames: false
        always_run: true
```

---

## CI/CD Pipeline Integration

### GitHub Actions

Example workflow in `.github/workflows/fiss-lint.yml`:

```yaml
name: FISS Validation

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  lint:
    name: Lint FISS Intellectual Space
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Install fiss-lint
        run: |
          curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Run fiss-lint in strict mode
        run: fiss-lint --strict .

      - name: Generate JSON report artifact
        if: always()
        run: fiss-lint --format json . > fiss-report.json || true

      - name: Upload validation report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: fiss-validation-report
          path: fiss-report.json
```

### GitLab CI

Example pipeline in `.gitlab-ci.yml`:

```yaml
stages:
  - lint

fiss_lint:
  stage: lint
  image: alpine:latest
  before_script:
    - apk add --no-cache bash curl
    - curl -fsSL https://raw.githubusercontent.com/AndreyVorozhko/fiss-lint/main/scripts/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
  script:
    - fiss-lint --strict .
  artifacts:
    when: always
    reports:
      junit: fiss-report.xml
    paths:
      - fiss-report.json
```

---

## Tooling Division of Responsibility in FISS Ecosystem

Within the FISS ecosystem, tools and AI agent skills operate with strictly demarcated responsibilities:

| Tool | Role | Mutates files? | Responsibility |
|---|---|:---:|---|
| **`fiss-lint`** | CLI Linter | **No** | Fast deterministic mechanical checks: AST parsing, link integrity, navigation syntax, and mandatory markers. |
| **`fiss-validate`** | AI Audit Skill | **No** | Qualitative semantic audit against the 7 architectural principles (7C: Compact, Context-aware, Context-first, Classified, Canonical, Continuous, Composable). |
| **`fiss-init`** | AI Bootstrap Skill | **Yes** | 0-to-1 bootstrapping of new FISS spaces from scratch: scaffolding root files (`INDEX.md`, `BOOTSTRAP.md`), agent entry points, and continuity mechanisms. |
| **`fiss-maintain`** | AI Maintenance Skill | **Yes** | Day 2 operational continuity: managing the 3-phase handoff gate (`Lock` -> `Prepare` -> `Release`), integrating outcomes across 7 context dimensions, and updating navigation. |

---

## License

This project is licensed under the terms of the MIT License.

---

Built with ღ and [FISS](https://fiss.vorozhko.ru)
