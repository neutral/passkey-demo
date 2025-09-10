# Purpose
Helper script to (re)generate the golden vectors JSON using the local codebase, ensuring a single command updates the committed file.

# Key Logic
- Builds `server/cmd/vectors` and runs it with `-fmt json`.
- Writes the output to `specs/goldens/tx-bundle-v1.json`.

# Interactions
- Invokes Go toolchain; writes to the repository under `specs/goldens/`.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

