# Step 42 — API examples (Done: 2025-09-12)

## Scope
- Provide copy/pasteable curl examples for the core endpoints to help developers exercise the API quickly without instrumenting WebAuthn ceremonies.

## Source to add/modify
- Add `docs/api-examples.md` with sections for:
  - Health: `GET /health`
  - Registration options: `POST /authn/passkey/registration/options`
  - Registration finish: payload skeleton (informational; not runnable)
  - Login options: `POST /authn/passkey/login/options`
  - Login finish: payload skeleton (informational; not runnable)
  - Account key: `GET /me/account_key` (requires cookie `sid` from login finish)
  - Tx list: `GET /tx/list` (requires cookie)
  - Tx signing options: `POST /tx/signing/options` (requires cookie)
  - Tx signing finish: payload skeleton (informational; not runnable)
- Include environment variable hints for `PORT` and base URL.

## Description files
- N/A (docs only).

## Request/response shape
- Show minimal request/response JSON bodies and important headers. Mark non-runnable ceremony steps as examples only.

## Algorithm
- Author concise curl commands using `-sS` and appropriate headers.
- Use `--cookie-jar` and `--cookie` to carry the session cookie between calls where relevant.

## Database interactions
- None.

## Policies & limits
- Keep examples privacy-safe; do not include secrets or raw binary beyond base64url examples.

## Sequencing
- None; standalone documentation.

## Tests
- Manual copy/paste from `docs/api-examples.md` to confirm commands render and basic endpoints respond.

## Verification
- `test -f docs/api-examples.md` returns 0.
- Spot-check `GET /health` example works when the server is running.

## User verification commands
```bash
# Verify docs file exists
test -f docs/api-examples.md && echo OK
# Health example
curl -sS localhost:8080/health -i
```

## Acceptance criteria
- `docs/api-examples.md` added with clear, runnable curls for non-ceremony endpoints and skeletons for ceremony finishes.
- Examples rely on defaults (`PORT=8080`) and include cookie handling guidance.

## Notes
- Ceremony finishes require the browser and are intentionally documented as non-runnable shapes.

## Refs
- Refs: requirement R-PLAT-2; requirement R-OPS-DEV

