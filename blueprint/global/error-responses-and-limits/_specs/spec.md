# R-ERR — Errors and Limits Spec

## Metadata
- Status: Approved
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Standardize error responses and apply rate/size limits at the router group level.

## Interfaces
- JSON error envelope and status codes per requirement.

## Data / Models
- N/A.

## Algorithms
- Enforce request size caps via middleware; apply token-bucket rate limiting; map verifier/policy/storage errors to appropriate status.

## Error Model
| HTTP | Code               | When                                                                                 |
| ---- | ------------------ | ------------------------------------------------------------------------------------ |
| 400  | `ERR_BAD_REQUEST`  | Malformed input, missing fields, base64 decode errors, signature parse/verify fails. |
| 401  | `ERR_UNAUTHORIZED` | No/invalid session for authenticated endpoints.                                      |
| 403  | `ERR_FORBIDDEN`    | Origin/RP mismatch; credential not linked to account; UV missing.                    |
| 409  | `ERR_CONFLICT`     | Nonce not monotonic; registration/login/tx session expired.                          |
| 413  | `ERR_TOO_LARGE`    | Payload exceeds max size (body limit).                                               |
| 429  | `ERR_RATE_LIMIT`   | Per-IP or per-session rate exceeded.                                                 |
| 500  | `ERR_INTERNAL`     | Unhandled errors.                                                                    |

## Security / Privacy
- Avoid echoing sensitive data in error bodies; include only thumbprints/ids where necessary.
- Allowlists: `RP_ID_ALLOWLIST`, `ORIGIN_ALLOWLIST`.

## Errors / Observability
- Standard envelope: `{ code, error, correlation_id? }`.
- Include `correlation_id` when a request id is present (via RequestID middleware).

## Limits
- TTL: registration/login/tx sessions expire in 5 minutes.
- Body size limit: 1 MiB for `/authn/*` and `/tx/*` groups.
- Rate: token bucket per remote IP; burst 20, target ~10 rps for demo.
- Nonce policy: `nonce` must increase per account.

## Testing Strategy
- Fuzz invalid inputs; large payload tests; rate-limit behavior.

## Open Questions
- None.

Refs: decision encoding-and-ceremony-guardrails; decision http-error-envelope; decision request-id-and-slog-json; decision router-builder-wiring; requirement R-ERR
