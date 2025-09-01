# Explainer: Registration Options (what and why)

Audience: Non-developers and stakeholders

What is this?
- When you press “Register”, the server creates a short-lived registration session and sends your browser a small JSON with a fresh random challenge and policy settings.
- Your browser turns this into the full WebAuthn `create()` request that your device’s passkey system understands.

Why these fields?
- `reg_session_id`: lets the server recognize which session (and challenge) your browser is answering later.
- `challenge`: a random value that prevents replay; your device signs data bound to this challenge.
- `rp_id`, `origin`: tie the passkey to the right site and web page (see RP ID vs Origin explainer).
- `uv_required`: requires biometric or PIN to prove it’s really you.
- `attestation: none`: we skip manufacturer trust chains to keep the demo simple.
- `expires_at`: the session goes stale after a few minutes for safety.

Safety notes
- Challenges and session IDs are random and short-lived. We don’t log them in plaintext.
- The next step (finish) verifies that the response matches this session and that the browser origin and site domain are correct.

