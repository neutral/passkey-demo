# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Done (One liner)

Completed steps are stored as individual files under `blueprint/done/`, with phase information prefixed in each filename (e.g., `phase-a-step-1-...md`). Simply add a single line summary here for all done steps with their number and link to the done file.

- [Step 1 — Initialize repo](blueprint/done/phase-a-step-1-initialize-repo.md)
- [Step 2 — Server Go module](blueprint/done/phase-a-step-2-server-go-module.md)
- [Step 3 — Add deps](blueprint/done/phase-a-step-3-add-deps.md)
- [Step 4 — Web app scaffold](blueprint/done/phase-a-step-4-web-app-scaffold.md)
- [Step 5 — Server config struct](blueprint/done/phase-b-step-5-server-config-struct.md)
- [Step 6 — DB init & migrations](blueprint/done/phase-b-step-6-db-init-and-migrations.md)
- [Step 7 — Types: COSE, WebAuthn, Bundle](blueprint/done/phase-b-step-7-types-cose-webauthn-bundle.md)
- [Step 8 — COSE → ECDSA helper](blueprint/done/phase-b-step-8-cose-to-ecdsa-helper.md)
- [Step 9 — Base64url utilities](blueprint/done/phase-b-step-9-base64url-utilities.md)
- [Step 10 — CBOR canonical codec](blueprint/done/phase-b-step-10-cbor-canonical-codec.md)
- [Step 11 — Parse authenticatorData](blueprint/done/phase-c-step-11-parse-authenticatordata.md)
- [Step 12 — ClientDataJSON validation](blueprint/done/phase-c-step-12-clientdatajson-validation.md)
- [Step 13 — Signature verify utility](blueprint/done/phase-c-step-13-signature-verify-utility.md)
- [Step 14 — RP ID & Origin checks](blueprint/done/phase-c-step-14-rp-origin-checks.md)
- [Step 15 — /authn/passkey/registration/options handler](blueprint/done/phase-d-step-15-registration-options-handler.md)
- [Step 16 — Attestation parsing (minimal)](blueprint/done/phase-d-step-16-attestation-parsing.md)
- [Step 17 — /authn/passkey/registration/finish handler](blueprint/done/phase-d-step-17-registration-finish-handler.md)
- [Step 18 — /authn/passkey/login/options handler](blueprint/done/phase-e-step-18-login-options-handler.md)
- [Step 19 — /authn/passkey/login/finish handler](blueprint/done/phase-e-step-19-login-finish-handler.md)
- [Step 20 — Bundle validation helper](blueprint/done/phase-f-step-20-bundle-validation-helper.md)
- [Step 21 — /tx/signing/options handler](blueprint/done/phase-f-step-21-signing-options-handler.md)
- [Step 22 — /tx/signing/finish handler](blueprint/done/phase-f-step-22-signing-finish-handler.md)
- [Step 23 — /tx/list handler](blueprint/done/phase-f-step-23-tx-list-handler.md)
- [Step 24 — Session middleware](blueprint/done/phase-g-step-24-session-middleware.md)
- [Step 25 — Rate limiting & limits](blueprint/done/phase-g-step-25-rate-limiting-limits.md)
- [Step 26 — CORS & cookies](blueprint/done/phase-g-step-26-cors-and-cookies.md)
- [Step 27 — Basic pages](blueprint/done/phase-h-step-27-basic-pages.md)
- [Step 28 — Base64url helpers (web)](blueprint/done/phase-h-step-28-base64url-helpers-web.md)
- [Step 29 — WebAuthn create() flow (Register)](blueprint/done/phase-h-step-29-webauthn-register.md)
- [Step 30 — WebAuthn get() flow (Login)](blueprint/done/phase-h-step-30-webauthn-login.md)
- [Step 31 — Dashboard: fetch list](blueprint/done/phase-h-step-31-dashboard-fetch-list.md)
- [Step 32 — Dashboard: build bundle](blueprint/done/phase-h-step-32-dashboard-build-bundle.md)
- [Step 33 — Dashboard: transaction signing](blueprint/done/phase-h-step-33-dashboard-transaction-signing.md)
- [Step 34 — Error Toasts](blueprint/done/phase-h-step-34-error-toasts.md)
- [Step 35 — Go Unit Tests](blueprint/done/phase-i-step-35-go-unit-tests.md)
- [Step 36 — Golden Vectors](blueprint/done/phase-i-step-36-golden-vectors.md)
- [Step 37 — Manual E2E](blueprint/done/phase-i-step-37-manual-e2e.md)
- [Step 37b — Refactoring](blueprint/done/phase-i-step-37b-refactoring-plan.md)
- [Step 37c — Router Builder, Session‑Only Auth, Error Envelope, TTL Stores, Repos, API Client](blueprint/done/phase-i-step-37c-fixes.md)
- [Step 37d — Robust Nonce Handling, High‑S Acceptance, Auto‑Filled Nonce UX](blueprint/done/phase-i-step-37d-fixes.md)

## Next

### Phase J — Hardening (Demo-grade)

---

39. **Session security**

    Scope

    - Establish demo-grade authenticated sessions: strong random IDs (≥128 bits), 1h expiry, sliding renewal on activity via middleware.
    - Issue cookie `sid` with `HttpOnly`, `SameSite=Lax`, and `Secure` only when `ORIGIN` is https; no client storage.
    - Ensure protected handlers rely on session middleware context; expired/missing sessions yield 401.

    Source to add/modify

    - Modify `server/internal/webauthn/login_finish.go`: confirm `sid` generation uses CSPRNG (≥24 bytes), DB insert writes `expires_at = now+1h`, and `setSessionCookie` sets `HttpOnly`, `SameSite=Lax`, `Secure` toggled by scheme.
    - Modify `server/internal/http/session.go`: keep sliding TTL refresh when less than half the TTL remains; do not attach session when expired/missing; load credential IDs.
    - Modify `server/internal/app/router.go`: ensure global `SessionMiddleware(db, true, time.Hour)` wraps groups before handlers.
    - Tests to add/update:
      - Add `server/internal/webauthn/login_finish_session_test.go`: assert DB `expires_at ≈ now+3600s` (±5s) and cookie attributes after login.
      - Keep `server/internal/http/session_test.go`: missing/expired cookie → 401 in protected next handler.
      - Keep `server/internal/http/session_refresh_test.go`: expiry extends on activity when < TTL/2 remains.

    Description files

    - Update `server/internal/webauthn/login_finish.go.desc.md`: call out 1h session TTL, cookie attributes, and DB session row fields.
    - Update `server/internal/http/session.go.desc.md`: note sliding refresh policy and that it attaches context only for valid sessions; mention 1h TTL default.
    - Update `server/internal/app/router.go.desc.md`: explicitly record middleware order and that session TTL is 1h with refresh enabled.
    - Ensure each updated description ends with accurate `Refs:` (see Refs below).

    Request/response shape

    - `POST /authn/passkey/login/finish` (on success):
      - Headers: `Set-Cookie: sid=<base64url>; HttpOnly; SameSite=Lax[; Secure]` (Secure only for https origin).
      - JSON: `{ account_thumb_hex: string, credential_id_b64: string }`.
    - Protected routes (e.g., `GET /me/account_key`, `/tx/*`) require cookie `sid=<value>`; no request body fields for session.

    Algorithm

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

    Database interactions

    - Insert on login: `INSERT INTO sessions(session_id, acct_cbor, expires_at, created_at)`.
    - Middleware read/update: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?`; optional `UPDATE sessions SET expires_at=? WHERE session_id=?`.
    - Foreign keys: `sessions.acct_cbor` references `accounts.acct_cbor` (ON DELETE CASCADE).

    Policies & limits

    - ID strength: `sidRaw` length ≥ 16 bytes; use 24 bytes (192 bits) CSPRNG.
    - TTL: 1 hour fixed; sliding refresh when < 30 minutes remain.
    - Cookie: `HttpOnly`, `SameSite=Lax`, `Secure` only for https `ORIGIN`; `Path=/`.
    - Authorization: missing/expired session → 401; do not fall back to cookie lookup in handlers.

    Sequencing

    - Do not change router layering from Step 24; keep CORS outermost, then session middleware, then group middlewares and handlers.
    - No new endpoints; focus on session issuance (login finish) and middleware behavior.
    - Keep DB schema unchanged (table `sessions` already created in migrations).

    Tests (happy path required, negative cases, invariants)

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

    Verification

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

    Acceptance criteria

    - Session IDs are CSPRNG-generated and at least 128 bits (24 bytes used).
    - New sessions expire in 1 hour (DB `expires_at` ≈ now+3600s) and renew on activity when < half TTL remains.
    - Cookie `sid` has `HttpOnly` and `SameSite=Lax`; `Secure` present only when `ORIGIN` is https.
    - Expired/missing sessions result in 401 on protected routes; re-login issues a fresh session and restores access.

    Notes

    - Keep cookie as a session cookie (no `Max-Age`/`Expires`) to align with demo goals; the DB `expires_at` still governs access.
    - No rotation of `sid` on refresh in this demo; a logout endpoint is out of scope for this step.

    Refs

    - Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-2; decision router-builder-wiring; spec session-cookies-usage-explainer

---

40. **Logging**

    - Structured logs with event names: `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`.
    - _Verify_: logs show account thumb and tx_id (hex).

---

41. **Build scripts**

    - Root Makefile: `make server`, `make web`, `make run`, `make clean`.
    - _Verify_: one command runs both.

### Phase K — Developer Experience

---

42. **API examples**

    - Add `docs/` with curl examples for each endpoint (sans WebAuthn ceremony).
    - _Verify_: docs render in repo.

---

43. **Postman / REST Client file**

    - Provide a collection with placeholders; helpful for observing JSON shapes.
    - _Verify_: collection can be imported.

---

44. **Env sample**

    - `.env.example` with `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`.
    - _Verify_: loads correctly.

---

45. **README**

    - Quickstart, limitations (no attestation trust), and demo notes (Touch ID prompts).
    - _Verify_: teammate can bootstrap in <10 minutes.

---

## Future

- Handler-level logging and error mapping integration

  - What: Wire a minimal JSON logger using Go `log/slog` in `server/cmd/api` and apply the `MapVerifyError` and `LogAssertion` utilities from `server/internal/webauthn` in the assertion-finish handler. Keep client responses generic while emitting structured, privacy-preserving logs with stable `error_kind` values from sentinel errors.
  - Why: Improves observability, incident triage, and auditability without leaking sensitive data. Cleanly separates transport concerns (HTTP codes) from cryptographic failure semantics via sentinel errors, enabling accurate metrics and alerts (e.g., spikes in `ErrMalformedDER`).
  - How: Initialize `slog` with a JSON handler for dev; in the handler, call `VerifyAssertion(...)`, map the error with `MapVerifyError`, log once via `LogAssertion` using hashed identifiers (`HashID`), and return an appropriate status code with a standard error envelope. This remains compatible with the existing plan’s later steps for endpoints and error envelopes.

- IDNA (punycode) normalization support
  - What: Normalize internationalized domain names to ASCII (punycode) for RP ID and origin comparisons using `golang.org/x/net/idna`.
  - Why: Prevent mismatches and policy bypass due to Unicode vs punycode inconsistencies; ensure consistent hashing for rpIdHash and accurate origin validation across i18n domains.
  - How: Add a deterministic normalization helper that converts Unicode hostnames to punycode before validation/comparison; guard with unit tests and vectors (e.g., `bücher.ch` ⇄ `xn--bcher-kva.ch`), and document deployment guidance to keep config values consistent.

## Acceptance Checks

- **Build**: `go build ./server/...` and `npm run build` in `/web` both succeed.
- **Register/Login**: Using macOS with Touch ID, both ceremonies prompt for fingerprint; login sets cookie.
- **Sign**: Create 2 messages with nonces 1 and 2; both appear in `/tx/list` and DB.
- **Ephemeral TTLs**: Registration/login/tx option sessions expire after 5 minutes; expired attempts return 409.
- **Replay/Nonce**: Re-submit nonce 2 → **409** conflict.
- **UV check**: If browser returns an assertion without UV (simulate by forcing options incorrectly) → **403** forbidden.
- **Origin/RP guard**: Change `origin` in request body → **403**.
- **Low‑S enforced**: Hand-craft a signature with high‑S (unit test) → **400**.

## Coverage & Refs (Traceability)

- R-FLOW-REG: Steps 11–17, 26, 38. Refs: requirement R-FLOW-REG; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails.
- R-FLOW-LOGIN: Steps 11–14, 18–19, 24, 26, 38–39. Refs: requirement R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations.
- R-FLOW-SIGN: Steps 20–23, 32–33, 26, 38. Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails.
- R-ID-KEY: Steps 7–8, 16–17, 19, 22–23. Refs: requirement R-ID-KEY; decision webauthn-corrections-and-standardizations.
- R-SCHEMA-LITE: Steps 10, 20, 32. Refs: requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails.
- R-UI-2BTN: Steps 27, 31–34. Refs: requirement R-UI-2BTN.
- R-PLAT-1: Steps 4, 27–34. Refs: requirement R-PLAT-1.
- R-PLAT-2: Steps 5, 11–26, 38–40. Refs: requirement R-PLAT-2.
- R-PLAT-3: Steps 6, 17, 19, 22–23, 34–35. Refs: requirement R-PLAT-3.
- R-OPS-DEV: Steps 5, 26, 41, 44–45. Refs: requirement R-OPS-DEV.
- R-ERR: Steps 25, 34, 38, 40. Refs: requirement R-ERR; decision encoding-and-ceremony-guardrails.
- R-NO-BROKER: Entire plan avoids brokers; synchronous calls. Refs: requirement R-NO-BROKER.
- R-PORTABLE: Steps 26, 27–34, 37. Refs: requirement R-PORTABLE.
- R-SEC-UV: Steps 14–15, 17–19, 21–22. Refs: requirement R-SEC-UV; decision webauthn-corrections-and-standardizations.
