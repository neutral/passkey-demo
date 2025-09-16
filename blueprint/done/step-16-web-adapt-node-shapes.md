### Step 16 — Web: adapt to Node server option/finish shapes

Scope

- Align Register/Login flows with Node `@simplewebauthn/server` JSON (camelCase, top-level options) while preserving `reg_session_id`/`login_session_id` metadata for finish requests.
- Update dashboard transaction signing to consume Node `/tx/signing/options` payloads (nested `options` object) and continue building finish payloads that bind to `tx_session_id`.
- Route API calls through `apiFetch`/`postJson`, confirm `API_BASE` points at the Node server origin, and surface structured `ApiError` details while always sending credentials for session-bound requests.
- Refresh Playwright config + targeted frontend tests to launch the Node server and assert the updated response shapes, error handling, and manual flows end-to-end.

Source to add/modify

- `web/src/config.ts` — keep the absolute Node API origin default, document ORIGIN expectations, and ensure helper guidance references credentialed CORS.
- `web/src/lib/webauthn.ts` — update adapters/types for Node camelCase JSON, add a `TxOptionsResponse` helper that unwraps nested `options`, and retain compatibility guards for any legacy snake_case payloads.
- `web/src/pages/Register.tsx` — switch to `postJson` for options/finish, parse the new option shape, and funnel failures through `ApiError` toasts.
- `web/src/pages/Login.tsx` — mirror the Register changes for assertion options/finish, enforcing credentialed fetches for the session cookie.
- `web/src/pages/Dashboard.tsx` — adapt signing flow to the Node `options` object, reuse the new helper to build `PublicKeyCredentialRequestOptions`, and tighten error handling (401 vs mismatch vs rate limits).
- `web/tests/register-adapter.spec.ts` — update fixtures to Node camelCase option JSON and validate adapter output.
- `web/tests/login-adapter.spec.ts` — same for assertion options.
- `web/tests/register-post-body.spec.ts` & `web/tests/login-post-body.spec.ts` — stub Node responses, assert finish payloads still include session ids + base64url fields.
- `web/tests/error-toasts.spec.ts` — refresh mocked responses (CamelCase, nested tx options) and keep toast expectations intact.
- `web/tests/tx-signing.spec.ts` & `web/tests/tx-signing-negative.spec.ts` — update mocked `/tx/signing/options` payloads to Node shape and verify helper usage + error messages.
- `web/playwright.config.ts` — replace the Go backend command with the Node server startup (`node src/server.js`) and keep timeouts/ports consistent.

Description files

- `web/src/config.ts.desc.md`
- `web/src/lib/webauthn.ts.desc.md`
- `web/src/pages/Register.tsx.desc.md`
- `web/src/pages/Login.tsx.desc.md`
- `web/src/pages/Dashboard.tsx.desc.md`
- `web/tests/register-adapter.spec.ts.desc.md`
- `web/tests/login-adapter.spec.ts.desc.md`
- `web/tests/register-post-body.spec.ts.desc.md`
- `web/tests/login-post-body.spec.ts.desc.md`
- `web/tests/error-toasts.spec.ts.desc.md`
- `web/tests/tx-signing.spec.ts.desc.md`
- `web/tests/tx-signing-negative.spec.ts.desc.md`
- `web/playwright.config.ts.desc.md` (new) — capture dual dev-server startup and Node API expectations.

Blueprint updates

- `blueprint/_user-flows/registration.md` — update response examples to the Node camelCase JSON (no nested `options`), keeping session metadata.
- `blueprint/_user-flows/login.md` — same adjustments for assertion options/finish and cookie expectations.
- `blueprint/_user-flows/transaction-signing.md` — note the nested `options` object and CamelCase fields (`rpId`, `userVerification`, `allowCredentials`).
- `blueprint/global/frontend-minimal-react/_specs/frontend-api-base-and-cors.md` — swap Go references for Node startup, emphasize `apiFetch` + credentialed CORS.
- `blueprint/global/easy-local-dev/_specs/spec.md` — revise local dev narrative to “Vite + Node server,” keeping ORIGIN/RP guidance.
- `blueprint/features/registration-flow/_specs/frontend-mapping-and-pitfalls.md` — describe mapping from Node JSON (camelCase `rp.id`, direct `user` block) instead of legacy `options.*` keys.
- `blueprint/features/login-flow/_specs/allowcredentials-semantics-and-omission-guidelines.md` — refresh examples to use `allowCredentials` camelCase and the new helper pipeline.
- `blueprint/features/login-flow/_specs/allowcredentials-and-discoverable-credentials-explainer.md` — mirror the naming updates so non-dev guidance matches the implemented API.

Request/response shape

- `POST /authn/passkey/registration/options` → 200 with `{ reg_session_id, expires_at }` plus the raw `PublicKeyCredentialCreationOptionsJSON` fields: `challenge` (base64url), `rp` (`{ id, name }`), `user` (`id` base64url, names), `pubKeyCredParams` (`alg: -7`), `authenticatorSelection` (`residentKey: 'required', userVerification: 'required'`), `attestation: 'none'`, optional `timeout`.
- `POST /authn/passkey/registration/finish` → `{ reg_session_id, id, rawId, type, response.{ attestationObject, clientDataJSON } }` base64url strings; expect 201 `{ account_thumb_hex, credential_id_b64 }`.
- `POST /authn/passkey/login/options` → 200 with `{ login_session_id, expires_at, challenge, rpId, userVerification, allowCredentials?, timeout? }`; each `allowCredentials[].id` is base64url.
- `POST /authn/passkey/login/finish` → `{ login_session_id, id, rawId, type, response.{ authenticatorData, clientDataJSON, signature, userHandle? } }`; must send cookies (`credentials: 'include'`), expect 200 `{ account_thumb_hex, credential_id_b64 }` and `Set-Cookie: sid`.
- `POST /tx/signing/options` → 200 `{ tx_session_id, challenge, tx_id_hex, expires_at, options: { challenge, rpId, userVerification, allowCredentials[], timeout?, origin } }`; strip `origin` before passing to `navigator.credentials.get`.
- `POST /tx/signing/finish` → `{ tx_session_id, id, rawId, type, response.{ authenticatorData, clientDataJSON, signature, userHandle } }`; expect 201 `{ tx_id_hex }`.

Algorithm

- Registration: use `postJson` to fetch options, store `reg_session_id` + expiry, ensure `user.id` populated (generate if missing), call `startRegistration(optionsJSON)`, then `postJson` finish with the untouched `RegistrationResponseJSON` + session id; surface `ApiError` (code + correlation) on failures.
- Login: `postJson` options, keep `login_session_id`, call `startAuthentication` with `toRequestOptionsJSON`, then `postJson` finish while sending credentials so cookies persist; translate 401 vs 403 vs 409 via `ApiError` for toasts.
- Transaction signing: `postJson` `/tx/signing/options`, feed the nested `options` through the new helper that removes `origin`, decodes `allowCredentials`, and returns `PublicKeyCredentialRequestOptions`; perform `navigator.credentials.get`, build finish payload via `buildTxFinish`, `postJson` finish, refresh list/nonce and reset bundle state.
- Consolidate API usage through `apiFetch` so every request sets `mode: 'cors'` + `credentials: 'include'`, automatically parsing the HTTP error envelope.

Database interactions

- None directly in the frontend; all persistence occurs via the Node backend endpoints exercised above.

Policies & limits

- Preserve UV/resident-key policy by asserting `authenticatorSelection.userVerification === 'required'` and `residentKey === 'required'` in adapters/tests.
- Ensure `allowCredentials` omission still results in discoverable credential UX; when present, only include non-empty IDs after base64url decode.
- Respect session TTL (5 minutes for reg/login, 5 minutes for tx) by continuing to send stored session ids once and clearing them after finish.
- Maintain 2 KB transaction bundle limit feedback: show 413/429 errors surfaced by the API (covered in tests).
- Always send credentialed requests for finish/list endpoints to honor SameSite=Lax session cookie policy.

Sequencing

- Depends on Steps 1–15 (Node backend endpoints + structured logging) already merged; ensure Node server builds cleanly before modifying the web client.
- Install frontend deps (`npm -C web ci`) and Playwright browsers before running tests; Node server must be available on :8080 when executing E2E specs.
- Coordinate with forthcoming Step 17 (golden parity) so shared helper changes remain backward compatible.

Tests

- Playwright unit-style specs: `npx playwright test register-adapter.spec.ts login-adapter.spec.ts register-post-body.spec.ts login-post-body.spec.ts`.
- Playwright dashboard specs: `npx playwright test tx-signing.spec.ts tx-signing-negative.spec.ts error-toasts.spec.ts` (ensures error mapping + bundle flow).
- Chromium E2E: `npx playwright test login-e2e.chromium.spec.ts webauthn-e2e.chromium.spec.ts` once Node server command updated.
- Optional smoke: `npm -C web run build` to ensure TypeScript succeeds.
- Repo policy: `bash tools/desc-check.sh HEAD~1 HEAD` after staging to confirm description coverage.

Verification

- Confirm `npm -C web run build` succeeds after code updates.
- Run the Playwright suites above until green; rerun targeted files after any fixes.
- Manual smoke: with Node server + Vite running, register, log in (cookie set), sign a transaction, and verify the list updates + logs show expected correlation ids.

User verification commands

```bash
npm -C node-server ci
npm -C web ci
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=node-server/demo.db node node-server/src/server.js &
API_PID=$!
npm -C web run dev -- --host &
WEB_PID=$!
npx playwright test register-adapter.spec.ts login-adapter.spec.ts register-post-body.spec.ts login-post-body.spec.ts
npx playwright test tx-signing.spec.ts tx-signing-negative.spec.ts error-toasts.spec.ts
npx playwright test login-e2e.chromium.spec.ts webauthn-e2e.chromium.spec.ts
# Manually visit http://localhost:5173 to register, log in, and sign a transaction
kill $WEB_PID || true
kill $API_PID || true
```

Acceptance criteria

- Web client consumes Node backend API shapes without interim mapping hacks; adapters/tests assert camelCase fields and session ids.
- Playwright suites pass with the Node server started via updated config, and manual flows succeed against the Node backend.
- All touched source files have synchronized `.desc.md`, and blueprint artifacts document the new shapes/policies.

Notes

- Keep adapters resilient by tolerating legacy snake_case keys during transition (fallback until all stubs updated).
- When updating Playwright config, ensure CI still reuses servers on reruns (respect existing `reuseExistingServer` flags).
- Coordinate with Step 17 to avoid duplicate golden-check scripts referencing outdated Go behavior.

Refs: goal passkey-registration-login-uv; goal server-derived-challenge-and-txid; requirement R-PLAT-1; requirement R-OPS-DEV; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations; decision http-error-envelope
