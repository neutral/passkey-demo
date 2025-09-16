# Step 13 — Node: Tx list (`/tx/list`) (Done)

Completed: 2025-09-16
Verification notes:
- `cd node-server && node --test test/tx-list.test.js`
- `cd node-server && npm test`
- `bash tools/desc-check.sh HEAD~1 HEAD`

### Step 13 — Node: Tx list (`/tx/list`)

Scope
- Implement authenticated `GET /tx/list` returning the requesting account’s transactions in reverse chronological order.
- Reuse the session middleware contract (expects `req.session.acct_cbor`) and existing error envelope helper for parity with prior steps.

Source to add/modify
- `node-server/src/tx/list.js` — new Express router exposing `/tx/list`, queries SQLite for the account’s transactions, and serializes them to JSON.
- `node-server/src/server.js` — mount the list router alongside signing routes so the handler is reachable during app bootstrap and tests.
- `node-server/test/tx-list.test.js` — unit tests covering success/empty/error/unauthorized flows via an in-memory SQLite database.

Description files
- `node-server/src/tx/list.js.desc.md` — document the handler’s responsibilities, auth expectations, ordering, and envelope semantics.
- `node-server/src/tx/tx.desc.md` — update overview to include the list route and how it complements options/finish.
- `node-server/src/server.js.desc.md` — note the additional `/tx/list` router wiring.

Blueprint updates
- None anticipated; existing R-UI-2BTN acceptance criteria already require the dashboard list and R-PLAT-3 covers SQLite persistence. During implementation, confirm those artifacts stay accurate and update if behavior diverges.

Request/response shape
- Request: `GET /tx/list` with HttpOnly `sid` cookie; no query/body payload.
- Success (`200 OK`): `{ "items": [{ "tx_id_hex": string, "nonce": number, "message": string, "created_at": number }, ...] }` with an empty array when no transactions exist.
- Errors: `401 Unauthorized` (`{ code: 'unauthorized', error: 'Unauthorized', correlation_id? }`) when session missing/invalid; `500 Internal` on DB failures (reuse existing envelope conventions).

Algorithm
- Require `req.session` and `req.session.acct_cbor`; otherwise return 401 via `writeError`.
- Normalize `acct_cbor` to a Buffer and prepare a reusable statement: `SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC`.
- Execute the query, mapping each row to `{ tx_id_hex, nonce, message, created_at }` (`tx_id_hex` via `Buffer.from(row.tx_id).toString('hex')`, `nonce` as `Number` keeping safe integers).
- Handle DB exceptions by logging (via existing logger if already wired) and returning a 500 envelope; otherwise respond with JSON and status 200.
- Ensure response uses `res.json`/`Content-Type: application/json` and does not expose raw bundle fields.

Database interactions
- Read-only `SELECT ... ORDER BY created_at DESC` against `transactions` filtered by `acct_cbor`.
- No mutations; use `better-sqlite3` prepared statement for efficiency and deterministic ordering.

Policies & limits
- Honor authentication policy: only sessions established via middleware may view transactions.
- Preserve strict ordering by `created_at DESC`; do not leak other accounts’ rows.
- Keep payload small—only expose public fields (hex id, nonce, message, timestamp); no pagination yet (acceptable given demo scope).

Sequencing
- Depends on prior steps that create transactions (Step 12) so list queries have data to return.
- Shares `/tx` mount path with options/finish; ensure router mount order doesn’t shadow existing handlers.
- Positioned before Step 14 (error envelope cleanup) and Step 15 (logging); keep error codes consistent with current helpers to avoid churn.

Tests
- `node-server/test/tx-list.test.js`
  - Happy path: authenticated request with two transactions having different `created_at` values returns items sorted DESC with expected hex/nonce/message fields.
  - Empty list: authenticated account with no transactions returns `{ items: [] }` and status 200.
  - Unauthorized: no `req.session` (or missing `acct_cbor`) → 401 with `code: 'unauthorized'`.
  - Database failure: stub `.all()` to throw to assert 500 envelope and no data leakage.
  - Optional: ensure non-buffer `acct_cbor` forms (Uint8Array from session middleware) are handled correctly.
- Run via `node --test test/tx-list.test.js` and include in `npm test` aggregate.

Verification
- Unit: `cd node-server && node --test test/tx-list.test.js`; fix issues and re-run until green.
- Aggregate: `cd node-server && npm test` to keep Step 11/12 suites passing alongside new tests.
- Manual smoke: seed a SQLite DB with an account/session/transaction, start the server with that DB, and curl `/tx/list` with and without the `sid` cookie validating 200 vs 401 responses.
- After changes, run `bash tools/desc-check.sh HEAD~1 HEAD` to ensure description coverage stays in sync.
- Re-run the full test suite after addressing any fixes to confirm stability before marking the step ready.

User verification commands
```bash
cd node-server
rm -f ../tmp/tx-list.db
sqlite3 ../tmp/tx-list.db <<'SQL'
CREATE TABLE IF NOT EXISTS accounts (acct_cbor BLOB PRIMARY KEY, acct_thumb BLOB, created_at INTEGER);
CREATE TABLE IF NOT EXISTS sessions (session_id TEXT PRIMARY KEY, acct_cbor BLOB, expires_at INTEGER, created_at INTEGER);
CREATE TABLE IF NOT EXISTS transactions (tx_id BLOB PRIMARY KEY, acct_cbor BLOB, nonce INTEGER, message TEXT, bundle_cbor BLOB, auth_data BLOB, client_data BLOB, signature BLOB, created_at INTEGER);
INSERT INTO accounts VALUES (X'0102', X'0304', strftime('%s','now'));
INSERT INTO sessions VALUES ('manual-sid', X'0102', strftime('%s','now') + 600, strftime('%s','now'));
INSERT INTO transactions VALUES (X'0A0B0C0D', X'0102', 42, 'demo message', X'00', X'01', X'02', X'03', strftime('%s','now'));
SQL
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=../tmp/tx-list.db node src/server.js &
API_PID=$!
sleep 1
curl -sS -b 'sid=manual-sid' http://127.0.0.1:8080/tx/list
curl -sS http://127.0.0.1:8080/tx/list -i
kill $API_PID || true
```

Acceptance criteria
- Authenticated calls return only that account’s transactions ordered by `created_at DESC`, matching `{ items: [...] }` shape consumed by the web dashboard.
- Unauthorized requests (missing/expired session) receive 401 with existing error envelope format.
- DB failures propagate as 500 with no partial responses.

Notes
- When adding `list.js`, mirror the coding style from `options.js`/`finish.js` (async error handling, Buffer normalization).
- Consider factoring shared helpers (e.g., `toAccountBuffer`) only if it keeps scope tight; otherwise inline within `list.js` for now.
- Keep logging minimal this step; Step 15 will introduce structured log events.

Refs: requirement R-UI-2BTN; requirement R-PLAT-3
