# Overview
Shared frontend utilities used across views and flows. This folder holds small, dependency-free helpers for encoding and (later) CBOR helpers used by WebAuthn and signing flows.

# Relations
- Imported by UI code and tests; used to convert between text and binary, and to encode/decode base64url for JSON payloads.
- Will be used by registration/login/signing steps to transform WebAuthn fields and CBOR bundles.

# Interfaces & Models
- Encoding helpers: base64url and UTF‑8 functions.
- Future: canonical CBOR encoding utilities for bundle construction.

# Refs
Refs: requirement R-PLAT-1; requirement R-ERR; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b

