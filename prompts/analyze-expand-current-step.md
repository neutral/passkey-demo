# Analyze & Expand Current Step

## Purpose
- Expand one concrete step at a time with actionable, verifiable checklists that an implementer can follow verbatim.
- Maintain traceability to blueprint artifacts; require description files to be updated for any modified sources.
- Require a complete unit-test plan that includes a happy path and negative cases.
- Important: Only expand the step in `blueprint/implementation.md`. Do not implement code or create files.
 - Plan updates to Blueprint artifacts (requirements/specs/ADRs/user-flows) so functional and non-functional requirements remain consistent with the step.

## Inputs
- `blueprint/implementation.md` (to locate and edit the current step).
- `blueprint/goals.md`, relevant requirements/specs/ADRs, and existing `*.desc.md` for context.

## Current Step Selection
- Current step = the lowest numbered step found in `blueprint/implementation.md`.
- If a `## Done` section exists listing completed steps, skip any step numbers explicitly recorded there and choose the lowest remaining step.

## Output
- A patch that ONLY edits `blueprint/implementation.md`, expanding the selected step in place so an implementer can execute it later without ambiguity.
- The expansion includes clear subsections:
  - Scope
  - Source to add/modify (files + short purpose)
  - Description files (to create and to update for modified sources)
  - Blueprint updates (requirements/specs/ADRs/user-flows to add/update)
  - Request/response shape (if applicable)
  - Algorithm (validation and processing logic)
  - Database interactions (queries/updates, if applicable)
  - Policies & limits (and how to verify them)
  - Sequencing (dependencies/workarounds)
  - Tests (happy path required, negative cases, invariants, mapping checks, commands)
  - Verification (unit + manual) and a "User verification commands" block
  - Acceptance criteria
  - Notes
  - Refs (explicit lines to related artifacts)

## Steps
1) Orient
- Read `blueprint/implementation.md` and locate the current step by number and title.
- Skim `blueprint/goals.md`, relevant requirements/specs/ADRs, and any existing `*.desc.md` for context.

2) Expand the Step (plan only)
- Source to add/modify: enumerate exact files to add/modify/delete with paths and short purposes. Include example commands when helpful.
- Description files: list the `<folder>.desc.md` or `<file>.<ext>.desc.md` to create AND explicitly note updates for any modified sources (e.g., `server/cmd/api/main.go.desc.md` when adding routes). Include one‑line purposes and the Refs you’ll include.
- Blueprint refs: include a “Refs:” line mapping to related goals/requirements/specs/decisions. Only mark a task In‑Progress when linked to Approved artifacts.
 - Blueprint updates: describe which Blueprint artifacts must change (or be created) because of this step, covering both functional requirements (FRs) and non‑functional requirements (NFRs):
   - Requirements: `blueprint/features/<feature>/requirement.md` (FR) or `blueprint/global/<topic>/requirement.md` (NFR). Specify fields to add/update (type, description, depends_on, scope, acceptance criteria, flows, interfaces, risks, date, owners). If a new requirement is needed, propose it (Draft) with a brief rationale.
   - Specs: `.../_specs/spec.md` and any related explainers to update/add (interfaces, data/models, algorithms, security/privacy, errors/observability, testing strategy, open questions). Align to interface/model changes planned for the step.
   - ADRs: `blueprint/_decisions/<decision-name>.md` to introduce or revise architectural decisions; state status (Draft/Approved) and consequences.
   - User flows: `blueprint/_user-flows/<kebab-name>.md` to update if UX/API sequencing changes.
   - Refs hygiene: ensure all updated artifacts include current `Refs:` pointing to goals/requirements/specs/decisions touched by this step.
- Request/response shape: specify minimal JSON fields (names, types) and binary encodings (e.g., base64url) when relevant.
- Algorithm: lay out validation, policy checks, and processing steps in order, including cryptographic digests, comparisons, and invariants.
- Database interactions: list exact queries/updates (tables/columns) with conditions and side effects.
- Policies & limits: restate critical policy bits (e.g., UV/residentKey/attestation; TTLs; size limits; allowlists) and how they will be verified.
- Sequencing: call out dependencies or temporary workarounds (e.g., inline session lookup before middleware; dev localhost exceptions for origin/secure cookies).
- Unit tests: define the testing strategy and list the test files to add (paths). Happy path is REQUIRED. Also cover edge/boundary cases, invalid inputs, policy failures, invariants (e.g., monotonic counters, canonical encodings, curve membership), and mapping to HTTP statuses. Include commands to run tests and expected outcomes.
- Verification: include concrete checks and commands (builds, curls), success criteria, and a "User verification commands" fenced block with copy/paste steps. Include a fix‑forward loop in the plan to re‑run tests until green during implementation.

3) Update the Plan Only (no code changes)
- Edit only `blueprint/implementation.md`, expanding the current step with clear subsections listed in Output.
- Do not add/modify any source files, test files, or description files outside of the plan.
- Include exact commands and file paths as instructions, but do not run them or create the files now.

4) Validate (as a plan)
- Ensure verification commands and success criteria are present and specific, but do not execute them.
- Confirm the expansion is self‑contained so an implementer can perform it later without ambiguity.
- Ensure the unit test plan is complete and runnable (commands included), includes a REQUIRED happy path, and asserts key invariants.
- Confirm the Verification section explicitly includes re‑running tests after fixes and an “all tests pass” exit criterion.
 - Confirm a concrete Blueprint updates section exists, listing exact requirement/spec/ADR/user‑flow files to update or create, with clear change intent. If any new requirements/specs/ADRs are proposed, ensure their key fields and Refs are specified as Draft.

5) Trace & Close
- Do not mark the step Done or change states based on analysis alone.
- Keep “Refs:” accurate within the step expansion. Later, when implemented, the step can be moved under `## Done` with date/notes/Refs.

## Constraints
- Submit a patch that ONLY edits `blueprint/implementation.md` by expanding the selected step in place.
- Do not modify any other files; do not create or delete files.
- Use concise bullets; prefer commands and acceptance checks over prose where it improves clarity.

## Step expansion must include
- Scope
- Source to add/modify
- Description files (to create AND updates for modified sources)
 - Blueprint updates (requirements/specs/ADRs/user-flows to update/add with exact paths)
- Request/response shape (if applicable)
- Algorithm
- Database interactions (if applicable)
- Policies & limits
- Sequencing
- Tests (happy path required, negative cases, invariants, mapping checks, commands)
- Verification (unit + manual) and "User verification commands" block
- Acceptance criteria
- Notes
- Refs

## Example
Step: 1 — Initialize repo
- Context: scaffold structure and meta files to unblock builds and docs.
- Structure: verify `server/`, `web/`, `blueprint/` exist; root `.gitignore` present.
- Source to add: `.editorconfig` (UTF‑8, LF, trim, final newline; 2 spaces for TS/MD/JSON; tabs for Go).
- Description files to add: `server/server.desc.md` (Refs: R-PLAT-2,R-PLAT-3,R-SEC-UV); `web/web.desc.md` (Refs: R-PLAT-1,R-UI-2BTN).
 - Blueprint updates: none in this step; future steps will add FR/NFRs and specs when interfaces or policies change.
- Blueprint updates: add Refs under step; do not mark other steps In‑Progress.
- Verification: `git status` shows only the three new files; `test -f .editorconfig`; `ls server server/internal web/src` exit 0.
- Notes: no code changes in this pass; actual file creation occurs during implementation.
