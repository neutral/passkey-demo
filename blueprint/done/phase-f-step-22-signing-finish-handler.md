### Step 22 — /tx/signing/finish handler (Done: 2025-09-04)

Verification notes
- Unit tests executed: `cd server && GOCACHE=$(pwd)/.gocache go test ./internal/tx -v` and `go test ./...` → all green.
- Happy path validates CDJ/AD/signature, signCount advance, transaction insert, and tx-session consumption.

22. **/tx/signing/finish handler (auth required)**

    Scope
    - Add `POST /tx/signing/finish` to complete the signing ceremony:
      - Requires `sid` cookie (auth session) and a valid `tx_session_id` from Step 21.
      - Verifies CDJ (`type=get`, `challenge` equals stored bytes, origin policy) and AD (rpIdHash, UV flag, signCount monotonic).
      - Verifies signature using the account’s COSE key; enforces low‑S and strict DER.
      - Persists the transaction row with `tx_id`, canonical `bundle_cbor` `B`, AD/CDJ/signature, and parsed `nonce`/`message`.

    Source to add/modify
    - Add `server/internal/tx/finish.go`: builder + handler
      - Types:
        - `type TxFinishInbound struct { TxSessionID string; ID string; RawID string; Type string; Response struct { AuthenticatorData string; ClientDataJSON string; Signature string; UserHandle string } }`
        - `type TxFinishResponse struct { TxIDHex string `json:"tx_id_hex"`; Stored bool `json:"stored"` }`
      - Builder:
        - `func BuildTxFinish(ctx context.Context, cfg *config.Config, txStore *TxSessionStore, db *sql.DB, sid string, in TxFinishInbound, now func() time.Time) (TxFinishResponse, error)`
          - Loads auth session by `sid` from DB; loads tx session by `in.TxSessionID` from `txStore` and verifies not expired.
          - Parses/validates CDJ and AD; enforces UV and rpIdHash; checks CDJ challenge equals stored `Challenge` bytes and origin policy.
          - Ensures `in.RawID` decodes to a known credential that belongs to `txSession.AcctCBOR`; `in.ID` matches `in.RawID`.
          - Checks `in.RawID` is in `txSession.CredentialIDs` allowlist.
          - Decodes account key from `txSession.AcctCBOR` to `types.CoseEC2`, converts to `ecdsa.PublicKey` via `crypto.ToECDSA`.
          - Verifies signature with `VerifyAssertion` and enforces `SignCount` strictly increasing (updates DB on success).
          - Recomputes `tx_id = SHA-256("TXIDv1" || B)` using `txSession.B` and decodes `B` → `types.Bundle` to extract `nonce` and `message`.
          - Inserts row into `transactions` and deletes the tx session (single‑use) on success.
      - Handler:
        - `func TxFinishHandler(cfg *config.Config, txStore *TxSessionStore, db *sql.DB) http.HandlerFunc` — parses cookie/body, calls `BuildTxFinish`, maps errors to statuses, and returns JSON `{ tx_id_hex, stored: true }`.
    - Tests: `server/internal/tx/finish_test.go` (see Tests below).
    - Update `server/internal/tx/tx.desc.md`: include finish handler responsibilities and invariants.

    Description files (to create AND updates for modified sources)
    - Create `server/internal/tx/finish.go.desc.md`: purpose, inputs/outputs, error mapping, DB writes, and invariants (UV, signCount, rpIdHash, challenge binding).
    - Update `server/internal/tx/tx.desc.md`: mention finish flow and persisted schema fields; note single‑use tx sessions.

    Request/response shape
    - Request (JSON; cookie `sid` required):
      - `tx_session_id: string`
      - `id: string` (base64url credentialId)
      - `rawId: string` (base64url credentialId; must equal `id`)
      - `type: "public-key"`
      - `response: { authenticatorData: string, clientDataJSON: string, signature: string, userHandle: string }` (base64url fields)
    - Response (JSON 200):
      - `tx_id_hex: string`
      - `stored: true`

    Algorithm
    - Authn: Read `sid` cookie → DB `sessions` lookup: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`; reject 401 if missing/expired.
    - Tx session: `txStore.Get(tx_session_id)`; reject 401 if missing/expired.
    - Parse CDJ: base64url decode → `ParseClientDataJSON`; require `type=get`; compare `cdj.Challenge` bytes to stored `Challenge`; origin policy via `CheckOrigin` (dev localhost OK) and `MapPolicyError`.
    - Parse AD: base64url decode → `ParseAuthData`; require `CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist)`; require `HasUV(ad.Flags)`.
    - Credential: decode `rawId`; require equals `id`; DB `SELECT acct_cbor_fk, sign_count FROM credentials WHERE credential_id=?`; require same account as `txSession.AcctCBOR`; require id present in `txSession.CredentialIDs`.
    - Account key: decode `txSession.AcctCBOR` → `types.CoseEC2`; convert with `crypto.ToECDSA`.
    - Verify: `VerifyAssertion(pub, adRaw, cdjRaw, sigRaw)`; map errors with `MapVerifyError`.
    - SignCount: enforce strictly increasing vs stored; update DB on success.
    - Persist: recompute `tx_id`, decode `B` → `types.Bundle`, then `INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at)`.
    - Cleanup: delete tx session on success; return `{ tx_id_hex, stored: true }`.

    Database interactions
    - Read: `sessions` (by cookie `sid`), `credentials` (by credential_id), and optional join validation on account.
    - Write: `UPDATE credentials SET sign_count = ? WHERE credential_id=?`; `INSERT INTO transactions (...)`.

    Policies & limits
    - TTL: tx sessions expire after 5 minutes (reject if expired/not found).
    - UV: required and asserted in AD; options also set `uv_required=true` (Step 21).
    - Origin/rpId: enforced as in login finish (`CheckOrigin`, `CheckRpIdHashAllowed`).
    - signCount: strictly increasing; equal/lower → 409.
    - Error mapping (until Step 38 envelopes):
      - Missing/expired cookie or tx session → 401.
      - CDJ type mismatch, bad base64/CBOR → 400.
      - Policy mismatches (origin/rpId/UV) → 403.
      - Sender/account mismatch or unexpected credential id → 401.
      - Verify errors: `ErrMalformedDER` → 400; `ErrHighS`/`ErrBadSignature` → 401.

    Sequencing
    - Step 24 middleware will unify session resolution; leave inline lookup for now.
    - Step 25/38 will add standardized error envelopes and body limits.

    Tests (happy path required, negative cases, invariants)
    - File: `server/internal/tx/finish_test.go`
    - Happy path:
      - Arrange: create account+credential (ECDSA P‑256), build a minimal Bundle and call `BuildTxOptions` to create a tx session; craft CDJ with returned `challenge`, AD with UV and rpIdHash; sign `SHA256(AD||SHA256(CDJ))` with the private key (normalize to low‑S); POST to handler with cookie `sid` and tx payload.
      - Assert: 200; response has `tx_id_hex`; transactions row exists with `acct_cbor`, `bundle_cbor=B`, `nonce`, `message`, and `created_at`; credential `sign_count` advanced; tx session consumed.
    - Negative cases:
      - Missing cookie or expired tx session → 401.
      - CDJ `type=create` or wrong `challenge` → 400/401.
      - Origin not allowed → 403.
      - rpIdHash mismatch → 403.
      - Missing UV → 403.
      - Unknown credentialId → 401.
      - Credential not linked to account or not in tx session allowlist → 401.
      - Non‑increasing signCount (equal/lower) → 409.
      - Malformed DER signature → 400; High‑S signature → 401.
    - Commands:
      - `cd server && go test ./internal/tx -run TestTxFinish_ -v`
      - `cd server && go test ./internal/tx -v && go test ./...`

    Verification
    - Unit: all tests above pass; insert verifies row content and signCount; session deleted.
    - Manual (optional): after routing, run finish with a valid browser assertion; expect 200 with `tx_id_hex`.
    - User verification commands
      ```bash
      cd server
      go test ./internal/tx -v   # includes finish handler tests
      go test ./...              # repo-wide sanity
      ```

    Acceptance criteria
    - Handler verifies CDJ/AD/signature, enforces policies, persists transaction, and advances `sign_count`.
    - Returns `{ tx_id_hex, stored: true }` and removes the tx session (single‑use).
    - Proper statuses for auth/policy/verify/signCount failures as specified.

    Notes
    - Keep builder isolated and deterministic by injecting `now func() time.Time`.
    - Error envelopes standardized later (Step 38); return plain statuses/messages now.

    Refs
    - Refs: goal server-derived-challenge-and-txid; goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations

