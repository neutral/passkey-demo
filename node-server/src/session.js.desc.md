# Purpose
Attach authenticated session context to requests by reading the `sid` cookie and looking up the session in SQLite. Optionally preload credential ids and perform a rolling TTL refresh.

# Key Logic
- Parse `Cookie` header (`sid`).
- DB lookup: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?` and ensure `expires_at > now`.
- Attach `req.session = { sid, acct_cbor, expires_at[, credential_ids] }`.
- Optional preload of `credential_ids` from `credentials` table; optional rolling refresh updates `expires_at`.
- Missing/expired/unknown sessions leave `req.session` undefined; middleware does not return 401.

# Interactions
- Wired in `src/server.js` after CORS; used by later protected routes.

# Refs
Refs: spec session-cookies-usage-explainer; requirement R-OPS-DEV

