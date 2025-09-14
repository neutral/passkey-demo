# Step 45 — README (Done: 2025-09-12)

## Scope
- Provide a concise README to bootstrap a teammate in under 10 minutes with quickstart, local run, test, and docs pointers.

## Source to add/modify
- Update `README.md` to include:
  - Quickstart prerequisites (Go, Node, npm) and one-liner to run server+web.
  - Environment sample usage via `.env.example`.
  - Build/test commands; short vs full Go tests.
  - Links to `docs/api-examples.md`, Postman collection, REST Client `.http`.
  - Notes on WebAuthn ceremonies and platform prompts (e.g., Touch ID).
  - Logging note (slog JSON) and how to set `LOG_LEVEL`/`LOG_FORMAT` for dev.

## Description files
- N/A (top-level docs only).

## Request/response shape
- N/A.

## Algorithm
- Author minimal, scannable sections with copy/paste commands that match existing Makefile targets and docs.

## Database interactions
- N/A.

## Policies & limits
- Avoid including secrets; refer to `.env.example` for local defaults.

## Sequencing
- Complements Steps 41–44 to provide a discoverable entrypoint.

## Tests
- Manual: follow the README Quickstart commands to confirm the experience.

## Verification
- `grep -q "Quickstart" README.md` returns 0.
- Copy/paste `make run` works given `.env.example` defaults.

## User verification commands
```bash
rg -n "Quickstart|API Examples|Postman|REST Client" README.md
```

## Acceptance criteria
- README contains Quickstart, local dev/run/test commands, and documentation links.

## Notes
- Keep README focused; deeper docs live in `docs/`.

## Refs
- Refs: requirement R-OPS-DEV; requirement R-PLAT-2

