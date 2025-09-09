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

## Next

### Phase H — Frontend (React) UI

---

32. **Dashboard: build bundle** — Moved to Done

    - Scope:

      - Enable building a canonical CBOR transaction bundle (B) on the client with fields `{ sender_key, nonce, message, valid_until? }`.
      - Fetch the logged-in account’s COSE EC2 public key (sender_key) from the backend (new authenticated endpoint) so the bundle’s `sender_key` matches the account key.
      - Encode the bundle deterministically (canonical CBOR) in the browser using a small library (`cbor-x`), produce base64url and hex previews, and keep results in local UI state (sent in Step 33).

    - Source to add/modify:

      - Add (frontend lib): `web/src/lib/cbor.ts` — Thin wrapper over `cbor-x` to encode/decode CBOR in canonical/deterministic mode.
      - Add (frontend lib): `web/src/lib/bundle.ts` — Helpers:
        - `type CoseEC2 = { kty: number; alg: number; crv: number; x: Uint8Array; y: Uint8Array }`
        - `buildBundle(sender: CoseEC2, nonce: number, message: string, validUntil?: number)` returns a JS object with integer-key map semantics matching the server’s CDDL (0: sender_key, 1: nonce, 2: message, 3?: valid_until).
        - `encodeBundleCanonical(b: any): Uint8Array` using `cbor-x` in canonical mode.
        - `bundleToB64Hex(u8: Uint8Array): { b64: string; hex: string }` using existing base64url util (Step 28).
      - Modify: `web/src/pages/Dashboard.tsx` —
        - Add inputs for `message` and `nonce`; add a “Build” button that:
          1) Ensures sender_key is loaded (see below)
          2) Builds the logical bundle from UI state
          3) Encodes canonical CBOR (B)
          4) Shows a read-only preview: `bundle_cbor_b64` and `hex(B)`
        - Add a lazy fetch of sender_key via new backend endpoint (see below) on first expansion of the signing area or when clicking “Build”.
        - Persist `bundle_cbor_b64` in component state so Step 33 can POST it to `/tx/signing/options`.
      - Add (backend): `server/internal/me/account_key.go` — Authenticated handler `GET /me/account_key` that:
        - Resolves the logged-in account from session (context or cookie), loads `acct_cbor`, decodes it to `types.CoseEC2`, and returns:
          - `acct_cbor_b64: string`
          - `sender_key: { kty, alg, crv, x, y }` with `x`/`y` base64url
      - Wire route (backend): `server/cmd/api/main.go` — mount `GET /me/account_key`.

    - Description files (create/update alongside code changes):

      - Create: `web/src/lib/cbor.ts.desc.md` — Canonical encoding usage and library notes; deterministic mode and integer-key mapping; Refs included.
      - Create: `web/src/lib/bundle.ts.desc.md` — Bundle shape, build helpers, invariants, and encoding; Refs included.
      - Update: `web/src/pages/Dashboard.tsx.desc.md` — Add signing section behavior (build/preview), sender_key fetch, and state management. Refs included.
      - Create: `server/internal/me/account_key.go.desc.md` — Authenticated endpoint purpose, response shape, and security considerations. Refs included.
      - Update: `server/cmd/api/main.go.desc.md` — Note new `/me/account_key` route wiring.

    - Request/response shape:

      - `GET {API_BASE}/me/account_key`
        - Requires session cookie `sid` (sent with `credentials: 'include'`)
        - 200 JSON:
          - `acct_cbor_b64: string` — base64url of canonical CBOR for COSE EC2 key
          - `sender_key: { kty: number, alg: number, crv: number, x: string(b64url), y: string(b64url) }`
        - 401 Unauthorized: no body required

      - Local bundle object (JS; not sent yet):
        - `{ 0: sender_key(COSE EC2 object), 1: nonce(uint), 2: message(string), 3?: valid_until(uint) }`
      - Encoded outputs for preview:
        - `bundle_cbor_b64: string` — base64url of canonical CBOR (B)
        - `bundle_hex: string` — hex(B) for visual verification only

    - Algorithm:

      - Fetch sender_key (lazy):
        - `fetch(apiUrl('/me/account_key'), { method: 'GET', mode: 'cors', credentials: 'include' })`
        - On 401 → prompt to Login; on 200 → parse and transform `x`/`y` from base64url to `Uint8Array`.
      - Build bundle:
        - Validate inputs: `message.trim().length > 0`; `nonce` is a positive integer (UI only; server enforces monotonicity later).
        - Construct object with integer keys {0,1,2,3?} mapping to `{ sender_key, nonce, message, valid_until? }`.
        - Encode canonical CBOR (deterministic) via `cbor-x`.
        - Produce base64url and hex previews; keep `bundle_cbor_b64` in state for Step 33.
      - Determinism check (dev-only):
        - Re-encode the same logical input and assert bytes equal; if not, show an inline warning.

    - Database interactions:

      - None on client. The backend `GET /me/account_key` performs a session lookup and decoding only (no writes).

    - Policies & limits:

      - Canonical CBOR encoding is required to match server hash anchors (challenge/tx_id).
      - ES256-only key (kty=2, alg=-7, crv=1); `x`/`y` must be 32 bytes. The server already enforces on registration; client just relays the key.
      - Optional UI limit: warn if `message.length > 1024` (server hard limit comes in Step 38).

    - Sequencing:

      - Depends on Step 30 (session cookie) for authenticated key fetch.
      - Precedes Step 33 (send bundle to `/tx/signing/options`).

    - Tests (happy path required, plus negatives):

      - Files to add (frontend): `web/tests/bundle-build.spec.ts`
        - Happy path:
          - Mock/fetch `GET /me/account_key` (Playwright route) with a sample COSE key; build bundle with nonce/message; ensure two encodes of the same logical object produce identical bytes; check base64url and hex are non-empty.
        - Negative:
          - Missing sender_key fetch (401) shows Login prompt; Build disabled.
          - Empty message or invalid nonce shows a validation message; bundle not built.
      - Files to add (backend minimal): optionally `server/internal/me/account_key_test.go` (unit tests):
        - 200 with valid session returns `acct_cbor_b64` and the expected COSE JSON fields.
        - 401 when session missing/expired.
      - Commands:
        - `npm -C web i cbor-x` (install lib)
        - `npm -C web run test:ui -- --project=chromium tests/bundle-build.spec.ts`
        - `go test ./server/internal/me -run AccountKey` (if backend test added)

    - Verification:

      - Manual dev:
        - Login, open Dashboard, enter a short message and a nonce, click Build; see b64/hex previews. Re-click Build and verify previews remain identical for the same input.
      - Determinism:
        - With the same `sender_key`, `nonce`, `message`, and optional `valid_until`, repeated encodes produce identical `bundle_cbor_b64` and hex.

      - User verification commands:

        ```bash
        # 1) Install frontend dependency
        npm -C web i cbor-x

        # 2) Start backend and web
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C server run ./cmd/api &
        SERVER_PID=$!
        npm -C web run dev &
        WEB_PID=$!

        # 3) In the app: Login → Dashboard → enter message+nonce → Build
        #    Verify bundle_cbor_b64 and hex(B) preview and determinism (build twice; values identical)

        # 4) Cleanup
        kill $WEB_PID || true
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - Dashboard can build a canonical CBOR bundle with `sender_key` equal to the logged-in account COSE key, plus `nonce` and `message` (and optional `valid_until`).
      - Deterministic encoding: identical inputs yield identical CBOR bytes across runs.
      - Previews show `bundle_cbor_b64` and hex(B) and are stable until inputs change.
      - Authenticated `GET /me/account_key` endpoint exists and returns the expected JSON shape; 401 when unauthorized.

    - Notes:

      - Keep UI minimal; validation is light-touch (server enforces hard limits later).
      - Sender key is public and not sensitive; fetching it per session avoids storing it persistently in the client.

    - Refs:
      - Refs: requirement R-FLOW-SIGN; requirement R-PLAT-1; requirement R-PLAT-2; requirement R-PORTABLE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer

---

33. **Transaction signing options**

    - `POST /tx/signing/options` with `bundle_cbor_b64` using absolute API URL and `credentials: 'include'`.
    - Receive `publicKey` options; call `navigator.credentials.get(...)`.
    - Send `/tx/signing/finish`.
    - On success: refresh list.
    - _Verify_: message appears in table with nonce and timestamp.

---

34. **Error toasts**

    - Best-effort now: render HTTP status and any JSON `{ error, code? }` if present; fall back to status text. Upgrade to standardized envelopes after Step 38.
    - _Verify_: cause a failure (e.g., bad nonce or origin) and see a visible inline alert with informative text.

### Phase I — Testing & Fixtures

---

35. **Unit tests (Go)**

    - COSE→ECDSA, CDJ parse, AD parse, low‑S check, canonical encoding, hash anchors.
    - _Verify_: `go test ./...` passes.

---

36. **Golden vectors**

    - Create a tiny CLI in `server/cmd/vectors` that:

      - Loads a known COSE key, builds a sample bundle, prints hex(B), challenge, tx_id.

    - _Verify_: consistent outputs between runs.

---

37. **Manual E2E**

    - Run server + web; perform registration, login, and sign two messages with nonces 1, 2 in Safari and Chrome on macOS.
    - _Verify_: `/tx/list` shows two entries; DB reflects persisted rows; flows succeed in both browsers with platform authenticators.

### Phase J — Hardening (Demo-grade)

---

38. **Input validation & errors**

    - Enforce message length ≤ 1024; nonce ≤ `2^53 - 1`; origin/rpId in allowlist; body size ≤ 64 KB.
    - Implement standardized error envelope `{ code, error, correlation_id? }` and map to HTTP 400/401/403/409/413/429/5xx per R-ERR.
    - _Verify_: oversize blocked with 400/413; invalid/replay/UV/origin issues map to correct codes; frontend displays `code`.

---

39. **Session security**

    - Random session IDs (≥128 bits), expiry 1h, renewal on activity.
    - _Verify_: expired session returns 401; re-login works.

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
