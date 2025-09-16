# Purpose
Serve authenticated `GET /tx/list` responses by reading the requesting account’s transactions from SQLite, ordering them newest-first, and serializing the minimal fields the dashboard expects.

# Key Logic
- Validates `req.session.acct_cbor`; unauthorized requests return the shared JSON envelope with `401`.
- Executes a prepared `SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC` and maps results to `{ tx_id_hex, nonce, message, created_at }`.
- Logs failures with `tx_list_error` and surfaces a `500` envelope when the query throws.

# Interactions
- Mounted under `/tx` by `server.js` alongside signing option/finish routes; depends on `session.js` to populate `req.session` and `error.js` for envelopes.
- Shares transaction storage created in Step 12; outputs feed the web dashboard (R-UI-2BTN).

# Refs
Refs: goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-3; requirement R-UI-2BTN; requirement R-FLOW-SIGN
