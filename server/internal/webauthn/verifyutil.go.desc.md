# Purpose
Provide handler-facing utilities to map `VerifyAssertion` sentinel errors to HTTP responses and log structured, safe audit events for WebAuthn assertion verification.

# API
- `MapVerifyError(err error) (status int, kind string)`: maps sentinel errors to status codes and stable kind strings.
- `MapPolicyError(err error) (status int, kind string)`: maps RP ID / Origin sentinel errors to status codes and stable kind strings.
- `HashID(id []byte) string`: returns hex(SHA256(id)) for safe identifier logging.
- `LogAssertion(l *slog.Logger, v VerifyLog)`: emits a structured `webauthn_assert_verify` event.

# Logging Schema
- Fields: outcome, error_kind, rp_id, origin, uv, up, sign_count, credential_id_hash, challenge_id, trace_id, latency_ms.
- Do not log raw AD/CDJ/signature or raw IDs; hash identifiers.

# Refs
Refs: requirement R-FLOW-LOGIN; decision observability-and-privacy; audit crypto-audit-report; specs/webauthn-rp-origin-verification-spec.md
