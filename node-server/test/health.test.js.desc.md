# Purpose
Verify the app boots and responds on `/health` with 200 and `{ status: 'ok' }` using an ephemeral port, without external test libraries.

# Key Logic
- Creates app via `createApp`, listens on port 0, uses Node fetch to query `/health`, asserts JSON payload, and closes the server.

# Interactions
- Imports `createApp`, `loadConfig`, and `openDB` from the Node server scaffold.

# Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-3

