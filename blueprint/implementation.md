# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Phase B — Server: Config, DB, Models

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

### Phase A — Repository & Tooling

### Step 1 — Initialize repo (Done: 2025-08-31)

- Structure

  - Ensure `server/`, `web/`, `blueprint/` exist (present in this repo) and root `.gitignore` exists (present).
  - Commands: `ls -la`, `git status` (confirm clean state before adding files).

- Source to add

  - Root `.editorconfig` with: UTF‑8; LF; trim trailing whitespace; insert final newline; 2 spaces for TS/TSX/JS/JSON/MD/YAML; tabs for Go.
    - Example sections: `[*.{ts,tsx,js,json,md,yml,yaml}] indent_size = 2`, `[*.go] indent_style = tab`.

- Description files to add

  - `server/server.desc.md`: backend overview (WebAuthn endpoints, SQLite persistence), relations to frontend and DB; key invariants (UV required, canonical CBOR, low‑S, signCount monotonic).
    - Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-PLAT-3; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails.
  - `web/web.desc.md`: SPA overview (Register/Login/Dashboard; WebAuthn invocations; CBOR bundle build), relations to backend.
    - Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; flows registration/login/transaction-signing.

- Blueprint updates

  - Add a short note under this step with “Refs: goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-PLAT-2; requirement R-OPS-DEV”.
  - Do not mark subsequent steps In‑Progress until their artifacts are Approved.

- Verify
  - `git status` shows only: `.editorconfig`, `server/server.desc.md`, `web/web.desc.md` as changes.
  - `test -f .editorconfig && echo ok` returns `ok`.
  - `ls server server/internal web/src >/dev/null` exits 0.
  - Editor/formatter recognizes `.editorconfig` (spot‑check by saving a TS file and a Go file).

### Step 2 — Server Go module (Done: 2025-08-31)

- Context

  - Initialize a minimal Go HTTP server with a health endpoint to unblock future backend steps.

- Structure

  - Ensure `server/` exists and contains a Go module.
    - If `server/go.mod` is missing, run: `cd server && go mod init txkit-demo && go mod tidy`.

- Source to add (instructions only)

  - `server/cmd/api/main.go`: package `main`; start `net/http` server on `:8080`; define `/health` handler returning `200 OK` with body `ok` (text/plain).
    - Use `http.NewServeMux()` and `http.ListenAndServe(":8080", mux)`; log a startup line `listening :8080`.
    - Keep constants in-file for now; Step 5 will introduce `config.go` and refactor port/origin.

- Description files to add (instructions only)

  - `server/cmd/api/main.go.desc.md`: Purpose (entrypoint; health), Key Logic (mux, handlers), Interactions (no DB yet), Refs.
    - Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV.

- Blueprint updates

  - Add “Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV” under this step after implementation.

- Verification (to run after implementation)

  - Build: `cd server && go build ./...` (expect exit 0).
  - Run dev: `cd server && go run ./cmd/api` (in a separate terminal).
  - Health: `curl -i http://localhost:8080/health` → `HTTP/1.1 200 OK` and body `ok`.

- User verification commands (copy/paste)

  ```bash
  # Build all server packages
  cd server && go build ./... && cd -

  # Run the API server in the background
  cd server
  go run ./cmd/api > /tmp/step2_api.log 2>&1 & echo $! > /tmp/step2_api.pid
  sleep 1

  # Verify health endpoint
  curl -i http://localhost:8080/health

  # Stop the server
  kill $(cat /tmp/step2_api.pid) && rm -f /tmp/step2_api.pid
  tail -n +1 /tmp/step2_api.log | sed -n '1,50p'
  cd -
  ```

- Notes
  - No CORS, cookies, DB, or config yet; these land in later steps (26, 24, 6, 5).
  - Keep the stub minimal to ensure fast builds and clear verification.

### Step 3 — Add deps (Done: 2025-08-31)

- Context

  - Add core dependencies for canonical CBOR encoding/decoding and SQLite persistence to support later server features.

- Structure

  - Work within `server/` Go module; ensure `server/go.mod` exists.
    - If missing, complete Step 2 prerequisites for module init.

- Source to add (instructions only)

  - Add CBOR library: `github.com/fxamacker/cbor/v2@v2.9.0` (canonical options support).
  - Add SQLite driver: `github.com/mattn/go-sqlite3@v1.14.32` (CGO‑based; acceptable for local dev).
  - Command sequence:
    - `cd server && go get github.com/fxamacker/cbor/v2@v2.9.0`
    - `cd server && go get github.com/mattn/go-sqlite3@v1.14.32`
    - `cd server && go mod tidy`

- Description files to add (instructions only)

  - None for this step; description files will accompany code that uses these deps (Steps 6–10 and DB/CBOR helpers).

- Blueprint updates

  - Refs to include upon implementation: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-PLAT-3; requirement R-SCHEMA-LITE.

- Verification (to run after implementation)

  - Inspect module files changed: `git diff -- server/go.mod server/go.sum` (shows added deps and checksums).
  - Build all packages: `cd server && go build ./...` (expect exit 0).
  - Optional: print versions resolved: `cd server && go list -m -json github.com/fxamacker/cbor/v2 github.com/mattn/go-sqlite3`.

- Notes

  - `github.com/mattn/go-sqlite3` requires CGO; macOS/Linux dev environments satisfy this by default. For CI or cross‑compile, consider build tags or `modernc.org/sqlite` in future ADRs (out of scope for demo).
  - Canonical CBOR usage will be implemented in Step 10; no code changes in this step beyond dependency resolution.

- User verification commands (copy/paste)

  ```bash
  # Ensure server module exists
  test -f server/go.mod && echo OK:server go.mod

  # Add dependencies (idempotent if already present)
  cd server
  go get github.com/fxamacker/cbor/v2@v2.9.0
  go get github.com/mattn/go-sqlite3@v1.14.32
  go mod tidy

  # Verify versions and build
  go list -m -json github.com/fxamacker/cbor/v2 github.com/mattn/go-sqlite3 | sed -n '1,80p'
  go build ./...
  cd -
  ```

### Phase B — Server: Config, DB, Models

### Step 4 — Web app scaffold (Done: 2025-08-31)

- Context

  - Initialize a minimal React + Vite SPA to support Register/Login/Dashboard in later steps.

- Structure

  - Ensure `web/` exists. If not present, scaffold from repo root.
    - New project: `npm create vite@latest web -- --template react`.
    - Existing folder: `cd web && npm create vite@latest . -- --template react`.

- Source to add (instructions only)

  - Ensure `package.json` contains scripts: `dev`, `build`, `preview` (Vite defaults).
  - Confirm `vite.config.ts` exists; keep defaults (proxy optional; CORS handled in Step 26).
  - Keep `src/App.tsx` as minimal root component; no additional pages yet (arrive in Step 27+).
  - Optional: add `.env.development` with `VITE_API_BASE=http://localhost:8080` for later use (not required yet).

- Description files to add (instructions only)

  - None; high-level `web/web.desc.md` already exists from Step 1. Add per-file descriptions when pages/components are implemented (Step 27+).

- Blueprint updates

  - Refs to include upon implementation: goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-OPS-DEV.

- Verification (to run after implementation)

  - Install deps: `cd web && npm ci`.
  - Dev server: `cd web && npm run dev` (expect server on http://localhost:5173).
  - Build: `cd web && npm run build` (expect dist/ output with no errors).
  - Optional: `curl -I http://localhost:5173` after dev server starts (expect 200 OK).

- User verification commands (copy/paste)

  ```bash
  # Install deps (if node_modules absent)
  test -d web/node_modules || (cd web && npm ci)

  # Start dev server in background and verify
  cd web
  npm run dev > /tmp/vite.log 2>&1 & echo $! > /tmp/vite.pid
  sleep 1
  curl -I http://localhost:5173
  kill $(cat /tmp/vite.pid) && rm -f /tmp/vite.pid

  # Production build
  npm run build
  cd -
  ```

- Notes
  - Do not wire API calls or pages yet; keep the scaffold minimal and compilable.
  - CORS is configured in Step 26. A Vite proxy is optional for local convenience and can be added later.

### Step 5 — Server config struct (Done: 2025-08-31)

- Context

  - Centralize runtime configuration (RP ID, origin, port, DB path, allowlists) to support secure WebAuthn checks and local dev.

- Structure

  - Add a `server/internal/config` package with a single `config.go` file, or a top-level `server/config.go` if keeping it flat (choose one; prefer `internal/config`).
  - Expose `type Config struct { RP_ID string; Origin string; Port string; DBPath string; RPAllowlist []string; OriginAllowlist []string }` and `func Load() (*Config, error)`.

- Source to add (instructions only)

  - `server/internal/config/config.go`:
    - Read env vars: `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`, `RP_ID_ALLOWLIST` (comma-separated), `ORIGIN_ALLOWLIST` (comma-separated).
    - Defaults: `RP_ID=localhost`, `ORIGIN=http://localhost:5173`, `PORT=8080`, `DB_PATH=server/demo.db`.
    - Normalize: trim spaces; lowercase `RP_ID`; ensure `Origin` has scheme and no trailing slash.
    - Derive `RPAllowlist` (include `RP_ID` if not present) and `OriginAllowlist` (include `Origin` if not present).
    - Validate: `RP_ID` non-empty; `Origin` parses as URL; `PORT` numeric; deny wildcard origins; allowlist entries must be exact matches (no globs).
    - Log (on startup) a concise summary (rp_id, origin, port, db_path) without secrets.
  - Integration note: refactor `cmd/api/main.go` later to call `config.Load()` and bind to `cfg.Port` (tracked in a later step to avoid scope creep here).

- Description files to add (instructions only)

  - `server/internal/config/config.go.desc.md`: Purpose (central config), Key Logic (env parsing, defaults, validation), Interactions (used by main and handlers), Refs.
    - Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-PORTABLE.

- Blueprint updates

  - Refs to include upon implementation: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-PORTABLE; requirement R-SEC-UV (policy alignment).

- Verification (to run after implementation)

  - Build: `cd server && go build ./...` (expect exit 0).
  - Run with defaults: `cd server && RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./cmd/api` and observe startup log contains rp_id/origin/port.
  - Invalid config: `cd server && ORIGIN=bad go run ./cmd/api` should log/return a clear error from `config.Load()`.

- User verification commands (copy/paste)

  ```bash
  # Build with config package present
  cd server && go build ./... && cd -

  # Run with explicit envs and observe log (Ctrl+C to stop)
  cd server
  RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo.db \
    go run ./cmd/api > /tmp/step5_server.log 2>&1 & echo $! > /tmp/step5_api.pid
  sleep 1
  grep -E "rp_id|origin|listening" -i /tmp/step5_server.log | sed -n '1,5p'
  kill $(cat /tmp/step5_api.pid) && rm -f /tmp/step5_api.pid
  cd -

  # Invalid origin should fail fast
  cd server && ORIGIN=bad go run ./cmd/api || echo "expected failure" && cd -
  ```

- Notes
  - Keep config minimal and focused; secrets are out of scope for the demo.
  - RP/Origin allowlists backstop later security checks in Step 14 and across ceremonies.

### Step 6 — DB init & migrations (Done: 2025-08-31)

- Context

  - Initialize SQLite persistence to support account, credential, session, and transaction storage per R-PLAT-3. Apply schema at startup with safe `CREATE TABLE IF NOT EXISTS` statements.

- Structure

  - Add `server/internal/storage` package (or `server/internal/db`): `storage.go` with `Open(cfg *config.Config) (*sql.DB, error)` and `Migrate(db *sql.DB) error`.
  - PRAGMAs on open: `foreign_keys=ON`, journal_mode=WAL, synchronous=NORMAL (demo-grade), busy_timeout=5000.

- Source to add (instructions only)

  - `server/internal/storage/storage.go`:
    - Open database at `cfg.DBPath` using `github.com/mattn/go-sqlite3` via `database/sql`.
    - Set connection pool (e.g., `SetMaxOpenConns(1)` for SQLite; `SetConnMaxIdleTime` reasonable).
    - Execute PRAGMAs and call `Migrate` with the schema:
      - Tables from R-PLAT-3 (accounts, credentials, sessions, transactions) with `IF NOT EXISTS` and `FOREIGN KEY` constraints.
    - Provide `Close()` responsibility to caller (main) later; for now, return `db`.
  - Integration note: wire into `cmd/api/main.go` later (when handlers need DB), keeping this step focused on package creation and migrations.

- Description files to add (instructions only)

  - `server/internal/storage/storage.go.desc.md`: Purpose (SQLite open + migrate), Key Logic (PRAGMAs, schema), Interactions (used by main/handlers), Refs.
    - Refs: goal simple-ui-and-storage; requirement R-PLAT-3; requirement R-PLAT-2; requirement R-ERR.

- Blueprint updates

  - Refs to include upon implementation: requirement R-PLAT-3; goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-NO-BROKER.

- Verification (to run after implementation)

  - Build: `cd server && go build ./...` (expect exit 0).
  - Run a tiny snippet (temporary or via main if already integrated) to call `storage.Open(cfg)` then `storage.Migrate(db)`.
  - Confirm DB file exists: `test -f server/demo.db`.
  - Optional (if `sqlite3` CLI available): `sqlite3 server/demo.db '.schema'` shows the four tables with expected columns.

- User verification commands (copy/paste)

  ```bash
  # Build server with storage package present
  cd server && go build ./... && cd -

  # Quick migration runner (inline Go) — does not modify app code
  cd server
  cat > /tmp/migrate.go <<'EOF'
  package main
  import (
    "log"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    store "github.com/neutral/passkey-demo/internal/storage"
  )
  func main(){
    cfg, err := cfgpkg.Load(); if err!=nil{ log.Fatal(err) }
    db, err := store.Open(cfg); if err!=nil{ log.Fatal(err) }
    defer db.Close()
    if err := store.Migrate(db); err!=nil { log.Fatal(err) }
    log.Println("migrated ok")
  }
  EOF
  PORT=0 go run /tmp/migrate.go
  test -f server/demo.db && echo OK:db-exists
  # Optional schema view
  command -v sqlite3 >/dev/null && sqlite3 server/demo.db '.schema' | sed -n '1,60p'
  cd -
  ```

- Notes
  - Keep PRAGMAs demo-grade; for production, review durability/performance trade-offs.
  - Foreign keys must be enabled for relational integrity; use `ON DELETE CASCADE` as specified.

### Step 7 — Types: COSE, WebAuthn, Bundle (Done: 2025-08-31)

- Context

  - Define core data shapes used across registration/login/signing: COSE EC2 public key, the signing `Bundle` (per CDDL), and minimal WebAuthn request/response payload shapes. Centralizing these types reduces duplication and mismatches across handlers.

- Structure

  - Add `server/internal/types/types.go` with plain structs and doc comments; no logic beyond JSON/CBOR tags where needed.
  - Keep enums/consts simple (e.g., ceremony types `"webauthn.create"|"webauthn.get"`).

- Source to add (instructions only)

  - `server/internal/types/types.go`:
    - `type CoseEC2 struct { Kty int \tAlg int \tCrv int \tX []byte \tY []byte }` — holds COSE EC2 public key fields parsed from attestation.
    - `type Bundle struct { SenderKey CoseEC2; Nonce uint64; Message string; ValidUntil *uint64 }` — mirrors the CDDL; used in signing.
    - Minimal WebAuthn payload shapes (JSON):
      - `type RegOptions struct { RP_ID string; Origin string; UVRequired bool; Attestation string }`
      - `type RegFinish struct { ID string; RawID string; Response struct{ AttestationObject string; ClientDataJSON string } }`
      - `type LoginOptions struct { RP_ID string; Origin string; UVRequired bool }`
      - `type LoginFinish struct { ID string; RawID string; Response struct{ AuthenticatorData string; ClientDataJSON string; Signature string; UserHandle string } }`
    - Note: Binary fields are base64url strings at the API surface; server decodes to bytes before verification.

- Description files to add (instructions only)

  - `server/internal/types/types.go.desc.md`: Purpose (shared models), Key Types (CoseEC2, Bundle, WebAuthn payloads), Interactions (used by handlers and helpers), Refs.
    - Refs: requirement R-ID-KEY; requirement R-SCHEMA-LITE; requirement R-PLAT-2.

- Blueprint updates

  - Refs to include upon implementation: goal key-first-identity-cose; goal minimal-cbor-bundle; requirement R-ID-KEY; requirement R-SCHEMA-LITE; requirement R-PLAT-2.

- Verification (to run after implementation)

  - Lint/build: `cd server && go vet ./... && go build ./...` (expect exit 0).
  - Import sanity: run `rg -n "package types" server` to confirm package compiles and is discoverable.

- User verification commands (copy/paste)

  ```bash
  cd server
  go vet ./... && go build ./...
  rg -n "package types|type CoseEC2|type Bundle" internal/types/types.go
  cd -
  ```

- Notes
  - Keep types minimal and transport‑oriented; parsing/validation logic belongs in subsequent steps (e.g., WebAuthn parsers, CBOR codec, crypto helpers).
