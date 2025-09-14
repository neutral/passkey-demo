# Step 39 — Session security (Done: 2025-09-12)

## Verification Notes
- Short-mode tests pass for http/webauthn/tx.
- Added `TestLoginFinish_SetsOneHourExpiry`: DB `expires_at ≈ now+3600s` and cookie issued.
- Added `TestLoginFinish_ReLoginAfterExpiry_Succeeds`: session expired in DB → fresh login issues new `sid`.

## Scope
- Establish demo-grade authenticated sessions: strong random IDs (≥128 bits), 1h expiry, sliding renewal on activity via middleware.
- Issue cookie `sid` with `HttpOnly`, `SameSite=Lax`, and `Secure` only when `ORIGIN` is https; no client storage.
- Ensure protected handlers rely on session middleware context; expired/missing sessions yield 401.

## Source to add/modify
- Modify `server/internal/webauthn/login_finish.go`: confirm `sid` generation uses CSPRNG (≥24 bytes), DB insert writes `expires_at = now+1h`, and `setSessionCookie` sets `HttpOnly`, `SameSite=Lax`, `Secure` toggled by scheme.
- Modify `server/internal/http/session.go`: keep sliding TTL refresh when less than half the TTL remains; do not attach session when expired/missing; load credential IDs.
- Modify `server/internal/app/router.go`: ensure global `SessionMiddleware(db, true, time.Hour)` wraps groups before handlers.
- Tests to add/update:
  - Add `server/internal/webauthn/login_finish_session_test.go`: assert DB `expires_at ≈ now+3600s` (±5s) and cookie attributes after login.
  - Keep `server/internal/http/session_test.go`: missing/expired cookie → 401 in protected next handler.
  - Keep `server/internal/http/session_refresh_test.go`: expiry extends on activity when < TTL/2 remains.

## Description files
- Update `server/internal/webauthn/login_finish.go.desc.md`: call out 1h session TTL, cookie attributes, and DB session row fields.
- Update `server/internal/http/session.go.desc.md`: note sliding refresh policy and that it attaches context only for valid sessions; mention 1h TTL default.
- Update `server/internal/app/router.go.desc.md`: explicitly record middleware order and that session TTL is 1h with refresh enabled.
- Ensure each updated description ends with accurate `Refs:` (see Refs below).

## Request/response shape
- `POST /authn/passkey/login/finish` (on success):
  - Headers: `Set-Cookie: sid=<base64url>; HttpOnly; SameSite=Lax[; Secure]` (Secure only for https origin).
  - JSON: `{ account_thumb_hex: string, credential_id_b64: string }`.
- Protected routes (e.g., `GET /me/account_key`, `/tx/*`) require cookie `sid=<value>`; no request body fields for session.

## Algorithm
- Login finish:
  - Generate `sidRaw = rand(24 bytes)`; `sid = base64url(sidRaw)`.
  - Insert `sessions(session_id=sid, acct_cbor, expires_at=now+1h, created_at=now)`.
  - Set cookie `sid` with attributes above; omit `Max-Age` for session-cookie semantics.
- Middleware:
  - Read `sid` cookie; `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?`.
  - If `now >= expires_at` → do not attach context (downstream returns 401).
  - If refresh enabled and `expires_at-now < (TTL/2)` → `UPDATE sessions SET expires_at=now+TTL` (sliding window).
  - Load `credential_id` list for account (`SELECT credential_id FROM credentials WHERE acct_cbor_fk=?`).
  - Attach `SessionContext{AcctCBOR, CredentialIDs}`; downstream handlers never read cookies directly.

## Database interactions
- Insert on login: `INSERT INTO sessions(session_id, acct_cbor, expires_at, created_at)`.
- Middleware read/update: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?`; optional `UPDATE sessions SET expires_at=? WHERE session_id=?`.
- Foreign keys: `sessions.acct_cbor` references `accounts.acct_cbor` (ON DELETE CASCADE).

## Policies & limits
- ID strength: `sidRaw` length ≥ 16 bytes; use 24 bytes (192 bits) CSPRNG.
- TTL: 1 hour fixed; sliding refresh when < 30 minutes remain.
- Cookie: `HttpOnly`, `SameSite=Lax`, `Secure` only for https `ORIGIN`; `Path=/`.
- Authorization: missing/expired session → 401; do not fall back to cookie lookup in handlers.

## Sequencing
- Do not change router layering from Step 24; keep CORS outermost, then session middleware, then group middlewares and handlers.
- No new endpoints; focus on session issuance (login finish) and middleware behavior.
- Keep DB schema unchanged (table `sessions` already created in migrations).

## Tests (happy path required, negative cases, invariants)
- `server/internal/webauthn/login_finish_session_test.go`:
  - Happy: successful login sets cookie with `HttpOnly`, `SameSite=Lax`, `Secure` toggled by https; DB `expires_at` within `[now+3595, now+3605]` seconds.
  - Negative: https origin toggles `Secure`; http origin must not set `Secure`.
- `server/internal/http/session_test.go`:
  - Negative: missing cookie → next returns 401; expired session row → 401.
  - Happy: valid unexpired session attaches context with account and one credential ID.
- `server/internal/http/session_refresh_test.go`:
  - Happy: with `expires_at=now+10s`, middleware refreshes to `≈ now+1h`.
  - Optional: no refresh when `expires_at-now >= TTL/2`.
- Commands:
  - `cd server && go test ./internal/webauthn -run Test(LoginFinish|SetSessionCookie).* -v`
  - `cd server && go test ./internal/http -run TestSessionMiddleware_.* -v && go test ./internal/http -run TestSessionMiddleware_RefreshExtendsExpiry -v`

## Verification
- Unit: all new and existing tests above pass; timings use tolerances to avoid flakes.
- Manual:
  - Start server; perform a normal login via the web UI (passkey). Confirm browser receives `sid` cookie with expected attributes.
  - Visit `/me/account_key` → 200 with account key JSON.
  - Force-expire the session in DB (set `expires_at` to past) and retry `/me/account_key` with same cookie → 401.
  - Login again; `/me/account_key` → 200 (re-login works). Optionally, hit `/me/account_key` again a few minutes later and check `sessions.expires_at` extended.
- User verification commands

```bash
# Backend tests
cd server
go test ./internal/webauthn -run Test(LoginFinish|SetSessionCookie).* -v
go test ./internal/http -run TestSessionMiddleware_.* -v
go test ./internal/http -run TestSessionMiddleware_RefreshExtendsExpiry -v

# Run server (separate terminal)
go build ./cmd/api && ./cmd/api

# Manual check: after logging in via the web UI, capture the sid value from the Set-Cookie header or browser devtools.
SID="<paste-sid-from-browser>"

# Check a protected endpoint with the cookie → 200
curl -i --cookie "sid=$SID" :8080/me/account_key

# Force-expire the session in the DB (default path server/demo.db)
sqlite3 server/demo.db "UPDATE sessions SET expires_at=(strftime('%s','now')-10) WHERE session_id='$SID';"

# Retry → 401 Unauthorized
curl -i --cookie "sid=$SID" :8080/me/account_key
```

## Acceptance criteria
- Session IDs are CSPRNG-generated and at least 128 bits (24 bytes used).
- New sessions expire in 1 hour (DB `expires_at` ≈ now+3600s) and renew on activity when < half TTL remains.
- Cookie `sid` has `HttpOnly` and `SameSite=Lax`; `Secure` present only when `ORIGIN` is https.
- Expired/missing sessions result in 401 on protected routes; re-login issues a fresh session and restores access.

## Notes
- Keep cookie as a session cookie (no `Max-Age`/`Expires`) to align with demo goals; the DB `expires_at` still governs access.
- No rotation of `sid` on refresh in this demo; a logout endpoint is out of scope for this step.

## Refs
- Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-2; decision router-builder-wiring; spec session-cookies-usage-explainer

