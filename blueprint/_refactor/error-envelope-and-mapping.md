# Standardized Error Envelope & Mapping

## Purpose
- Establish a uniform error response envelope and centralized mapping helpers so all handlers return consistent errors with stable codes and messages.

## Context
- Current handlers use `http.Error` with ad hoc strings; some mapping exists (`MapVerifyError`, `MapPolicyError`) but responses aren’t standardized.
- Quality gate R-ERR calls for consistent envelopes and HTTP status mapping.

## Proposal
- Introduce `internal/httpx/errors` with:
  - Envelope: `{ code: string, error: string, correlation_id?: string }`.
  - `Write(w, status, code, msg)` to write JSON with correct headers.
  - `WriteReq(w, r, status, code, msg)` to include `correlation_id` when present.
  - Mappers: consolidate `MapVerifyError` and `MapPolicyError` into `Map(status, code)` (bridge in future step).
- Conventions:
  - Codes are PascalCase matching sentinel names where applicable (e.g., `ErrNonceNotMonotonic`).
  - User-facing `error` strings are concise; details live in logs.

## Migration Plan
1) Add `internal/httpx/errors` and unit tests.
2) Adopt in `/tx/signing/options`; expand to other handlers incrementally.
3) Keep existing policy/verify mappers for now; centralize mapping in a follow-up change.
4) Add Request ID middleware and include `correlation_id` in envelopes.

## Risks & Mitigations
- Over-specifying codes: keep mapping minimal and aligned to sentinels; evolve via ADRs if needed.
- Backward compatibility: demo only; surface-level change safe.

## Testing Strategy
- Unit: mapper returns expected `(status, code)` for known sentinels; JSON shape exact match.
- Integration: handler tests assert envelope fields and statuses across success/failure paths.

## Acceptance Criteria
- Error envelope used in modified endpoints; unit tests for envelope shape.
- Mapping consistency documented; consolidation planned via ADR.

## Refs
- Refs: requirement R-ERR; goal server-derived-challenge-and-txid; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; requirement R-PLAT-2
