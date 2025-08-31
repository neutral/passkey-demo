# Purpose
Centralized runtime configuration for the server. Reads environment variables, applies defaults, normalizes, validates, and exposes allowlists for RP IDs and Origins used by verification.

# Key Logic
- Env vars: `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`, `RP_ID_ALLOWLIST`, `ORIGIN_ALLOWLIST`.
- Defaults: `RP_ID=localhost`, `ORIGIN=http://localhost:5173`, `PORT=8080`, `DB_PATH=server/demo.db`.
- Normalizes values (lowercase RP, trim origin/trailing slash), ensures primary values are present in allowlists, and validates formats (URL scheme/host, numeric port, no wildcards).
- API: `Config` struct and `Load()` constructor.

# Interactions
- Called by `cmd/api/main.go` at startup to obtain `cfg.Port` and to log effective settings.
- Future steps use `RPAllowlist`/`OriginAllowlist` in security checks.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-PORTABLE; requirement R-SEC-UV
