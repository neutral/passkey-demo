package config

import (
    "testing"
)

func TestLoad_DefaultsAndNormalize(t *testing.T) {
    t.Setenv("RP_ID", "")
    t.Setenv("ORIGIN", "")
    t.Setenv("PORT", "")
    t.Setenv("DB_PATH", "")
    t.Setenv("RP_ID_ALLOWLIST", "")
    t.Setenv("ORIGIN_ALLOWLIST", "")
    cfg, err := Load()
    if err != nil {
        t.Fatalf("Load defaults: %v", err)
    }
    if cfg.RP_ID != "localhost" || cfg.Origin != "http://localhost:5173" || cfg.Port != "8080" {
        t.Fatalf("unexpected defaults: %+v", cfg)
    }
    // allowlists should include primary values
    if !contains(cfg.RPAllowlist, cfg.RP_ID) || !contains(cfg.OriginAllowlist, cfg.Origin) {
        t.Fatalf("allowlists missing primary values: %+v", cfg)
    }
}

func TestLoad_InvalidOrigin(t *testing.T) {
    t.Setenv("ORIGIN", "bad")
    // clear others to use defaults
    for _, k := range []string{"RP_ID", "PORT", "DB_PATH", "RP_ID_ALLOWLIST", "ORIGIN_ALLOWLIST"} {
        t.Setenv(k, "")
    }
    if _, err := Load(); err == nil {
        t.Fatalf("expected error for invalid origin")
    }
}

func TestLoad_NormalizesAndRejectsWildcards(t *testing.T) {
    // Lowercases RP_ID and trims trailing slash on Origin; rejects wildcards in allowlists
    t.Setenv("RP_ID", "LocalHost")
    t.Setenv("ORIGIN", "http://localhost:5173/")
    t.Setenv("PORT", "8080")
    t.Setenv("DB_PATH", "server/demo.db")
    // valid allowlists
    t.Setenv("RP_ID_ALLOWLIST", "example.com, localhost")
    t.Setenv("ORIGIN_ALLOWLIST", "http://localhost:5173")
    cfg, err := Load()
    if err != nil {
        t.Fatalf("Load normalize: %v", err)
    }
    if cfg.RP_ID != "localhost" { // lowercased
        t.Fatalf("expected rp_id lowercase, got %s", cfg.RP_ID)
    }
    if cfg.Origin != "http://localhost:5173" { // no trailing slash
        t.Fatalf("expected origin trimmed, got %s", cfg.Origin)
    }
    // wildcards should be rejected
    t.Setenv("RP_ID_ALLOWLIST", "*")
    if _, err := Load(); err == nil {
        t.Fatalf("expected error for wildcard in RP_ID_ALLOWLIST")
    }
    t.Setenv("RP_ID_ALLOWLIST", "localhost")
    t.Setenv("ORIGIN_ALLOWLIST", "http://*.example.com")
    if _, err := Load(); err == nil {
        t.Fatalf("expected error for wildcard in ORIGIN_ALLOWLIST")
    }
}
