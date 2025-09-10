# Agents Guide

## Blueprint Strategy

## Purpose

- Centralize goals, requirements, specs, decisions, and the implementation plan.
- Ensure traceability: Requirement → Spec → Decision → Task → Code → Verification.

## Conventions

- Kebab‑case names for files and folders.
- Add a “Refs:” line linking related artifacts (goal/requirement/spec/decision).
- Standard states: Draft, Reviewed, Approved, In‑Progress, Merged, Verified.

## Folder Layout

- `blueprint/goals.md`: Project goals and design tenets.
- `blueprint/_decisions/`: ADRs (architecture decisions).
- `blueprint/global/`: Non‑functional requirements (NFRs).
- `blueprint/features/`: Functional requirements (FRs), grouped by feature.
- `blueprint/_user-flows/`: Canonical user flows (one per file, kebab‑case names).
- `blueprint/implementation.md`: Implementation task list tied to requirements/specs.

## Goals (blueprint/goals.md)

- Include title, date, owners.
- List concise, testable goals; add design tenets and non‑goals.
- Define success metrics and change control (changes require an ADR).
- Traceability: ADRs and requirements reference related goals via “Refs: goal <name>”.

## Requirements & Specs

- Requirement file: `blueprint/(features|global)/<name>/requirement.md`.
  - Fields: type (fr|nfr), description, depends_on, scope, acceptance criteria, flows, interfaces, risks, date, owners.
- Spec file: `blueprint/(features|global)/<name>/_specs/spec.md`.
  - Cover: overview, interfaces, data/models, algorithms, security/privacy, errors/observability, testing strategy, open questions.
- Decision (ADR): `blueprint/_decisions/<decision-name>.md` with status, context, decision, consequences, alternatives, references.
- Always add “Refs: requirement/spec/decision/goal” to connect artifacts.

## Implementation Plan (blueprint/implementation.md)

- Single source of implementation tasks derived from Approved requirements/specs.
- Entry format:
  - `[ ] <area>: <requirement-name> → <action>`
    - `Spec: <spec-name>` | `Decision: <decision-name>`
    - `Refs: goal <goal-name>; requirement <name>`
    - `Notes: <constraints or codegen hints>`
    - `State: Draft|Reviewed|Approved|In‑Progress|Merged|Verified`
- Rules:
  - No orphan tasks: every task references at least one requirement (and typically a spec).
  - Keep “Refs” current when goals/requirements/specs/ADRs change.

## Quality Gates

- Requirements: testable acceptance criteria; clear scope and risks.
- Specs: cover interfaces, data, security, errors, observability, and testing.
- ADRs: explicit consequences and trade‑offs; status maintained.
- Tasks: only In‑Progress when linked to Approved artifacts; PRs reference artifact names and confirm criteria are met.

## Migration Note

- When present, `blueprint/_draft/` contains raw notes/specs. Migrate its content into the structured folders (`global/`, `features/`, `_user-flows/`, `_decisions/`, `implementation.md`).

## Description Strategy

## Purpose

- Ensure every source folder and source file has a human‑readable description.
- Complement, not replace, exhaustive inline code comments within source files.

## Scope & Naming

- Applies to all folders that contain source code and all source code files.
- Description files are Markdown.
- For a folder: add a file with the same folder name, placed inside that folder, with extension `.desc.md`.
  - Example: `server/internal/webauthn/webauthn.desc.md` describes `server/internal/webauthn/`.
- For a source file: add a file with the full source filename (including its extension) plus `.desc.md` appended, colocated with the source file.
  - Example: `server/cmd/api/main.go.desc.md` describes `server/cmd/api/main.go`.
  - Example: `web/src/pages/Register.tsx.desc.md` describes `web/src/pages/Register.tsx`.

## Content Requirements

- High‑level overview of the logic and responsibilities of the folder or file.
- How it relates to other parts of the repo (dependencies, callers, data flow).
- Key invariants, assumptions, and important algorithms or protocols used.
- Interfaces and models it defines or consumes; notable errors/edge cases.
- Pointers to relevant blueprint artifacts using “Refs:” (goals/requirements/specs/decisions).
- Keep concise but informative enough to understand without reading the code first.

## Usage Rules

- Description files are mandatory for new code generated or added by agents.
- Update description files alongside code changes to keep them accurate.
- Do not omit or reduce inline code comments; they remain exhaustive within source files.
- Co‑locate description files with their targets; do not centralize elsewhere.

## Minimal Templates

Folder (`<folder>.desc.md`):

```
# Overview
<What lives here; responsibilities; how components interact>

# Relations
<Who uses this; what it depends on; data flow>

# Interfaces & Models
<Key handlers/types/schemas>

# Refs
Refs: goal <name>; requirement <name>; spec <name>; decision <name>
```

Source file (`<filename>.<ext>.desc.md`):

```
# Purpose
<What this file does at a high level>

# Key Logic
<Important functions/components, algorithms, invariants>

# Interactions
<Called by/depends on; external I/O; errors>

# Refs
Refs: goal <name>; requirement <name>; spec <name>; decision <name>
```

## Agent Workflow

- Orient: read `blueprint/goals.md` and relevant `*.desc.md` to anchor priorities and current design intent.
- Plan: pull tasks from `blueprint/implementation.md` with clear “Refs”; identify folders/files to touch and whether required `.desc.md` files exist.
- Prepare: create or update description files (`<folder>.desc.md`, `<file>.<ext>.desc.md`) alongside planned code, including Refs to goals/requirements/specs/decisions.
- Build: scaffold from spec-defined interfaces and data models; consult `blueprint/_user-flows/` to drive UX/API; implement with exhaustive inline code comments.
- Document: keep description files high-level and accurate (logic, relations, invariants, interfaces, errors); co-locate with code and update as behavior evolves.
- Trace: carry artifact names through branches, commits, and PRs; reference both blueprint artifacts and the description files updated/created.
- Verify: execute testing strategy from specs; confirm acceptance criteria; mark tasks Verified; ensure description files and comments reflect final behavior.
- Migrate: if `blueprint/_draft/` exists, migrate notes into structured folders and update related description files.

## Step Expansion Guidelines

Use the analyze and expand current step prompt from `prompts/analyze-expand-current-step.md` to review guidelines for analysis and expansion. Always wait for user to review the expanded analysis and plan before proceeding to implementation.

### Plan Maintenance

- When a step is completed and verified, move the entire expanded step content verbatim into a new file under `blueprint/done/` (preserve all headings, sub-bullets, verification criteria, and Refs). Do not summarize.
- File naming: prefix the filename with the phase to aid organization, then the step number and a short kebab name, e.g., `phase-a-step-1-initialize-repo.md`, `phase-b-step-6-db-init-and-migrations.md`.
- Remove the completed step from the active list so it exists only under `blueprint/done/` (no duplication). Add a single line description of the step to the Done section with the step number.
- Keep “Refs:” lines accurate when moving; add the completion date and any verification notes at the top of the moved step.

## Step Implementation Guidelines

This process applies after a step has been expanded in `blueprint/implementation.md`. Follow it to implement the step end‑to‑end while preserving traceability and quality.

### Prerequisites

- Read the expanded step and related artifacts (requirements/specs/ADRs, goals, user flows).
- Confirm dependencies are in place (prior steps complete; required tools installed).

### Branching

- Create a feature branch named after the step: `step-<n>-<short-name>`.

```bash
git checkout -b step-<n>-<kebab-name>
```

### Implement (scope exactly as planned)

- Create/modify only the files listed in the step’s “Source to add” and follow the stated policies (e.g., UV/residentKey/attestation, TTLs, limits).
- Keep changes minimal; avoid unrelated refactors. Generate a Refactor file with details and add it to `blueprint/_refactor` folder for future consideration.
- Maintain exhaustive inline code comments; align with repo style.
- Ensure all generated code is well commented within the source files, explaining core logic, invariants, assumptions, and error handling.

### Description Files

- Create/update all `<folder>.desc.md` and `<file>.<ext>.desc.md` listed in the step.
- Include accurate overviews, relations, invariants, interfaces, notable errors; keep “Refs:” lines up‑to‑date.
- When modifying existing source files, update their corresponding `<filename>.<ext>.desc.md` in the same commit to reflect the changes (routes added, logic shifts, new invariants, errors).

### Verification

- Execute every command in the step’s “Verification” section.
- Add/execute unit tests where specified; capture outputs (status codes, fields) and confirm success criteria.
- Provide a "User verification commands" fenced code block with copy/paste shell commands that validate the step end‑to‑end (build, run, curl/tests, cleanup). Keep it minimal and idempotent.

```bash
go build ./... || npm run build
curl :8080/health -i
```

### Documentation & Blueprint Updates

- If the step mentions docs (API examples, env sample), add/update them accordingly.
- Update `blueprint/implementation.md` with any verification notes explicitly called for by the step.

### Commit & Push

- Stage only the files relevant to this step.

```bash
git add <files>
git commit -m "step-<n>: <concise summary>" -m "Refs: goal ..., requirement ..., decision ..."
git push -u origin step-<n>-<kebab-name>
```

### Pull Request

- Open a PR linking the step title and “Refs:” from the plan.
- In the PR description, confirm acceptance criteria and verification outcomes.

### Move to Done

- After merge/verification, move the entire expanded step content verbatim to `## Done` in `blueprint/implementation.md` (add date + notes) and remove it from the active list. Do not leave any pointers behind.

## Markdown Conventions

- Title: every `.md` file starts with a single H1 (`# Title`) that clearly names the document.
- Purpose: include a short `## Purpose` section near the top stating why the doc exists.
- Headings: use a consistent hierarchy (H1 once; H2 for main sections; H3 for subsections). Do not skip levels.
- Code blocks: use fenced code blocks with language hints for clarity and syntax highlighting (e.g., `bash, `json, `go, `ts, `sql, `cddl). Prefer blocks over inline for multi-line commands or code.
- Inline code: wrap commands, file paths, env vars, constants, and identifiers in backticks.
- Lists: use bullets for concise, scannable items; keep each bullet verifiable and single-purpose.
- Refs: include a final `Refs:` line in blueprint artifacts linking goals/requirements/specs/decisions where applicable.
- Consistency: prefer present tense, active voice; avoid fluff; keep sections short and self-contained.

## Tools

Write scripts in the `tools/` folder to automate repetitive actions.

### Semantic Search (ck)

- Install check: run `ck --version` or `command -v ck`; if missing, install with `cargo install ck-search`.
- Index status: run `ck --status .` or check `./.ck/`; build with `ck --index .` (safe to delete/rebuild `.ck/`).
- When to use:
  - Exact text/regex → use default `ck "pattern" path/` (grep-compatible).
  - Intent-level concept → use `ck --sem "query" path/` (requires index).
  - Balance precision + recall → use `ck --hybrid "query" path/`.
  - Need whole functions/classes → add `--full-section`; for scripts add `--json`.
- See `tools/ck.md` for detailed flags, examples, and workflows.

## Refactoring Strategy

- Refer to `prompts/refactoring-strategy.md` for the reusable prompt and checklist on identifying, documenting, and prioritizing refactoring opportunities. Capture ideas under `blueprint/_refactor/` and do not implement them unless explicitly prioritized.
