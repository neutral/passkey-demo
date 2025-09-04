### Step 23 — /tx/list handler (Done: 2025-09-04)

Verification notes
- Unit tests executed: `cd server && GOCACHE=$(pwd)/.gocache go test ./internal/tx -v` and `go test ./...` → all green.
- Phase F E2E confirms: options → finish → list returns newly persisted transaction.

23. **/tx/list handler (auth required)**

    Scope
    - Add `GET /tx/list` to return transactions for the authenticated account (scoped by session cookie `sid`).
    - Response includes: `tx_id_hex`, `nonce`, `message`, `created_at` for each item.
    - Minimal policy for demo; no pagination yet (keep interface simple for Phase F).

    Source to add/modify
    - Add `server/internal/tx/list.go`: builder + handler
      - `type TxListItem struct { TxIDHex string; Nonce uint64; Message string; CreatedAt int64 }`
      - `type TxListResponse struct { Items []TxListItem }`
      - `func BuildTxList(ctx context.Context, db *sql.DB, sid string) (TxListResponse, error)` — resolves auth session (`sid` → `acct_cbor`), queries `transactions` by `acct_cbor`, builds items with hex(tx_id).
      - `func TxListHandler(db *sql.DB) http.HandlerFunc` — reads cookie `sid`, calls builder, returns JSON.
    - Add tests: `server/internal/tx/list_test.go` (see Tests below).
    - No router changes here; handler tested via `httptest`.

    Description files (to create AND updates for modified sources)
    - Create `server/internal/tx/list.go.desc.md`: purpose, auth scoping by session, fields returned, and Refs.
    - Update `server/internal/tx/tx.desc.md`: document list handler and its relation to finish (records returned are those persisted in Step 22).

    Request/response shape
    - Request: `GET /tx/list` with cookie `sid`.
    - Response (JSON 200): `{ "items": [ { "tx_id_hex": string, "nonce": number, "message": string, "created_at": number } ] }`.

    Algorithm
    - Authn: read cookie `sid`; if missing → 401.
    - Session lookup: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`.
      - If not found or expired → 401.
    - Query transactions: `SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC`.
    - Convert `tx_id` (BLOB) to lowercase hex; copy other fields as-is; return as `items`.

    Database interactions
    - Read: `sessions` (by `sid`), `transactions` (by `acct_cbor`).
    - No writes.

    Policies & limits
    - Auth required: session cookie must be valid and unexpired.
    - Data exposure: returns only current account’s transactions; no cross-account leakage.
    - Ordering: newest first (`created_at DESC`).
    - Pagination: omitted for Phase F (could add `?limit` later; out of scope here).

    Sequencing
    - Completes Phase F; depends on Step 22 persisting transactions and Step 21 creating tx sessions.
    - Step 24 middleware will later unify session resolution; keep inline lookup now.

    Tests (happy path required, negative cases, invariants)
    - File: `server/internal/tx/list_test.go`
    - Setup: in-memory SQLite with `storage.Migrate`; insert two accounts; insert sessions for both; insert several transactions for account A and one for account B.
    - Happy path: request with cookie for account A → 200; items length equals number of A’s transactions; each item has non-empty `tx_id_hex`; items ordered by `created_at DESC`.
    - Negative: missing cookie → 401; unknown/expired session → 401.
    - Isolation: request with account B’s cookie only returns B’s single transaction; A’s items never appear.
    - Commands:
      - `cd server && go test ./internal/tx -run TestTxList_ -v`
      - `cd server && go test ./internal/tx -v && go test ./...`

    Verification
    - Unit: tests above pass; list returns only the authenticated account’s transactions with correct shapes and ordering.
    - Manual (optional): after signing a message (Step 22), call `GET /tx/list` in a browser session; see the new entry.
    - User verification commands
      ```bash
      cd server
      go test ./internal/tx -v   # includes list handler tests
      go test ./...              # repo-wide sanity
      ```

    Acceptance criteria
    - Returns 200 with `items` for valid cookie/session; 401 without/expired session.
    - Items include `tx_id_hex`, `nonce`, `message`, `created_at` and are scoped to the account.

    Notes
    - Keep handler simple and cohesive with Step 21/22 patterns; defer pagination and filtering to later phases.

    Refs
    - Refs: goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations

    Phase-wide additional tests (Phase F)
    - File: `server/internal/tx/phase_f_e2e_test.go`
    - End-to-end: options → finish → list
      - Build bundle for account A; call options; sign assertion; call finish; then `GET /tx/list` and verify the inserted item appears with matching `nonce`/`message` and a deterministic `tx_id_hex`.
    - Determinism: rebuild the same bundle and assert `tx_id_hex` equals the previous value; attempting options with the same nonce again yields 409 (from Step 21).
    - Allowlist: ensure `allowCredentials` from options includes the credential used in finish; list subsequently contains the persisted transaction.
    - Policy cross-checks: induce wrong origin or rpId during finish and assert proper 403 mapping; list remains unchanged.

