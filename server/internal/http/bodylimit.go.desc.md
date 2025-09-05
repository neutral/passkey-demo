# Purpose
Limit request body sizes for JSON endpoints to prevent resource exhaustion.

# Key Logic
- Wraps handler with `http.MaxBytesReader` at a configured byte size.
- If the body exceeds the limit while reading, returns 413 Payload Too Large.

# Interactions
- Applied to `/authn/*` and `/tx/*` endpoints (router wiring TBD).

# Refs
Refs: requirement R-ERR; requirement R-PLAT-2

