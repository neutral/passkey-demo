### Step 32 — Dashboard: build bundle (Done: 2025-09-09)

Verification notes
- Frontend: Dashboard builds canonical CBOR bundle with sender_key; previews base64url and hex; determinism verified by repeated builds.
- Backend: `/me/account_key` returns account COSE EC2 key for authenticated sessions; 401 when unauthorized.
- Build: `npm -C web run build` OK; `cd server && go build ./...` OK.

32. **Dashboard: build bundle**

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

