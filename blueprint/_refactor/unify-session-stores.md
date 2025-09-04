# Unify In‑Memory Session Stores (Reg/Login/Tx)

## Purpose
- Reduce duplication across `RegSessionStore`, `LoginSessionStore`, and `TxSessionStore` by introducing a single, generic, TTL‑aware in‑memory store.
- Centralize lifecycle policies (capacity, expiry, GC) and error semantics to simplify handlers and improve reliability.

## Context
- Three separate stores exist today with near‑identical mutex + map patterns and capacity checks.
- TTL handling is manual at call sites; no background GC for expired items.
- `randBytes`/ID generation and TTL values are scattered in multiple files.

## Proposal
- Create `server/internal/session` (or `internal/store/session`) with a reusable store:
  - API:
    - `type Store[T any] struct { Put(id string, v T) error; Get(id string) (T, bool); Delete(id string); Size() int }`
    - Constructor: `New[T](capacity int, ttl time.Duration, gcInterval time.Duration)`
    - Internal: `map[string]entry[T]` with `expiresAt`; single `sync.Mutex`.
    - Background GC goroutine removing expired entries at `gcInterval` (stoppable on context cancel for clean shutdown).
  - Typed wrappers for clarity (optional): `RegSessions = Store[RegSession]`, `LoginSessions = Store[LoginSession]`, `TxSessions = Store[TxSession]`.
- Consolidate ID generation:
  - Move `randBytes` to `internal/crypto/randutil` or `internal/util/randutil` with a single exported `Bytes(n int) ([]byte, error)`.
- Centralize TTL constants:
  - `const RegTTL = 5*time.Minute`, `LoginTTL = 5*time.Minute`, `TxTTL = 5*time.Minute` in a shared `internal/session/policy.go`.

## Migration Plan (Incremental)
1) Introduce the generic store alongside existing stores (no behavior change).
2) Migrate `RegSessionStore` internals to the generic store; keep the public API stable.
3) Repeat for `LoginSessionStore` and `TxSessionStore`.
4) Remove legacy store implementations once all call sites use the generic.
5) Replace duplicate `randBytes` with `randutil.Bytes` across reg/login/tx options.

## Risks & Mitigations
- GC overhead: keep `gcInterval` coarse (e.g., 1–5 minutes) and bound capacity; document trade‑offs for the demo.
- API churn: shim types preserve current call sites to avoid broad refactors.
- Concurrency correctness: retain the proven `sync.Mutex` pattern; add tests for races and expiry.

## Testing Strategy
- Unit tests for the generic store:
  - Put/Get/Delete semantics; capacity limit; expiry behavior; GC removes expired entries.
  - Concurrency sanity via `-race` in CI/local.
- Regression tests for reg/login/tx handlers remain green after switching to generic store.

## Acceptance Criteria
- Single generic session store used by registration, login, and tx flows.
- No behavior regressions; TTL and capacity policies unchanged (5 minutes, same capacity if set).
- All tests pass; no duplicated `randBytes`/store code remains.

## Notes
- This is a future improvement; current implementation intentionally sticks to minimal, localized changes per step.

## Refs
- Refs: goal server-derived-challenge-and-txid; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails; requirement R-PLAT-2

