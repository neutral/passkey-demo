# R-FLOW-LOGIN — `allowCredentials` Semantics and Omission Guidelines (Technical)

## Metadata
- Status: Draft
- Date: 2025-09-05
- Owners: passkey-demo maintainers

## Purpose
- Provide developer-focused guidance on building `PublicKeyCredentialRequestOptions` for assertion (login), clarifying when and how to use `allowCredentials`, and why omitting it (not sending an empty array) is the correct way to enable discoverable-credential UX.

## Typical Flows
- Discoverable (resident) credentials / passkeys:
  - Omit `publicKey.allowCredentials` entirely.
  - Browser/authenticator performs account discovery for the RP ID and presents the passkey picker.
- Non‑discoverable (server‑resolved) credentials:
  - Include `publicKey.allowCredentials` with one or more credential descriptors `{ type: 'public-key', id: BufferSource }` that the server resolved for the user.

## Why Omit Instead of Sending `[]`
- Semantics: An empty array can be interpreted as “no credentials allowed,” not “no filter.”
- Cross‑browser behavior: Inconsistent; some implementations fail immediately or show “No passkeys available.”
- Intent clarity: Omission explicitly signals discoverable‑credential mode, aligning with WebAuthn’s account discovery behavior.

## Builder Pattern (Recommended)
- Build the list programmatically, then set the property only if non‑empty:

```ts
const ids = (resp.options.allow_credentials || [])
  .map(b64 => ({ type: 'public-key', id: base64urlToBytes(b64) }))
  .filter(d => d.id.byteLength > 0)

const publicKey: PublicKeyCredentialRequestOptions = {
  challenge: base64urlToBytes(resp.challenge),
  userVerification: 'required',
  ...(resp.options.rp_id ? { rpId: resp.options.rp_id } : {}),
  ...(ids.length > 0 ? { allowCredentials: ids } : {}),
}
```

## Edge Cases & Notes
- Discoverable credentials absent: The browser may show “No passkeys available.” Provide user guidance or fallback to a username flow to populate `allowCredentials`.
- Multiple credentials on record: Using `allowCredentials` targets a subset to reduce user choice or enforce policy.
- UV required: Keep `userVerification: 'required'` regardless of `allowCredentials` usage.
- Cookies: For finish requests, use `credentials: 'include'` and rely on HttpOnly cookies; do not read `sid` via JS.

## Testing Guidance
- Unit: Assert that builders omit `allowCredentials` when the decoded list is empty, and include it when non‑empty.
- E2E: With Chromium Virtual Authenticator, treat 401 (credential not recognized) as acceptable when the DB is unseeded; it still proves the browser → server flow.

## Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec allowcredentials-and-discoverable-credentials-explainer; spec-a; spec-b

