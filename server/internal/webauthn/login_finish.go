package webauthn

import (
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "net/url"
    "strings"
    "time"

    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    cryptoutil "github.com/neutral/passkey-demo/internal/crypto"
    types "github.com/neutral/passkey-demo/internal/types"
    randutil "github.com/neutral/passkey-demo/internal/util/randutil"
    errx "github.com/neutral/passkey-demo/internal/httpx/errors"
    "log/slog"
    mid "github.com/neutral/passkey-demo/internal/httpx/middleware"
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
            errx.WriteReq(w, r, http.StatusMethodNotAllowed, errx.CodeMethodNotAllowed, "method not allowed")
            return
        }
        var in loginFinishInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        // Lookup and validate session (single-use on success/expiry)
        sess, ok := store.Get(in.LoginSessionID)
        if !ok || time.Now().After(sess.ExpiresAt) {
            if ok { store.Delete(in.LoginSessionID) }
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }

        // Decode clientDataJSON and validate ceremony type and challenge
        cdjRaw, err := b64.Decode(in.Response.ClientDataJSON)
        if err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        cdj, err := ParseClientDataJSON(cdjRaw)
        if err != nil || !IsGet(cdj) {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        if string(cdj.Challenge) != string(sess.Challenge) {
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }
        // Origin policy: allow dev localhost exception when configured origin is localhost
        devLocal := strings.HasPrefix(cfg.Origin, "http://localhost")
        if err := CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocal); err != nil {
            status, _ := MapPolicyError(err)
            code := errx.CodeForbidden
            if status == http.StatusBadRequest { code = errx.CodeBadRequest }
            errx.WriteReq(w, r, status, code, "policy violation")
            return
        }

        // Decode authenticatorData and signature
        adRaw, err := b64.Decode(in.Response.AuthenticatorData)
        if err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
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
            code := errx.CodeForbidden
            if status == http.StatusBadRequest { code = errx.CodeBadRequest }
            errx.WriteReq(w, r, status, code, "policy violation")
            return
        }
        // Require User Verification (UV)
        if !HasUV(ad.Flags) {
            errx.WriteReq(w, r, http.StatusForbidden, errx.CodeForbidden, "forbidden")
            return
        }

        // Identify account by credential ID
        credID, err := b64.Decode(in.RawID)
        if err != nil || len(credID) == 0 {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        // Query credential -> account
        var acctCBOR []byte
        var storedCount int64
        row := db.QueryRow(`SELECT acct_cbor_fk, sign_count FROM credentials WHERE credential_id = ?`, credID)
        if err := row.Scan(&acctCBOR, &storedCount); err != nil {
            // 401 to avoid credential enumeration
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }
        // Load account COSE key and convert to ecdsa.PublicKey
        var cose types.CoseEC2
        if err := enc.DecodeCanonical(acctCBOR, &cose); err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        pub, err := cryptoutil.ToECDSA(&cose)
        if err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }

        // Verify assertion signature over ad || SHA256(cdj)
        if err := VerifyAssertion(pub, adRaw, cdjRaw, sigRaw); err != nil {
            // If the only failure is high-S, try the relaxed verifier (normalize S)
            if !errors.Is(err, ErrHighS) || VerifyAssertionAllowHighS(pub, adRaw, cdjRaw, sigRaw) != nil {
                status, kind := MapVerifyError(err)
                // Emit a single structured verification log (no raw materials)
                v := VerifyLog{
                    Outcome:          "failure",
                    ErrorKind:        kind,
                    RP_ID:            cfg.RP_ID,
                    Origin:           cfg.Origin,
                    UV:               HasUV(ad.Flags),
                    UP:               HasUP(ad.Flags),
                    SignCount:        ad.SignCount,
                    CredentialIDHash: HashID(credID),
                }
                if reqID, ok := mid.FromContext(r.Context()); ok {
                    v.TraceID = reqID
                }
                LogAssertion(slog.Default(), v)
                code := errx.CodeUnauthorized
                if status == http.StatusBadRequest { code = errx.CodeBadRequest }
                errx.WriteReq(w, r, status, code, "verification failed")
                return
            }
            // else: accepted high-S after normalization; continue
        }

        // Enforce signCount policy
        // If the authenticator reports 0, treat as "counter not supported" and do not enforce monotonicity.
        // Otherwise, require strictly increasing vs stored.
        if ad.SignCount == 0 {
            // Leave stored count as-is
        } else {
            if int64(ad.SignCount) <= storedCount {
                errx.WriteReq(w, r, http.StatusConflict, errx.CodeConflict, "conflict")
                return
            }
            if _, err := db.Exec(`UPDATE credentials SET sign_count = ? WHERE credential_id = ?`, int64(ad.SignCount), credID); err != nil {
                errx.WriteReq(w, r, http.StatusInternalServerError, errx.CodeInternal, "internal error")
                return
            }
        }

        // Create server session and set cookie
        sidRaw, err := randutil.BytesE(24)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        sid := b64.Encode(sidRaw)
        now := time.Now().Unix()
        exp := time.Now().Add(1 * time.Hour).Unix()
        if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)`, sid, acctCBOR, exp, now); err != nil {
            errx.WriteReq(w, r, http.StatusInternalServerError, errx.CodeInternal, "internal error")
            return
        }
        setSessionCookie(w, cfg, sid)

        // Single-use: delete login session after success
        store.Delete(in.LoginSessionID)

        // Success log and response
        thumb := acctThumb(acctCBOR)
        // Structured success log (context handler injects correlation_id)
        slog.InfoContext(r.Context(), "login_finish",
            slog.String("account_thumb_hex", hex.EncodeToString(thumb)),
            slog.String("credential_id_hash", HashID(credID)),
            slog.Uint64("sign_count", uint64(ad.SignCount)),
        )
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
