# Explainer: Registration Finish (what happens on the server)

Audience: Non‑developers and stakeholders

Summary
- This is the final step of creating a passkey (registration). Your browser returns a signed package from your device (the “authenticator”). The server checks it carefully and then stores your new passkey so you can log in later without a password.

What we receive
- reg_session_id: An opaque token created earlier when you pressed Register. It binds this response to a short‑lived server session.
- clientDataJSON: A small JSON blob from the browser with the ceremony type, the random challenge we issued, and the page’s origin (site URL).
- attestationObject: A CBOR (binary) structure from the authenticator that contains the authenticator data, including your new public key and a credential identifier.

What we verify (safety checks)
- Session: The reg_session_id exists, is not expired, and is single‑use.
- Challenge: The challenge in clientDataJSON matches the one we issued for this session.
- Origin: The origin (scheme + host + port) in clientDataJSON matches our configured site (or an explicit allowlist for other environments).
- RP ID hash: The hash stored by the authenticator for the relying party (our domain) matches our configured RP ID.
- User Verification (UV): The authenticator indicates you confirmed with biometric/PIN (UV required by policy).
- Attestation policy: We accept only fmt: "none" (no manufacturer trust chain), which is standard for a simple demo.

What we extract and store
- Public key (COSE EC2, P‑256): This is the passkey’s public half; it lets us verify future signatures.
- Credential ID: A unique identifier the authenticator uses to locate the passkey later.
- Sign counter: A number that increases with each use; later we’ll use it to detect clones.
- Account identity: We use your public key as your account identity. We store the key as canonical CBOR (a compact binary format) and compute an account “thumbprint” (a short hex string) as SHA‑256("ACCTK1" || acct_cbor) for display.
- Database rows:
  - accounts: acct_cbor (binary public key), acct_thumb (hash), created_at.
  - credentials: credential_id, acct_cbor_fk (foreign key), sign_count, optional aaguid, created_at.

What we return to the app
- Success (201): account_thumb_hex and the credential ID (encoded), which the frontend can show or log for debugging.
- Failures: Clear, generic errors (e.g., session expired, origin mismatch, invalid data). We do not expose sensitive details in responses.

Privacy & security notes
- No private keys ever leave your device; the server only stores public data.
- Session tokens and challenges are short‑lived and random; we avoid logging them in plaintext.
- Origin/RP ID checks prevent mix‑ups and phishing. UV ensures you approved the operation on your device.

Why this matters
- These checks bind your new passkey to the correct site and page, and ensure you actually approved the creation. With the public key safely stored, future logins can be verified without a password.

