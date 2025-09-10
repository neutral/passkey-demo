# ck (Semantic Grep) — Usage Guide

## Purpose

Provide agents and developers with clear, actionable guidance to install, verify, index, and use `ck` for semantic and hybrid search in this repository.

## Install & Verify

- Check install: `ck --version` or `command -v ck`
- Install via Cargo: `cargo install ck-search`
- Build from source:
  - `git clone https://github.com/BeaconBay/ck && cd ck`
  - `cargo install --path ck-cli`

## Index Status & Management

- Where: `ck` stores its index in `./.ck/` (safe to delete and rebuild).
- Quick status: `ck --status .`
- Visual check: `test -d ./.ck && echo "ck index present" || echo "ck index missing"`
- Build index: `ck --index .` (index once; search many)
- Clean index: `ck --clean .` then rebuild with `ck --index .`
- Incremental add: `ck --add <path/to/file-or-dir>`

## Search Modes (choose by need)

- Default/Regex: `ck "pattern" path/` — drop-in `grep` compatibility, no index needed.
- Semantic: `ck --sem "concept or intent" path/` — meaning over keywords; requires index.
- Hybrid: `ck --hybrid "query" path/` — combines semantic + keyword for precision.

Tips:

- Full sections: add `--full-section` to return entire functions/classes around matches.
- Threshold: filter by score with `--threshold <0..1>` (e.g., `--threshold 0.05`).
- Top-K: limit results with `--topk N`.
- JSON for automation: add `--json` and pipe to `jq`.

## Practical Examples

- Intent-only discovery: `ck --sem "webauthn registration flow" .`
- Combine precision + meaning: `ck --hybrid "resident key|attestation" .`
- Find error handling: `ck --sem --full-section "error handling" server/`
- Exact tokens (fast): `ck -n "credentialId" web/` (regex/grep-compatible)
- Structured output: `ck --json --sem "Relying Party" . | jq '.[].file'`

## File Filtering & Ignores

- Respects `.gitignore` by default; use `--no-ignore` to include ignored files.
- Common caches/build dirs are excluded by default (`node_modules`, `target`, `build`, `__pycache__`, `.fastembed_cache`, etc.).
- Add/override: `--exclude <glob>` or `--no-default-excludes`.

## When To Use What

- Known exact text/regex or filenames → use default `ck` (grep-compatible).
- Conceptual intent across synonyms/patterns → use `--sem`.
- Reduce false positives but keep anchors → use `--hybrid`.
- Need whole function/class for analysis → add `--full-section`.
- Automating with scripts/agents → add `--json` (+ `--topk`, `--threshold`).

## Performance & Notes

- Indexing is a one-time cost per repo (safe to run repeatedly).
- Searches are sub-second on typical projects; entirely offline.
- `.ck/` is a cache; delete anytime to reclaim space and rebuild as needed.

## Repo Scenarios: grep vs ck

This repository includes a Go backend (`server/`), a TypeScript web app (`web/`), and extensive blueprint/spec documents (`blueprint/`, `specs/`). Use the following patterns to choose the best search mode.

### When grep (regex) is better

- Exact symbol/identifier lookup, definitions, or straightforward references.
  - Find a Go function definition: `ck -n '^func ParseAuthData' server/internal/webauthn/`
  - Locate a struct/type by name: `ck -n '^type CoseEC2' server/internal/types/`
  - Find constant/flag usage: `ck -n 'FlagUV|FlagUP|FlagAT' server/internal/webauthn/`
- Literal string/constants in code or tests.
  - Find exact error messages: `ck -n 'too many requests' server/internal/http/`
  - Find base64url helpers: `ck -n 'DecodeCanonical|EncodeCanonical' server/internal/encoding/`
- Quick TODO/FIXME sweeps, file globs, or filename-based targeting.
  - `ck -R -n 'TODO|FIXME' .`
  - `ck -n 'login_finish' server/internal/webauthn/*.go`

Why: Regex is immediate, precise, and requires no index; best when you know the token or pattern.

### When ck semantic search is better

- Conceptual intent across synonyms where exact words may differ.
  - WebAuthn origin verification logic: `ck --sem 'relying party origin verification' server/ specs/`
  - CBOR canonical encoding behavior: `ck --sem 'cbor canonical encoding rules' server/internal/encoding/ specs/`
  - Attestation parsing flow: `ck --sem --full-section 'webauthn attestation parsing' server/internal/webauthn/`
  - Session cookie lifecycle: `ck --sem 'session cookie middleware and auth vs session' server/internal/http/ blueprint/`
  - Rate limiting middleware: `ck --sem 'token bucket rate limiter' server/internal/http/`
  - Discoverable credentials semantics: `ck --sem 'resident keys discoverable allowCredentials semantics' blueprint/features/login-flow/_specs/ server/internal/webauthn/`
- Navigating design docs by topic, not exact phrasing.
  - `ck --sem 'rp origin test vectors and risks' specs/`
  - `ck --sem 'transaction signing bundle shape and account binding' blueprint/features/message-signing-flow/_specs/`

Why: Captures meaning (“attestation parsing”, “origin verification”) even when code/docs use varied wording.

### When ck hybrid is better

- You know an anchor token but want semantically related results ranked well.
  - Login flow and allowCredentials: `ck --hybrid 'allowCredentials login flow' server/internal/webauthn/ blueprint/features/login-flow/`
  - Nonce policy and replay protection: `ck --hybrid 'nonce replay protection' server/internal/tx/ blueprint/features/message-signing-flow/_specs/`
  - Base64url and clientDataJSON handling: `ck --hybrid 'clientDataJSON base64url decode' server/internal/webauthn/ server/internal/encoding/`
- Reduce false positives in large areas while keeping a must-have keyword.
  - `ck --hybrid --threshold 0.02 'attestation fmt none packed' server/internal/webauthn/`
  - `ck --hybrid --topk 10 'rate limit middleware' server/internal/http/`

Why: Blends precision of a known token with semantic ranking to surface the right context first.

### Repo-tailored examples

- Get entire functions for error handling insights in HTTP layer:
  - `ck --sem --full-section 'error handling http middleware' server/internal/http/`
- Inspect WebAuthn registration finish logic end-to-end:
  - `ck --hybrid --full-section 'registration finish attestation parse authData' server/internal/webauthn/`
- Find where session cookies are set/cleared across code and docs:
  - `ck --sem 'session cookie set clear httpOnly sameSite' server/internal/http/ blueprint/global/single-go-backend/_specs/`
- Trace CBOR/COSE key handling across modules:
  - `ck --hybrid 'COSE EC2 key parse' server/internal/webauthn/ server/internal/types/`
- Map frontend WebAuthn calls to backend handlers:
  - `ck --sem 'webauthn registration options endpoint' web/src/ server/`

### JSON and scripted consumption

- Return structured results for automation:
  - `ck --json --sem --topk 8 'origin verification' server/ specs/ | jq '.[] | {file, line, preview}'`
  - `ck --json --hybrid --threshold 0.03 'nonce policy' blueprint/ server/ | jq -r '.[].file' | sort -u`

### Quick decision guide

- Know the exact token/regex → use regex (grep-compatible `ck`).
- Know the concept but not the wording → use `--sem`.
- Have an anchor token but want smarter ranking/context → use `--hybrid`.
