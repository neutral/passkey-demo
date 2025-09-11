package app

import (
    "database/sql"
    "net/http"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    httpx "github.com/neutral/passkey-demo/internal/http"
    httpmid "github.com/neutral/passkey-demo/internal/httpx/middleware"
    me "github.com/neutral/passkey-demo/internal/me"
    tx "github.com/neutral/passkey-demo/internal/tx"
    repos "github.com/neutral/passkey-demo/internal/repos"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
)

// Deps defines runtime dependencies constructed by main and injected here.
// This enables central route wiring without hard-coding constructors.
type Deps struct {
    RegStore   *webauthn.RegSessionStore
    LoginStore *webauthn.LoginSessionStore
    TxStore    *tx.TxSessionStore
    CredsRepo  *repos.CredentialsRepo
    TxRepo     *repos.TransactionsRepo
}

// BuildRouter assembles the HTTP mux with grouped middlewares.
// Ordering: CORS (outermost) → Session middleware → group middlewares → handlers.
// Group limits: body size and rate limits applied to /authn/* and /tx/*.
func BuildRouter(cfg *cfgpkg.Config, db *sql.DB, deps Deps) http.Handler {
    // Root mux
    root := http.NewServeMux()

    // Health endpoint (no group limits)
    root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    // Build group wrappers
    // Choose conservative but permissive limits to avoid breaking legitimate requests.
    // Authentication endpoints carry WebAuthn payloads; keep limits ~1MiB.
    authnBodyLimit := int64(1 << 20) // 1 MiB
    txBodyLimit := int64(1 << 20)    // 1 MiB
    authnRL := httpx.NewRateLimiter(20, 10) // burst 20, ~10 rps per key
    txRL := httpx.NewRateLimiter(20, 10)

    // Helper to wrap a handler with group middlewares
    wrapGroup := func(limit int64, rl *httpx.RateLimiter, h http.Handler) http.Handler {
        return httpx.RateLimitMiddleware(rl, httpx.RemoteIP)(
            httpx.BodyLimitMiddleware(limit)(h),
        )
    }

    // /authn/* sub-mux
    authnMux := http.NewServeMux()
    authnMux.Handle("/authn/passkey/registration/options", webauthn.RegistrationOptionsHandler(cfg, deps.RegStore))
    authnMux.Handle("/authn/passkey/registration/finish", webauthn.RegistrationFinishHandler(cfg, deps.RegStore, db))
    authnMux.Handle("/authn/passkey/login/options", webauthn.LoginOptionsHandler(cfg, deps.LoginStore))
    authnMux.Handle("/authn/passkey/login/finish", webauthn.LoginFinishHandler(cfg, deps.LoginStore, db))
    root.Handle("/authn/", wrapGroup(authnBodyLimit, authnRL, authnMux))

    // /tx/* sub-mux
    txMux := http.NewServeMux()
    txMux.Handle("/tx/signing/options", tx.TxOptionsHandler(cfg, deps.TxStore, deps.CredsRepo, db))
    txMux.Handle("/tx/signing/finish", tx.TxFinishHandler(cfg, deps.TxStore, db, deps.CredsRepo))
    txMux.Handle("/tx/list", tx.TxListHandler(deps.TxRepo))
    root.Handle("/tx/", wrapGroup(txBodyLimit, txRL, txMux))

    // /me/* (no extra group limits beyond session)
    root.Handle("/me/account_key", me.AccountKeyHandler(db))

    // Global wrapper: CORS (outermost), then session middleware.
    // Request ID outermost for correlation
    return httpmid.RequestID(
        httpx.CORSMiddleware(cfg)(
            httpx.SessionMiddleware(db, true, time.Hour)(root),
        ),
    )
}
