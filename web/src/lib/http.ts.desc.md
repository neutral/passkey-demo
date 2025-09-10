# Purpose
Lightweight helpers to normalize server errors into a consistent structure for UI toasts.

# Key Logic
- `parseHttpError(Response)`: builds `{ status, code?, title, detail }`; attempts JSON parse and prefers `error | message | detail | title` fields; falls back to `statusText` and `HTTP <status>`.
- `normalizeError(e)`: converts thrown errors and unknown values into `{ title, detail }`.

# Interactions
- Pages call `parseHttpError` on non-OK responses and `normalizeError` in catch blocks. Output feeds `ErrorToast`.

# Refs
Refs: global/error-responses-and-limits; requirement R-PLAT-1; goal simple-ui-and-storage

