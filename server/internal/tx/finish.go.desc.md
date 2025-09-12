# Purpose
Complete the transaction signing ceremony: validate CDJ/AD against the tx session and account policy, verify the ES256 assertion, and persist the transaction with canonical `bundle_cbor`.

# Key Logic
- Auth: requires session middleware; account context is taken from request context (no cookie/DB fallback).
- Tx session: load from `TxSessionStore` by `tx_session_id`; ensure not expired; single-use on success.
- Validate: parse `clientDataJSON` and `authenticatorData`; enforce `type=get`, `challenge` equality, origin policy, `rpIdHash` match, and UV.
- Credential: ensure presented `credential_id` exists (via `CredentialsRepo`), belongs to `acct_cbor`, and appears in the tx-session allowlist.
- SignCount policy: if `ad.SignCount == 0`, treat as counter‑not‑supported and do not enforce monotonicity or update stored count; otherwise require strictly increasing.
- Verify: use `VerifyAssertion` (strict DER). If the only failure is high‑S, normalize S and accept (compatibility with some authenticators). Advance `sign_count` when increasing.
- Persist: compute `tx_id = SHA-256("TXIDv1" || B)`; decode `B` to extract `nonce`/`message`; insert into `transactions`.

# Interactions
- Called by `POST /tx/signing/finish`; depends on `TxSessionStore` and the account session established by login.
- Uses `internal/webauthn` for CDJ/AD parsing, policy checks, and signature verification.

# Refs
Refs: goal server-derived-challenge-and-txid; goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations
