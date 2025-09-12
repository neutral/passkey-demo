# Purpose
Handle `POST /authn/passkey/login/finish`: verify a WebAuthn assertion for login, enforce UV and policy checks, and establish a session.

# Key Logic
- Accepts `login_session_id` and WebAuthn assertion fields; decodes base64url inputs.
- Validates CDJ (`type=get`, challenge equality, origin policy) and AD (rpIdHash, UV flag).
- Looks up credential → account, converts account COSE to EC key, and verifies ES256 signature (low‑S enforced; high‑S normalized for compatibility).
- Enforces `signCount` strictly increasing when non-zero; updates stored counter.
- On success: creates a server session row with `expires_at = now + 3600s` (1 hour TTL) and sets `sid` cookie (HttpOnly, SameSite=Lax, Secure when `origin` is https).
- Errors: uses JSON error envelope `{code,error,correlation_id?}` mapped to 400/401/403/409/5xx.
- Logging: on verify failure, emits a single `webauthn_assert_verify` log via `LogAssertion` with `error_kind` and safe attributes (no raw materials). On success, emits `login_finish` with `account_thumb_hex`, `credential_id_hash`, and `sign_count` plus `correlation_id` when present.

# Interactions
- Reads/writes SQLite `credentials` and `sessions` tables; consumes `internal/encoding` and `internal/crypto`.
- Uses `internal/webauthn` helpers for parsing, policy, and verification; uses `internal/httpx/errors` for envelopes.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-SEC-UV; requirement R-ERR; decision http-error-envelope; decision webauthn-corrections-and-standardizations
