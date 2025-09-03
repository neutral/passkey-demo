# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Phase E — Login Endpoints

19. [Done] /authn/passkey/login/finish handler — see blueprint/done/phase-e-step-19-login-finish-handler.md

### Step 19 — /authn/passkey/login/finish handler (Expanded)

Scope

- Complete the login ceremony: validate the login session, verify CDJ and AD (UV, rpIdHash), identify the account by credential ID, verify the assertion signature, enforce monotonic signCount, persist updates, and establish a server session via cookie.

Source to add/modify

- `server/internal/webauthn/login_finish.go`: Implements POST handler and helpers.
- `server/internal/webauthn/login_finish.go.desc.md`: High-level description, relations, invariants, and Refs.
- `server/internal/webauthn/login_finish_test.go`: Unit tests (negative paths and shape) and a required happy path with a generated ES256 key.
- `server/cmd/api/main.go`: Mount `POST /authn/passkey/login/finish`.
- (Reuse) `server/internal/webauthn`: `ParseClientDataJSON`, `IsGet`, `CheckOrigin`, `CheckRpIdHashAllowed`, `VerifyAssertion`, `HasUV`.
- (Reuse) `server/internal/encoding`: base64url and canonical CBOR for COSE decode.
- (Reuse) `server/internal/crypto`: COSE→ECDSA conversion for public key verification.

Description files

- `server/internal/webauthn/login_finish.go.desc.md`
  - Purpose: Verify WebAuthn assertion for login and issue app session cookie.
  - Key logic: session lookup/TTL, CDJ type=get, origin allowlist, AD rpIdHash + UV, DB lookups, signature verify, signCount update, session insert, cookie set.
  - Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; spec R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations.
- Update description files for any modified sources in this step, including:
  - `server/cmd/api/main.go.desc.md` to document the new `POST /authn/passkey/login/finish` route and relations.
  - Any other touched files’ `<filename>.<ext>.desc.md` to keep behavior and relations accurate.

Request/response shape

- Request JSON (minimal):
  - `login_session_id` (string; from options)
  - WebAuthn fields per `types.LoginFinish`:
    - `id`, `rawId`, `type`, `response.{authenticatorData, clientDataJSON, signature, userHandle}` (binary as base64url strings)
- Response: 200 JSON `{ account_thumb_hex, credential_id_b64 }`; also sets `Set-Cookie` with server session ID. Error responses return clear HTTP status codes (see below).

Algorithm

- Load login session by `login_session_id`; reject if missing/expired; single-use (delete on use or on expiration).
- Parse and validate `clientDataJSON`:
  - `type == webauthn.get`, challenge matches session, origin allowed (`CheckOrigin` with dev localhost exception).
- Parse `authenticatorData` (AD header):
  - `CheckRpIdHashAllowed(ad.rpIdHash, cfg.RP_ID, cfg.RPAllowlist)`; require `HasUV(flags)`.
- Identify account by credential ID:
  - Decode `rawId` (base64url) → `credID`.
  - Query `credentials` to get `{acct_cbor_fk, sign_count}`; if not found → 401 (avoid credential enumeration).
  - Load `accounts.acct_cbor` and decode to `types.CoseEC2`; convert to `ecdsa.PublicKey`.
- Verify signature:
  - Compute over `ad || SHA256(cdj)` using `VerifyAssertion` (low‑S, strict DER, P‑256 enforced).
- Enforce signCount policy:
  - Require `ad.signCount > stored.sign_count`; else 409 (conflict) and do not update.
  - On success, update `credentials.sign_count = ad.signCount`.
- Create server session:
  - Generate random 24‑byte `session_id`; set expiry = now + 1h; insert into `sessions (session_id, acct_cbor, expires_at, created_at)`.
  - Set cookie `sid=<session_id>` with attributes: `HttpOnly`, `Path=/`, `SameSite=Lax`, `Secure` when `cfg.Origin` is https; dev localhost may omit `Secure`.

Policies and limits

- Methods: POST only; others → 405.
- Session TTL: 5 minutes (from Step 18); consume (delete) after a successful finish attempt (regardless of outcome, delete if expired).
- UV required; UP is implied by platform flows and may be optionally checked.
- Origin policy: exact allowlist; dev localhost exception supported.
- RP ID policy: exact allowlist; no wildcards/eTLD+1.
- Error mapping:
  - 401: missing/expired login session; unknown credential; signature mismatch or high‑S; challenge mismatch.
  - 403: policy failures (rpIdHash/Origin) via `MapPolicyError`.
  - 409: signCount not strictly increasing.
  - 400: malformed inputs (bad base64/JSON/DER) via `MapVerifyError` and basic parsing checks.
  - 500: unexpected DB/internal errors.

Database interactions

- Lookup credential: `SELECT acct_cbor_fk, sign_count FROM credentials WHERE credential_id = ?`.
- Load account key: `SELECT acct_cbor FROM accounts WHERE acct_cbor = ?` (or join on previous query’s `acct_cbor_fk`).
- Update signCount: `UPDATE credentials SET sign_count = ? WHERE credential_id = ?`.
- Insert session: `INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)`.

Sequencing

- No auth middleware yet; cookie issued inline here; Step 24 will add middleware to consume DB sessions.
- CORS and cookie flags refined in Step 26; for dev over HTTP localhost, omit `Secure`.

Tests (to add)

- Happy path (required):
  - Generate an ES256 keypair; create COSE EC2 for the public key; insert `accounts` and `credentials` rows with `sign_count = n`.
  - Build a login session (store) and craft CDJ (`type=webauthn.get`, challenge from session, origin = cfg.Origin).
  - Build AD with correct `rpIdHash` for cfg.RP_ID, `flags` with UV set, and `signCount = n+1`.
  - Sign over `ad || SHA256(cdj)`; send finish payload with `rawId`, `authenticatorData`, `clientDataJSON`, `signature`.
  - Expect 200; `Set-Cookie` present; `credentials.sign_count` updated to `n+1`; login session consumed (cannot reuse); cookie attributes correct (HttpOnly, SameSite=Lax; Secure set for https origin, omitted for dev http localhost).
- Negative cases:
  - Missing/expired login session → 401.
  - CDJ type mismatch (`webauthn.create`) → 400.
  - Challenge mismatch → 401.
  - Origin not allowed (host/port/scheme) → 403.
  - rpIdHash mismatch → 403.
  - Missing UV bit → 403.
  - Unknown credential ID → 401 (avoid enumeration).
  - Non‑increasing signCount (equal or lower) → 409; DB not updated.
  - Malformed DER signature → 400 (MapVerifyError: ErrMalformedDER).
  - High‑S signature → 401 (MapVerifyError: ErrHighS).
  - Unsupported curve/alg in account key (bad COSE) → 400.
  - Base64 errors: bad `rawId`/`authenticatorData`/`clientDataJSON`/`signature` → 400.
- Policy & mapping checks:
  - Validate MapPolicyError and MapVerifyError produce expected HTTP codes for representative failures (400/401/403/409).
- Session semantics:
  - Single‑use: after a successful finish, reusing the same `login_session_id` returns 401 and store entry is gone.

Verification

- Unit tests: `cd server && GOCACHE=$(pwd)/.gocache go test ./...` → all pass.
- Manual (browser):
  - Register an account (Step 17), then call login options and finish from the browser. Expect a passkey prompt and, on success, a `Set-Cookie` header.
  - Subsequent protected endpoint (added in a later step) recognizes the session.
- Quick negative checks (curl):
  - POST finish with a bogus `login_session_id` returns 401.

User verification commands

```bash
# Run tests (after implementation)
cd server && GOCACHE=$(pwd)/.gocache go test ./...

# Start the API
go run ./cmd/api &
API_PID=$!
sleep 0.5

# Quick negative: bogus session id yields 401
curl -sS -X POST :8080/authn/passkey/login/finish \
  -H 'Content-Type: application/json' \
  -d '{"login_session_id":"bogus","id":"","rawId":"","type":"public-key","response":{"authenticatorData":"","clientDataJSON":"","signature":"","userHandle":""}}' -i | sed -n '1,20p'

# Cleanup
kill $API_PID
```

Acceptance criteria

- Handler verifies CDJ/AD and signature using account COSE key; updates signCount; inserts server session; sets cookie; returns 200 on success.
- Negative cases mapped to correct statuses (400/401/403/409) with stable behavior.

Notes

- Standardized error envelopes and structured logging will be added in later steps (38–40). Keep error text simple and avoid leaking sensitive data.

## Phase F — Transaction Signing (Server-supplied options)

20. **Bundle validation helper**

    - `bundle.go`: decode base64 → CBOR → `Bundle`; re-encode canonical CBOR; rebuild `B`; compute `tx_id`, `challenge = SHA256("CHALv1"||B)`.
    - Check `SenderKey` matches logged-in account’s key.
    - Check nonce monotonic (query last known or max nonce in `transactions` for account).
    - _Verify_: tests for encoding roundtrip and hashing stable.

21. **/tx/signing/options handler (auth required)**

    - Parse `bundle_cbor_b64`; call helper.
    - Read cookie and resolve server session to an account (inline until middleware in step 24 is added).
    - Create `tx_session_id`, store `{B, challenge, acct_cbor, expected_cred_id, expiresAt}` with TTL = 5 minutes.
    - Respond with **assertion options** (challenge, rpId, `userVerification: required`, `allowCredentials` = logged-in user’s credentialId).
    - _Verify_: browser can call; returns options with `expires_at`.

22. **/tx/signing/finish handler (auth required)**

    - Load `tx_session_id`; verify not expired; verify CDJ (`type=get`, challenge, origin).
    - Verify AD (rpIdHash, UV, signCount monotonic).
    - Verify signature using **account’s COSE key**.
    - Persist transaction (tx_id, bundle_cbor, AD, CDJ, signature, nonce, message).
    - _Verify_: returns 200 with `tx_id_hex`.

23. **/tx/list handler (auth required)**

    - Query `transactions` by `acct_cbor`; return `tx_id_hex`, `nonce`, `message`, `created_at`.
    - _Verify_: shows inserted records.

## Phase G — Sessions & Middleware

24. **Session middleware**

    - Read cookie; look up session; attach `acct_cbor` and `credential_id` to request context.
    - Enforce expiry; refresh if desired.
    - _Verify_: protected routes return 401 without cookie; 200 with cookie.

25. **Rate limiting & limits**

    - Simple token bucket per-IP in memory for `/authn/*` and `/tx/*`.
    - Max body size middleware (e.g., 64 KB).
    - _Verify_: exceed limits → 429/413.

26. **CORS & cookies**

    - Allow `http://localhost:5173`; set `SameSite=Lax`; for dev over HTTP, skip `Secure`.
    - _Verify_: cross-origin works from Vite.

## Phase H — Frontend (React) UI

27. **Basic pages**

    - `Register.tsx`, `Login.tsx`, `Dashboard.tsx`; a simple router (or conditional rendering).
    - Home screen presents only two primary actions: Register and Login (post-login shows Dashboard).
    - CORS configured in step 26; alternatively, set a Vite dev proxy to backend (`/api` → `http://localhost:8080`) during development.
    - _Verify_: SPA renders pages; home shows exactly two buttons.

28. **Base64url helpers (web)**

    - JS utils: ArrayBuffer ⇄ base64url; UTF‑8 encoder/decoder.
    - _Verify_: unit test in browser console.

29. **WebAuthn create() flow (Register)**

    - `POST /authn/passkey/registration/options`; convert JSON fields to `PublicKeyCredentialCreationOptions` (transform b64url→ArrayBuffer).
    - Call `navigator.credentials.create(...)`.
    - Send `registration/finish` payload (b64url encode ArrayBuffers).
    - On success: show account thumb, route to Login.
    - _Verify_: end-to-end registration completes.

30. **WebAuthn get() flow (Login)**

    - `POST /authn/passkey/login/options`; convert to `PublicKeyCredentialRequestOptions`.
    - Call `navigator.credentials.get(...)`.
    - Send `login/finish`; on success: set “logged in” UI state (no global store, just local state) and route to Dashboard.
    - _Verify_: end-to-end login completes; cookie present.

31. **Dashboard: fetch list**

    - Call `GET /tx/list`; render in table.
    - _Verify_: empty initially.

32. **Dashboard: build bundle**

    - UI collects `message` (string) and computes a **nonce**:

      - Option A (simpler): user enters nonce manually for demo.
      - Option B: call `GET /tx/next-nonce` (optional helper) to get `last_nonce+1`. (If not implemented, client tracks increment.)

    - Build `Bundle` object in JS compatible with the CDDL.
    - Install a small CBOR lib (e.g., `cbor-x`) and encode **canonical CBOR** in browser.
    - _Verify_: preview CBOR (hex) in console.

33. **Transaction signing options**

    - `POST /tx/signing/options` with `bundle_cbor_b64`.
    - Receive `publicKey` options; call `navigator.credentials.get(...)`.
    - Send `/tx/signing/finish`.
    - On success: refresh list.
    - _Verify_: message appears in table with nonce and timestamp.

34. **Error toasts**

    - Render server error messages (JSON `{ code, error }`) as inline alerts.
    - _Verify_: cause a failure (bad origin or nonce) and observe mapped error code per server envelope.

## Phase I — Testing & Fixtures

35. **Unit tests (Go)**

    - COSE→ECDSA, CDJ parse, AD parse, low‑S check, canonical encoding, hash anchors.
    - _Verify_: `go test ./...` passes.

36. **Golden vectors**

    - Create a tiny CLI in `server/cmd/vectors` that:

      - Loads a known COSE key, builds a sample bundle, prints hex(B), challenge, tx_id.

    - _Verify_: consistent outputs between runs.

37. **Manual E2E**

    - Run server + web; perform registration, login, and sign two messages with nonces 1, 2 in Safari and Chrome on macOS.
    - _Verify_: `/tx/list` shows two entries; DB reflects persisted rows; flows succeed in both browsers with platform authenticators.

## Phase J — Hardening (Demo-grade)

38. **Input validation & errors**

    - Enforce message length ≤ 1024; nonce ≤ `2^53 - 1`; origin/rpId in allowlist; body size ≤ 64 KB.
    - Implement standardized error envelope `{ code, error, correlation_id? }` and map to HTTP 400/401/403/409/413/429/5xx per R-ERR.
    - _Verify_: oversize blocked with 400/413; invalid/replay/UV/origin issues map to correct codes; frontend displays `code`.

39. **Session security**

    - Random session IDs (≥128 bits), expiry 1h, renewal on activity.
    - _Verify_: expired session returns 401; re-login works.

40. **Logging**

    - Structured logs with event names: `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`.
    - _Verify_: logs show account thumb and tx_id (hex).

41. **Build scripts**

    - Root Makefile: `make server`, `make web`, `make run`, `make clean`.
    - _Verify_: one command runs both.

## Phase K — Developer Experience

42. **API examples**

    - Add `docs/` with curl examples for each endpoint (sans WebAuthn ceremony).
    - _Verify_: docs render in repo.

43. **Postman / REST Client file**

    - Provide a collection with placeholders; helpful for observing JSON shapes.
    - _Verify_: collection can be imported.

44. **Env sample**

    - `.env.example` with `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`.
    - _Verify_: loads correctly.

45. **README**

    - Quickstart, limitations (no attestation trust), and demo notes (Touch ID prompts).
    - _Verify_: teammate can bootstrap in <10 minutes.

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

## Done

Completed steps are stored as individual files under `blueprint/done/`, with phase information prefixed in each filename (e.g., `phase-a-step-1-...md`).

## Future

- Handler-level logging and error mapping integration

  - What: Wire a minimal JSON logger using Go `log/slog` in `server/cmd/api` and apply the `MapVerifyError` and `LogAssertion` utilities from `server/internal/webauthn` in the assertion-finish handler. Keep client responses generic while emitting structured, privacy-preserving logs with stable `error_kind` values from sentinel errors.
  - Why: Improves observability, incident triage, and auditability without leaking sensitive data. Cleanly separates transport concerns (HTTP codes) from cryptographic failure semantics via sentinel errors, enabling accurate metrics and alerts (e.g., spikes in `ErrMalformedDER`).
  - How: Initialize `slog` with a JSON handler for dev; in the handler, call `VerifyAssertion(...)`, map the error with `MapVerifyError`, log once via `LogAssertion` using hashed identifiers (`HashID`), and return an appropriate status code with a standard error envelope. This remains compatible with the existing plan’s later steps for endpoints and error envelopes.

- IDNA (punycode) normalization support
  - What: Normalize internationalized domain names to ASCII (punycode) for RP ID and origin comparisons using `golang.org/x/net/idna`.
  - Why: Prevent mismatches and policy bypass due to Unicode vs punycode inconsistencies; ensure consistent hashing for rpIdHash and accurate origin validation across i18n domains.
  - How: Add a deterministic normalization helper that converts Unicode hostnames to punycode before validation/comparison; guard with unit tests and vectors (e.g., `bücher.ch` ⇄ `xn--bcher-kva.ch`), and document deployment guidance to keep config values consistent.
