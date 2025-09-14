### Step 18 — /authn/passkey/login/options handler (Done: 2025-09-02)

Scope
- Issue WebAuthn assertion options with a fresh 32-byte challenge and create a short-lived login session bound to RP ID and Origin.
- Keep `allowCredentials` empty (discoverable credentials) for the demo; filtering can be added later.

Source to add/modify
- `server/internal/webauthn/login_options.go`: Login session model/store; `BuildLoginOptions`; `LoginOptionsHandler` (POST).
- `server/internal/webauthn/login_options.go.desc.md`: High-level description, relations, invariants, and Refs.
- `server/internal/webauthn/login_options_test.go`: Unit tests for builder (entropy, TTL) and handler JSON shape.
- `server/cmd/api/main.go`: Instantiate `LoginSessionStore(10000)`; mount `POST /authn/passkey/login/options`.
- `server/internal/types/types.go`: Ensure `LoginOptions` includes `AllowCredentials []string`.
- `server/internal/types/types.go.desc.md`: Note `allow_credentials` field and Refs to login requirement/spec.

Description files
- `server/internal/webauthn/login_options.go.desc.md`
  - One-liner: Issues assertion options and tracks short-lived login sessions (rpId, origin, challenge, TTL).
  - Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; spec R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations

Blueprint refs
- Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; spec R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations

Policies and limits
- Session ID: 24 random bytes (base64url in JSON).
- Challenge: 32 random bytes (base64url in JSON).
- TTL: 5 minutes; sessions stored in-memory with capacity 10k (demo-grade).
- Method: POST only; respond 405 for others.
- JSON: `login_session_id`, `challenge`, `options{ rp_id, origin, uv_required=true, allow_credentials=[] }`, `expires_at` (unix seconds).

Sequencing
- `allowCredentials` kept empty for discoverable credentials; optional filtering by known credential IDs can be added after Step 19 introduces login finish and DB lookups.
- No middleware yet; in-memory store created in `main.go` alongside registration store.

Verification
- Builder tests: verify 32B challenge, ≥16B entropy for session id, TTL ≈ now+5m, store contains matching entry, options fields correct.
- Handler test: POST returns 200 + JSON with non-empty `login_session_id`, `challenge`, `options` (`rp_id`, `origin`, `uv_required=true`, `allow_credentials=[]`), and `expires_at`.
- Manual `curl` check against running server.

User verification commands
```bash
# Run unit tests
cd server && GOCACHE=$(pwd)/.gocache go test ./...

# Start the API (env defaults OK for demo)
go run ./cmd/api &
API_PID=$!
sleep 0.5

# Request login options
curl -sS -X POST :8080/authn/passkey/login/options | jq \
  '.login_session_id, .challenge, .options.rp_id, .options.origin, .options.uv_required, .options.allow_credentials, .expires_at'

# Expect: non-empty ids, rp_id/origin match config, uv_required=true, allow_credentials=[], expires_at ≈ now+300s

# Cleanup
kill $API_PID
```

Acceptance criteria
- `go test ./...` passes; handler responds 200 with correct JSON shape and TTL; sessions stored in memory.

Notes
- Error mapping and structured logging are out-of-scope here; integrate with MapPolicyError/Verify logging in later steps.

