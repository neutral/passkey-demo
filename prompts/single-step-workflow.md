# Single‑Step Workflow

## Purpose
Provide a repeatable, end‑to‑end process for advancing one implementation step at a time with clear planning, code, tests, verification, and traceability.

## Prerequisites
- Read `blueprint/goals.md`, the relevant requirement/spec/ADR, and any `*.desc.md` for context.
- Open a local, step‑scoped branch named `step-<n>-<kebab-name>` (e.g., `git checkout -b step-8-cose-to-ecdsa`) unless asked to work directly on `main`.
- Plan to merge this branch back into `main` with a merge commit when done, then delete it locally and on remote.

## Naming & Traceability
- Branch: `step-<n>-<kebab-name>` (e.g., `step-8-cose-to-ecdsa`).
- Refs: carry “Refs: goal/requirement/spec/decision” in plan entries, description files, and commit messages.
- Done artifacts: move completed step content into `blueprint/done/phase-<letter>-step-<n>-<kebab>.md`.

## Analyze & Expand (plan‑only)
- Edit only `blueprint/implementation.md` to expand the current step (lowest pending number):
  - Structure: folders/modules you’ll touch.
  - Source to add: exact file paths, responsibilities, commands.
  - Description files to add: `<folder>.desc.md` / `<file>.<ext>.desc.md` with Refs.
  - Blueprint updates: Refs/state guidance.
  - Verification: build/run/curl and expected results; include fix‑forward loop (re‑run until passes).
  - Unit tests: testing strategy, test file paths, cases (happy path, edge/error), invariants to assert, and commands to run tests.
  - Notes: scope limits, sequencing/workarounds, policy/limit restatements.
- Do not change code yet in this phase.

### Confirmation Gate (Stop for approval)
- After posting the expanded step (plan‑only) in `blueprint/implementation.md`, pause and ask for user confirmation before proceeding with implementation.
- Only continue to “Implement” once the user explicitly approves the plan.

## Implement (code + desc files)
- Create/modify only the files listed in the plan.
- Keep code minimal and focused; avoid unrelated refactors.
- Add/update description files alongside code (purpose, relations, invariants, interfaces, Refs).

### Generate Explaininers for New Concepts
- When introducing or working with new concepts/technologies (e.g., SQLite usage, common types, CORS, cookies, sessions), generate a non‑developer explainer under the relevant requirement’s `_specs/` using `prompts/non-dev-explainer-prompt.md`.
- Link the explainer in the plan’s Refs, and reference it from relevant description files.

## Unit Tests & Invariants
- Add tests specified in the plan; cover the entire surface:
  - Valid paths and error paths; boundary and malformed inputs; invariants (e.g., counters monotonicity, canonical encoding, curve membership, PRAGMAs enabled).
- Run tests and fix forward until all pass: `go test ./...` (or equivalent).

## Verification (copy/paste commands)
- Execute the plan’s verification commands (builds, dev runs, curls); ensure expected outputs/status.
- Keep a short, idempotent “User verification commands” block in the plan for manual checks.

## Done Artifacts
- Move the entire expanded step content verbatim into a new file under `blueprint/done/` with a phase‑prefixed filename.
- Remove the step from the active list (no pointers left behind). Keep the `## Done` section minimal and pointing to the done folder.

## Merge & Cleanup
- If working on a branch:
  - Rebase on the latest `main` if needed.
  - Merge to `main` with a merge commit; push `main`.
  - Delete the feature branch locally (`git branch -d step-<n>-...`) and on remote (`git push origin --delete step-<n>-...`).
- If working on `main`, commit directly after tests pass and verification completes.

## Non‑Developer Explainers (as needed)
- For subsystems (SQLite, common types, CORS, sessions), add a non‑developer explainer under the relevant requirement’s `_specs/` using `prompts/non-dev-explainer-prompt.md`.

## Conventions
- Markdown: H1 title, `## Purpose`, consistent H2/H3; fenced code blocks with language hints.
- Refs: always include “Refs: goal/requirement/spec/decision” where relevant.
- Logging: do not log raw binary; prefer thumbprints/IDs when needed.

## Quick Checklist
1) Create branch `step-<n>-<kebab-name>`.
2) Expand step in `blueprint/implementation.md` (plan only; include tests and fix‑forward verification).
3) Implement code + description files per plan.
4) Add unit tests; run `go test ./...`; fix until green.
5) Run verification commands; fix forward.
6) Move full step content into `blueprint/done/...md`; remove from active list.
7) Merge to `main`; delete branch.
