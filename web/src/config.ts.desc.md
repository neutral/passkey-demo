# Purpose
Centralizes absolute API base configuration for the frontend and forbids reliance on a dev proxy. Ensures all network calls target `API_BASE` and work under CORS per Step 26.

# Key Logic
- `API_BASE` from `import.meta.env.VITE_API_BASE` with fallback `http://localhost:8080`.
- `apiUrl(path)` returns an absolute URL via `new URL(path, API_BASE)`.

# Interactions
- Imported by future steps to construct absolute fetch URLs and to pass `credentials: 'include'` as needed.

# Refs
Refs: goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b

