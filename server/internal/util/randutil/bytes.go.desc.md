# Purpose
Tiny helper for generating cryptographically secure random bytes, shared across packages to remove duplicated `randBytes` helpers.

# Key Logic
- `Bytes(n int) []byte` panics on CSPRNG failure.
- `BytesE(n int) ([]byte, error)` returns an error instead of panicking.

# Interactions
- Used by WebAuthn registration/login options and tx options to create session IDs and challenges.

# Refs
Refs: goal simple-ui-and-storage; decision session-store-refactor

