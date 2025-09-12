# Purpose
Handle `POST /authn/passkey/registration/finish`: validate attestation, enforce origin/rpId/UV policies, and persist the account + credential.

# Key Logic
- Accepts `reg_session_id` and attestation fields; decodes base64url inputs.
- Validates CDJ (`type=create`, challenge equality, origin policy) and parses `attestationObject` to extract AD, AAGUID, credential id, and COSE EC2 key.
- Enforces `rpIdHash` match and requires UV; validates COSE → EC public key conversion.
- Persists account (idempotent on `acct_cbor`) and credential (conflicts produce 409).
- Deletes the registration session after success (single-use).
- Errors: uses JSON error envelope `{code,error,correlation_id?}` mapped to 400/401/403/409/5xx.

# Interactions
- Writes `accounts` and `credentials` tables; uses `internal/encoding` and `internal/crypto`.
- Uses `internal/webauthn` helpers for parsing and policy; uses `internal/httpx/errors` for envelopes.

# Refs
Refs: requirement R-FLOW-REG; requirement R-SEC-UV; requirement R-ERR; decision http-error-envelope; decision webauthn-corrections-and-standardizations

