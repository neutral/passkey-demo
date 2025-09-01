# Purpose
Issue WebAuthn registration options and a fresh challenge, and store a short-lived registration session in memory.

# API
- `type RegSession{Challenge []byte, RP_ID, Origin string, ExpiresAt time.Time}` — server-side state.
- `type RegSessionStore` — concurrency-safe in-memory map with optional capacity.
- `BuildRegistrationOptions(cfg, store, now)` — creates a session and returns JSON-friendly response.
- `RegistrationOptionsHandler(cfg, store)` — `POST /authn/passkey/registration/options` handler.

# Behavior
- Session ID: 24 random bytes (base64url), Challenge: 32 random bytes (base64url).
- TTL: 5 minutes; stored alongside rp_id and origin from config.
- Options: `uv_required=true`, `attestation="none"`, echo `rp_id` and `origin`.

# Refs
Refs: specs/webauthn-registration-options-spec.md; specs/registration-options-test-vectors.md; requirement R-SEC-UV; requirement R-PLAT-2

