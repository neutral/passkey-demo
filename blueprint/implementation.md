# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Phase B — Server: Config, DB, Models

8. **COSE → ECDSA helper**

   - See Done section (Step 8) for full details and verification commands.

9. **Base64url utilities**

   - `b64url.go`: encode/decode with/without padding (tolerant decode).
   - _Verify_: tests roundtrip random bytes.

10. **CBOR canonical codec**

    - `cbor.go`: instantiate fxamacker encoder with canonical options; `EncodeCanonical(v any) ([]byte, error)` and `Decode()` helpers.
    - _Verify_: test deterministic encoding order for a map (compare hex to expected).

## Phase C — WebAuthn Core Verification

11. **Parse authenticatorData**

    - `webauthn_ad.go`: parse AD → `rpIdHash []byte`, `flags byte`, `signCount uint32`.
    - _Verify_: test with known binary sample; ensure big‑endian parse.

12. **ClientDataJSON validation**

    - `webauthn_cdj.go`: parse JSON; expose `Type`, `Challenge` (decoded), `Origin`.
    - _Verify_: test tolerant base64url decode; reject bad encodings.

13. **Signature verify utility**

    - `webauthn_sig.go`: `VerifyAssertion(pub *ecdsa.PublicKey, ad, cdj, sigDER []byte) error` building `SHA256(ad || SHA256(cdj))` and verifying low‑S.
    - _Verify_: test with synthetic keypair & generated signature.

14. **RP ID & Origin checks**

    - `security.go`: `CheckRpIdHash(adRpHash, rpId string)`, `CheckOrigin(origin string, allow []string)`.
    - _Verify_: unit tests.

## Phase D — Registration Endpoints

15. **/authn/passkey/registration/options handler**

    - Generate `reg_session_id`, random 32B challenge, temp `user.id` (random 32B), build options JSON.
    - Set policy: `userVerification: "required"`, `residentKey: "required"`, `attestation: "none"`.
    - Save `{reg_session_id, challenge, rpId, origin, expiresAt}` with TTL = 5 minutes in memory map (and optional table if persisting).
    - _Verify_: `curl` returns options; check challenge length & fields; policy flags present; `expires_at` ≈ now+5m.

16. **Attestation parsing (minimal)**

    - `webauthn_att.go`: decode `attestationObject` CBOR; extract:

      - `authData` → parse AAGUID, credentialId, **COSE key**.

    - _Verify_: with a real registration, ensure you can parse returned attestation.

17. **/authn/passkey/registration/finish handler**

    - Load session; verify not expired; verify CDJ (`type=create`, challenge, origin).
    - Verify `rpIdHash` in AD; check UV flag; extract COSE key, credentialId, initial signCount, and optional AAGUID.
    - **Persist**: insert into `accounts` (acct_cbor = canonical CBOR of COSE key; acct_thumb = SHA256("ACCTK1"||acct_cbor)) and `credentials` (credential_id, acct_cbor_fk, sign_count, aaguid).
    - _Verify_: register via browser → handler returns 201 with `account_thumb_hex`.

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
