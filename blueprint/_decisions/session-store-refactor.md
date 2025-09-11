Title: Unify Reg/Login/Tx Stores via Generic TTL Store
Status: Accepted
Date: 2025-09-11

Context:
- Three nearly-identical in-memory stores existed with mutex+map boilerplate and no GC.
- TTL and capacity semantics were scattered and duplicated.

Decision:
1) Introduce generic TTL store `internal/util/ttlstore` with capacity bound and background GC.
2) Wrap with typed stores keeping public APIs: `RegSessionStore`, `LoginSessionStore`, `TxSessionStore`.
3) Replace duplicate `randBytes` with `internal/util/randutil`.

Consequences:
- Less duplication, clearer policies, and tests around expiry/capacity.

Alternatives:
- Keep ad hoc stores: increases maintenance cost and risk of inconsistent behavior.

References:
- Implementation: `server/internal/util/ttlstore/*`, `server/internal/util/randutil/*`, adoption in reg/login/tx options handlers.

Refs: goal simple-ui-and-storage; requirement R-PLAT-2

