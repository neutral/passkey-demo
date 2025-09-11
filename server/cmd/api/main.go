package main

import (
    "log"
    "net/http"
    "context"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
    app "github.com/neutral/passkey-demo/internal/app"
    repos "github.com/neutral/passkey-demo/internal/repos"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
    tx "github.com/neutral/passkey-demo/internal/tx"
)

func main() {
    cfg, err := cfgpkg.Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }
    // DB
    db, err := storepkg.Open(cfg)
    if err != nil { log.Fatalf("db open: %v", err) }
    if err := storepkg.Migrate(db); err != nil { log.Fatalf("db migrate: %v", err) }
    // In-memory stores
    regStore := webauthn.NewRegSessionStore(10000)
    loginStore := webauthn.NewLoginSessionStore(10000)
    txStore := tx.NewTxSessionStore(10000)

    // Prepare read-only repositories (prepared statements)
    credsRepo, err := repos.NewCredentials(context.Background(), db)
    if err != nil { log.Fatalf("repos credentials: %v", err) }
    txRepo, err := repos.NewTransactions(context.Background(), db)
    if err != nil { log.Fatalf("repos transactions: %v", err) }
    deps := app.Deps{RegStore: regStore, LoginStore: loginStore, TxStore: txStore, CredsRepo: credsRepo, TxRepo: txRepo}

    // Build router with group middlewares and global wrappers
    handler := app.BuildRouter(cfg, db, deps)

    log.Printf("rp_id=%s origin=%s port=%s db=%s", cfg.RP_ID, cfg.Origin, cfg.Port, cfg.DBPath)
    log.Printf("server listening on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
