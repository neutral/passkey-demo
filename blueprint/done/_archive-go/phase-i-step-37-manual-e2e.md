### Step 37 — Manual E2E (Done: 2025-09-10)

Verification notes
- Safari and Chrome on macOS: registration, login, and transaction signing succeeded.
- The sender_key mismatch was resolved by server fallback decode + logical field comparison and client embedding of canonical COSE from `acct_cbor_b64`.

37. **Manual E2E**

    - Purpose

      - Validate the full demo flow on macOS using real platform authenticators (Touch ID / iCloud Keychain) in Safari and Chrome. Covers registration, login, and transaction signing bound to bundle content.

    - Dependencies

      - Server endpoints: Steps 15–23 complete and passing.
      - Middleware & policy: Steps 24–26 (sessions, rate limits, CORS/cookies) in place.
      - Web app: Steps 27–33 present and configured to call the API via absolute URLs.
      - Vectors/tests: Step 36 helps sanity-check anchoring if needed.

    - Environment

      - macOS with Touch ID enabled (or iCloud Keychain), Safari ≥ 17, Chrome ≥ 120.
      - Node 18+, npm; Go 1.22+.
      - Default dev settings: `RP_ID=localhost`, `ORIGIN=http://localhost:5173`, `PORT=8080`, `DB_PATH=server/demo.db`.

    - Procedure

      1) Start backend

         - Option A (run):
           ```bash
           RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo.db \\
           go -C server run ./cmd/api
           ```
         - Option B (build):
           ```bash
           go -C server build ./cmd/api && RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 \
             DB_PATH=server/demo.db ./server/api
           ```

      2) Start frontend (in another terminal)

         ```bash
         cd web
         npm install
         npm run dev
         # Vite dev server at http://localhost:5173
         ```

      3) Register (Safari, then Chrome)

         - Open `http://localhost:5173`.
         - Click Register and complete the platform authenticator prompt.
         - Expect success toast; server logs show `reg_options` then `reg_finish`.

      4) Login (Safari, then Chrome)

         - Click Login; complete Touch ID prompt.
         - Expect success toast; an `sid` cookie appears for the origin.

      5) Sign two messages (nonces 1 and 2)

         - Go to Dashboard; enter message `hello` with nonce `1`, click Sign.
         - Repeat with nonce `2`.
         - Expect two success toasts; server logs show `tx_options`/`tx_finish` pairs.

    - Verification

      - List API (authenticated via browser session):
        - Use the app’s “Transactions” panel; two entries visible with nonces 1 and 2.

      - DB check (optional):
        ```bash
        sqlite3 server/demo.db \
          "SELECT COUNT(*) FROM transactions;" \
          "SELECT nonce, message FROM transactions ORDER BY id;"
        ```

      - Backend health:
        ```bash
        curl :8080/health -i
        ```

      - Acceptance:
        - Safari: registration, login, and both sign actions succeed.
        - Chrome: registration, login, and both sign actions succeed.
        - `/tx/list` shows two entries for the logged-in account; DB rows reflect nonce/message and derived `tx_id`.

    - Troubleshooting

      - If CORS blocks requests, confirm `ORIGIN` matches the exact Vite URL and that credentials are allowed.
      - If login succeeds but `/tx/list` 401s via curl, note it requires the browser session cookie; use the app UI or a cookie jar.
      - If Touch ID prompts don’t appear, ensure platform authenticator is enabled for the browser profile and that iCloud Keychain or Touch ID is set up.

    - Refs

      - Refs: requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails; goal server-derived-challenge-and-txid

    - Appendix — CBOR/COSE Interop (sender_key mismatch)

      - Symptom:
        - `POST /tx/signing/options` intermittently returned 401 "sender_key mismatch" despite valid session and CORS.
        - Inspection showed typed struct decode of the nested COSE under bundle key `0` yielded zero values for `kty/alg/crv/x/y`, while a generic decode surfaced a valid EC2 key (kty=2, alg=-7, crv=1) with 32-byte X/Y matching the UI.

      - Root Cause:
        - The frontend encoder (`cbor-x`) produced a nested COSE map shape that fxamacker/cbor’s strict typed struct decoder didn’t bind in some variants. Differences included key typing and wrappers (e.g., CBOR tag wrapping byte strings) that left Go struct fields at their zero values during direct struct unmarshal.
        - Comparing raw CBOR bytes (or relying solely on typed decode) made equality brittle across encoder variants, even when the logical COSE fields matched.

      - Server Changes (decoding methodology):
        - Identity check now compares logical COSE fields only: `kty`, `alg`, `crv`, and the byte arrays `x`, `y`.
        - Robust fallback when typed decode of the nested COSE is empty:
          - Unmarshal the top-level bundle as `map[int64]cbor.RawMessage` (then `map[uint64]`/`map[any]`) and extract key `0` as a `cbor.RawMessage`.
          - Attempt typed unmarshal of that raw value into a struct with numeric keys (`1, 3, -1, -2, -3`).
          - If still empty, unmarshal into `map[any]any`, coerce keys to integers, and unwrap `cbor.Tag` where needed to obtain `[]byte` for X/Y.
        - Canonically re-encode the logical bundle to form `B` and proceed with anchors and nonce checks.

      - Client Changes (structural alignment):
        - Fetch `/me/account_key`, decode `acct_cbor_b64`, and normalize into a `Map<number, any>` with numeric COSE keys (`1, 3, -1, -2, -3`).
        - Build the bundle as `new Map([[0, senderCoseObj], [1, nonce], [2, message]])`, then encode canonically. This eliminates structural drift between encoders by embedding the server’s canonical COSE.

      - Mapping/Decoding Intricacies (JS ↔ Go):
        - JS `Map` vs Go struct binding: Go struct tags require exact numeric keys; if a nested COSE map arrives with key types/shapes the decoder doesn’t match (e.g., wrapped bytes, unexpected key typing), fields remain zero and no error is raised.
        - Generic-first recovery: decoding to `map[any]any` and coercing keys (`1,3,-1,-2,-3`) plus unwrapping `cbor.Tag` ensures robust extraction of X/Y and parameters across encoder variants.
        - Never compare raw CBOR for identity: use logical field equivalence; canonical re-encoding is only for computing anchors (`challenge`, `tx_id`).

      - Verification:
        - UI Build embeds canonical COSE → `POST /tx/signing/options` returns 200; `finish` succeeds; `/tx/list` shows the transaction(s).
        - Golden vectors remain valid; the server accepts bundles whose logical COSE matches the account key, regardless of benign encoder differences.

      - Refs: requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer; goal server-derived-challenge-and-txid; goal minimal-cbor-bundle

