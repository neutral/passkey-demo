# Purpose
Verify WebAuthn assertion (login) and establish an application session via a secure cookie. Validates session, ClientDataJSON, authenticatorData (UV, rpIdHash), and signature, enforces monotonic signCount, updates DB, and issues a cookie.

# Key Logic
- Load `login_session_id` from `LoginSessionStore`; require not expired; single‑use after success.
- Parse `clientDataJSON` (`type=webauthn.get`, challenge match, origin allowlist with dev localhost exception).
- Parse `authenticatorData` header; require `rpIdHash` match and UV flag set; read `signCount`.
- Identify account by `rawId` (credential ID) → credentials row → account COSE key; convert to ECDSA and verify signature over `ad || SHA256(cdj)`.
- Enforce signCount policy: if the authenticator reports `signCount == 0`, treat it as "counter not supported" and do not enforce monotonicity; otherwise require strictly increasing and update the stored count. Create server session (1h) and set `sid` cookie (HttpOnly, SameSite=Lax, Secure for https origin).

# Interactions
- Uses `internal/encoding` for base64url and canonical CBOR; `internal/crypto` for COSE→ECDSA; `internal/storage` tables `accounts`, `credentials`, `sessions`.
- Error mapping via `MapVerifyError` and `MapPolicyError` for consistent HTTP statuses.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; spec R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations; spec session-cookies-usage-explainer
