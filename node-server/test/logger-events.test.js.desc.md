# Purpose
Verify that the structured logging helpers emit the expected payloads, including correlation id, event names, and hashed identifiers.

# Key Logic
- Temporarily stub `logger.info`/`logger.error` to capture arguments when invoking helper functions.
- Assert success helpers add correlation id, rp/origin, and contextual fields; error helpers attach reasons/messages and propagate error objects.
- Confirm `logWebauthnVerifyFailure` emits the `webauthn_assert_verify` event with hashed identifiers.

# Interactions
- Imports helper exports and the shared `logger` instance from `src/logger.js`.

# Refs
Refs: requirement R-ERR; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines
