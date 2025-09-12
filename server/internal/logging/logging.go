package logging

import (
    "context"
    "log/slog"
    "os"
    "strings"
    slogctx "github.com/veqryn/slog-context"
)

// GetLevelFromEnv parses LOG_LEVEL into a slog.Level.
// Accepted values: debug|info|warn|error (case-insensitive). Defaults to info.
func GetLevelFromEnv() slog.Level {
    switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
    case "debug":
        return slog.LevelDebug
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}

// NewWithLevelVar constructs a logger using the provided level var and format.
// format: "json" (default) or "text". When addSource is true, source locations are added.
func NewWithLevelVar(lv *slog.LevelVar, format string, addSource bool) *slog.Logger {
    if lv == nil {
        var tmp slog.LevelVar
        tmp.Set(slog.LevelInfo)
        lv = &tmp
    }
    opts := &slog.HandlerOptions{Level: lv, AddSource: addSource}
    var h slog.Handler
    switch strings.ToLower(format) {
    case "text":
        h = slog.NewTextHandler(os.Stdout, opts)
    default:
        h = slog.NewJSONHandler(os.Stdout, opts)
    }
    // Wrap with context-aware handler so attributes in context are included
    ch := slogctx.NewHandler(h, nil)
    return slog.New(ch)
}

// FromContext returns the default logger for now.
// Placeholder to adopt a context-aware handler in the future.
func FromContext(ctx context.Context) *slog.Logger { return slog.Default() }
