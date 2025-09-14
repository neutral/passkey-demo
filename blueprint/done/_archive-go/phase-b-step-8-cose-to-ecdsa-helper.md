### Step 8 — COSE → ECDSA helper (Done: 2025-08-31)

   - Context
     - Convert a COSE EC2 public key (from WebAuthn) into a Go `ecdsa.PublicKey` to verify signatures for login and signing. Enforce algorithm/curve checks (ES256/P‑256) and basic well‑formedness.

   - Structure
     - Add `server/internal/crypto/crypto_cose.go` with a single helper:
       - `func ToECDSA(k *types.CoseEC2) (*ecdsa.PublicKey, error)`.

   - Source to add (instructions only)
     - `server/internal/crypto/crypto_cose.go`:
       - Validate `k != nil`, `k.Kty==2` (EC2), `k.Alg==-7` (ES256), `k.Crv==1` (P‑256).
       - Require `len(k.X)==32 && len(k.Y)==32`; construct big.Ints; create `ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}`.
       - Check `curve.IsOnCurve(X,Y)`; return error if not.
       - No private key handling in this helper.

   - Description files to add (instructions only)
     - `server/internal/crypto/crypto_cose.go.desc.md`: Purpose (COSE→ECDSA conversion), Key Logic (parameter checks, on‑curve test), Interactions (used by signature verifier), Refs.
       - Refs: requirement R-ID-KEY; requirement R-PLAT-2.

   - Blueprint updates
     - Refs to include upon implementation: goal key-first-identity-cose; requirement R-ID-KEY; requirement R-PLAT-2.

   - Verification (to run after implementation)
     - Build: `cd server && go build ./...` (expect exit 0).
     - Quick sanity (optional inline runner): construct a known valid P‑256 point (or derive from an ecdsa keypair) and ensure `ToECDSA` returns a key and `IsOnCurve` passes.

   - User verification commands (copy/paste)

     ```bash
     cd server && go build ./... && cd -
     # Optional: run a tiny inline program to exercise ToECDSA when implemented.
     ```

   - Notes
     - Keep scope narrow: conversion and validation only. Signature verification and low‑S checks live in Step 13.

