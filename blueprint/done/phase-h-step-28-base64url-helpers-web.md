### Step 28 — Base64url helpers (web) (Done: 2025-09-05)

Verification notes
- Playwright: 6 passed (encoding specs + UI smoke)
- Encoding vectors: RFC 4648 known strings matched; multibyte UTF‑8 round-trip; invalid input rejected; 4 KiB random round-trip OK
- No app/UI wiring yet (consumed in later steps)

28. **Base64url helpers (web)**

    - Scope:

      - Add browser-safe utilities for binary/text conversions used across WebAuthn and signing flows:
        - ArrayBuffer/Uint8Array ⇄ base64url (RFC 4648 URL-safe, no padding).
        - UTF‑8 string ⇄ Uint8Array using `TextEncoder`/`TextDecoder`.
      - Centralize in a small `lib` module and keep zero external deps.

    - Source to add/modify:

      - Add: `web/src/lib/encoding.ts` — Implements:
        - `bytesToBase64url(input: ArrayBuffer | Uint8Array): string`
        - `base64urlToBytes(s: string): Uint8Array`
        - `utf8ToBytes(s: string): Uint8Array`
        - `bytesToUtf8(b: ArrayBuffer | Uint8Array): string`
        - Implementation notes: works in browser and Node test env; use `btoa/atob` when available, otherwise Node `Buffer` fallback; strip `=` padding on encode; accept optional padding on decode.
      - Add (tests): `web/tests/encoding.spec.ts` — Unit tests via Playwright runner to assert conversions and invariants.
      - (No changes to app/UI yet; consumers will be wired in Steps 29–33.)

    - Description files (create/update alongside code changes):

      - Create: `web/src/lib/lib.desc.md` — Folder overview for shared frontend utilities (encoding, later CBOR).
      - Create: `web/src/lib/encoding.ts.desc.md` — High-level description of functions, invariants (no padding; URL-safe), and environment fallbacks. Refs included.
      - Update: `web/web.desc.md` — Mention centralized encoding utilities used by registration/login/signing flows.

    - Request/response shape:

      - Binary fields in JSON use base64url (no `=` padding). Examples used later:
        - Registration finish: `id`, `rawId`, `response.clientDataJSON`, `response.attestationObject`.
        - Login/signing finish: `id`, `rawId`, `response.authenticatorData`, `response.clientDataJSON`, `response.signature`, `response.userHandle?`.
      - UTF‑8 encoding used for string-to-bytes where needed (e.g., message → CBOR, or debugging hex previews in Step 32).

    - Algorithm:

      - bytesToBase64url:
        - Ensure `Uint8Array` view of input.
        - Encode to base64 using `btoa(String.fromCharCode(...bytes))` in browser; fallback to `Buffer.from(bytes).toString('base64')` in Node.
        - Convert to URL-safe: replace `+`→`-`, `/`→`_`, strip trailing `=`.
      - base64urlToBytes:
        - Normalize: replace `-`→`+`, `_`→`/`; add `=` padding to length % 4 → 0.
        - Decode with `atob` (browser) or `Buffer.from(b64, 'base64')` (Node).
        - Return `Uint8Array` of bytes; reject inputs containing non-base64 characters (throw `TypeError`).
      - utf8ToBytes / bytesToUtf8:
        - Use `new TextEncoder().encode(s)` and `new TextDecoder('utf-8', {fatal:true}).decode(bytes)`.
      - Invariants:
        - Encoding output contains only `[A-Za-z0-9_-]`; no padding.
        - Round-trip: `base64urlToBytes(bytesToBase64url(x))` deep-equals original bytes for a variety of inputs, including empty, ASCII, and multibyte strings.

    - Database interactions:

      - None.

    - Policies & limits:

      - RFC 4648 URL-safe alphabet; do not emit `=` padding on encode.
      - Be liberal in what you accept: decode may accept padded input but always emit unpadded on encode.
      - No global polyfills (Buffer, atob/btoa); implement small environment fallback checks inside the module.

    - Sequencing:

      - Step 27 completed UI scaffolding; Step 28 prepares helpers needed by Steps 29–33.
      - Consumers: Step 29 uses decode (b64url→ArrayBuffer) for options; Steps 29–30 use encode for finish payloads; Steps 32–33 may use UTF‑8 for message → bytes before CBOR.

    - Tests (happy path required, plus negatives):

      - Files to add:
        - `web/tests/encoding.spec.ts`
      - Happy path:
        - bytes ⇄ base64url round-trip on sample byte arrays: `[]`, `[0]`, `[0,255]`, `[1,2,3,4,5]`.
        - utf8 ⇄ bytes round-trip on `""`, `"foo"`, `"Hello, world!"`, and multibyte `"π🙂"`.
        - Known vectors (RFC 4648): `"f" → Zg`, `"fo" → Zm8`, `"foo" → Zm9v`, `"foob" → Zm9vYg`, `"fooba" → Zm9vYmE`, `"foobar" → Zm9vYmFy` (no padding).
      - Negative/boundary cases:
        - base64url decode rejects strings with invalid characters (e.g., `*` or whitespace) with `TypeError`.
        - Decoder accepts padded input (e.g., `"Zg=="`) and yields same bytes as `"Zg"`.
        - Large input sanity: encode/decode a 4 KiB random byte array.
      - Commands:
        - `npm -C web run test:ui` (Playwright runner executes `encoding.spec.ts`).

    - Verification:

      - Unit tests pass with the cases listed above.
      - Manual spot-check in browser console after building:
        - `bytesToBase64url(new TextEncoder().encode('Hello')) === 'SGVsbG8'`.
        - `new TextDecoder().decode(base64urlToBytes('SGVsbG8')) === 'Hello'`.
      - Fix-forward loop: adjust implementation on failures and re-run tests until green.

      - User verification commands:

        ```bash
        # 1) Install deps (if not already)
        npm -C web ci || npm -C web install

        # 2) Implement encoding helpers and tests (per files above), then run UI tests
        npm -C web run test:ui --silent

        # 3) Manual quick check in a running dev session (optional)
        npm -C web run dev &
        WEB_PID=$!
        # Open http://localhost:5173 and verify in console using the examples in Verification
        kill $WEB_PID || true
        ```

    - Acceptance criteria:

      - `web/src/lib/encoding.ts` exports the four helpers; code works in both browser and Playwright (Node) environments.
      - Base64url outputs are URL-safe and unpadded; decoder accepts padded or unpadded.
      - Playwright tests in `web/tests/encoding.spec.ts` pass (happy path + negatives above).
      - Description files for `lib/` and `encoding.ts` exist and include Refs.

    - Notes:

      - Keep helpers small and dependency-free. Do not introduce polyfills or large utility libraries.
      - Prefer throwing `TypeError` for invalid base64url inputs to surface errors clearly to callers.

    - Refs:
      - Refs: requirement R-PLAT-1; requirement R-ERR; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b

