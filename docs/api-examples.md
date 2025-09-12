# API Examples (curl)

These examples assume the server runs on `localhost:8080` and the web app on `localhost:5173`. Adjust `PORT` if needed.

## Health
```bash
curl -sS localhost:8080/health -i
```

## Registration — Options
Issue options to begin a registration ceremony.
```bash
curl -sS -X POST localhost:8080/authn/passkey/registration/options -H 'Content-Type: application/json'
```
Response (example):
```json
{
  "reg_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {
    "rp_id": "localhost",
    "origin": "http://localhost:5173",
    "uv_required": true,
    "attestation": "none"
  },
  "expires_at": 1735689600
}
```

## Registration — Finish (payload skeleton)
Browser posts authenticator response back to the server. Example shape (not runnable):
```json
{
  "reg_session_id": "<b64>",
  "id": "<b64>",
  "rawId": "<b64>",
  "type": "public-key",
  "response": {
    "attestationObject": "<b64>",
    "clientDataJSON": "<b64>"
  }
}
```

## Login — Options
```bash
curl -sS -X POST localhost:8080/authn/passkey/login/options -H 'Content-Type: application/json'
```
Response (example):
```json
{
  "login_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {"rp_id":"localhost","origin":"http://localhost:5173","uv_required":true,"allow_credentials":[]},
  "expires_at": 1735689600
}
```

## Login — Finish (payload skeleton)
Browser posts authenticator response back to the server. Example shape (not runnable):
```json
{
  "login_session_id": "<b64>",
  "id": "<b64>",
  "rawId": "<b64>",
  "type": "public-key",
  "response": {
    "authenticatorData": "<b64>",
    "clientDataJSON": "<b64>",
    "signature": "<b64>",
    "userHandle": "<b64>"
  }
}
```

On success, the server sets a session cookie `sid` (HttpOnly). Use a cookie jar to persist it for subsequent calls.

## Account Key (requires cookie)
```bash
curl -sS -b cookies.txt -c cookies.txt \
  localhost:8080/me/account_key -i
```
Response (example):
```json
{
  "sender_key": {"kty": 2, "alg": -7, "crv": 1, "x": "<b64>", "y": "<b64>"},
  "acct_cbor_b64": "<b64>"
}
```

## Transactions — List (requires cookie)
```bash
curl -sS -b cookies.txt -c cookies.txt \
  localhost:8080/tx/list
```

## Transactions — Signing Options (requires cookie)
Send a base64url-encoded canonical CBOR bundle.
```bash
curl -sS -X POST -H 'Content-Type: application/json' \
  -b cookies.txt -c cookies.txt \
  localhost:8080/tx/signing/options \
  -d '{"bundle_cbor_b64":"<b64>"}'
```
Response (example):
```json
{
  "tx_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {"rp_id": "localhost", "origin": "http://localhost:5173", "uv_required": true, "allow_credentials": ["<b64>"]},
  "tx_id_hex": "deadbeef...",
  "expires_at": 1735689600
}
```

## Transactions — Signing Finish (payload skeleton)
Example shape (not runnable):
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
    "userHandle": "<b64>"
  }
}
```

## Notes
- For ceremony endpoints (`*/finish`), the browser constructs these payloads via WebAuthn APIs.
- Use a cookie jar (`-b cookies.txt -c cookies.txt`) to carry the `sid` cookie between calls requiring authentication.

