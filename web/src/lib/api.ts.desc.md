# Purpose
Lightweight API client wrapper around `fetch` that enforces `credentials: 'include'`, JSON content type, and decodes the backend error envelope into a typed `ApiError`.

# Key Logic
- `apiFetch<T>(url, init)` parses `{code,error,correlation_id}` on non-2xx and throws `ApiError`.
- `postJson<T>(url, body)` convenience helper for JSON POST.

# Interactions
- Pages can centralize error-to-toast mapping based on `ApiError.code` (e.g., `ERR_UNAUTHORIZED`, `ERR_CONFLICT`).

# Refs
Refs: requirement R-ERR; decision http-error-envelope

