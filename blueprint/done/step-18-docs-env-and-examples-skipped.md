# Step 18 — Docs, env, and examples (Skipped)

## Completion
- Date: 2025-09-16
- Outcome: Skipped after verification; existing docs already match Node backend workflows.

## Purpose
- Record documentation audit results affirming parity with Node server configuration and tooling.

### Step 18 — Docs, env, and examples

#### Scope
- Audit developer-facing references (.env sample, README quickstart, docs/api-examples, REST/Postman assets) to ensure they match the Node backend flows and config; only modify files when mismatches are found.
- Confirm instructions cover both npm-based workflow and Makefile shortcuts, including how to run the Node server + Vite UI together and reset the SQLite DB.

#### Source to add/modify
- `.env.example`: align comments/defaults with current Node server configuration (RP_ID, ORIGIN, PORT, DB_PATH, optional logging).
- `README.md`: update configuration, run, tests, and troubleshooting sections to reference Node-specific commands, make targets, and cookie/WebAuthn caveats.
- `docs/api-examples.md`, `docs/rest-client/api.http`: refresh sample requests/responses to reflect Node API shapes (session ids, challenge fields, error envelopes).
- `docs/postman/passkey-demo.postman_collection.json`, `docs/postman/local.postman_environment.json`: ensure base URLs, auth flow, and example bodies align with new backend.

#### Description files
- None (documents only; no related `.desc.md` files required).

#### Blueprint updates
- Update `blueprint/global/easy-local-dev/requirement.md` acceptance criteria to note that env samples + docs describe the Node server workflow.
- Extend `blueprint/global/easy-local-dev/_specs/spec.md` testing/documentation strategy to reference the refreshed README/API assets.
- No ADR/user-flow changes expected; document any discovered gaps in existing artifacts instead of creating new ADRs.

#### Request/response shape
- Capture canonical examples for `/authn/passkey/registration|login/options`, `/tx/signing/options`, and `/tx/signing/finish` showing: session id (base64url), challenge (base64url), `tx_id_hex`, and WebAuthn option fields (`rpId`, `userVerification`, `allowCredentials`).
- Include reminder that finish payloads mirror WebAuthn `PublicKeyCredential` JSON produced by the browser.

#### Algorithm
- Spin up Node server + Vite UI (using `make run` or explicit npm scripts).
- Exercise flows via browser or curl to capture real responses for docs; ensure cookies (sid) persist when fetching account key/transactions.
- Document reset steps: stop services, delete `node-server/demo.db` (or dedicated docs DB), restart.

#### Database interactions
- For documentation capture, use a throwaway SQLite file (e.g., `node-server/docs.db`) to avoid polluting primary dev data; include clean-up guidance in docs.

#### Policies & limits
- Emphasize UV-required configuration, secure cookie attributes, and request size limits (link back to Step 14 error envelope/limits behavior).
- Note golden vector dependency for transaction signing (bundle must match canonical CBOR).

#### Sequencing
- Review `.env.example` and README first before touching examples to ensure terminology consistency.
- Update API examples/Postman/REST client after verifying actual responses, so artifacts stay in sync.
- Re-run quickstart instructions end-to-end after edits to validate accuracy before closing the step.

#### Tests
- Happy path: follow README quickstart (copy `.env.example`, `make run`, register/login/sign) while noting any documentation gaps.
- Negative: intentionally misconfigure env (e.g., mismatch `ORIGIN`) to ensure README troubleshooting covers error responses.
- Command checks: `npm -C node-server test`, `npm -C web test` (or `npm run test:ui --prefix web`) mentioned where appropriate to validate docs.

#### Verification
- Execute README quickstart verbatim; record discrepancies for doc updates, rerun until the flow succeeds without ad-hoc fixes.
- Run curl commands from `docs/api-examples.md` against live server to confirm response shapes; update samples if diffs observed.
- Validate Postman collection imports and environment values, adjusting documentation upon success.

#### User verification commands
```bash
# Prep environment
cp -n .env.example .env || true
set -a; source .env 2>/dev/null || true; set +a

# Start services (separate terminals if preferred)
make run &
APP_PID=$!

# Exercise API samples
curl -sS localhost:8080/health
curl -sS -X POST localhost:8080/authn/passkey/registration/options -H 'Content-Type: application/json'

# Shutdown
kill $APP_PID || true
```

#### Acceptance criteria
- `.env.example`, README, and API docs mirror the Node server commands/config with no references to the deprecated Go backend.
- Postman/REST Client artifacts operate against the Node API without manual tweaks beyond documented steps.
- Blueprint requirement/spec updates reflect that documentation coverage satisfies R-OPS-DEV.

#### Notes
- Audit on 2025-09-16 found README/.env.example/API docs already aligned with Node server; no edits required.
- Verification commands executed (health + registration options curls) succeeded against temporary DB `node-server/docs.db`.
- Node unit suite currently reports `SQLITE_CONSTRAINT_PRIMARYKEY` in session tests due to shared `:memory:` path; tracked separately (pre-existing).
- Coordinate with Step 17 log if additional verification scripts are referenced (avoid duplication).

#### Refs
Refs: requirement R-OPS-DEV; step-41/42/43 analogs; decision http-error-envelope

Refs: requirement R-OPS-DEV; step-41/42/43 analogs; decision http-error-envelope
