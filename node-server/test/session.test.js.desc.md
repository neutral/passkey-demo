# Purpose
Validate session middleware: reads `sid` cookie, loads from DB, verifies expiry, and attaches `req.session`.

# Key Logic
- Inserts rows into `sessions` with future/past expiry; uses a test route `/whoami` to echo `hasSession` and `sid`.
- Asserts that only a valid, unexpired cookie results in a present session context.

# Refs
Refs: spec session-cookies-usage-explainer; requirement R-OPS-DEV

