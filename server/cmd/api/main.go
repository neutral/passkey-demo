package main

import (
    "log"
    "net/http"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
)

func main() {
    cfg, err := cfgpkg.Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }
    mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// TODO: mount /authn/passkey/*, /tx/*, /transaction/list

    log.Printf("rp_id=%s origin=%s port=%s db=%s", cfg.RP_ID, cfg.Origin, cfg.Port, cfg.DBPath)
    log.Printf("server listening on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
