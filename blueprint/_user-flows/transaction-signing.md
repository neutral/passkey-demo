# Transaction Signing (Post-Login)

State: Draft
Refs: goal transaction-content-signing; goal minimal-cbor-bundle; goal server-derived-challenge-and-txid

## Goal

Allow an authenticated user to create a simple “transaction” (a message payload) that is signed using their passkey. The server verifies the signature and stores the message, associating it with the user’s account.

## Flow Summary

The client prepares the transaction bundle and sends it to the server to obtain WebAuthn assertion options whose challenge is deterministically derived from the bundle. The client then invokes WebAuthn with those options, and finally submits the resulting assertion to the server for verification and storage.

## Actors

- User
- Client (Browser/UI)
- Authenticator (Platform/Security Key)
- Server (Backend API)

## Preconditions

- User is authenticated (session established via login flow).
- A credential is registered and linked to the account; COSE key available to the client.
- Origin and `rpId` are configured and allowlisted; HTTPS in production.

## Step-by-step

1. **User Input:** On the dashboard (post-login UI), the user enters a message (or minimal structured data) and clicks “Sign and Submit”.
2. **Client Builds Transaction Bundle:** The client constructs a **Bundle** (CBOR) containing:

   - `sender_key`: the user’s **COSE** public key (from registration; Model 2 identity).
   - `nonce`: a numeric nonce (recommend per-account monotonic counter).
   - `message`: the user-typed content.
   - (Optional) `valid_until`: expiry timestamp.
     Encode to **canonical CBOR** to produce bytes `B`.

3. **Client Requests **Server-Supplied Options**:** The client sends `B` (base64url) to `POST /tx/signing/options`.

   - **Server recomputes canonical CBOR** (`B`) and derives:
     `challenge = SHA-256("CHALv1" || B)` and **`tx_id = SHA-256("TXIDv1" || B)`** _(correction #3)_.
   - **Server enforces nonce monotonicity per account** (reject if `nonce` ≤ last seen) _(correction #5)_.
   - Server returns `PublicKeyCredentialRequestOptions` with that **challenge**, `rpId`, `userVerification: "required"`, and `allowCredentials` containing the user’s credential ID.

4. **Client Creates Assertion:** The client calls `navigator.credentials.get({ publicKey: options })`. The authenticator:

   - Prompts for biometric/PIN (UV).
   - Produces `authenticatorData`, `clientDataJSON`, and `signature` over `authenticatorData || SHA-256(clientDataJSON)`.

5. **Client Submits Assertion:** The client POSTs to `POST /tx/signing/finish` with:

   - `credential.id`, `response.clientDataJSON`, `response.authenticatorData`, `response.signature`, and the `tx_session_id` (or equivalent) from step 3.

6. **Server Verifies and Stores:**

   - Load stored `B` and server-derived challenge.
   - Verify `clientDataJSON` (`type="webauthn.get"`, challenge equals stored challenge, origin allowlisted).
   - Verify `authenticatorData` (`rpIdHash`, UV flag, and `signCount` strictly increasing) and signature (low‑S ECDSA P-256) using the account’s public key.
   - On success: compute/store `tx_id`, persist the transaction record (account, nonce, message, raw `B`, AD, CDJ, signature, timestamp).
   - Return success (e.g., 200 + `tx_id`).

## Important Details

- The **server** generates the WebAuthn assertion options (and thus the challenge) for transaction signing, but does so **deterministically from client-provided content** (`challenge = SHA-256("CHALv1" || B)`), ensuring the signature is bound to the exact bundle content _(correction #1)_.
- The **`tx_id`** used for record identity is standardized in this demo as **`SHA-256("TXIDv1" || B)`** _(correction #3)_.
- The server **enforces nonce monotonicity** per account to prevent replay _(correction #5)_.
- UV is required for signing; the server verifies the UV flag.
- Enforce **low‑S** ECDSA signatures when verifying assertions.

## Security Notes

- Verify `type == "webauthn.get"` and challenge equality; authenticate only from allowlisted `origin` and correct `rpIdHash`.
- Require UV in options and confirm UV flag in authenticator data.
- Enforce nonce monotonicity per account; reject duplicates or regressions.
- Enforce ES256 low‑S signatures; reject invalid DER or non‑canonical encodings.

## Request/Response Examples

- Request: `POST /tx/signing/options`

```json
{ "bundle_cbor_b64": "<b64url-cbor-B>" }
```

- Response: 200

```json
{
  "tx_session_id": "<b64>",
  "challenge": "<b64>",
  "tx_id_hex": "<hex>",
  "expires_at": 1735689600,
  "options": {
    "rpId": "example.com",
    "origin": "https://example.com",
    "userVerification": "required",
    "challenge": "<b64>",
    "allowCredentials": [
      { "type": "public-key", "id": "<b64>" }
    ],
    "timeout": 60000
  }
}
```

- Request: `POST /tx/signing/finish`

```json
{
  "tx_session_id": "<b64>",
  "id": "<b64>",
  "rawId": "<b64>",
  "type": "public-key",
  "response": {
    "authenticatorData": "<b64>",
    "clientDataJSON": "<b64>",
    "signature": "<b64>",
    "userHandle": ""
  }
}
```

- Response: 200

```json
{ "tx_id_hex": "<hex>", "stored": true }
```

## Errors and Observability

- 400: invalid bundle encoding or challenge mismatch → envelope `{code: "ERR_BAD_REQUEST"}`.
- 401: unauthorized (missing session), sender_key mismatch, or invalid signature/origin → `{code: "ERR_UNAUTHORIZED"}`.
- 403: origin or rpId policy violation → `{code: "ERR_FORBIDDEN"}`.
- 409: nonce not strictly increasing or no credentials on account → `{code: "ERR_CONFLICT"}`.
- Correlation: responses may include `correlation_id` for tracing.
- Metrics: `tx_sign_attempts`, `tx_sign_success`, `tx_sign_failures` with reason; latency per endpoint.
- Logs: include `tx_session_id`, `tx_id`, `credential_id`, `nonce`, `acct_thumb`; exclude raw `B` from logs; store in DB only.

## Postconditions

- Transaction persisted with `tx_id`; record includes account, nonce, message, raw `B`, AD/CDJ, signature, timestamp.
- Account’s last seen `nonce` advanced; credential `signCount` advanced.

## Outputs

- 200 OK; returns `tx_id` for the persisted transaction.
- Database: new transaction record with raw `B`, AD, CDJ, signature, and metadata.

## Verification

- Build canonical CBOR bundle; call options; run `navigator.credentials.get({...})`; submit finish; expect 200 with `tx_id`.
- Query list endpoint or DB to confirm record exists with matching `tx_id` and `nonce`.
