# Purpose
List transactions for the authenticated account. Scopes results by the session cookie `sid` and returns `tx_id_hex`, `nonce`, `message`, and `created_at`.

# Key Logic
- Requires session middleware; obtain `acct_cbor` from request context (no cookie/DB fallback).
- If no session, return 401.
- Query `transactions` via prepared repo by `acct_cbor` ordered by `created_at DESC`.
- Hex‑encode `tx_id` (BLOB) and return items in a JSON envelope.

# Interactions
- Used after signing (Step 22) to show persisted transactions.
- Shares the same session scoping model as options/finish (Steps 21–22).

# Refs
Refs: goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations
