# Deterministic FISS Rules Registry

This registry defines the mechanical validation rules enforced by `fiss-lint` based on the normative requirements of [FISS v1.0.0](https://fiss.vorozhko.ru/v1.0.0/llms.txt) and the official [Conformance Checklist](https://fiss.vorozhko.ru/v1.0.0/en/conformance.html).

## Severity Classification
- **Error:** Triggered by violations of normative `MUST`, `MUST NOT`, or `REQUIRED` requirements. Produces exit code `1`.
- **Warning:** Triggered by deviations from `SHOULD`, `SHOULD NOT`, or `RECOMMENDED` practices. Produces exit code `0` by default, or `1` under `--strict`.

## Rule Table

| Rule ID | Severity | Category | Normative Requirement |
|---|---|---|---|
| `FISS-R001` | Error | Root Structure | The project root MUST contain the `FISS/` directory. |
| `FISS-R002` | Error | Root Structure | The `FISS/` directory MUST contain `INDEX.md` and `BOOTSTRAP.md`. |
| `FISS-R003` | Error | Navigation | `FISS/INDEX.md` MUST contain a link targeting `BOOTSTRAP.md`. |
| `FISS-R004` | Error | Navigation | The link to `BOOTSTRAP.md` MUST have an attached read condition requiring reading before project work. |
| `FISS-R005` | Error | Syntax | Every navigation entry in any `INDEX.md` MUST follow the two-line format: `- [Title](path)` followed by `  Read when: <condition>` with an indented read condition marker and non-empty text. The marker MUST NOT be localized or overridden. |
| `FISS-R006` | Error | Link Integrity | Every relative link in an index MUST resolve to an existing physical `.md` file or `INDEX.md` of a composite area. Links to bare directories without `INDEX.md` are prohibited. |
| `FISS-R007` | Error | Topology | A composite area (a directory representing an area referenced in navigation) MUST contain its own `INDEX.md`. |
| `FISS-R008` | Error | Topology | Every used area MUST be reachable from `FISS/INDEX.md` through indexes. Standalone `.md` files in `FISS/` not reachable from the index hierarchy are flagged. |
| `FISS-R009` | Error | Overrides Entry | If `FISS/overrides/` exists, it MUST contain `INDEX.md`, and `FISS/INDEX.md` MUST link to `FISS/overrides/INDEX.md` with a read condition requiring it before skill use. |
| `FISS-R010` | Error | Agent Entry | If `AGENTS.md` (or root agent instruction file) exists in the project root, it MUST direct agents to `FISS/INDEX.md` without duplicating space content. |
| `FISS-R011` | Warning | Standard Reference | `FISS/INDEX.md` SHOULD contain a reference to the official standard website (`https://fiss.vorozhko.ru`). |
| `FISS-R012` | Error | Knowledge Partitioning | If `FISS/knowledge/` exists, it MUST function strictly as a container grouping `subject` and/or `project` areas. Loose files directly in `FISS/knowledge/` are prohibited. |
| `FISS-R013` | Error | Human Partitioning | If `FISS/human/` exists, it MUST function strictly as a container grouping `knowledge` and/or `hmm` areas. Loose files directly in `FISS/human/` are prohibited. |
| `FISS-R014` | Error | State Registry | Registries of state items in `FISS/state/` (such as ADRs, risks, open questions) MUST NOT be scattered loosely; they MUST be structured as composite areas with `INDEX.md` or consolidated single files. |
| `FISS-R015` | Error | Derivation Syntax | Derived knowledge MUST use the exact non-localized marker `Derived from:` followed by Markdown links to source materials. Local links MUST resolve to existing physical files. |
| `FISS-R016` | Error | HMM Derivation | Every content material in `FISS/human/hmm/` (and `FISS/human/hmm.md`) MUST be explicitly marked as derived (`Derived from:`) and link to source materials. |
| `FISS-R017` | Error | Override Routing | Override navigation in `FISS/overrides/` MUST be organized by the subject of a rule, NOT by tool or skill name. |
| `FISS-R018` | Error | Handoff Validity | The resolved FISS handoff record MUST declare a valid `fiss synchronization` state (`synchronized`, `pending`, or `unresolved`) and identify the current work item. |

## Formatting Invariants for `FISS-R005`
A valid navigation entry consists of exactly two lines:
1. Markdown list item: `- [Title](target.md)`
2. Indented read condition: `  Read when: non-empty condition text` (indented by 2 spaces).
The read condition marker MUST always be the exact English text `Read when:`. It MUST NOT be localized, translated, or overridden.
