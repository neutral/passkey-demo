# Analyze & Expand Current Step

## Purpose
- Drive implementation forward by expanding one concrete step at a time with actionable, verifiable checklists.
- Maintain traceability to blueprint artifacts and keep description files in sync with code.
- Important: Only expand the step in `blueprint/implementation.md`. Do not implement code or create files.

## Inputs
- `blueprint/implementation.md` (to locate and edit the current step).
- `blueprint/goals.md`, relevant requirements/specs/ADRs, and existing `*.desc.md` for context.

## Current Step Selection
- Current step = the lowest numbered step found in `blueprint/implementation.md`.
- If a `## Done` section exists listing completed steps, skip any step numbers explicitly recorded there and choose the lowest remaining step.

## Output
- A patch that ONLY edits `blueprint/implementation.md`, expanding the selected step in place so an implementer can execute it later without ambiguity.
- The expansion includes clear subsections (Structure, Source to add, Description files to add, Blueprint updates, Verification, Notes) and explicit “Refs: …”.

## Steps
1) Orient
- Read `blueprint/implementation.md` and locate the current step by number and title.
- Skim `blueprint/goals.md`, relevant requirements/specs/ADRs, and any existing `*.desc.md` for context.

2) Expand the Step (plan only)
- Source changes: enumerate exact files to add/modify/delete with paths and short purposes. Include example commands when helpful.
- Description files: list the `<folder>.desc.md` or `<file>.<ext>.desc.md` that must be created/updated; add one‑line purpose and the Refs you’ll include.
- Blueprint refs: include a “Refs:” line mapping to related goals/requirements/specs/decisions. Only mark a task In‑Progress when linked to Approved artifacts.
- Verification: specify concrete checks and commands (builds, curls, tests) with success criteria (status codes, output fields, file presence). Include a fix‑forward loop: run the tests, address any failures in implementation, and re‑run until all tests pass.
- Sequencing: call out dependencies or temporary workarounds (e.g., inline session lookup before middleware, Vite proxy before CORS).
- Policies & limits: restate critical policy bits enforced by this step (e.g., UV/residentKey/attestation; TTLs; size limits) and how to verify them.
- Unit tests: define the testing strategy for this step and list the test files to add (paths, names). Cover the full surface area: happy‑path, edge/boundary cases, invalid inputs, and error paths. Call out invariants to assert (e.g., monotonic counters, canonical encodings, curve membership) and any golden vectors/fixtures. Include the command to run tests (e.g., `go test ./...`) and expected outcomes.

3) Update the Plan Only (no code changes)
- Edit only `blueprint/implementation.md`, expanding the current step with clear subsections:
  - Structure, Source to add, Description files to add, Blueprint updates, Verification, Notes.
- Do not add/modify any source files, test files, or description files outside of the plan.
- Include exact commands and file paths as instructions, but do not run them or create the files now.

4) Validate (as a plan)
- Ensure verification commands and success criteria are present and specific, but do not execute them.
- Confirm the expansion is self‑contained so an implementer can perform it later without ambiguity.
- Ensure the unit test plan is complete and runnable (commands included) and asserts key invariants.
 - Confirm the Verification section explicitly includes re‑running tests after fixes and an “all tests pass” exit criterion.

5) Trace & Close
- Do not mark the step Done or change states based on analysis alone.
- Keep “Refs:” accurate within the step expansion. Later, when implemented, the step can be moved under `## Done` with date/notes/Refs.

## Constraints
- Submit a patch that ONLY edits `blueprint/implementation.md` by expanding the selected step in place.
- Do not modify any other files; do not create or delete files.
- Use concise bullets; prefer commands and acceptance checks over prose where it improves clarity.

## Step expansion must include
- Structure, Source to add, Description files to add, Blueprint updates, Verification, Unit tests (files, cases, invariants, commands), Notes.

## Example
Step: 1 — Initialize repo
- Context: scaffold structure and meta files to unblock builds and docs.
- Structure: verify `server/`, `web/`, `blueprint/` exist; root `.gitignore` present.
- Source to add: `.editorconfig` (UTF‑8, LF, trim, final newline; 2 spaces for TS/MD/JSON; tabs for Go).
- Description files to add: `server/server.desc.md` (Refs: R-PLAT-2,R-PLAT-3,R-SEC-UV); `web/web.desc.md` (Refs: R-PLAT-1,R-UI-2BTN).
- Blueprint updates: add Refs under step; do not mark other steps In‑Progress.
- Verification: `git status` shows only the three new files; `test -f .editorconfig`; `ls server server/internal web/src` exit 0.
- Notes: no code changes in this pass; actual file creation occurs during implementation.
