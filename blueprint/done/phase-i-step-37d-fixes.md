# Phase I — Step 37d — Fixes and UX Improvements

## Purpose
Address observed issues after the 37b refactor rollout: intermittent 401 on `/tx/signing/finish`, incorrect nonce recorded as 0, and reduce user error by auto-filling the next nonce in the UI.

## Changes

- Backend: Robust nonce decoding and validation
  - `server/internal/tx/bundle.go`
    - Added generic CBOR fallback to extract `nonce` (key 1) and `message` (key 2) when typed decode is insufficient.
    - Reject zero nonce (must be a positive integer) early; previously a zero-value could slip through for first transactions.
  - Spec update: `blueprint/features/message-signing-flow/_specs/nonce-policy-explainer.md` now explicitly states nonce > 0.

- Backend: Accept high‑S ECDSA signatures for transaction signing
  - `server/internal/tx/finish.go`
    - Strict verification attempted first; when the only failure is `ErrHighS`, normalize S (S' = N − S) and verify again.
    - Keeps all other checks unchanged (P‑256, strict DER, UV, rpIdHash/origin policy, allowlist, signCount policy).
  - ADR: `blueprint/_decisions/webauthn-accept-high-s-signing-too.md` (Accepted).
  - ADR superseded: `blueprint/_decisions/webauthn-accept-high-s-login-only.md` marked Superseded.
  - Description updated: `server/internal/tx/finish.go.desc.md` clarifies verification policy.

- Frontend: Auto-fill next nonce, read-only
  - `web/src/pages/Dashboard.tsx`
    - After loading the list, compute `next = max(nonce) + 1` (or `1` if none) and set the nonce input to this value.
    - Make nonce input read-only; users no longer type the nonce.
    - On successful sign, refresh the list (which auto-advances the nonce for the next bundle).
  - Description updated: `web/src/pages/Dashboard.tsx.desc.md` documents the auto-fill behavior.

## Verification

- Build/tests
  - `go build ./server/...`
  - `go test ./server/internal/tx -run TestTxFinish_Happy -v` → PASS

- Manual (dev)
  - Start server: `RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C server run ./cmd/api`
  - Login via UI.
  - Dashboard shows suggested nonce `1`; sign a message. List shows the new row with nonce = 1.
  - Nonce auto-advances to `2`; sign again. List shows nonce = 2.
  - No intermittent 401 on `/tx/signing/finish` observed; finish succeeds on first attempt.
  - Attempt to build a bundle with nonce 0 (manually simulate via API): `/tx/signing/options` responds `400` (invalid bundle) per server-side guard.

## Descriptions Updated

- `server/internal/tx/finish.go.desc.md`
- `web/src/pages/Dashboard.tsx.desc.md`

## Refs

Refs: requirement R-FLOW-SIGN; goal transaction-content-signing; decision encoding-and-ceremony-guardrails; decision weboauthn-corrections-and-standardizations; decision weboauthn-accept-high-s-signing-too
