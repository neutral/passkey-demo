# Step 9 — Node: GET `/me/account_key` (Done)

### Step 9 — Node: GET `/me/account_key` (authenticated)

Scope

- Return the logged-in account’s COSE EC2 public key in JSON and base64url CBOR.

Source to add

- `node-server/src/me.js`: route using session context; CBOR decode to extract x/y for convenience, plus `acct_cbor_b64`.

Verification

- With valid `sid`, returns 200 and expected JSON; without `sid`, 401.

Acceptance criteria

- Shape matches current web expectations for bundle building.

Refs: requirement R-UI-2BTN; requirement R-SCHEMA-LITE; spec bundle-shape-and-client-production-explainer

---


