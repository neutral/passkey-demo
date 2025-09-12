Title: Structured Logging with log/slog — Approach and Guidelines
Status: Accepted
Date: 2025-09-12

Context:
- The codebase uses a JSON error envelope and Request ID middleware for correlation but still mixes `log.Printf` with early `slog` utilities in `internal/webauthn`.
- We need robust, efficient, and clean logging with consistent structure, privacy guardrails, and low operator friction.
- Insights synthesized from “Logging in Go with Slog: A Practitioner’s Guide” (Dash0) and the associated HN thread; the HN item currently has minimal comments, so guidance is primarily from the Dash0 article.

Decision:
1) Adopt log/slog as the unified logging API with JSON output by default.
   - Initialize a process‑wide default logger in `server/cmd/api` using `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ Level: &levelVar })`.
   - Control verbosity via env `LOG_LEVEL` (debug|info|warn|error) mapped to a `slog.LevelVar`; allow dynamic changes without restart.
   - Prefer `stdout`/`stderr` over direct file writes; rely on the runtime (or containers) for collection/rotation.

2) Enforce typed attributes; forbid ad‑hoc key/value pairs.
   - Always log with `slog.Attr` helpers (e.g., `slog.String`, `slog.Int`, `slog.GroupValue`) or `LogAttrs`.
   - Add `sloglint` via `golangci-lint` with `attr-only: true` to prevent `!BADKEY` incidents and keep schema consistent.

3) Define a stable event schema and reserved attribute keys.
   - Message: use event‑style messages like `webauthn_assert_verify` rather than prose sentences.
   - Reserved keys: `time`, `level`, `msg`, plus our domain: `correlation_id`, `trace_id`, `component`, `event`, `error_kind`, `outcome`, `latency_ms`, `rp_id`, `origin`, `uv`, `up`, `sign_count`, `credential_id_hash`, `challenge_id`.
   - Keep keys lower_snake_case; avoid duplicates. If dedup is needed, wrap with a de‑dup handler (e.g., `slog-dedup`).

4) Context propagation and correlation.
   - Continue to issue/accept `X-Request-ID` and place it in `context.Context` (existing middleware).
   - Prefer `Logger.InfoContext(ctx, ...)` where available, and include `correlation_id` as an attribute.
   - Optionally in Phase B, integrate a context‑aware handler (e.g., `slog-context`) to automatically read `correlation_id` from `ctx` to reduce boilerplate.

5) Privacy and data minimization.
   - Do not log secrets or raw binary materials (credential IDs, signatures, CBOR blobs). Use hashed identifiers via `webauthn.HashID`.
   - For domain structs, implement `slog.LogValuer` to control safe projections (e.g., only `id`), preventing accidental PII leakage.

6) Error logging and mapping.
   - Use existing `MapVerifyError` to emit a stable `error_kind` and consistent HTTP mapping; log a single structured line per failure with `LogAssertion`.
   - Stack traces: only include in debug builds or behind an env feature flag, using a `ReplaceAttr` function or a dedicated error type; omit in production by default.

7) Performance and volume control.
   - Guard expensive debug logging with `logger.Enabled(ctx, slog.LevelDebug)` before computing payloads.
   - Avoid `AddSource` in production; allow in dev for convenience.
   - For chatty paths, use a sampling handler (e.g., `samber/slog-sampling`) composed around the JSON handler. Consider fan‑out/multi handlers only when required.

8) Configuration surface.
   - `LOG_LEVEL` (debug|info|warn|error) → initial level, applied to a `slog.LevelVar` for live tuning.
   - `LOG_FORMAT` (json|text) → default `json` for prod; `text` permitted for local dev (optionally with colorized `tint`).

9) Migration plan.
   - Phase A: Introduce default JSON logger and migrate `log.Printf` callsites in HTTP handlers and `webauthn` to structured `slog` with the reserved schema.
   - Phase B: Add `sloglint` to CI; fix violations and require typed attributes.
   - Phase C: Introduce context‑aware correlation (optional) and sampling/dedup where needed.
   - Phase D: Consider OpenTelemetry bridge (`otelslog`) for log/trace correlation in future.

Consequences:
- Pros
  - Consistent, machine‑parsable logs; easier incident triage and analytics.
  - Safer logging by construction via typed attributes and `LogValuer` for domain types.
  - Tunable verbosity at runtime; reduced prod overhead (no source, optional sampling).
- Cons
  - Slightly higher per‑log overhead vs zap/zerolog; acceptable for our throughput. Can mitigate with sampling and careful attribute use.
  - Additional lint and handler dependencies; minor maintenance burden.

Alternatives:
- Keep `log.Printf`: rejected due to lack of structure and correlation.
- Use zap/zerolog directly: highest performance, but fragments API and sacrifices stdlib ergonomics and ecosystem convergence around `slog`.
- Put logger in context: not recommended; prefer context‑aware handlers and explicit dependency injection or `slog.SetDefault`.

Evaluation (Step 40 fit and comparisons):
- Fit for Step 40: Good. `slog` matches our event-style schema, integrates with existing `VerifyLog/MapVerifyError`, keeps deps minimal, and allows runtime tuning via `LevelVar`. Ecosystem handlers (sampling, dedup, tint) cover dev/prod needs.
- Zap (uber-go/zap): Very fast and mature with strong typing. External dep and non-stdlib API. Consider only if profiling shows log hot paths dominate latency/CPU.
- Zerolog: Fastest and low-alloc with builder-style API. Non-stdlib and opinionated ergonomics. Best for ultra-high-volume services.
- Legacy (logrus/go-kit/apex): Slower, older; not recommended for new adoption.
- Hybrids: Keep `slog` API, swap handlers (e.g., `slog-json`, sampling, fanout) or bridge to OpenTelemetry (`otelslog`) for trace correlation without changing call sites.

Trade-offs and net:
- `slog` is slower than zap/zerolog but “good enough” for this demo-grade scope; privacy controls and stdlib stability are priority. If later profiling shows material overhead, add sampling or reconsider zap/zerolog for the hottest paths.

References:
- Dash0: Logging in Go with Slog: A Practitioner’s Guide — https://www.dash0.com/guides/logging-in-go-with-slog
- HN discussion: https://news.ycombinator.com/item?id=45167739
- Sampling handler: https://github.com/samber/slog-sampling
- Multi/fan‑out: https://github.com/samber/slog-multi
- Context handler: https://github.com/veqryn/slog-context
- JSON v2 handler: https://github.com/veqryn/slog-json
- Dedup handler: https://github.com/veqryn/slog-dedup
- OpenTelemetry bridge: https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog

Refs: requirement R-ERR; decision http-error-envelope; decision request-id-and-slog-json; spec base64url-usage-explainer
