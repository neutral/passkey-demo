# Purpose
Convenience runner that executes the Go test suite package-by-package with clear progress output, and supports a fast “short” mode.

# Key Logic
- Lists server packages via `go list ./...` and runs `go test` per package.
- Short mode adds `-short` so tests that honor `testing.Short()` can skip heavy paths (e.g., E2E).
- Echoes progress as `[#/N] <package>` for visibility.

# Interactions
- Relies on the Go toolchain and the repository’s `server/` module layout.

# Refs
Refs: goal developer-velocity; decision encoding-and-ceremony-guardrails; requirement R-FLOW-SIGN

