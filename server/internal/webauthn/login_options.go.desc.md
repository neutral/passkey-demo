# Purpose
Issue WebAuthn assertion (login) options and manage short‑lived login sessions bound to RP ID and Origin with a fresh challenge.

# Key Logic
- `LoginSessionStore(capacity)`: backed by generic `ttlstore` (capacity-bounded, 5m TTL with GC) holding `{challenge, rp_id, origin, expires_at}` keyed by `login_session_id`.
- `BuildLoginOptions`: 24‑byte session id, 32‑byte challenge (via `randutil`), TTL 5m; persists session and returns JSON with `options`.
- `LoginOptionsHandler`: POST endpoint that returns `rp_id`, `origin`, `uv_required=true`, `allow_credentials=[]`.

# Interactions
- Reads config from `internal/config` for `RP_ID`, `Origin`.
- Encodes binary fields using `internal/encoding` base64url helpers.
- Consumed by frontend to call `navigator.credentials.get(...)`.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; spec R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations
