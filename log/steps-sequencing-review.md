Sequencing Review

- Build order is mostly coherent: server scaffolding → config/DB/models → WebAuthn helpers → registration → login → signed tx → sessions/
  limits → UI → tests → hardening → DX.
- Each server phase compiles independently and introduces only used deps. Frontend work starts after core APIs exist.

Blocking Issues Found (and fixed)

- Vite scaffold path: adjusted to avoid nesting web/web so step 4 works from repo root or inside web/.
- CORS before UI: added guidance in step 26 to use a Vite proxy or permissive dev CORS so frontend can call API before hardening step 37.
- Session middleware timing: clarified in steps 21/22 to perform inline cookie→session lookup until middleware (step 24) is added,
  avoiding a gap where “auth required” routes would fail.
- TTL enforcement: added “verify not expired” checks to registration/login/tx finish steps so the 5‑minute TTL is actually enforced when
  implemented.
- Frontend CBOR dependency: added explicit install of a small CBOR lib in step 31 to build canonical CBOR bundles.

Order Soundness by Phase

- Phase A: OK. Step 4 now works without nesting issues.
- Phase B: OK. DB migrations precede any persistence; types/crypto/base64/CBOR helpers precede their use.
- Phase C: OK. Verification utilities ready before endpoints use them.
- Phase D/E: OK. Registration then login; persistence and session creation happen after schema is in place.
- Phase F: OK. Signing options/finish can run with inline session checks; DB schema supports nonce/tx storage.
- Phase G: OK. Middleware and limits can be added after handlers; handlers already work with inline checks.
- Phase H: OK. UI can run with proxy/CORS note; APIs already exist.
- Phases I–K: Tests, hardening, and DX steps don’t block earlier phases.
