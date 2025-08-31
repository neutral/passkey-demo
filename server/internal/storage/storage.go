package storage

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "time"

    _ "github.com/mattn/go-sqlite3"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
)

// Open opens (and creates if needed) the SQLite database at cfg.DBPath,
// applies recommended PRAGMA settings for a demo environment, and returns the DB handle.
func Open(cfg *cfgpkg.Config) (*sql.DB, error) {
    // Ensure directory exists for DB path (if any)
    if dir := filepath.Dir(cfg.DBPath); dir != "." && dir != "" {
        if err := os.MkdirAll(dir, 0o755); err != nil {
            return nil, fmt.Errorf("mkdir %s: %w", dir, err)
        }
    }
    dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_foreign_keys=on", cfg.DBPath)
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, err
    }
    // Conservative connection pool settings for SQLite.
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)
    db.SetConnMaxIdleTime(2 * time.Minute)

    // Apply PRAGMAs not covered by DSN flags.
    pragmas := []string{
        "PRAGMA journal_mode=WAL",
        "PRAGMA synchronous=NORMAL",
        // foreign_keys covered by DSN but set again explicitly in session
        "PRAGMA foreign_keys=ON",
    }
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    for _, p := range pragmas {
        if _, err := db.ExecContext(ctx, p); err != nil {
            _ = db.Close()
            return nil, fmt.Errorf("apply pragma %q: %w", p, err)
        }
    }
    return db, nil
}

// Migrate creates the schema tables if they do not exist.
func Migrate(db *sql.DB) error {
    stmts := []string{
        // accounts table
        `CREATE TABLE IF NOT EXISTS accounts (
            acct_cbor    BLOB PRIMARY KEY,
            acct_thumb   BLOB NOT NULL,
            created_at   INTEGER NOT NULL
        )`,
        // credentials table
        `CREATE TABLE IF NOT EXISTS credentials (
            credential_id BLOB PRIMARY KEY,
            acct_cbor_fk  BLOB NOT NULL,
            sign_count    INTEGER NOT NULL,
            aaguid        BLOB,
            created_at    INTEGER NOT NULL,
            FOREIGN KEY(acct_cbor_fk) REFERENCES accounts(acct_cbor) ON DELETE CASCADE
        )`,
        // sessions table
        `CREATE TABLE IF NOT EXISTS sessions (
            session_id  TEXT PRIMARY KEY,
            acct_cbor   BLOB NOT NULL,
            expires_at  INTEGER NOT NULL,
            created_at  INTEGER NOT NULL,
            FOREIGN KEY(acct_cbor) REFERENCES accounts(acct_cbor) ON DELETE CASCADE
        )`,
        // transactions table
        `CREATE TABLE IF NOT EXISTS transactions (
            tx_id       BLOB PRIMARY KEY,
            acct_cbor   BLOB NOT NULL,
            nonce       INTEGER NOT NULL,
            message     TEXT NOT NULL,
            bundle_cbor BLOB NOT NULL,
            auth_data   BLOB NOT NULL,
            client_data BLOB NOT NULL,
            signature   BLOB NOT NULL,
            created_at  INTEGER NOT NULL,
            FOREIGN KEY(acct_cbor) REFERENCES accounts(acct_cbor) ON DELETE CASCADE
        )`,
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    for _, s := range stmts {
        if _, err := db.ExecContext(ctx, s); err != nil {
            return err
        }
    }
    return nil
}
