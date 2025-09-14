# Step 43 — Postman / REST Client file (Done: 2025-09-12)

## Scope
- Provide a Postman collection and a REST Client `.http` file with ready-to-run requests for core endpoints.
- Speed up manual testing and onboarding without building ceremony payloads by hand.

## Source to add/modify
- Add `docs/postman/passkey-demo.postman_collection.json` with folders/requests:
  - Health: `GET {{baseUrl}}/health`
  - Registration options: `POST {{baseUrl}}/authn/passkey/registration/options`
  - Registration finish (payload skeleton; example body only)
  - Login options: `POST {{baseUrl}}/authn/passkey/login/options`
  - Login finish (payload skeleton; example body only)
  - Account key (requires cookie): `GET {{baseUrl}}/me/account_key`
  - Tx list (requires cookie): `GET {{baseUrl}}/tx/list`
  - Tx signing options (requires cookie): `POST {{baseUrl}}/tx/signing/options`
  - Tx signing finish (payload skeleton; example body only)
- Add `docs/postman/local.postman_environment.json` with `baseUrl=http://localhost:8080`.
- Add `docs/rest-client/api.http` for VS Code REST Client users with parallel requests.

## Description files
- N/A (docs only).

## Request/response shape
- Collection includes example JSON responses for options endpoints and example request bodies for finish endpoints.
- REST Client file shows discrete requests with headers and notes about cookie handling.

## Algorithm
- Author collection with a `baseUrl` variable to simplify switching environments.
- Keep finish endpoints as non-runnable examples; focus runnable requests on health, options, account_key, tx list, tx options.

## Database interactions
- None.

## Policies & limits
- No secrets in the collection; use local defaults only.

## Sequencing
- Standalone documentation assets.

## Tests
- Manual: import the Postman collection and environment; send health/options requests; ensure 200/201-like responses. Use Postman's cookie store for authenticated calls after login finish is performed via the UI.
- VS Code: open `docs/rest-client/api.http`; run the health and options requests.

## Verification
- `test -f docs/postman/passkey-demo.postman_collection.json` returns 0.
- `test -f docs/rest-client/api.http` returns 0.

## User verification commands
```bash
# Verify files exist
ls -1 docs/postman/passkey-demo.postman_collection.json docs/postman/local.postman_environment.json docs/rest-client/api.http
```

## Acceptance criteria
- Postman collection and environment JSONs are present and importable.
- REST Client `.http` file present with parallel requests.

## Notes
- Postman automatically handles `Set-Cookie`/`Cookie` for subsequent requests in a session. Finish endpoints remain examples only.

## Refs
- Refs: requirement R-PLAT-2; requirement R-OPS-DEV

