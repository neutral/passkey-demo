package tx

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "strings"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    cryptoutil "github.com/neutral/passkey-demo/internal/crypto"
    types "github.com/neutral/passkey-demo/internal/types"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
    httpctx "github.com/neutral/passkey-demo/internal/http"
    repos "github.com/neutral/passkey-demo/internal/repos"
    errx "github.com/neutral/passkey-demo/internal/httpx/errors"
)

// Inbound payload for /tx/signing/finish.
type TxFinishInbound struct {
    TxSessionID string `json:"tx_session_id"`
    ID          string `json:"id"`
    RawID       string `json:"rawId"`
    Type        string `json:"type"`
    Response    struct {
        AuthenticatorData string `json:"authenticatorData"`
        ClientDataJSON    string `json:"clientDataJSON"`
        Signature         string `json:"signature"`
        UserHandle        string `json:"userHandle"`
    } `json:"response"`
}

// Response payload for /tx/signing/finish.
type TxFinishResponse struct {
    TxIDHex string `json:"tx_id_hex"`
    Stored  bool   `json:"stored"`
}

// Sentinel errors for handler mapping.
var (
    ErrAuthSession   = errors.New("auth session not found or expired")
    ErrTxSession     = errors.New("tx session not found or expired")
    ErrBadJSON       = errors.New("bad json")
    ErrBadBase64     = errors.New("bad base64")
    ErrTypeMismatch  = errors.New("clientDataJSON type mismatch")
    ErrChallenge     = errors.New("challenge mismatch")
    ErrCredUnknown   = errors.New("credential not recognized")
    ErrCredMismatch  = errors.New("credential/account mismatch")
    ErrAllowlist     = errors.New("credential not in allowlist")
    ErrSignCount     = errors.New("signCount not increasing")
)

// BuildTxFinish performs validation and persistence for a signing finish request.
func BuildTxFinish(ctx context.Context, cfg *cfgpkg.Config, txStore *TxSessionStore, db *sql.DB, sid string, in TxFinishInbound, now func() time.Time) (TxFinishResponse, error) {
    // Auth session lookup
    var acctCBOR []byte
    var exp int64
    if err := db.QueryRowContext(ctx, `SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`, sid).Scan(&acctCBOR, &exp); err != nil {
        return TxFinishResponse{}, ErrAuthSession
    }
    if now().Unix() >= exp {
        return TxFinishResponse{}, ErrAuthSession
    }
    // Tx session lookup
    txSess, ok := txStore.Get(in.TxSessionID)
    if !ok || now().After(txSess.ExpiresAt) {
        return TxFinishResponse{}, ErrTxSession
    }
    // Parse CDJ
    cdjRaw, err := b64.Decode(in.Response.ClientDataJSON)
    if err != nil {
        return TxFinishResponse{}, ErrBadBase64
    }
    cdj, err := webauthn.ParseClientDataJSON(cdjRaw)
    if err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    if !webauthn.IsGet(cdj) {
        return TxFinishResponse{}, ErrTypeMismatch
    }
    if string(cdj.Challenge) != string(txSess.Challenge) {
        return TxFinishResponse{}, ErrChallenge
    }
    // Origin policy
    devLocal := strings.HasPrefix(cfg.Origin, "http://localhost")
    if err := webauthn.CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocal); err != nil {
        return TxFinishResponse{}, err
    }
    // Parse AD and signature
    adRaw, err := b64.Decode(in.Response.AuthenticatorData)
    if err != nil {
        return TxFinishResponse{}, ErrBadBase64
    }
    sigRaw, err := b64.Decode(in.Response.Signature)
    if err != nil {
        return TxFinishResponse{}, ErrBadBase64
    }
    ad, _, err := webauthn.ParseAuthData(adRaw)
    if err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    // rpId policy
    if err := webauthn.CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist); err != nil {
        return TxFinishResponse{}, err
    }
    // UV required
    if !webauthn.HasUV(ad.Flags) {
        return TxFinishResponse{}, webauthn.ErrOriginNotAllowed // map to 403 via policy mapper
    }
    // Credential id
    credID, err := b64.Decode(in.RawID)
    if err != nil || len(credID) == 0 {
        return TxFinishResponse{}, ErrBadBase64
    }
    if in.ID != in.RawID { // mismatch between id and rawId
        return TxFinishResponse{}, ErrBadJSON
    }
    // Load credential
    var storedAcct []byte
    var storedCount int64
    if err := db.QueryRowContext(ctx, `SELECT acct_cbor_fk, sign_count FROM credentials WHERE credential_id = ?`, credID).Scan(&storedAcct, &storedCount); err != nil {
        return TxFinishResponse{}, ErrCredUnknown
    }
    if !bytes.Equal(storedAcct, txSess.AcctCBOR) || !bytes.Equal(storedAcct, acctCBOR) {
        return TxFinishResponse{}, ErrCredMismatch
    }
    // Must be in allowlist from options
    allowed := false
    for _, id := range txSess.CredentialIDs {
        if bytes.Equal(id, credID) { allowed = true; break }
    }
    if !allowed {
        return TxFinishResponse{}, ErrAllowlist
    }
    // Account public key
    var cose types.CoseEC2
    if err := enc.DecodeCanonical(txSess.AcctCBOR, &cose); err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    pub, err := cryptoutil.ToECDSA(&cose)
    if err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    // Verify assertion signature with strict low-S; accept high-S by normalization for compatibility
    if err := webauthn.VerifyAssertion(pub, adRaw, cdjRaw, sigRaw); err != nil {
        if !errors.Is(err, webauthn.ErrHighS) || webauthn.VerifyAssertionAllowHighS(pub, adRaw, cdjRaw, sigRaw) != nil {
            return TxFinishResponse{}, err
        }
        // accepted high-S after normalization; continue
    }
    // Enforce signCount policy
    // If the authenticator reports 0, treat as counter-not-supported and do not enforce monotonicity or update stored count.
    // Otherwise, require strictly increasing vs stored.
    if ad.SignCount == 0 {
        // leave stored count unchanged
    } else {
        if int64(ad.SignCount) <= storedCount {
            return TxFinishResponse{}, ErrSignCount
        }
        if _, err := db.ExecContext(ctx, `UPDATE credentials SET sign_count = ? WHERE credential_id = ?`, int64(ad.SignCount), credID); err != nil {
            return TxFinishResponse{}, err
        }
    }
    // Compute tx_id from canonical B and persist transaction
    // Decode bundle to extract nonce/message
    var bun types.Bundle
    if err := enc.DecodeCanonical(txSess.B, &bun); err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    txID := sha256.Sum256(append([]byte(anchorTxIDPrefix), txSess.B...))
    nowUnix := now().Unix()
    if _, err := db.ExecContext(ctx, `INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        txID[:], txSess.AcctCBOR, int64(bun.Nonce), bun.Message, txSess.B, adRaw, cdjRaw, sigRaw, nowUnix,
    ); err != nil {
        return TxFinishResponse{}, err
    }
    // Single-use: delete tx session
    txStore.Delete(in.TxSessionID)

    return TxFinishResponse{TxIDHex: hex.EncodeToString(txID[:]), Stored: true}, nil
}

// BuildTxFinishWithAcct is like BuildTxFinish but uses the provided acctCBOR
// from session middleware and does not consult the sessions table.
func BuildTxFinishWithAcct(ctx context.Context, cfg *cfgpkg.Config, txStore *TxSessionStore, db *sql.DB, creds *repos.CredentialsRepo, acctCBOR []byte, in TxFinishInbound, now func() time.Time) (TxFinishResponse, error) {
    // Tx session lookup
    txSess, ok := txStore.Get(in.TxSessionID)
    if !ok || now().After(txSess.ExpiresAt) {
        return TxFinishResponse{}, ErrTxSession
    }
    // Parse CDJ
    cdjRaw, err := b64.Decode(in.Response.ClientDataJSON)
    if err != nil {
        return TxFinishResponse{}, ErrBadBase64
    }
    cdj, err := webauthn.ParseClientDataJSON(cdjRaw)
    if err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    if !webauthn.IsGet(cdj) {
        return TxFinishResponse{}, ErrTypeMismatch
    }
    if string(cdj.Challenge) != string(txSess.Challenge) {
        return TxFinishResponse{}, ErrChallenge
    }
    // Origin policy
    devLocal := strings.HasPrefix(cfg.Origin, "http://localhost")
    if err := webauthn.CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocal); err != nil {
        return TxFinishResponse{}, err
    }
    // Parse AD and signature
    adRaw, err := b64.Decode(in.Response.AuthenticatorData)
    if err != nil {
        return TxFinishResponse{}, ErrBadBase64
    }
    sigRaw, err := b64.Decode(in.Response.Signature)
    if err != nil {
        return TxFinishResponse{}, ErrBadBase64
    }
    ad, _, err := webauthn.ParseAuthData(adRaw)
    if err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    // rpId policy
    if err := webauthn.CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist); err != nil {
        return TxFinishResponse{}, err
    }
    // UV required
    if !webauthn.HasUV(ad.Flags) {
        return TxFinishResponse{}, webauthn.ErrOriginNotAllowed // map to 403 via policy mapper
    }
    // Credential id
    credID, err := b64.Decode(in.RawID)
    if err != nil || len(credID) == 0 {
        return TxFinishResponse{}, ErrBadBase64
    }
    if in.ID != in.RawID { // mismatch between id and rawId
        return TxFinishResponse{}, ErrBadJSON
    }
    // Load credential
    storedAcct, storedCount, err := creds.GetAccountAndCount(ctx, credID)
    if err != nil {
        return TxFinishResponse{}, ErrCredUnknown
    }
    if !bytes.Equal(storedAcct, txSess.AcctCBOR) || !bytes.Equal(storedAcct, acctCBOR) {
        return TxFinishResponse{}, ErrCredMismatch
    }
    // Must be in allowlist from options
    allowed := false
    for _, id := range txSess.CredentialIDs {
        if bytes.Equal(id, credID) { allowed = true; break }
    }
    if !allowed {
        return TxFinishResponse{}, ErrAllowlist
    }
    // Account public key
    var cose types.CoseEC2
    if err := enc.DecodeCanonical(txSess.AcctCBOR, &cose); err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    pub, err := cryptoutil.ToECDSA(&cose)
    if err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    // Verify assertion signature with strict low-S; accept high-S by normalization for compatibility
    if err := webauthn.VerifyAssertion(pub, adRaw, cdjRaw, sigRaw); err != nil {
        if !errors.Is(err, webauthn.ErrHighS) || webauthn.VerifyAssertionAllowHighS(pub, adRaw, cdjRaw, sigRaw) != nil {
            return TxFinishResponse{}, err
        }
        // accepted high-S after normalization; continue
    }
    // Enforce signCount policy
    if ad.SignCount == 0 {
        // leave stored count unchanged
    } else {
        if int64(ad.SignCount) <= storedCount {
            return TxFinishResponse{}, ErrSignCount
        }
        if _, err := db.ExecContext(ctx, `UPDATE credentials SET sign_count = ? WHERE credential_id = ?`, int64(ad.SignCount), credID); err != nil {
            return TxFinishResponse{}, err
        }
    }
    // Compute tx_id from canonical B and persist transaction
    var bun types.Bundle
    if err := enc.DecodeCanonical(txSess.B, &bun); err != nil {
        return TxFinishResponse{}, ErrBadJSON
    }
    txID := sha256.Sum256(append([]byte(anchorTxIDPrefix), txSess.B...))
    nowUnix := now().Unix()
    if _, err := db.ExecContext(ctx, `INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        txID[:], txSess.AcctCBOR, int64(bun.Nonce), bun.Message, txSess.B, adRaw, cdjRaw, sigRaw, nowUnix,
    ); err != nil {
        return TxFinishResponse{}, err
    }
    txStore.Delete(in.TxSessionID)
    return TxFinishResponse{TxIDHex: hex.EncodeToString(txID[:]), Stored: true}, nil
}

// TxFinishHandler handles POST /tx/signing/finish.
func TxFinishHandler(cfg *cfgpkg.Config, txStore *TxSessionStore, db *sql.DB, creds *repos.CredentialsRepo) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            errx.WriteReq(w, r, http.StatusMethodNotAllowed, errx.CodeMethodNotAllowed, "method not allowed")
            return
        }
        s, ok := httpctx.FromSession(r.Context())
        if !ok {
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }
        var in TxFinishInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        resp, err := BuildTxFinishWithAcct(r.Context(), cfg, txStore, db, creds, s.AcctCBOR, in, time.Now)
        if err != nil {
            switch {
            case errors.Is(err, ErrAuthSession):
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrTxSession):
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrCredUnknown):
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrCredMismatch):
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrAllowlist):
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrBadJSON), errors.Is(err, ErrBadBase64), errors.Is(err, ErrTypeMismatch):
                errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
                return
            case errors.Is(err, ErrChallenge):
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrSignCount):
                errx.WriteReq(w, r, http.StatusConflict, errx.CodeConflict, "conflict")
                return
            default:
                // Map policy and verify errors if applicable
                if status, _ := webauthn.MapPolicyError(err); status != http.StatusOK && status != http.StatusInternalServerError {
                    // Map policy errors to 403 (or 400 when returned by mapper) using envelope
                    code := errx.CodeForbidden
                    if status == http.StatusBadRequest { code = errx.CodeBadRequest }
                    errx.WriteReq(w, r, status, code, "policy violation")
                    return
                }
                if status, _ := webauthn.MapVerifyError(err); status != http.StatusOK && status != http.StatusInternalServerError {
                    // Map verify errors to 400 or 401 depending on mapper
                    code := errx.CodeUnauthorized
                    if status == http.StatusBadRequest { code = errx.CodeBadRequest }
                    errx.WriteReq(w, r, status, code, "verification failed")
                    return
                }
                errx.WriteReq(w, r, http.StatusInternalServerError, errx.CodeInternal, "internal error")
                return
            }
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}
