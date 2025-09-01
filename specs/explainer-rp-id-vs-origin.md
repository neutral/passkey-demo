# Explainer: RP ID vs Origin (why both checks matter)

Audience: Non-developers and stakeholders

Summary
- RP ID and Origin are two different anchors the server checks to be confident a passkey assertion is for the right site, from the right browser context.

RP ID
- Think of RP ID as the site’s domain name (like `example.com`).
- Your passkey is tied to this RP ID. The authenticator includes a hash of it (`rpIdHash`) with every assertion.
- The server compares that hash to its expected RP ID to ensure the credential belongs to this site.

Origin
- Origin is more specific: it’s the full web address context (scheme + host + port), like `https://example.com` or `https://app.example.com:8443`.
- Browsers enforce strong rules about origins; we check the origin string we receive matches what we expect (or is explicitly allowed).

Why we need both
- RP ID confirms “this passkey belongs to this site”.
- Origin confirms “this request came from a browser page we trust (correct scheme, host, and port)”.
- Together, they stop cross-site and phishing-style mix-ups.

Development exception
- For local development only, WebAuthn allows `http://localhost`. We keep this as a narrow exception; everything else uses `https`.

What happens on failure
- If the RP ID hash doesn’t match, we reject the request (it’s for a different site).
- If the origin isn’t allowed, we reject the request (it’s not coming from the right web page).

