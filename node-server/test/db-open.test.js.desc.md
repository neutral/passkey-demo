# Purpose
Exercise `openDB` to confirm in-memory paths remain transient and filesystem paths still auto-create parent directories.

# Key Logic
- Verifies `:memory:` databases do not create on-disk files and each connection is isolated.
- Ensures file-backed paths generate parent directories and accept schema creation.

# Interactions
- Calls `openDB` directly; no HTTP server spin-up. Uses temporary directories via `fs`/`os`.

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV; decision encoding-and-ceremony-guardrails
