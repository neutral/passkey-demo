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
    // Verify assertion signature
    if err := webauthn.VerifyAssertion(pub, adRaw, cdjRaw, sigRaw); err != nil {
        return TxFinishResponse{}, err
    }
    // Enforce strictly increasing signCount
    if int64(ad.SignCount) <= storedCount {
        return TxFinishResponse{}, ErrSignCount
    }
    if _, err := db.ExecContext(ctx, `UPDATE credentials SET sign_count = ? WHERE credential_id = ?`, int64(ad.SignCount), credID); err != nil {
        return TxFinishResponse{}, err
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

// TxFinishHandler handles POST /tx/signing/finish.
func TxFinishHandler(cfg *cfgpkg.Config, txStore *TxSessionStore, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        c, err := r.Cookie("sid")
        if err != nil || c.Value == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        var in TxFinishInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            http.Error(w, "bad json", http.StatusBadRequest)
            return
        }
        resp, err := BuildTxFinish(r.Context(), cfg, txStore, db, c.Value, in, time.Now)
        if err != nil {
            switch {
            case errors.Is(err, ErrAuthSession), errors.Is(err, ErrTxSession), errors.Is(err, ErrCredUnknown), errors.Is(err, ErrCredMismatch), errors.Is(err, ErrAllowlist):
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            case errors.Is(err, ErrBadJSON), errors.Is(err, ErrBadBase64), errors.Is(err, ErrTypeMismatch):
                http.Error(w, "bad request", http.StatusBadRequest)
                return
            case errors.Is(err, ErrChallenge):
                http.Error(w, "challenge mismatch", http.StatusUnauthorized)
                return
            case errors.Is(err, ErrSignCount):
                http.Error(w, "conflict", http.StatusConflict)
                return
            default:
                // Map policy and verify errors if applicable
                if status, _ := webauthn.MapPolicyError(err); status != http.StatusOK && status != http.StatusInternalServerError {
                    http.Error(w, "forbidden", status)
                    return
                }
                if status, _ := webauthn.MapVerifyError(err); status != http.StatusOK && status != http.StatusInternalServerError {
                    http.Error(w, "verification failed", status)
                    return
                }
                http.Error(w, "internal error", http.StatusInternalServerError)
                return
            }
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}

