# Purpose
Ensure a user can log in again after their previous session has expired, resulting in a fresh `sid` cookie being issued.

# Key Logic
- Perform a successful login to create a session and capture `sid`.
- Force-expire that session by setting `sessions.expires_at` to the past.
- Perform a second login; assert 200 and a new `sid` distinct from the expired one.

# Interactions
- Uses `BuildLoginOptions` and `LoginFinishHandler` to exercise full login finish logic.
- Directly updates the `sessions` table to simulate expiry between logins.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-2; spec session-cookies-usage-explainer; decision router-builder-wiring

