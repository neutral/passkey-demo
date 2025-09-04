### Step 21 — /tx/signing/options handler (Done: 2025-09-04)

Verification notes
- Unit tests executed: `cd server && GOCACHE=$(pwd)/.gocache go test ./internal/tx -v` and `go test ./...` → all green.
- Happy path validates response fields and tx-session storage; negatives cover missing/expired cookie, bad base64, sender key mismatch, nonce conflict, and no credentials.

21. **/tx/signing/options handler (auth required)**

    Scope
    - Add an authenticated handler `POST /tx/signing/options` that:
      - Reads session cookie `sid` to resolve the logged-in account (inline DB lookup; middleware arrives in Step 24).
      - Validates the client-provided bundle via Step 20 helper; derives `challenge` and `tx_id` from canonical `B`.
      - Creates a short-lived tx session (5 minutes) storing `B`, `challenge`, `acct_cbor`, and expected credential ids.
      - Returns WebAuthn assertion options with `challenge`, `rpId`, `userVerification: required`, and `allowCredentials` for the account’s credentials, plus `tx_session_id`, `tx_id_hex`, and `expires_at`.

    Source to add/modify
    - Add `server/internal/tx/options.go`: tx session model/store, builder, and handler
      - Types:
        - `type TxSession struct { B []byte; Challenge []byte; AcctCBOR []byte; CredentialIDs [][]byte; ExpiresAt time.Time }`
        - `type TxSessionStore` with `NewTxSessionStore(capacity int)`, `Put/Get/Delete` (mutex-protected; capacity-checked like reg/login stores).
      - Inbound/response JSON:
        - `type TxOptionsInbound { BundleCBOR string \t`json:"bundle_cbor_b64"`\t }`
        - `type TxOptionsResponse { TxSessionID string; Challenge string; Options types.LoginOptions; TxIDHex string; ExpiresAt int64 }`
      - Builder:
        - `func BuildTxOptions(ctx context.Context, cfg *config.Config, store *TxSessionStore, db *sql.DB, acctCBOR []byte, bundleB64 string, now func() time.Time) (TxOptionsResponse, error)`
          - Uses `tx.ValidateAndAnchorBundle(...)` to parse/anchor; selects all credential IDs for `acctCBOR` from DB; creates `sid` (24B rand), TTL = 5m; stores session; returns response with base64url `challenge`, `options` (rpId, origin, uv_required, allowCredentials as base64url credential IDs), `tx_id_hex`, and `expires_at`.
      - Handler:
        - `func TxOptionsHandler(cfg *config.Config, txStore *TxSessionStore, db *sql.DB) http.HandlerFunc`
          - Requires `sid` cookie; loads session row from DB: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?`; checks expiry (`expires_at > now`), else 401.
          - Parses JSON body → `TxOptionsInbound`; calls `BuildTxOptions(...)` with `acctCBOR` from session.
          - Writes 200 JSON response.
    - Add tests `server/internal/tx/options_test.go` (see Tests below).
    - No route wiring changes in this step (handler is unit-testable via httptest; integration wiring can happen where server routes are defined).

    Description files (to create AND updates for modified sources)
    - Create `server/internal/tx/options.go.desc.md`: purpose, flow, error mapping (see Policies), session model, and Refs.
    - Update `server/internal/tx/tx.desc.md`: note the tx session store and handler; list invariants (TTL=5m; challenge derived from B; allowCredentials populated from DB).

    Request/response shape
    - Request (JSON):
      - `bundle_cbor_b64: string` — base64url of canonical CBOR bytes `B`.
      - Auth: cookie `sid` (session id) must be present and valid.
    - Response (JSON 200):
      - `tx_session_id: string`
      - `challenge: string` (base64url of 32 bytes)
      - `options: { rp_id: string; origin: string; uv_required: bool; allow_credentials: string[] }`
      - `tx_id_hex: string` (lowercase hex)
      - `expires_at: number` (unix seconds)

    Algorithm
    - Authn: Read `sid` cookie → query DB `sessions` for `{acct_cbor, expires_at}`; reject 401 if not found/expired.
    - Parse: JSON body (`bundle_cbor_b64`).
    - Validate & anchor: `tx.ValidateAndAnchorBundle(ctx, db, acct_cbor, bundle_cbor_b64)` → `B`, `challenge`, `tx_id`, `bundle`.
    - Credentials: `SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?` → list; if empty, 409 (inconsistent account state).
    - Session: generate tx session id (24B rand → base64url), TTL = `now + 5m`; store `{B, challenge, acct_cbor, expected_cred_ids, expiresAt}` in `TxSessionStore`.
    - Build options: `rpId = cfg.RP_ID`, `origin = cfg.Origin`, `uv_required = true`, `allow_credentials = base64url(credential_id[])`.
    - Respond: include `challenge` (base64url), `tx_session_id`, `tx_id_hex` (hex of anchor), and `expires_at`.

    Database interactions
    - `SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?` — validate auth session and obtain account identity.
    - `SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?` — enumerate account’s credential IDs.
    - No writes to DB in this step; tx sessions are in-memory (store capacity bounded).

    Policies & limits
    - TTL: tx sessions expire after 5 minutes; expired/unknown `tx_session_id` will be rejected by Step 22.
    - UV: `uv_required = true` in options (enforced during finish).
    - Allowlist: `allow_credentials` includes all credential IDs for the account.
    - Error mapping (temporary until Step 38 envelopes):
      - No/invalid cookie or expired session → 401.
      - `ErrBundleBase64` or `ErrBundleCBOR` → 400.
      - `ErrSenderKeyMismatch` → 401.
      - `ErrNonceNotMonotonic` → 409.
      - No credentials for account → 409.

    Sequencing
    - Step 24 will replace inline session lookup with middleware injecting `acct_cbor` into request context.
    - Step 22 will consume `tx_session_id` to verify CDJ/AD and persist transaction using the stored `B` and `challenge`.
    - Step 25/38 will add rate limiting, body size limits, and standardized error envelopes.

    Tests (happy path required, negative cases, invariants)
    - File: `server/internal/tx/options_test.go`
    - Setup helpers: create in-memory SQLite (use `storage.Migrate`), insert `accounts` row and at least one `credentials` row; insert `sessions` row with `session_id = sid`, `acct_cbor`, and non-expired `expires_at`.
    - Happy path: build bundle for that account (canonical CBOR), set cookie `sid`, POST JSON `{bundle_cbor_b64}` to handler; assert 200; response contains non-empty `tx_session_id`, base64url `challenge` of 32B, `options` with rpId/origin and uv_required=true, allow_credentials containing the account’s credentialId (base64url), `tx_id_hex`, and `expires_at ≈ now+5m`; store has entry for `tx_session_id` with matching `challenge`/`B`/`acct_cbor`.
    - Negative: missing cookie → 401.
    - Negative: expired or unknown session in DB → 401.
    - Negative: bad base64 in `bundle_cbor_b64` → 400.
    - Negative: bundle sender_key ≠ account key → 401.
    - Negative: nonce not strictly increasing (preload transactions with higher/equal nonce) → 409.
    - Negative: no credentials for account → 409.
    - Invariants: repeated call with same bundle and session yields same `challenge`/`tx_id`; different message/nonce changes both.
    - Commands:
      - `cd server && go test ./internal/tx -run TestTxOptionsHandler_ -v`
      - `cd server && go test ./internal/tx -v && go test ./...`

    Verification
    - Run the tests above; ensure all pass and error statuses match expectations.
    - Manual (optional): after route wiring, curl the endpoint with a valid cookie and bundle; confirm 200 and fields.
    - User verification commands
      ```bash
      cd server
      go test ./internal/tx -v   # includes handler tests
      go test ./...              # repo-wide sanity
      ```

    Acceptance criteria
    - Authenticated handler returns assertion options with a server-derived challenge bound to canonical `B` and includes `tx_session_id`, `tx_id_hex`, and `expires_at`.
    - Tx session stored in-memory with `B`, `challenge`, `acct_cbor`, and expected credential ids; TTL = 5 minutes.
    - Proper status codes for auth failures (401), malformed bundle (400), key mismatch (401), and nonce violations (409).

    Notes
    - Keep handler pure and testable; prefer a `BuildTxOptions` that accepts `now func() time.Time` for deterministic tests.
    - Error envelopes and rate limiting will be added in later steps; return plain statuses/messages now to keep scope tight.

    Refs
    - Refs: goal server-derived-challenge-and-txid; goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations

