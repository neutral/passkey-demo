# Nonce Monotonicity Policy Explainer

## Purpose
Describe the per-account nonce policy used to prevent replay and enforce ordering when signing transactions.

## Overview
- Every account maintains a strictly increasing `nonce` in the Bundle.
- Server looks up `MAX(nonce)` in `transactions` for the account and rejects any bundle with `nonce <= last_nonce`.
- Policy is enforced during options issuance so clients receive early, clear feedback (e.g., HTTP 409 at the handler layer).

## Details
- Lookup: `SELECT MAX(nonce) FROM transactions WHERE acct_cbor = ?`.
- Accept only if `bundle.nonce > last_nonce` (no equals).
- Negative cases: duplicates or regressions are rejected; client should retry with a higher nonce.
- Concurrency: for this demo, options issuance guards monotonicity; no cross-process contention in scope.

## Refs
Refs: goal transaction-content-signing; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails

