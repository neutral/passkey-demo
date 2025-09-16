# Implementation Plan — simplewebauthn + Node Server

## Purpose

- Migrate the Vite+React web app to use `@simplewebauthn/browser` for WebAuthn flows.
- Add a new `node-server/` that replaces the original Go backend using `@simplewebauthn/server`, matching functionality and policies from `blueprint/done`.
- Preserve data model, security policies (UV required), cookies/CORS, error envelopes, logging, and transaction signing semantics (bundle anchors: CHALv1/TXIDv1).

## Notes

- Keep API shapes JSON-friendly; the browser lib consumes `PublicKeyCredential*OptionsJSON`.
- Parity targets the endpoints and behaviors implemented by done steps: registration, login, session, CORS, limits, logging, error envelopes, transaction signing (options/finish/list), and account key exposure.
- Environment variables retain names from Go server: `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`, `RP_ID_ALLOWLIST`, `ORIGIN_ALLOWLIST`.

## Done

- Step 16 — Web: adapt to Node server option/finish shapes — see `blueprint/done/step-16-web-adapt-node-shapes.md`
- Step Refactor-16 — Web: remove legacy Go backend compatibility — see `blueprint/done/step-refactor-16-web-remove-go-compat.md`
- Step Refactor-17 — Repo: remove Go server artifacts — see `blueprint/done/step-refactor-17-remove-go-server.md`

- Step 1 — Web: adopt `@simplewebauthn/browser` for Register/Login — see `blueprint/done/step-1-web-adopt-simplewebauthn-browser.md`
- Step 2 — Node: scaffold `node-server/` (Express + SQLite + Pino) — see `blueprint/done/step-2-node-scaffold-express-sqlite-pino.md`
- Step 3 — Node: DB schema and config parity — see `blueprint/done/step-3-node-db-schema-and-config-parity.md`
- Step 4 — Node: CORS, cookies, and session middleware — see `blueprint/done/step-4-node-cors-cookies-and-session-middleware.md`
- Step 5 — Node: Registration options (`/authn/passkey/registration/options`) — see `blueprint/done/step-5-node-registration-options.md`
- Step 6 — Node: Registration finish (`/authn/passkey/registration/finish`) — see `blueprint/done/step-6-node-registration-finish.md`
- Step 7 — Node: Login options (`/authn/passkey/login/options`) — see `blueprint/done/step-7-node-login-options.md`
- Step 8 — Node: Login finish (`/authn/passkey/login/finish`) — see `blueprint/done/step-8-node-login-finish.md`
- Step 9 — Node: GET `/me/account_key` (authenticated) — see `blueprint/done/step-9-node-get-me-account-key.md`
- Step 10 — Node: Bundle anchors and validation helpers — see `blueprint/done/step-10-node-bundle-anchors-and-validation-helpers.md`
- Step 11 — Node: Tx signing options (`/tx/signing/options`) — see `blueprint/done/step-11-node-tx-signing-options.md`
- Step 12 — Node: Tx signing finish (`/tx/signing/finish`) — see `blueprint/done/step-12-node-tx-signing-finish.md`
- Step 13 — Node: Tx list (`/tx/list`) — see `blueprint/done/step-13-node-tx-list.md`
- Step 14 — Node: Error envelopes and limits — see `blueprint/done/step-14-node-error-envelopes-and-limits.md`

---

### Step 17 — E2E and golden parity

Scope

- Add Node-side test script to validate bundle golden vectors (`specs/goldens/tx-bundle-v1.json`).
- Smoke E2E via Playwright: run Node server and Vite; Virtual Authenticator; assert all flows.

Source to add

- `node-server/scripts/golden-check.js` reading golden JSON and comparing outputs.
- `web/playwright.config.ts` updates to run Node server for E2E project.

Verification

- `node node-server/scripts/golden-check.js` passes; Playwright E2E green.

Acceptance criteria

- Bundle anchors and flows verified end-to-end against goldens.

Refs: requirement R-FLOW-SIGN; spec golden-vectors; spec manual-e2e

---

### Step 18 — Docs, env, and examples

Scope

- Update `.env.example` with Node server variables; update README with Node start commands.
- Provide REST client examples for Node endpoints; ensure Postman or REST Client entries reflect JSON shapes.

Source to add/modify

- `node-server/README.md`, `.env.example` additions; `docs/` updates; API examples under `docs/`.

Verification

- Quick curl scripts work against Node server; README steps reproduce register → login → sign → list.

Acceptance criteria

- Documentation parity and clear local dev instructions.

Refs: requirement R-OPS-DEV; step-41/42/43 analogs; decision http-error-envelope

---

## User Verification Commands (after all steps)

```bash
# 1) Install deps
npm -C web ci || npm -C web i
npm -C node-server ci || npm -C node-server i

# 2) Start Node server (dev)
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=node-server/demo.db \
  node node-server/src/server.js &
API_PID=$!

# 3) Start Vite
npm -C web run dev &
WEB_PID=$!

# 4) Manual E2E
# Visit http://localhost:5173 → Register → Login → Dashboard sign → See list item

# 5) Golden check
node node-server/scripts/golden-check.js

# 6) Cleanup
kill $WEB_PID || true
kill $API_PID || true
```

## Refs

Refs: goal passkey-registration-login-uv; goal server-derived-challenge-and-txid; goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-PLAT-3; requirement R-SEC-UV; requirement R-ERR; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-OPS-DEV; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision http-error-envelope; decision request-id-and-slog-json; decision cbor-cose-interop-and-decoding-fallbacks; decision webauthn-corrections-and-standardizations; decision webauthn-accept-high-s-login-only; decision webauthn-accept-high-s-signing-too
