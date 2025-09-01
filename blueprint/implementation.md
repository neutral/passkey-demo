# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Phase C — WebAuthn Core Verification

## Phase D — Registration Endpoints

17. **/authn/passkey/registration/finish handler**

    - Context
      - Complete registration by validating the browser’s attestation response against the stored session and persisting the new account and credential.
      - Policy: `fmt: none`, `userVerification: required`, strict RP/Origin checks, and minimal, auditable persistence.

    - Structure
      - Route: `POST /authn/passkey/registration/finish`
      - Request (JSON):
        - `reg_session_id` (string): returned from options step.
        - WebAuthn payload (from browser): matches `types.RegFinish` (`id`, `rawId`, `type`, `response.attestationObject`, `response.clientDataJSON`).
      - Response (201 JSON): `{ account_thumb_hex, credential_id_b64 }` (and optional diagnostic fields in dev).

    - Steps
      1) Look up session by `reg_session_id`; reject if not found or expired; delete on success or terminal failure (single-use).
      2) Decode `clientDataJSON` via `ParseClientDataJSON`; require `IsCreate`.
      3) Challenge match: compare decoded challenge bytes to session `challenge`.
      4) Origin check: `CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocalhostOK)`.
      5) Decode attestationObject and extract data: `ExtractRegistrationData(attObjB)` → `ad`, `aaguid`, `credID`, `cose`.
      6) RP ID hash: `CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist)`.
      7) UV/UP policy: require `HasUV(ad.Flags)`; optionally log `HasUP`.
      8) COSE key validation: `crypto.ToECDSA(&cose)`; must be ES256/P-256; derive Go pub key for future verification.
      9) Persist:
         - `acct_cbor = encoding.EncodeCanonical(cose)` and `acct_thumb = SHA256("ACCTK1" || acct_cbor)`.
         - Insert into `accounts` if not present; OK if exists with identical `acct_cbor`.
         - Insert into `credentials` with `credential_id`, `acct_cbor_fk`, `sign_count = ad.SignCount`, `aaguid`.
         - On unique violation (`credential_id` already exists): return 409 Conflict.
     10) Return 201 with `account_thumb_hex` and `credential_id_b64`.

    - Error handling & mapping
      - 400: malformed inputs (bad base64, malformed CBOR, invalid JSON), `ErrRpIdInvalid`, unsupported attestation fmt.
      - 401: expired or missing session; challenge mismatch.
      - 403: origin/RP policy failures (use `MapPolicyError`).
      - 409: duplicate credential id.
      - 500: storage/internal errors.

    - Tests
      - Unit: synthetic end-to-end builder that creates a valid attestationObject with fmt none, matching challenge, valid COSE EC2; handler returns 201 and DB rows appear.
      - Negative cases: expired/missing session (401); wrong challenge (401); wrong origin (403); rpIdHash mismatch (403); no UV flag (403); unsupported fmt (400); bad CBOR (400); duplicate credential (409).
      - Ensure `reg_session_id` is single-use (second call fails with 401/409 depending on flow).

    - Acceptance criteria
      - Successful finish produces idempotent account row (no dupes) and a new credential row with correct fields; returns `account_thumb_hex` and the credential id.
      - All policy and parsing failures map to the expected HTTP status using sentinel-based mapping.
      - Session is consumed (deleted) after a terminal outcome.

    - Notes
      - Keep response minimal; do not return raw keys or binary blobs; use base64url/hex for identifiers.
      - Log structured event `reg_finish` (Future logging item) with hashed identifiers.

## Phase E — Login Endpoints

18. **/authn/passkey/login/options handler**

    - Generate `login_session_id`, random 32B challenge.
    - Return options (`rpId`, `challenge`, `userVerification: required`, `allowCredentials: []`).
    - Store `{login_session_id, challenge, rpId, origin, expiresAt}` with TTL = 5 minutes in memory.
    - _Verify_: `curl` returns proper JSON; includes `expires_at` ≈ now+5m.

19. **/authn/passkey/login/finish handler**

    - Load session; verify not expired; verify CDJ (`type=get`, challenge, origin).
    - Parse AD: verify rpIdHash, UV; read `signCount`.
    - **Identify account** by `credential_id` from request (look up in `credentials`).
    - Verify signature using **account’s COSE key** (ECDSA).
    - Check `signCount` monotonic; update DB.
    - Create server **session** (random token) stored in `sessions`; set cookie.
    - _Verify_: browser login works; `Set-Cookie` present; subsequent authed call returns 200.

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
