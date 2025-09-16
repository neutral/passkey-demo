# Registration (Passkey Creation)

State: Draft
Refs: goal passkey-registration-login-uv; goal webauthn-policy-defaults; goal key-first-identity-cose

## Goal

Create a new user account and a corresponding passkey credential on the user’s authenticator. The passkey’s public key will be stored as the user’s identity on the server.

## Flow Summary

The client obtains a registration challenge/options from the server, invokes the WebAuthn API to create a credential, then sends the result to the server for verification and account creation.

## Actors

- User
- Client (Browser/UI)
- Authenticator (Platform/Security Key)
- Server (Backend API)

## Preconditions

- Origin and `rpId` are configured and allowlisted; HTTPS in production.
- Browser supports WebAuthn and user verification (UV) ceremonies.
- No existing credential duplication for this account on the authenticator.

## Step-by-step

1. **Client Requests Options:** The user clicks “Register” on the React app. The client (browser) sends a request to the backend (e.g. `POST /authn/passkey/registration/options`) to get a **`PublicKeyCredentialCreationOptions`** object for registration. The request may include parameters like `rpId` (relying party ID, typically the web domain), `origin` (expected origin), and whether UV is required.
2. **Server Generates Challenge:** The backend generates a random **challenge** (a cryptographically random byte sequence ≥16 bytes) and builds the `publicKey` options for WebAuthn:

   - `rp` (Relying Party): identifies the server (e.g. `id` = domain, `name` = app name).
   - `user`: a temporary user handle (for demo, this can be a random byte ID since we have no username yet; it’s needed to satisfy the WebAuthn API, though in a passkey scenario this could be a dummy if using discoverable credentials).
   - `pubKeyCredParams`: allowed algorithms (use ECDSA w/ SHA-256, i.e. **COSE alg -7**, since we want an ES256 key).
   - `challenge`: the random challenge.
   - **`authenticatorSelection`: set `residentKey` to `"required"`** _(correction #4)_ **and** `userVerification` to `"required"` to enforce biometric/PIN verification during registration.
   - `attestation`: `"none"` (we do not need attestation data in this demo).
     The server stores this challenge and a temporary registration session ID in a server-side table (to verify later). It returns the options (typically in JSON) to the client along with the session ID.

3. **Client Initiates Credential Creation:** The React app uses the returned options to call `navigator.credentials.create({ publicKey: options })`. This prompts the authenticator (e.g. the browser’s built-in authenticator like Touch ID) to create a new key pair:

   - The user verifies their identity (e.g. fingerprint on MacBook Touch ID, since UV is required).
   - The authenticator generates a new private–public key pair scoped to our application (RP).
   - It signs the challenge and some metadata (client data and authenticator data) with the new private key, producing an **attestation** response.
   - The result is returned as a `PublicKeyCredential` object containing:

     - `id` (credential ID, a unique identifier for the credential, often an opaque byte array).
     - `rawId` (the same ID in byte form).
     - `response.attestationObject` (binary data with attestation info, including the public key and AAGUID).
     - `response.clientDataJSON` (JSON with the challenge, origin, and type).

   - This is all handled by the browser; the client code simply awaits the promise from `create()`.

4. **Client Sends Attestation to Server:** The browser then sends a `POST /authn/passkey/registration/finish` request to the backend with the new credential data and the `reg_session_id` from step 2. (All binary data like `attestationObject`, `clientDataJSON`, etc. are base64url-encoded in transit.)
5. **Server Verifies and Registers:**

   - The server retrieves the stored challenge and expected `rpId`/origin using `reg_session_id`.
   - It **parses `clientDataJSON`** and checks:

     - `type` is `"webauthn.create"` (registration ceremony).
     - The challenge in `clientDataJSON` matches the one it generated (ensuring this response is for the original challenge).
     - The origin matches the expected web origin (and is on an allowlist).

   - It **parses `attestationObject`** (CBOR data) to extract:

     - The new public key (in **COSE Key** format).
     - The credential ID.
     - Flags like whether the credential is user-verifying (UV), backup eligible, etc., and the initial signature counter.

   - It verifies the attestation signature if attestation format isn’t “none” (in our case attestation is "none", so we skip detailed attestation trust chain verification, simplifying the flow).
   - **Account Creation:** The server now creates a new user account record:

     - The **account ID** is set to the canonical CBOR encoding of the **COSE public key** (this serves as an identity in Model 2).
     - It may also store a hash (thumbprint) of this key for convenience (e.g. `acct_thumb = SHA-256("ACCTK1" || coseKeyBytes)`) for logs or references.

   - **Credential Storage:** The server saves a credential record **linked to that account**:

     - The credential’s `id` (as received from authenticator).
     - The **signCount** (initial signature counter from authenticator data) for future use.
     - (Audit fields like AAGUID and resident/backup flags may be stored as needed.)

     > **Correction #2:** In the demo, **store the COSE public key only in the `accounts` record** (as the identity). **Do not duplicate the public key in `credentials`**; credentials reference the account.

   - After successful verification and storage, the server responds with a success (e.g. HTTP 201 Created), possibly returning:

     - `account_id` (e.g. a hex representation of the account’s key thumbprint).
     - `credential_id` (base64url for reference).
     - (Optionally) the public key or key handle for debugging.

   - At this point, the user is **registered** with a passkey. The credential is now associated with the new account on the server.

## Important Details

- UV is required; server must enforce `userVerification: "required"` in options and verify the UV flag in authenticator data.
- Resident credentials are required; set `residentKey: "required"` to support username‑less flows.
- Attestation policy is `none`; skip trust-chain validation while still parsing attestation for the public key and flags.
- Store the COSE public key only on the account; credentials reference the account (no key duplication).
- Persist the initial `signCount` for later monotonic checks during login/signing.

## Security Notes

The registration flow ensures the public key is bound to the correct origin and challenge. By requiring user verification (UV), we ensure the credential creation involved a biometric/PIN, preventing silent or automated registrations. Attestation is set to none to avoid complexity; this means we don’t verify device provenance, which is acceptable for a demo.

## Request/Response Examples

- Request: `POST /authn/passkey/registration/options`

```json
{}
```

- Response: 200

```json
{
  "reg_session_id": "<b64>",
  "expires_at": 1735689600,
  "challenge": "<b64>",
  "rp": { "id": "example.com", "name": "Passkey Demo" },
  "user": {
    "id": "<base64url-32-bytes>",
    "name": "passkey-user",
    "displayName": "Passkey User"
  },
  "pubKeyCredParams": [{ "type": "public-key", "alg": -7 }],
  "authenticatorSelection": {
    "residentKey": "required",
    "requireResidentKey": true,
    "userVerification": "required"
  },
  "attestation": "none",
  "timeout": 60000
}
```

- Request: `POST /authn/passkey/registration/finish`

```json
{
  "reg_session_id": "<uuid>",
  "id": "<b64url>",
  "rawId": "<b64url>",
  "type": "public-key",
  "response": {
    "attestationObject": "<b64url>",
    "clientDataJSON": "<b64url>"
  }
}
```

- Response: 201

```json
{ "account_thumb_hex": "<hex-thumbprint>", "credential_id_b64": "<b64>" }
```

## Errors and Observability

- 400: invalid/missing fields or challenge mismatch; no retry.
- 401/403: origin not allowlisted; reject and log `origin`, `rpId`.
- 409: credential already registered; return conflict.
- Metrics: `registration_attempts`, `registration_success`, `registration_failures` with reason.
- Logs: include `reg_session_id`, `credential_id` (b64url), `acct_thumb` (if derived); exclude PII.

## Postconditions

- Account record created with COSE public key as identity.
- Credential record stored and linked to account; initial `signCount` persisted.
- No session is required by this flow unless specified by the application.

## Outputs

- 201 Created; returns `account_thumb_hex` and `credential_id_b64`.
- Database: new account and credential records with initial `signCount`.

## Verification

- Trigger options; run `navigator.credentials.create({...})` in the app; submit finish; expect 201 with `account_thumb_hex` and `credential_id_b64`.
- Check DB for new account and credential; confirm `signCount` stored and `residentKey` policy enforced.
