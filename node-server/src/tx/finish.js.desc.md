# Purpose
Implements `POST /tx/signing/finish`, consuming the transaction session created by the options handler, verifying the WebAuthn assertion, enforcing policies, and persisting the signed bundle/anchors.

# Key Logic
- Validates request shape, loads tx session from `TxSessionStore`, and ensures the authenticated account matches.
- Calls `verifyAuthenticationResponse` with stored challenge/origin/rpId, requires UV, checks signature counter monotonicity, and maps failures to envelopes (401/403/409).
- Decodes canonical bundle bytes to extract nonce/message, updates credential `sign_count`, inserts a row into `transactions` with bundle/authenticator/client data/signature, deletes the tx session, and logs `tx_finish` events.

# Interactions
- Depends on `tx/options.js` session store, `bundle.js` canonical bytes, `logger.js` for success/error logs, and SQLite for credential lookups and inserts. Shares DB state with login/session flows for counters.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-FLOW-SIGN; decision webauthn-accept-high-s-signing-too; decision http-error-envelope; decision encoding-and-ceremony-guardrails
