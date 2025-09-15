# Step 9 — Node: GET `/me/account_key` (Done)

Completed: 2025-09-15
Verification notes:
- `npm -C node-server test`

### Step 9 — Node: GET `/me/account_key` (authenticated)

Scope

- Implement GET `/me/account_key` to return the authenticated account’s COSE public key (canonical CBOR) and helper fields (thumbs, coordinates) for the dashboard.
- Require a valid session cookie (`sid`); respond with JSON envelope 401 when missing/expired.
- Maintain parity with Go’s response shape: base64url `acct_cbor_b64`, lowercase hex `account_thumb_hex`, hexified x/y coordinates, `created_at` for display.
- Update web dashboard to consume the Node response without additional reshaping.

Source to add/modify

- Add `node-server/src/me.js` containing an Express router (or handler) that reads `req.session`, queries the DB, and formats the response.
- Update `node-server/src/server.js` to mount the route under `/me` after session middleware.
- Add `node-server/test/me-account-key.test.js` covering authenticated success, unauthorized (no session), and missing account edge case.
- Update web dashboard (`web/src/pages/Dashboard.tsx`, tests) to handle Node JSON (if different from Go output).

Description files

- Add `node-server/src/me.js.desc.md` describing the route responsibilities, DB reads, logging, and error mapping.
- Update `node-server/src/server.js.desc.md` noting the new `/me` router.
- Update `node-server/src/webauthn/webauthn.desc.md` if shared helpers (e.g., thumb hashing) move/expand.
- Update `web/src/pages/Dashboard.tsx.desc.md` documenting Node server response expectations.

Blueprint updates

- Update `blueprint/features/passkey-first-identity/_specs/spec.md` (or relevant doc) with the Node `/me/account_key` response structure and session requirements.
- Update `blueprint/features/passkey-first-identity/requirement.md` acceptance criteria to mention Node parity (same response fields, secure session usage).
- If a frontend mapping doc exists, note that Node returns canonical JSON matching Go.

Request/response shape

- Request: GET `/me/account_key` with `Cookie: sid=<session>`; expect JSON.
- Response (200): `{ account_thumb_hex, acct_cbor_b64, pubkey_x_hex, pubkey_y_hex, created_at }`.
- Error: 401 envelope `{ code: 'unauthorized', error: 'Unauthorized', correlation_id? }` when session invalid; 404 or 500 for missing account/DB failure as defined.

Algorithm

- Verify `req.session` exists (populated by middleware). If absent/expired, return 401 envelope.
- Query `accounts` by `acct_cbor` (or `acct_thumb`) from session; optionally derive x/y coordinates via COSE decode.
- Compute thumb (if not stored) using `SHA-256('ACCTK1'||acct_cbor)`; convert to hex.
- Respond with JSON matching Go format; include `created_at` from DB.
- Optional: log `me_account_key` access with hashed identifiers.

Database interactions

- `SELECT acct_cbor, acct_thumb, created_at FROM accounts WHERE acct_cbor = ?`.
- No writes.

Policies & limits

- Endpoint requires authenticated session; rely on session middleware for TTL enforcement.
- Response should not leak additional metadata beyond existing Go implementation.

Sequencing

- Depends on Step 8 (login finish) for session establishment.
- Dashboard features in later steps rely on this endpoint for account context.

Tests

- `node-server/test/me-account-key.test.js` using in-memory DB:
  - *Happy path*: seed account + session, request with cookie, assert 200 JSON fields.
  - *Unauthorized*: no cookie or unknown session → 401 envelope.
  - *Account missing*: expect 404 or 500 (choose behavior, e.g., 401 to avoid leaks).
- Update web tests (if any) to verify dashboard fetch uses Node API.
- Command: `npm -C node-server test`; `npm -C web test:ui` if dashboard flows covered.

Verification

- Manual: start Node server, login via web, hit `/me/account_key` with browser (should populate dashboard); or curl with `sid` cookie.
- Confirm response fields match expected keys/format.

Acceptance criteria

- Endpoint returns account key data for authenticated sessions; 401 envelope otherwise.
- Web dashboard works with Node backend.
- Blueprint/docs updated accordingly.

Notes

- Consider centralizing COSE decoding helpers if used by other endpoints.
- Ensure logging (if added) aligns with Step 15 structured logging fields.

Refs: goal passkey-registration-login-uv; goal simple-ui-and-storage; requirement R-PLAT-3; requirement R-UI-2BTN; spec passkey-first-identity/spec.md; decision request-id-and-slog-json; decision http-error-envelope
