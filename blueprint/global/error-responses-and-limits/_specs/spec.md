# R-ERR — Errors and Limits Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Standardize error responses and apply basic rate/size limits.

## Interfaces
- JSON error structure; status codes per requirement.

## Data / Models
- N/A.

## Algorithms
- Enforce request size caps; apply rate limiting (token bucket or middleware placeholder); map verifier/storage errors to appropriate status.

## Error Model
| HTTP | Code               | When                                                                                 |
| ---- | ------------------ | ------------------------------------------------------------------------------------ |
| 400  | `ERR_BAD_REQUEST`  | Malformed input, missing fields, base64 decode errors, signature parse/verify fails. |
| 401  | `ERR_UNAUTHORIZED` | No/invalid session for authenticated endpoints.                                      |
| 403  | `ERR_FORBIDDEN`    | Origin/RP mismatch; credential not linked to account; UV missing.                    |
| 409  | `ERR_CONFLICT`     | Nonce not monotonic; registration/login/tx session expired.                          |
| 413  | `ERR_TOO_LARGE`    | Payload exceeds max size.                                                            |
| 429  | `ERR_RATE_LIMIT`   | Per-IP or per-session rate exceeded.                                                 |
| 500  | `ERR_INTERNAL`     | Unhandled errors.                                                                    |

## Security / Privacy
- Avoid echoing sensitive data in error bodies; include only thumbprints/ids where necessary.
- Allowlists: `RP_ID_ALLOWLIST`, `ORIGIN_ALLOWLIST`.

## Errors / Observability
- Include error code, message, and optional correlation id.

## Limits
- TTL: registration/login/tx sessions expire in 5 minutes.
- Body size limit: ≤ 64 KB; message length ≤ 1 KB.
- Rate: suggest 10 requests/minute per IP for auth endpoints in the demo.
- Nonce policy: `nonce` must increase per account.

## Testing Strategy
- Fuzz invalid inputs; large payload tests; rate-limit behavior.

## Open Questions
- Whether to include standardized error codes.

Refs: decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; requirement R-ERR
