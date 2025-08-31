package storage

import (
    "database/sql"
    "path/filepath"
    "testing"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
)

func TestOpenAndMigrateCreatesDBAndTables(t *testing.T) {
    t.Setenv("RP_ID", "")
    t.Setenv("ORIGIN", "")
    t.Setenv("PORT", "")
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")
    t.Setenv("DB_PATH", dbPath)
    cfg, err := cfgpkg.Load()
    if err != nil {
        t.Fatalf("load cfg: %v", err)
    }
    db, err := Open(cfg)
    if err != nil {
        t.Fatalf("open: %v", err)
    }
    defer db.Close()
    if err := Migrate(db); err != nil {
        t.Fatalf("migrate: %v", err)
    }
    // verify PRAGMAs
    var jm string
    if err := db.QueryRow("PRAGMA journal_mode").Scan(&jm); err != nil {
        t.Fatalf("pragma journal_mode: %v", err)
    }
    if jm != "wal" {
        t.Fatalf("expected journal_mode=wal, got %s", jm)
    }
    var fkon int
    if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkon); err != nil {
        t.Fatalf("pragma foreign_keys: %v", err)
    }
    if fkon != 1 {
        t.Fatalf("expected foreign_keys=1, got %d", fkon)
    }
    var sync int
    if err := db.QueryRow("PRAGMA synchronous").Scan(&sync); err != nil {
        t.Fatalf("pragma synchronous: %v", err)
    }
    if sync != 1 { // NORMAL
        t.Fatalf("expected synchronous=NORMAL(1), got %d", sync)
    }
    // verify tables exist
    mustHave := []string{"accounts", "credentials", "sessions", "transactions"}
    for _, tbl := range mustHave {
        if !tableExists(t, db, tbl) {
            t.Fatalf("expected table %s to exist", tbl)
        }
    }
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
    t.Helper()
    var n int
    if err := db.QueryRow("SELECT count(1) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&n); err != nil {
        t.Fatalf("query sqlite_master: %v", err)
    }
    return n == 1
}
