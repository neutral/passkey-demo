### Step 19 — /authn/passkey/login/finish handler (Done: 2025-09-03)

Verification notes
- Unit tests executed locally with module cache pinned: `cd server && GOCACHE=$(pwd)/.gocache go test ./...` → all packages OK.
- Targeted tests passed: `TestLoginFinish_Happy`, `TestLoginOptionsHandler_JSON`, `TestRegistrationFinish_Happy`.
- Local server bind checks are environment-dependent; unit tests validate the full happy-path logic end-to-end without network.

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

