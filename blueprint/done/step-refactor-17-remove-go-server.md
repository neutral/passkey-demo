### Step Refactor-17 — Repo: remove Go server artifacts

Scope

- Delete the legacy `server/` Go backend (source, tooling, docs) now that the Node stack is the single implementation.
- Update tooling, docs, and CI references that still point to Go binaries or make targets.

Source to modify/delete

- Remove `server/` directory (Go source) and any Go-specific tooling under `tools/`/`docs/` referencing it.
- Update root `Makefile`, scripts, or npm commands that still launch the Go server.
- Clean up `.gitignore`, `.vscode/`, or CI configs referencing Go artifacts (`go.sum`, binaries).

Description files

- Delete `server/*.desc.md` alongside source removal.
- Update root/project descriptions (`Makefile.desc.md`, `docs/*.desc.md`) to reflect Node-only backend.

Blueprint updates

- `blueprint/implementation.md` Done section (once the step completes) noting removal.
- Any remaining requirements/specs referencing Go (e.g., archived done steps) should include pointer to archival status; if live specs reference Go, update them to state “historical context only”.

Policies & limits

- Ensure removing Go artifacts does not affect historical references required for compliance (if necessary, leave archived docs in `blueprint/done/_archive-go`).

Sequencing

- Should run after Refactor-16 so the web client no longer depends on Go.
- Coordinate with documentation Step 18 to avoid duplicate README edits.

Tests

- `npm -C node-server run lint` / `npm -C node-server run test` (or equivalent) to confirm Node backend unaffected.
- `npm -C web run build` / Playwright smoke to ensure entire repo still builds without Go.
- `git grep` to verify no references to Go server remain in active docs/scripts.

Verification

- Repository builds/tests pass without Go toolchain installed.
- Manual `git status` shows `server/` removal only alongside planned doc/script updates.

Acceptance criteria

- Go backend code and entrypoints removed; Node backend is sole implementation.
- Scripts/docs reflect Node-only workflow.

Notes

- Preserve historical Go artifacts under `blueprint/done/_archive-go` or docs if needed for reference.

Refs: goal simple-ui-and-storage; requirement R-OPS-DEV; requirement R-PLAT-3; decision router-builder-wiring (update note if needed)
