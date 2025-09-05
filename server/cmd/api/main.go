package main

import (
    "log"
    "net/http"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    httpx "github.com/neutral/passkey-demo/internal/http"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
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

    // Wrap with outermost CORS middleware
    handler := httpx.CORSMiddleware(cfg)(mux)

    log.Printf("rp_id=%s origin=%s port=%s db=%s", cfg.RP_ID, cfg.Origin, cfg.Port, cfg.DBPath)
    log.Printf("server listening on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
