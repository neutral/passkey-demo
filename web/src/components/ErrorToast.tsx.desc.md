# Purpose
Reusable, dismissible UI banner for surfacing errors (role="alert"). Used across Register, Login, and Dashboard to show HTTP failures and thrown error messages without blocking flows.

# Key Logic
- Props: `title`, `detail`, optional `code` and `status`, and `onClose`.
- Renders a fixed-position toast at top-right; includes a close button.
- Accessible via `role="alert"` and `aria-live="assertive"`.

# Interactions
- Consumed by page components. Works with `parseHttpError` and `normalizeError` from `web/src/lib/http.ts`.

# Refs
Refs: global/error-responses-and-limits; requirement R-PLAT-1; goal simple-ui-and-storage

