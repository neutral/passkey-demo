Title: JSON Error Envelope and Centralized Mapping
Status: Accepted
Date: 2025-09-11

Context:
- Handlers previously used `http.Error` with ad hoc strings; mapping of policy/verify errors was inconsistent.
- Frontend needs stable `code`s for UX and tests; observability benefits from correlation ids.

Decision:
1) Standardize JSON error envelope: `{ code: string, error: string, correlation_id?: string }`.
2) Provide helpers under `internal/httpx/errors`:
   - `Write(w, status, code, message)`
   - `WriteReq(w, r, status, code, message)` which includes `correlation_id` when the request carries one.
3) Adopt envelope incrementally (first: `/tx/signing/options`); keep existing policy/verify mappers in `webauthn` and bridge over time.

Consequences:
- Consistent client experience; easier end-to-end assertions in tests; improved tracing with correlation ids.

Alternatives:
- Free-form error JSON or text bodies: rejected; breaks consistency and automated handling.

References:
- Implementation: `server/internal/httpx/errors/*`, first adoption in `server/internal/tx/options.go`.

Refs: requirement R-ERR; decision router-builder-wiring; decision request-id-and-slog-json

