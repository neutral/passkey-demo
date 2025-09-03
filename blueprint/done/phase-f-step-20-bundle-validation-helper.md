### Step 20 — Bundle validation helper (Done: 2025-09-03)

Verification notes
- Unit tests passed locally: `go test ./internal/tx -v` and `go test ./...` under `server/`.
- Anchors (CHALv1/TXIDv1) stable across repeated runs; negative cases return expected sentinel errors.

20. **Bundle validation helper**

    Scope
    - Server-side helper to validate a client-supplied Bundle (canonical CBOR), derive the signing anchors, and enforce account/key binding and nonce policy. No HTTP wiring yet.
    - Inputs: base64url-encoded CBOR `B` (`bundle_cbor_b64`) and the logged-in account identity (`acct_cbor`, canonical CBOR of COSE key) supplied by caller (Step 21 resolves session inline until Step 24 middleware).
    - Outputs: parsed `types.Bundle`, canonical bytes `B`, `challenge = SHA256("CHALv1" || B)`, `tx_id = SHA256("TXIDv1" || B)`.

    Source to add/modify
    - Add `server/internal/tx/bundle.go`: core helper and sentinels
      - `type AnchoredBundle struct { Bundle types.Bundle; B []byte; Challenge [32]byte; TxID [32]byte }`
      - `func ValidateAndAnchorBundle(ctx context.Context, db *sql.DB, acctCBOR []byte, bundleCBORBase64 string) (*AnchoredBundle, error)` — decodes base64→CBOR→`types.Bundle`, re-encodes canonical CBOR to form `B`, compares `SenderKey` to `acctCBOR`, checks per-account nonce monotonicity from DB, computes `challenge` and `tx_id`.
      - Sentinel errors: `ErrBundleBase64`, `ErrBundleCBOR`, `ErrSenderKeyMismatch`, `ErrNonceNotMonotonic`.
      - Constants: `anchorChallengePrefix = "CHALv1"`, `anchorTxIDPrefix = "TXIDv1"`.
    - Add tests `server/internal/tx/bundle_test.go`: unit tests listed below.
    - No changes to existing schemas or handlers in this step.

    Description files (to create AND updates for modified sources)
    - Create `server/internal/tx/tx.desc.md`: overview of the tx package and its role in signing; invariants (canonical CBOR, monotonic nonce); relations (used by `/tx/*` handlers); Refs to artifacts below.
    - Create `server/internal/tx/bundle.go.desc.md`: purpose of the helper; inputs/outputs; error semantics; DB lookups; hashing anchors; Refs.
    - Update `server/internal/types/types.go.desc.md`: note `types.Bundle` is consumed by `internal/tx` and hashed canonically for anchors; Refs.

    Request/response shape (helper I/O)
    - Input to helper:
      - `bundle_cbor_b64: string` — base64url (padding optional) of CBOR bytes `B` sent by client.
      - `acct_cbor: []byte` — canonical CBOR of the logged-in account’s COSE public key (from session/account lookup).
    - Output from helper:
      - `B: []byte` (canonical CBOR), `challenge: [32]byte`, `tx_id: [32]byte`, and parsed `types.Bundle` fields for caller reuse (nonce, message).

    Algorithm
    - Decode base64url using `internal/encoding` (accept padded/unpadded) → raw CBOR.
    - Decode CBOR into `types.Bundle` with robust error mapping.
    - Re-encode `Bundle` with `internal/encoding.EncodeCanonical(...)` to derive canonical bytes `B`.
    - Compute anchors:
      - `challenge = SHA256("CHALv1" || B)`
      - `tx_id     = SHA256("TXIDv1" || B)`
    - Verify account binding:
      - Canonically CBOR-encode `Bundle.SenderKey` and compare bytes to `acct_cbor` (exact match required) → else `ErrSenderKeyMismatch`.
    - Enforce nonce policy:
      - Query DB: `SELECT COALESCE(MAX(nonce), -1) FROM transactions WHERE acct_cbor = ?` → `last_nonce`.
      - Require `bundle.Nonce > last_nonce`; otherwise `ErrNonceNotMonotonic`.
    - Return `AnchoredBundle` containing `Bundle`, `B`, and anchors.

    Database interactions
    - Read-only query against `transactions(acct_cbor, nonce)` to find `MAX(nonce)` for the account.
    - No schema changes; no writes in this step.

    Policies & limits
    - Canonical CBOR determinism: server recomputes `B` regardless of client encoding; hashes derive from canonical `B` only.
    - Account binding: `sender_key` must equal the logged-in account COSE key (canonical CBOR bytes compare equal).
    - Nonce strictly increasing per account enforced here; collision/regression returns a deterministic error.
    - Size limits and error envelopes are formalized in Step 38; this helper returns typed errors for handler mapping in Step 21.

    Sequencing
    - Step 21 will call this helper inside `/tx/signing/options`; until Step 24 middleware lands, the handler resolves `acct_cbor` from session inline.
    - Step 22 consumes the same anchors (`tx_id`, `B`) via the tx session created in Step 21.

    Tests (happy path required, negative cases, invariants)
    - File: `server/internal/tx/bundle_test.go`
    - Happy path: non-canonical input CBOR decodes and re-encodes canonically; `challenge`/`tx_id` stable and match golden vectors; sender key matches account; nonce monotonic with empty history.
    - Negative: base64 decode error → `ErrBundleBase64`.
    - Negative: CBOR decode error (truncated map) → `ErrBundleCBOR`.
    - Negative: sender key mismatch (different key or altered X/Y) → `ErrSenderKeyMismatch`.
    - Negative: nonce <= last seen in DB → `ErrNonceNotMonotonic`.
    - Invariants: canonical `B` equality on repeated calls; hash stability (same inputs → same anchors); differing message/nonce → different anchors.
    - Commands: `cd server && go test ./internal/tx -v` and `go test ./...` should pass.

    Verification
    - Unit: run tests above; ensure all pass and error kinds match expectations.
    - Manual (optional now): none required; endpoint wiring occurs in Step 21.
    - User verification commands
      ```bash
      cd server
      go test ./internal/tx -v
      go test ./...    # sanity across the module
      ```

    Acceptance criteria
    - Helper returns canonical `B`, `challenge`, and `tx_id` deterministically for the same logical bundle.
    - Sender key equality check enforces account binding using canonical CBOR bytes.
    - Nonce policy rejects non-increasing values based on DB history.
    - Typed sentinel errors are surfaced for handler mapping in Step 21.

    Notes
    - No HTTP, cookies, or sessions added here; Step 21 wires the handler and tx-session cache.
    - Error-to-HTTP mapping and limits (message length, body size) will be enforced in later steps (25, 38).

    Refs
    - Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations

