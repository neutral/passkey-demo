# Purpose
Normalizes the Node SimpleWebAuthn backend responses into the shapes expected by `@simplewebauthn/browser`. Handles camelCase JSON only (registration, login, and transaction signing) and builds finish payloads for dashboard signing.

# Key Logic
- `flattenOptions` lifts the nested `options` object returned by `/tx/signing/options` so login/signing share the same normalization path.
- `toCreationOptionsJSON`/`normalizeCreationOptions` validate RP/user metadata, enforce ES256, and retain server-provided challenges/user ids.
- `toRequestOptionsJSON`/`toRequestOptions` convert login and signing requests into native WebAuthn parameters (base64url→ArrayBuffer) while stripping fields the browser must not see (e.g., `origin`).
- `mapDomException` and `buildTxFinish` provide consistent error surfaces and finish payload encoding for dashboard flows.

# Interactions
- Registration/Login pages call these adapters before `startRegistration`/`startAuthentication`, then `postJson` the resulting JSON with session ids.
- Dashboard reuses the helpers to turn `/tx/signing/options` data into `navigator.credentials.get` parameters and to build finish payloads.
- Tests in `web/tests/*` ensure the adapters, POST bodies, and signing flows stay aligned with Node responses.

# Refs
Refs: requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec frontend-api-base-and-cors
