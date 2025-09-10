# Purpose
End-to-end test for the transaction signing flow (options → finish → list) using in-memory HTTP handlers and an ephemeral SQLite DB.

# Key Logic
- Creates an account, credential, and session directly in the DB.
- Builds a canonical bundle B and derives challenge via options builder.
- Constructs CDJ and AD, signs the digest with a generated ECDSA key, finishes the tx, and verifies it appears in `/tx/list`.
- Skips when `-short` is supplied to avoid running E2E in “short” mode.

# Interactions
- Uses `httptest` to exercise HTTP handlers without a live server.
- Depends on `internal/storage`, `internal/encoding`, and `internal/tx` handlers.

# Refs
Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails; goal server-derived-challenge-and-txid

