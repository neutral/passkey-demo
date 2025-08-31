# R-PLAT-2 — Base64/Base64url Usage Explainer (Non‑Developer)

## Purpose & Audience
- Explain why and how the backend uses base64/base64url encodings for binary data in JSON and URLs, in practical terms for technical collaborators who aren’t writing code.

## What Base64/Base64url Is
- A way to represent binary data (bytes) as readable text using a small alphabet.
- “base64url” is the URL‑safe variant: it replaces `+` with `-` and `/` with `_` so values can be used in URLs and JSON without escaping.

## Why We Use It Here
- JSON does not natively carry binary data; base64url lets us send keys, signatures, and CBOR bytes as strings.
- URL‑safe: tokens can travel in query strings or paths without breaking on reserved characters.
- Interoperability: browsers and servers across platforms handle base64url consistently.

## How We Use It in This App
- Binary fields (credential IDs, keys, signatures, CBOR bundles) are exchanged as base64url strings in JSON requests/responses.
- Policy (guardrails):
  - Encode without padding (`=`) for shorter, cleaner strings.
  - Decode tolerantly: accept inputs with or without padding when reading.
- Internally, the server decodes base64url strings back to bytes for verification and storage.

## Examples (Conceptual)
- A credential ID (binary) → `A1bC...` base64url string in JSON; server decodes it before lookup.
- A signed bundle (CBOR bytes) → base64url string in the request; server decodes it before hashing/verification.

## Performance & Safety Settings
- Encoding/decoding is fast and part of the standard library.
- Tolerant decode reduces integration friction between different client stacks.

## Security & Privacy
- Only public or derived binary data is encoded for transport (e.g., public keys, signatures, transaction bundles).
- We avoid logging raw encoded contents. Logs use short identifiers (thumbprints) where necessary.
- Base64url is not encryption or obfuscation — it’s just an encoding. Sensitive material must still be protected at rest and in transit (TLS).

## Operating It Day‑to‑Day
- No operator action is required. Encodings happen automatically at API boundaries.
- If an integration sends incorrect strings (wrong alphabet or padding), the server returns a clear error.

## Limitations & When to Adjust
- Some external tools produce padded base64; others don’t. Our tolerant decoder accepts both.
- If a downstream system requires padding, the client can re‑add it when needed without changing server behavior.

## Errors & Observability
- Malformed inputs produce “bad request” style errors (invalid alphabet/length); logs include a concise cause but not the data itself.

## Glossary
- base64: textual representation of binary data using `A–Z a–z 0–9 + /` plus `=` for padding.
- base64url: URL‑safe variant using `- _` instead of `+ /`; padding is optional.
- padding: extra `=` characters sometimes added to make lengths align; we omit them on encode.

## Refs
- Refs: decision encoding-and-ceremony-guardrails; requirement R-PLAT-2; requirement R-ERR
