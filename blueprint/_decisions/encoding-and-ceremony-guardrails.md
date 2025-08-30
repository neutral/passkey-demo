Title: Encoding and Ceremony Guardrails
Status: Accepted
Date: 2025-08-29
Context:
- Draft notes require strict handling of binary encodings and session binding for WebAuthn ceremonies to avoid interoperability pitfalls.
Decision:
- Use canonical CBOR on both client and server when building the signing bundle.
- Treat binary identifiers (credentialId, acct_cbor) as opaque; represent as base64url in JSON.
- Convert ArrayBuffer ⇄ base64url correctly; strip '=' padding on encode; accept both with/without padding on decode.
- Tie all verifications to server-supplied option sessions (registration/login/tx) to prevent replay/mismatch.
- Use hex only for human display (thumbprints, tx_ids), not for binary verification.
Consequences:
- Fewer encoding-related bugs; predictable hashing for anchors; safer verification lifecycle.
References:
- Refs: requirement schema-lightweight-cbor-bundle; requirement ops-local-dev-experience

