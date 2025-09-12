# Purpose
End-to-end test for the transaction signing flow (options → finish → list) using in-memory HTTP handlers and an ephemeral SQLite DB.

# Key Logic
- Creates an account, credential, and session directly in the DB.
- Builds a canonical bundle B and derives challenge via options builder.
- Constructs CDJ and AD, signs the digest with a generated ECDSA key, finishes the tx, and verifies it appears in `/tx/list`.
- Marked as a LONG test via build tag. The file has `//go:build long`, so it is excluded from default `go test ./...` runs and only executes when explicitly enabled with `-tags=long`.

# Interactions
- Uses `httptest` to exercise HTTP handlers without a live server.
- Depends on `internal/storage`, `internal/encoding`, and `internal/tx` handlers.

# Refs
Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails; goal server-derived-challenge-and-txid
