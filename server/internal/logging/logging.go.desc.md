# Purpose
Initialize and provide the process-wide structured logger using `log/slog`. Exposes helpers to parse `LOG_LEVEL`, construct a JSON or text handler, wrap with a context-aware handler, and return a default logger. This module centralizes logging configuration and keeps runtime selection (level, format, AddSource) in one place.

# Key Logic
- `GetLevelFromEnv() slog.Level`: maps `LOG_LEVEL` env to debug|info|warn|error (default info).
- `NewWithLevelVar(lv, format, addSource) *slog.Logger`: builds a `slog.Logger` with a `slog.LevelVar` for dynamic runtime updates; chooses JSON (default) or text handler and optionally adds source locations; wraps with `slog-context` so attributes in context (e.g., `correlation_id`) are injected when using `InfoContext`.
- `FromContext(ctx)`: returns the default logger (context injection handled by the handler itself).

# Interactions
- Called from `server/cmd/api/main.go` to set `slog.SetDefault(...)` using `LOG_LEVEL` and `LOG_FORMAT`.
- Handlers emit event-style logs; prefer `InfoContext(ctx, ...)` so context attributes (like `correlation_id`) are injected by the handler.

# Refs
Refs: requirement R-ERR; decision request-id-and-slog-json; decision http-error-envelope; decision structured-logging-with-slog-guidelines
