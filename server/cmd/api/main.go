package main

import (
    "log"
    "net/http"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    httpx "github.com/neutral/passkey-demo/internal/http"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
    tx "github.com/neutral/passkey-demo/internal/tx"
    me "github.com/neutral/passkey-demo/internal/me"
)

func main() {
    cfg, err := cfgpkg.Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }
    mux := http.NewServeMux()
    // DB
    db, err := storepkg.Open(cfg)
    if err != nil { log.Fatalf("db open: %v", err) }
    if err := storepkg.Migrate(db); err != nil { log.Fatalf("db migrate: %v", err) }
    // In-memory stores
    regStore := webauthn.NewRegSessionStore(10000)
    loginStore := webauthn.NewLoginSessionStore(10000)
    txStore := tx.NewTxSessionStore(10000)

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

    // Registration options
    mux.Handle("/authn/passkey/registration/options", webauthn.RegistrationOptionsHandler(cfg, regStore))
    // Registration finish
    mux.Handle("/authn/passkey/registration/finish", webauthn.RegistrationFinishHandler(cfg, regStore, db))
    // Login options
    mux.Handle("/authn/passkey/login/options", webauthn.LoginOptionsHandler(cfg, loginStore))
    // Login finish
    mux.Handle("/authn/passkey/login/finish", webauthn.LoginFinishHandler(cfg, loginStore, db))
    // Tx signing options/finish
    mux.Handle("/tx/signing/options", tx.TxOptionsHandler(cfg, txStore, db))
    mux.Handle("/tx/signing/finish", tx.TxFinishHandler(cfg, txStore, db))
    // Transactions list (authenticated)
    mux.Handle("/tx/list", tx.TxListHandler(db))
    // Account key (authenticated)
    mux.Handle("/me/account_key", me.AccountKeyHandler(db))

    // Wrap with outermost CORS middleware
    // Layer: CORS (outermost) → Session middleware → mux
    handler := httpx.CORSMiddleware(cfg)(httpx.SessionMiddleware(db, true, time.Hour)(mux))

    log.Printf("rp_id=%s origin=%s port=%s db=%s", cfg.RP_ID, cfg.Origin, cfg.Port, cfg.DBPath)
    log.Printf("server listening on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
