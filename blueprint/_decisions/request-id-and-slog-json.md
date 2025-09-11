Title: Request IDs and Structured Logging
Status: Accepted (Request IDs); Planned (slog JSON)
Date: 2025-09-11

Context:
- Troubleshooting cross-tier issues benefits from a correlation id per request; error envelopes can expose this id.
- Logging format is currently basic; future alignment to structured JSON (`slog`) is desired.

Decision:
1) Add `RequestID` middleware under `internal/httpx/middleware`:
   - Accepts `X-Request-ID` or generates a base64url id.
   - Stores id in request context; mirrors it on `X-Request-ID` response header.
2) Error envelope uses `WriteReq` to include `correlation_id` when present.
3) Defer full slog JSON logging rollout to a follow-up step.

Consequences:
- Immediate correlation across logs and API responses; unlocks better triage.
- Minimal runtime overhead.

Alternatives:
- No request id: poorer debuggability.

References:
- Implementation: `server/internal/httpx/middleware/request_id.go`; usage in `server/internal/app/router.go`.

Refs: requirement R-ERR; goal simple-ui-and-storage; decision http-error-envelope

