# Purpose
Provide base64url (URL-safe, no padding) and UTF‑8 conversion helpers for the browser and Playwright (Node) environments without external dependencies.

# Key Logic
- `bytesToBase64url` encodes bytes to base64, maps `+`→`-`, `/`→`_`, and strips trailing `=` padding.
- `base64urlToBytes` validates input, accepts optional padding, restores base64 and decodes to bytes.
- `utf8ToBytes`/`bytesToUtf8` use `TextEncoder`/`TextDecoder('utf-8', { fatal: true })`.
- Environment fallback: prefer `btoa/atob`; fallback to Node `Buffer` when unavailable.

# Interactions
- Consumed by registration/login/signing flows for WebAuthn binary fields and message/CBOR conversions in later steps.
- Tested via Playwright by dynamically importing the module from the Vite dev server.

# Refs
Refs: requirement R-PLAT-1; requirement R-ERR; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b

