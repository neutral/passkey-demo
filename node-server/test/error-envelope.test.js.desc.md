# Purpose
Validate the error helper module: direct responders emit the standardized envelope with correlation ids and the Express error handler maps JSON parse failures to `ERR_BAD_REQUEST`.

# Key Logic
- Spins up lightweight Express apps; one calls `respondBadRequest`, the other triggers a body-parser syntax error and relies on `buildExpressErrorHandler()`.
- Asserts HTTP statuses, `code` fields, and `correlation_id` propagation.

# Interactions
- Exercises `src/error.js` in isolation without involving WebAuthn routes.

# Refs
Refs: requirement R-ERR; decision http-error-envelope
