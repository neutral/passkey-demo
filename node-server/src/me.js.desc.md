# Purpose
Serve authenticated account metadata under `/me/account_key`, exposing the canonical COSE EC2 key (base64url + derived hex helpers) for dashboard bundle building while keeping session enforcement centralized on the server.

# Key Logic
- Requires `req.session` (populated by cookie middleware); loads the account row by `acct_cbor`, falls back to recomputing the thumb when missing, and canonicalizes CBOR via `cbor-x` to surface base64url x/y coordinates.
- Emits structured logs `me_account_key`/`me_account_key_error` with `account_thumb_hex` and `correlation_id` and maps unauthenticated/missing rows to `ERR_UNAUTHORIZED`.
- Guards against malformed CBOR or missing coordinates by returning `ERR_INTERNAL` envelopes rather than leaking partial state.

# Interactions
Mounted by `src/server.js` after session middleware so the cookie loader runs first. Depends on `logger.js` for structured logging, `error.js` for JSON envelopes, and SQLite via prepared statements to read `accounts` rows.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-PLAT-3; requirement R-UI-2BTN; spec passkey-first-identity/spec.md; decision http-error-envelope
