Title: WebAuthn Corrections and Standardizations
Status: Accepted
Date: 2025-08-29
Context:
- The draft specs identified inconsistencies that need normalization for a secure demo.
Decision:
1) Server-supplied options for transaction signing: derive challenge = SHA-256("CHALv1"||B).
2) Single source of truth for public key: store COSE key in `accounts` only; credentials reference accounts.
3) Transaction identity anchor: standardize on "TXIDv1".
4) Resident credentials: set `residentKey = "required"` to support username-less login.
5) Nonce policy: enforce strictly increasing nonce per account during signing options.
6) SignCount policy: require strictly increasing signCount; reject equal/lower.
Consequences:
- Aligns WebAuthn flows with best practices; simplifies data model; strengthens replay and downgrade protections.
References:
- Refs: requirement security-user-verification-required; requirement schema-lightweight-cbor-bundle; requirement identity-passkey-first-cose-key

