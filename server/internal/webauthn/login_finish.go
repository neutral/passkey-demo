package webauthn

import (
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "log"
    "net/http"
    "net/url"
    "strings"
    "time"

    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    cryptoutil "github.com/neutral/passkey-demo/internal/crypto"
    types "github.com/neutral/passkey-demo/internal/types"
)

// loginFinishInbound mirrors the shape sent by the browser for login finish.
// Binary fields arrive as base64url strings (unpadded or padded accepted).
type loginFinishInbound struct {
    LoginSessionID string `json:"login_session_id"`
    ID             string `json:"id"`
    RawID          string `json:"rawId"`
    Type           string `json:"type"`
    Response       struct {
        AuthenticatorData string `json:"authenticatorData"`
        ClientDataJSON    string `json:"clientDataJSON"`
        Signature         string `json:"signature"`
        UserHandle        string `json:"userHandle"`
    } `json:"response"`
}

type loginFinishResponse struct {
    AccountThumbHex string `json:"account_thumb_hex"`
    CredentialIDB64 string `json:"credential_id_b64"`
}

// LoginFinishHandler handles POST /authn/passkey/login/finish.
func LoginFinishHandler(cfg *cfgpkg.Config, store *LoginSessionStore, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var in loginFinishInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            http.Error(w, "bad json", http.StatusBadRequest)
            return
        }
        // Lookup and validate session (single-use on success/expiry)
        sess, ok := store.Get(in.LoginSessionID)
        if !ok || time.Now().After(sess.ExpiresAt) {
            if ok { store.Delete(in.LoginSessionID) }
            http.Error(w, "session expired or not found", http.StatusUnauthorized)
            return
        }

        // Decode clientDataJSON and validate ceremony type and challenge
        cdjRaw, err := b64.Decode(in.Response.ClientDataJSON)
        if err != nil {
            http.Error(w, "bad clientData", http.StatusBadRequest)
            return
        }
        cdj, err := ParseClientDataJSON(cdjRaw)
        if err != nil || !IsGet(cdj) {
            http.Error(w, "invalid clientData", http.StatusBadRequest)
            return
        }
        if string(cdj.Challenge) != string(sess.Challenge) {
            http.Error(w, "challenge mismatch", http.StatusUnauthorized)
            return
        }
        // Origin policy: allow dev localhost exception when configured origin is localhost
        devLocal := strings.HasPrefix(cfg.Origin, "http://localhost")
        if err := CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocal); err != nil {
            status, _ := MapPolicyError(err)
            http.Error(w, "origin not allowed", status)
            return
        }

        // Decode authenticatorData and signature
        adRaw, err := b64.Decode(in.Response.AuthenticatorData)
        if err != nil {
            http.Error(w, "bad authenticatorData", http.StatusBadRequest)
            return
        }
        sigRaw, err := b64.Decode(in.Response.Signature)
        if err != nil {
            http.Error(w, "bad signature", http.StatusBadRequest)
            return
        }
        ad, _, err := ParseAuthData(adRaw)
        if err != nil {
            http.Error(w, "bad authenticatorData", http.StatusBadRequest)
            return
        }
        // rpIdHash must match configured RP ID or allowlist
        if err := CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist); err != nil {
            status, _ := MapPolicyError(err)
            http.Error(w, "rpId not allowed", status)
            return
        }
        // Require User Verification (UV)
        if !HasUV(ad.Flags) {
            http.Error(w, "user verification required", http.StatusForbidden)
            return
        }

        // Identify account by credential ID
        credID, err := b64.Decode(in.RawID)
        if err != nil || len(credID) == 0 {
            http.Error(w, "bad credential id", http.StatusBadRequest)
            return
        }
        // Query credential -> account
        var acctCBOR []byte
        var storedCount int64
        row := db.QueryRow(`SELECT acct_cbor_fk, sign_count FROM credentials WHERE credential_id = ?`, credID)
        if err := row.Scan(&acctCBOR, &storedCount); err != nil {
            // 401 to avoid credential enumeration
            http.Error(w, "credential not recognized", http.StatusUnauthorized)
            return
        }
        // Load account COSE key and convert to ecdsa.PublicKey
        var cose types.CoseEC2
        if err := enc.DecodeCanonical(acctCBOR, &cose); err != nil {
            http.Error(w, "invalid account key", http.StatusBadRequest)
            return
        }
        pub, err := cryptoutil.ToECDSA(&cose)
        if err != nil {
            http.Error(w, "invalid public key", http.StatusBadRequest)
            return
        }

        // Verify assertion signature over ad || SHA256(cdj)
        if err := VerifyAssertion(pub, adRaw, cdjRaw, sigRaw); err != nil {
            status, kind := MapVerifyError(err)
            // Emit a structured debug log for triage (dev-friendly; no raw material)
            log.Printf("login_finish verify: kind=%s status=%d rp_id=%s origin=%s uv=%t up=%t sc=%d cred_hash=%s",
                kind, status, cfg.RP_ID, cfg.Origin, HasUV(ad.Flags), HasUP(ad.Flags), ad.SignCount, HashID(credID))
            http.Error(w, "assertion verification failed", status)
            return
        }

        // Enforce signCount policy
        // If the authenticator reports 0, treat as "counter not supported" and do not enforce monotonicity.
        // Otherwise, require strictly increasing vs stored.
        if ad.SignCount == 0 {
            // Leave stored count as-is
        } else {
            if int64(ad.SignCount) <= storedCount {
                http.Error(w, "signCount not increasing", http.StatusConflict)
                return
            }
            if _, err := db.Exec(`UPDATE credentials SET sign_count = ? WHERE credential_id = ?`, int64(ad.SignCount), credID); err != nil {
                http.Error(w, "db error", http.StatusInternalServerError)
                return
            }
        }

        // Create server session and set cookie
        sidRaw, err := randBytes(24)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        sid := b64.Encode(sidRaw)
        now := time.Now().Unix()
        exp := time.Now().Add(1 * time.Hour).Unix()
        if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)`, sid, acctCBOR, exp, now); err != nil {
            http.Error(w, "db error", http.StatusInternalServerError)
            return
        }
        setSessionCookie(w, cfg, sid)

        // Single-use: delete login session after success
        store.Delete(in.LoginSessionID)

        // Success response
        thumb := acctThumb(acctCBOR)
        resp := loginFinishResponse{AccountThumbHex: hex.EncodeToString(thumb), CredentialIDB64: b64.Encode(credID)}
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}

// setSessionCookie writes a session cookie with appropriate attributes.
func setSessionCookie(w http.ResponseWriter, cfg *cfgpkg.Config, sid string) {
    secure := false
    if u, err := url.Parse(cfg.Origin); err == nil {
        if strings.EqualFold(u.Scheme, "https") {
            secure = true
        }
    }
    c := &http.Cookie{
        Name:     "sid",
        Value:    sid,
        Path:     "/",
        HttpOnly: true,
        Secure:   secure,
        SameSite: http.SameSiteLaxMode,
    }
    http.SetCookie(w, c)
}
