# Purpose
Authenticate requests via cookie `sid`, load the server session and related account data from the database, and attach a `SessionContext` (account CBOR and credential IDs) to the request context for downstream handlers.

# Key Logic
- Read `sid` cookie; look up `{acct_cbor, expires_at}` in `sessions`.
- Enforce expiry; optionally refresh expiry when less than half of the TTL remains.
- Load `credential_id` list for the account; attach data via `WithSession`.
- If no valid session, pass through without context; protected handlers should return 401.

# Interactions
- Wraps `/tx/*` routes. Tx handlers (`options`, `finish`, `list`) can retrieve `acct_cbor` from `FromSession(ctx)` instead of re‑querying by cookie.

# Refs
Refs: goal passkey-registration-login-uv; goal transaction-content-signing; requirement R-PLAT-2; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations

