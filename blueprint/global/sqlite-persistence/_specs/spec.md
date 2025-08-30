# R-PLAT-3 — SQLite Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Embed SQLite for persistence; store binary fields as BLOB; expose base64url at API layer.

## Interfaces
- DB access via Go `database/sql`; migrations/CREATE IF NOT EXISTS on startup.

## Data / Models
- Tables: accounts, credentials, transactions as per requirement SQL.
- Sessions: optional; may be in-memory for the demo or persisted via the `sessions` table.

## Algorithms
- Simple CRUD and FK enforcement; cascade delete accounts→credentials/transactions; monotonic nonce checks at options stage; sign_count updates at login/signing.

## Security / Privacy
- Treat raw BLOBs as sensitive; avoid dumping in logs; use hex/thumb only for display.

## Errors / Observability
- Map driver errors to HTTP status; expose minimal messages.

## Testing Strategy
- Initialize fresh DB; run flows; verify rows and constraints; check cascade behavior.

## Open Questions
- None for demo.

Refs: decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; requirement R-PLAT-3
