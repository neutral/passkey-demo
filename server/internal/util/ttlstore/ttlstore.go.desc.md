# Purpose
Generic, concurrency-safe TTL store with optional capacity bound and background GC. Used to back in-memory flow session stores without duplicating map+mutex logic.

# Key Logic
- `New[K,V](capacity, ttl, gc)` constructs a store and starts periodic GC if `gc=true`.
- `Put/Get/Delete` manage entries; expired entries are treated as missing.
- Capacity bound enforced on `Put` (after opportunistic GC); returns `ErrCapacity` when full.

# Interactions
- Wrapped by typed stores for registration, login, and tx sessions.

# Refs
Refs: goal simple-ui-and-storage; decision session-store-refactor

