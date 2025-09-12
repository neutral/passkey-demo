# Purpose
Initialize and provide the process-wide structured logger using `log/slog`. Exposes helpers to parse `LOG_LEVEL`, construct a JSON or text handler, and return a default logger. This module centralizes logging configuration and keeps runtime selection (level, format, AddSource) in one place.

# Key Logic
- `GetLevelFromEnv() slog.Level`: maps `LOG_LEVEL` env to debug|info|warn|error (default info).
- `NewWithLevelVar(lv, format, addSource) *slog.Logger`: builds a `slog.Logger` with a `slog.LevelVar` for dynamic runtime updates; chooses JSON (default) or text handler and optionally adds source locations.
- `FromContext(ctx)`: returns the default logger; placeholder for future context-aware handler (e.g., to pull `correlation_id`).

# Interactions
- Called from `server/cmd/api/main.go` to set `slog.SetDefault(...)` using `LOG_LEVEL` and `LOG_FORMAT`.
- Handlers call `slog.Default()` (or `logging.FromContext`) and emit event-style logs.

# Refs
Refs: requirement R-ERR; decision request-id-and-slog-json; decision http-error-envelope; decision structured-logging-with-slog-guidelines

