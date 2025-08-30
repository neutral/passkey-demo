# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Phase A — Repository & Tooling

1. **Initialize repo**

   - Create directories: `server/`, `web/`.
   - Add root `.editorconfig`, `.gitignore` (Go, Node, SQLite db file).
   - _Verify_: `git status` shows clean structure.

2. **Server Go module**

   - `cd server && go mod init txkit-demo && go mod tidy`.
   - Create `main.go` with HTTP server stub (health endpoint).
   - _Verify_: `go build ./...` succeeds; `curl :8080/health` → `200 OK`.

3. **Add deps**

   - Choose CBOR lib (e.g., `github.com/fxamacker/cbor/v2`), SQLite driver (`github.com/mattn/go-sqlite3`).
   - _Verify_: `go get` and `go build` succeed.

4. **Web app scaffold**

   - `cd ../web && npm create vite@latest web -- --template react`.
   - Install deps: none needed beyond React for now.
   - _Verify_: `npm run dev` launches Vite dev server; open `http://localhost:5173`.

## Phase B — Server: Config, DB, Models

5. **Server config struct**

   - Add `config.go` with `RP_ID`, `ORIGIN`, `PORT`, DB path, and allowlists. Load from env with defaults (`localhost`, `http://localhost:5173`).
   - _Verify_: log config on startup.

6. **DB init & migrations**

   - `db.go`: open/create SQLite file, PRAGMA settings, apply DDL (see §2.1) on startup.
   - _Verify_: on start, DB file exists and tables present (`sqlite3 demo.db '.schema'`).

7. **Types: COSE, WebAuthn, Bundle**

   - `types.go`: define structs

     - `CoseEC2 {Kty, Alg, Crv, X, Y}`
     - `Bundle { SenderKey CoseEC2; Nonce uint64; Message string; ValidUntil *uint64 }`
     - WebAuthn request/response JSON types.

   - _Verify_: `go vet`, `go build` OK.

8. **COSE → ECDSA helper**

   - `crypto_cose.go`: `ToECDSA(*CoseEC2) (*ecdsa.PublicKey, error)` with curve checks.
   - _Verify_: unit test with a known valid point.

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
    - Save `{reg_session_id, challenge, rpId, origin, expiresAt}` in memory map (and optional table if persisting).
    - _Verify_: `curl` returns options; check challenge length & fields.

16. **Attestation parsing (minimal)**

    - `webauthn_att.go`: decode `attestationObject` CBOR; extract:

      - `authData` → parse AAGUID, credentialId, **COSE key**.

    - _Verify_: with a real registration, ensure you can parse returned attestation.

17. **/authn/passkey/registration/finish handler**

    - Load session; verify CDJ (`type=create`, challenge, origin).
    - Verify `rpIdHash` in AD; check UV flag; extract COSE key, credentialId, initial signCount.
    - **Persist**: insert into `accounts` (acct_cbor, acct_thumb) and `credentials` (1:1).
    - _Verify_: register via browser → handler returns 201 with `account_thumb_hex`.

## Phase E — Login Endpoints

18. **/authn/passkey/login/options handler**

    - Generate `login_session_id`, random 32B challenge.
    - Return options (`rpId`, `challenge`, `userVerification: required`, `allowCredentials: []`).
    - Store `{login_session_id, challenge, rpId, origin}` in memory.
    - _Verify_: `curl` returns proper JSON.

19. **/authn/passkey/login/finish handler**

    - Load session; verify CDJ (`type=get`, challenge, origin).
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
    - Create `tx_session_id`, store `{B, challenge, acct_cbor, expected_cred_id}`.
    - Respond with **assertion options** (challenge, rpId, UV required, allowCredentials = logged-in user’s credentialId).
    - _Verify_: browser can call; returns options.

22. **/tx/signing/finish handler (auth required)**

    - Load `tx_session_id`; verify CDJ (`type=get`, challenge, origin).
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

## Phase H — Frontend (React) UI

26. **Basic pages**

    - `Register.tsx`, `Login.tsx`, `Dashboard.tsx`; a simple router (or conditional rendering).
    - _Verify_: SPA renders pages.

27. **Base64url helpers (web)**

    - JS utils: ArrayBuffer ⇄ base64url; UTF‑8 encoder/decoder.
    - _Verify_: unit test in browser console.

28. **WebAuthn create() flow (Register)**

    - `POST /authn/passkey/registration/options`; convert JSON fields to `PublicKeyCredentialCreationOptions` (transform b64url→ArrayBuffer).
    - Call `navigator.credentials.create(...)`.
    - Send `registration/finish` payload (b64url encode ArrayBuffers).
    - On success: show account thumb, route to Login.
    - _Verify_: end-to-end registration completes.

29. **WebAuthn get() flow (Login)**

    - `POST /authn/passkey/login/options`; convert to `PublicKeyCredentialRequestOptions`.
    - Call `navigator.credentials.get(...)`.
    - Send `login/finish`; on success: set “logged in” UI state (no global store, just local state) and route to Dashboard.
    - _Verify_: end-to-end login completes; cookie present.

30. **Dashboard: fetch list**

    - Call `GET /tx/list`; render in table.
    - _Verify_: empty initially.

31. **Dashboard: build bundle**

    - UI collects `message` (string) and computes a **nonce**:

      - Option A (simpler): user enters nonce manually for demo.
      - Option B: call `GET /tx/next-nonce` (optional helper) to get `last_nonce+1`. (If not implemented, client tracks increment.)

    - Build `Bundle` object in JS compatible with the CDDL.
    - Encode **canonical CBOR** in browser (use a small CBOR lib).
    - _Verify_: preview CBOR (hex) in console.

32. **Transaction signing options**

    - `POST /tx/signing/options` with `bundle_cbor_b64`.
    - Receive `publicKey` options; call `navigator.credentials.get(...)`.
    - Send `/tx/signing/finish`.
    - On success: refresh list.
    - _Verify_: message appears in table with nonce and timestamp.

33. **Error toasts**

    - Render server error messages (JSON `error`, `code`) as inline alerts.
    - _Verify_: cause a failure (bad origin or nonce) and observe message.

## Phase I — Testing & Fixtures

34. **Unit tests (Go)**

    - COSE→ECDSA, CDJ parse, AD parse, low‑S check, canonical encoding, hash anchors.
    - _Verify_: `go test ./...` passes.

35. **Golden vectors**

    - Create a tiny CLI in `server/cmd/vectors` that:

      - Loads a known COSE key, builds a sample bundle, prints hex(B), challenge, tx_id.

    - _Verify_: consistent outputs between runs.

36. **Manual E2E**

    - Run server + web; perform registration, login, and sign two messages with nonces 1, 2.
    - _Verify_: `/tx/list` shows two entries; DB reflects persisted rows.

## Phase J — Hardening (Demo-grade)

37. **CORS & cookies**

    - Allow `http://localhost:5173`; set `SameSite=Lax`; for dev over HTTP, skip `Secure`.
    - _Verify_: cross-origin works from Vite.

38. **Input validation**

    - Enforce message length ≤ 1024; nonce ≤ `2^53 - 1`; origin/rpId in allowlist.
    - _Verify_: oversize blocked with 400.

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
- **Replay/Nonce**: Re-submit nonce 2 → **409** conflict.
- **UV check**: If browser returns an assertion without UV (simulate by forcing options incorrectly) → **403** forbidden.
- **Origin/RP guard**: Change `origin` in request body → **403**.
- **Low‑S enforced**: Hand-craft a signature with high‑S (unit test) → **400**.

