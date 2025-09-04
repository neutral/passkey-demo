# Standardized Error Envelope & Mapping

## Purpose
- Establish a uniform error response envelope and centralized mapping helpers so all handlers return consistent errors with stable codes and messages.

## Context
- Current handlers use `http.Error` with ad hoc strings; some mapping exists (`MapVerifyError`, `MapPolicyError`) but responses aren’t standardized.
- Quality gate R-ERR calls for consistent envelopes and HTTP status mapping.

## Proposal
- Introduce `internal/httpx/errors` with:
  - Envelope: `{ code: string, error: string, correlation_id?: string }`.
  - `func Write(w http.ResponseWriter, status int, code, msg string)` to write JSON with correct headers.
  - `func WithCorrelation(ctx) (ctx, id)` + `ExtractCorrelation(ctx)` to plumb an optional correlation ID.
  - Mappers: consolidate `MapVerifyError` and `MapPolicyError` into `Map` that accepts sentinel errors and returns `(status, code)`.
- Conventions:
  - Codes are PascalCase matching sentinel names where applicable (e.g., `ErrNonceNotMonotonic`).
  - User-facing `error` strings are concise; details live in logs.

## Migration Plan
1) Add `internal/httpx/errors` and unit tests.
2) Update registration/login/tx handlers to use `errors.Write` and `errors.Map`.
3) Remove ad hoc `http.Error` call sites and duplicate mappers.
4) Add correlation IDs to logs, propagate from request context.

## Risks & Mitigations
- Over-specifying codes: keep mapping minimal and aligned to sentinels; evolve via ADRs if needed.
- Backward compatibility: demo only; surface-level change safe.

## Testing Strategy
- Unit: mapper returns expected `(status, code)` for known sentinels; JSON shape exact match.
- Integration: handler tests assert envelope fields and statuses across success/failure paths.

## Acceptance Criteria
- All error responses use the standard envelope and centralized mapper.
- Handler tests validate correct status and `code` for representative failures.

## Refs
- Refs: requirement R-ERR; goal server-derived-challenge-and-txid; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; requirement R-PLAT-2

