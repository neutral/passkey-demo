# Purpose
Exposes the `POST /tx/signing/options` handler for the Node server. Validates transaction bundles, derives signing anchors, issues short-lived tx sessions, and returns WebAuthn assertion options bound to the server-derived challenge.

# Key Logic
- Verifies `req.session` exists (populated by cookie middleware) and parses `{ bundle_cbor_b64 }` from JSON.
- Calls `validateAndAnchorBundle` to canonicalize the bundle, enforce sender-key/nonce policies, and compute deterministic `challenge`/`tx_id` anchors; maps typed `BundleValidationError` kinds to `ERR_BAD_REQUEST`/`ERR_UNAUTHORIZED`/`ERR_CONFLICT` envelopes via `error.js`.
- Loads credential IDs for the authenticated account from SQLite, prunes an in-memory `TxSessionStore`, generates a 24-char base64url session id with collision retries, and stores `{ canonical B, challenge, txId, credentialIds, acctCbor, expiresAt }` for Step 12.
- Builds WebAuthn request options (`rpId`, `origin`, `timeout`, `userVerification`, `allowCredentials`) and returns JSON `{ tx_session_id, challenge, options, tx_id_hex, expires_at }` while logging `tx_options`/`tx_options_error` events.

# Interactions
- Depends on `bundle.js` for bundle validation, `logger.js` for success/error logs, `error.js` for envelopes, `limits.js` (through the `/tx` group) for 1 MiB body cap + rate limiter, and SQLite for account credential lookups. Exported `TxSessionStore`/`createTxSessionStore` are shared with the signing finish handler.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-FLOW-SIGN; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
