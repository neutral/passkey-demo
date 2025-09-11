# Purpose
Unit test for the JSON error envelope writer. Verifies status code and JSON shape with `code` and `error` fields.

# Key Logic
- Calls `Write` with 401/`ERR_UNAUTHORIZED` and checks decoded `Envelope`.

# Interactions
- Exercises `internal/httpx/errors` only; no external I/O.

# Refs
Refs: requirement R-ERR; decision http-error-envelope

