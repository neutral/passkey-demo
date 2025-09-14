### Step 15 — /authn/passkey/registration/options handler (Done: 2025-09-01)

Scope
- Issue WebAuthn registration options with a fresh 32-byte challenge and create a short-lived registration session bound to RP ID and Origin.

Artifacts
- Code: `server/internal/webauthn/reg_options.go`
  - `RegSession`, `RegSessionStore(capacity)`, `BuildRegistrationOptions`, `RegistrationOptionsHandler`.
  - Session ID: 24 random bytes (base64url). Challenge: 32 random bytes (base64url). TTL: 5 minutes.
  - Options: `rp_id`, `origin`, `uv_required=true`, `attestation="none"`.
- Wiring: `server/cmd/api/main.go` mounts `POST /authn/passkey/registration/options` with in-memory store (capacity 10k).
- Docs: `server/internal/webauthn/reg_options.go.desc.md`.
- Tests: `server/internal/webauthn/reg_options_test.go` (builder and handler shape, TTL window, store contents).
- Specs: `specs/webauthn-registration-options-spec.md`, `specs/registration-options-test-vectors.md`, `specs/registration-options-risks-and-ambiguities.md`, `specs/explainer-registration-options.md`.

Verification
- `cd server && go test ./...` passes; handler returns JSON with expected fields; store contains corresponding entry.

Notes
- Single-use and consumption logic will be enforced in Step 17 (finish handler).
- Capacity limit is set to 10k for demo; consider rate limiting in hardening steps.

Refs
- requirement R-SEC-UV; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations

