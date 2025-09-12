package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "strings"

    app "github.com/neutral/passkey-demo/internal/app"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    repos "github.com/neutral/passkey-demo/internal/repos"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
    tx "github.com/neutral/passkey-demo/internal/tx"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
    logx "github.com/neutral/passkey-demo/internal/logging"
)

func main() {
    // Initialize slog default logger from env
    var lv slog.LevelVar
    lv.Set(logx.GetLevelFromEnv())
    // Add source only in dev when LOG_FORMAT=text or when explicitly requested via DEV_ADD_SOURCE=1
    addSource := strings.ToLower(os.Getenv("DEV_ADD_SOURCE")) == "1"
    logger := logx.NewWithLevelVar(&lv, os.Getenv("LOG_FORMAT"), addSource)
    slog.SetDefault(logger)

    cfg, err := cfgpkg.Load()
    if err != nil {
        slog.Error("config_error", slog.Any("error", err))
        os.Exit(1)
    }
    // DB
    db, err := storepkg.Open(cfg)
    if err != nil {
        slog.Error("db_open_error", slog.Any("error", err))
        os.Exit(1)
    }
    if err := storepkg.Migrate(db); err != nil {
        slog.Error("db_migrate_error", slog.Any("error", err))
        os.Exit(1)
    }
    // In-memory stores
    regStore := webauthn.NewRegSessionStore(10000)
    loginStore := webauthn.NewLoginSessionStore(10000)
    txStore := tx.NewTxSessionStore(10000)

    // Prepare read-only repositories (prepared statements)
    credsRepo, err := repos.NewCredentials(context.Background(), db)
    if err != nil {
        slog.Error("repos_credentials_error", slog.Any("error", err))
        os.Exit(1)
    }
    txRepo, err := repos.NewTransactions(context.Background(), db)
    if err != nil {
        slog.Error("repos_transactions_error", slog.Any("error", err))
        os.Exit(1)
    }
    deps := app.Deps{RegStore: regStore, LoginStore: loginStore, TxStore: txStore, CredsRepo: credsRepo, TxRepo: txRepo}

    // Build router with group middlewares and global wrappers
    handler := app.BuildRouter(cfg, db, deps)

    slog.Info("server_start",
        slog.String("rp_id", cfg.RP_ID),
        slog.String("origin", cfg.Origin),
        slog.String("port", cfg.Port),
        slog.String("db_path", cfg.DBPath),
    )
    if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
        slog.Error("server_error", slog.Any("error", err))
        os.Exit(1)
    }
}
