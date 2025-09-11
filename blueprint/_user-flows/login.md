# Login (Passkey Authentication)

State: Draft
Refs: goal passkey-registration-login-uv

## Goal

Allow an existing user to authenticate with the previously registered passkey, establishing a session (or simply confirming their identity to proceed to the app).

## Flow Summary

The client requests a login challenge, uses the authenticator to sign the challenge (WebAuthn assertion), and the server verifies the signature using the stored public key.

## Actors

- User
- Client (Browser/UI)
- Authenticator (Platform/Security Key)
- Server (Backend API)

## Preconditions

- Origin and `rpId` configured and allowlisted; HTTPS in production.
- Browser supports WebAuthn and user verification (UV).
- A credential was previously registered and stored for the account.

## Step-by-step

1. **Client Requests Assertion Options:** When the user clicks “Login”, the client sends a request (e.g. `POST /authn/passkey/login/options`) to the server for a **`PublicKeyCredentialRequestOptions`** (assertion options). Because this demo uses username-less (passkey) login, the client may not supply a username; instead, the server can allow the authenticator to select the credential. The server generates (dev note: `rpId` and `origin` must be allowlisted):

   - A fresh random **challenge** (bytes).
   - `rpId` set to the relying party ID (domain).
   - An **allowCredentials** list: for username-less, this can be left empty to allow any credential recognized by the authenticator for this RP. Alternatively, if the user had identified which account to use, the server could include that user’s credential ID. For demo simplicity, we may omit `allowCredentials` or include a single credential ID.
   - `userVerification`: `"required"` (we enforce biometric/PIN on login as well).
     The server stores the challenge (and possibly a temporary session context if needed) and returns these options to the client.

2. **Client Initiates Assertion:** The client calls `navigator.credentials.get({ publicKey: options })` using the received options. This triggers the authenticator to find the matching credential and prompt the user:

   - The user provides biometric/PIN (UV required).
   - The authenticator uses the private key from the previously created credential to sign the new challenge and relevant data.
   - It returns a `PublicKeyCredential` with:

     - `id` (credential ID used – identifies which key was used).
     - `response.authenticatorData` (binary data with flags and the new signature counter).
     - `response.clientDataJSON` (JSON with the challenge, origin, and type).
     - `response.signature` (the signature over the concatenated auth data and hash of client data).
     - (Optionally `response.userHandle` if a resident credential; not critical here).

3. **Client Sends Assertion to Server:** The client sends a `POST /authn/passkey/login/finish` request with the credential ID and response fields (base64url encoded).
4. **Server Verifies Assertion:**

   - Look up the user account by the `credential_id` (stored at registration).
   - Retrieve the stored public key (from the account record).
   - Verify **clientDataJSON**:

     - `type` == `"webauthn.get"`.
     - The challenge in `clientDataJSON` matches the server-generated challenge.
     - The `origin` matches our allowlisted origin (e.g. `https://example.com`).

   - Parse **authenticatorData**:

     - Check that `rpIdHash` equals `SHA-256(rpId)`.
     - Verify the **UV flag** is set (since UV is required).
     - **Check `signCount` is strictly increasing** relative to stored value; **reject** if not _(correction #6)_; update stored counter on success.

   - Compute `signatureBase = authenticatorData || SHA-256(clientDataJSON)` and verify the ECDSA P-256 signature (**enforce low‑S**) with the stored public key.
   - If all checks pass, establish a session (e.g. secure cookie) and return HTTP 200.

The user is now logged in without a password, purely via cryptographic authentication, and can view the dashboard (previous messages and a form to add a new one).

## Important Details

- Username‑less login supported; `allowCredentials` may be omitted or include a single credential ID.
- UV is required in assertion options and must be verified via the UV flag.
- `signCount` must be strictly increasing; update the stored counter on success.
- Signature algorithm: ES256; enforce low‑S during verification.
- Origin allowlist and `rpIdHash` checks are mandatory.

## Security Notes

- Authenticate only against allowlisted `origin` and correct `rpIdHash`.
- Verify `type == "webauthn.get"` and challenge match in `clientDataJSON`.
- Enforce UV and monotonic `signCount` to mitigate phishing and replay.

## Request/Response Examples

- Request: `POST /authn/passkey/login/options` (empty body)

```json
{}
```

- Response: 200

```json
{
  "login_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {
    "rp_id": "example.com",
    "origin": "https://example.com",
    "uv_required": true,
    "allow_credentials": []
  },
  "expires_at": 1735689600
}
```

- Request: `POST /authn/passkey/login/finish`

```json
{
  "login_session_id": "<uuid>",
  "credential": {
    "id": "<b64url>",
    "type": "public-key",
    "response": {
      "authenticatorData": "<b64url>",
      "clientDataJSON": "<b64url>",
      "signature": "<b64url>",
      "userHandle": "<b64url>"
    }
  }
}
```

- Response: 200 (sets `sid` HttpOnly cookie)

```json
{ "account_thumb_hex": "ab12...fe", "credential_id_b64": "<b64>" }
```

## Errors and Observability

- 400: missing fields or malformed payload → envelope `{code: "ERR_BAD_REQUEST"}`.
- 401: invalid assertion (challenge/origin/signature) → `{code: "ERR_UNAUTHORIZED"}`.
- 403: origin or rpId policy violation → `{code: "ERR_FORBIDDEN"}`.
- 409: non‑monotonic `signCount` → `{code: "ERR_CONFLICT"}`.
- Responses may include `correlation_id` for tracing.
- Metrics: `login_attempts`, `login_success`, `login_failures` with reason.
- Logs: include `login_session_id`, `credential_id`, `acct_thumb`; exclude PII.

## Postconditions

- Session established (HttpOnly cookie `sid`) for the authenticated account.
- Stored `signCount` advanced to the value from authenticator data.

## Outputs

- 200 OK; session established for the account.
- Updated credential record with advanced `signCount`.

## Verification

- Trigger options; run `navigator.credentials.get({...})` in the app; submit finish; expect 200 and session cookie.
- Confirm session-protected endpoint (e.g., `GET /me`) returns account context.
