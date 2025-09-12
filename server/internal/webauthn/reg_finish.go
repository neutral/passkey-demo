package webauthn

import (
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"

    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    cryptoutil "github.com/neutral/passkey-demo/internal/crypto"
    errx "github.com/neutral/passkey-demo/internal/httpx/errors"
)

type regFinishInbound struct {
    RegSessionID string `json:"reg_session_id"`
    ID           string `json:"id"`
    RawID        string `json:"rawId"`
    Type         string `json:"type"`
    Response     struct {
        AttestationObject string `json:"attestationObject"`
        ClientDataJSON    string `json:"clientDataJSON"`
    } `json:"response"`
}

type regFinishResponse struct {
    AccountThumbHex   string `json:"account_thumb_hex"`
    CredentialIDB64   string `json:"credential_id_b64"`
}

func acctThumb(acctCBOR []byte) []byte {
    h := sha256.Sum256(append([]byte("ACCTK1"), acctCBOR...))
    return h[:]
}

// RegistrationFinishHandler handles POST /authn/passkey/registration/finish.
func RegistrationFinishHandler(cfg *cfgpkg.Config, store *RegSessionStore, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            errx.WriteReq(w, r, http.StatusMethodNotAllowed, errx.CodeMethodNotAllowed, "method not allowed")
            return
        }
        var in regFinishInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        sess, ok := store.Get(in.RegSessionID)
        if !ok || time.Now().After(sess.ExpiresAt) {
            // single-use: remove if present
            if ok { store.Delete(in.RegSessionID) }
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }
        // Parse CDJ
        cdjRaw, err := b64.Decode(in.Response.ClientDataJSON)
        if err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        cdj, err := ParseClientDataJSON(cdjRaw)
        if err != nil || !IsCreate(cdj) {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        // Challenge match
        if string(cdj.Challenge) != string(sess.Challenge) {
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }
        // Origin check (allow dev localhost if configured origin is localhost)
        devLocal := strings.HasPrefix(cfg.Origin, "http://localhost")
        if err := CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocal); err != nil {
            status, _ := MapPolicyError(err)
            code := errx.CodeForbidden
            if status == http.StatusBadRequest { code = errx.CodeBadRequest }
            errx.WriteReq(w, r, status, code, "policy violation")
            return
        }
        // Parse attestation object
        attB, err := b64.Decode(in.Response.AttestationObject)
        if err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        ad, aaguid, credID, cose, err := ExtractRegistrationData(attB)
        if err != nil {
            // Map expected errors to 400
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        // RP ID hash
        if err := CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist); err != nil {
            status, _ := MapPolicyError(err)
            code := errx.CodeForbidden
            if status == http.StatusBadRequest { code = errx.CodeBadRequest }
            errx.WriteReq(w, r, status, code, "policy violation")
            return
        }
        // UV required
        if !HasUV(ad.Flags) {
            errx.WriteReq(w, r, http.StatusForbidden, errx.CodeForbidden, "forbidden")
            return
        }
        // Validate COSE EC2 key to Go ecdsa.PublicKey
        if _, err := cryptoutil.ToECDSA(&cose); err != nil {
            // Debug-only metadata to triage failures without printing raw key material
            log.Printf("reg_finish: ToECDSA failed: %v; cose{kty=%d alg=%d crv=%d xlen=%d ylen=%d}", err, cose.Kty, cose.Alg, cose.Crv, len(cose.X), len(cose.Y))
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad request")
            return
        }
        // Persist account and credential
        acctCBOR, err := enc.EncodeCanonical(cose)
        if err != nil {
            errx.WriteReq(w, r, http.StatusInternalServerError, errx.CodeInternal, "internal error")
            return
        }
        thumb := acctThumb(acctCBOR)
        now := time.Now().Unix()
        // Insert account (idempotent)
        if _, err := db.Exec(`INSERT OR IGNORE INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, thumb, now); err != nil {
            http.Error(w, "db error", http.StatusInternalServerError)
            return
        }
        // Insert credential
        _, err = db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, int64(ad.SignCount), aaguid[:], now)
        if err != nil {
            // naive unique detection
            if strings.Contains(strings.ToLower(err.Error()), "unique") {
                errx.WriteReq(w, r, http.StatusConflict, errx.CodeConflict, "conflict")
                return
            }
            errx.WriteReq(w, r, http.StatusInternalServerError, errx.CodeInternal, "internal error")
            return
        }
        // Success: delete session (single-use)
        store.Delete(in.RegSessionID)

        resp := regFinishResponse{AccountThumbHex: hex.EncodeToString(thumb), CredentialIDB64: b64.Encode(credID)}
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        _ = json.NewEncoder(w).Encode(resp)
    }
}
