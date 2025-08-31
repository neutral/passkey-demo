# Blueprint Strategy

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

# Description Strategy

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

# Agent Workflow

- Orient: read `blueprint/goals.md` and relevant `*.desc.md` to anchor priorities and current design intent.
- Plan: pull tasks from `blueprint/implementation.md` with clear “Refs”; identify folders/files to touch and whether required `.desc.md` files exist.
- Prepare: create or update description files (`<folder>.desc.md`, `<file>.<ext>.desc.md`) alongside planned code, including Refs to goals/requirements/specs/decisions.
- Build: scaffold from spec-defined interfaces and data models; consult `blueprint/_user-flows/` to drive UX/API; implement with exhaustive inline code comments.
- Document: keep description files high-level and accurate (logic, relations, invariants, interfaces, errors); co-locate with code and update as behavior evolves.
- Trace: carry artifact names through branches, commits, and PRs; reference both blueprint artifacts and the description files updated/created.
- Verify: execute testing strategy from specs; confirm acceptance criteria; mark tasks Verified; ensure description files and comments reflect final behavior.
- Migrate: if `blueprint/_draft/` exists, migrate notes into structured folders and update related description files.
