# Purpose
List transactions for the authenticated account. Scopes results by the session cookie `sid` and returns `tx_id_hex`, `nonce`, `message`, and `created_at`.

# Key Logic
- Read `sid` cookie; resolve `acct_cbor` and expiry from `sessions`.
- If missing or expired, return 401.
- Query `transactions` by `acct_cbor` ordered by `created_at DESC`.
- Hex‑encode `tx_id` (BLOB) and return items in a JSON envelope.

# Interactions
- Used after signing (Step 22) to show persisted transactions.
- Shares the same session scoping model as options/finish (Steps 21–22).

# Refs
Refs: goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations

