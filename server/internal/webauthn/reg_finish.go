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
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var in regFinishInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            http.Error(w, "bad json", http.StatusBadRequest)
            return
        }
        sess, ok := store.Get(in.RegSessionID)
        if !ok || time.Now().After(sess.ExpiresAt) {
            // single-use: remove if present
            if ok { store.Delete(in.RegSessionID) }
            http.Error(w, "session expired or not found", http.StatusUnauthorized)
            return
        }
        // Parse CDJ
        cdjRaw, err := b64.Decode(in.Response.ClientDataJSON)
        if err != nil {
            http.Error(w, "bad clientData", http.StatusBadRequest)
            return
        }
        cdj, err := ParseClientDataJSON(cdjRaw)
        if err != nil || !IsCreate(cdj) {
            http.Error(w, "invalid clientData", http.StatusBadRequest)
            return
        }
        // Challenge match
        if string(cdj.Challenge) != string(sess.Challenge) {
            http.Error(w, "challenge mismatch", http.StatusUnauthorized)
            return
        }
        // Origin check (allow dev localhost if configured origin is localhost)
        devLocal := strings.HasPrefix(cfg.Origin, "http://localhost")
        if err := CheckOrigin(cdj.Origin, cfg.Origin, cfg.OriginAllowlist, devLocal); err != nil {
            status, _ := MapPolicyError(err)
            http.Error(w, "origin not allowed", status)
            return
        }
        // Parse attestation object
        attB, err := b64.Decode(in.Response.AttestationObject)
        if err != nil {
            http.Error(w, "bad attestationObject", http.StatusBadRequest)
            return
        }
        ad, aaguid, credID, cose, err := ExtractRegistrationData(attB)
        if err != nil {
            // Map expected errors to 400
            http.Error(w, "invalid attestation", http.StatusBadRequest)
            return
        }
        // RP ID hash
        if err := CheckRpIdHashAllowed(ad.RpIDHash, cfg.RP_ID, cfg.RPAllowlist); err != nil {
            status, _ := MapPolicyError(err)
            http.Error(w, "rpId not allowed", status)
            return
        }
        // UV required
        if !HasUV(ad.Flags) {
            http.Error(w, "user verification required", http.StatusForbidden)
            return
        }
        // Validate COSE EC2 key to Go ecdsa.PublicKey
        if _, err := cryptoutil.ToECDSA(&cose); err != nil {
            // Debug-only metadata to triage failures without printing raw key material
            log.Printf("reg_finish: ToECDSA failed: %v; cose{kty=%d alg=%d crv=%d xlen=%d ylen=%d}", err, cose.Kty, cose.Alg, cose.Crv, len(cose.X), len(cose.Y))
            http.Error(w, "invalid public key", http.StatusBadRequest)
            return
        }
        // Persist account and credential
        acctCBOR, err := enc.EncodeCanonical(cose)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
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
                http.Error(w, "credential exists", http.StatusConflict)
                return
            }
            http.Error(w, "db error", http.StatusInternalServerError)
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
