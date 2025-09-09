package me

import (
    "encoding/json"
    "net/http"

    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    httpctx "github.com/neutral/passkey-demo/internal/http"
    "database/sql"
)

type accountKeyResponse struct {
    AcctCBORB64 string `json:"acct_cbor_b64"`
    SenderKey   struct {
        Kty int    `json:"kty"`
        Alg int    `json:"alg"`
        Crv int    `json:"crv"`
        X   string `json:"x"`
        Y   string `json:"y"`
    } `json:"sender_key"`
}

// AccountKeyHandler serves GET /me/account_key (authenticated).
// It returns the logged-in account's COSE EC2 public key as JSON with base64url x/y.
func AccountKeyHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Resolve account from session context (preferred) or via sid cookie lookup
        var acctCBOR []byte
        if s, ok := httpctx.FromSession(r.Context()); ok {
            acctCBOR = s.AcctCBOR
        } else {
            c, err := r.Cookie("sid")
            if err != nil || c.Value == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            var exp int64
            row := db.QueryRow(`SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`, c.Value)
            if err := row.Scan(&acctCBOR, &exp); err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
        }

        // Decode acct_cbor to COSE EC2 and return as JSON
        var cose struct {
            Kty int    `cbor:"1" json:"kty"`
            Alg int    `cbor:"3" json:"alg"`
            Crv int    `cbor:"-1" json:"crv"`
            X   []byte `cbor:"-2" json:"x"`
            Y   []byte `cbor:"-3" json:"y"`
        }
        if err := enc.DecodeCanonical(acctCBOR, &cose); err != nil {
            http.Error(w, "invalid account key", http.StatusInternalServerError)
            return
        }
        var resp accountKeyResponse
        resp.AcctCBORB64 = b64.Encode(acctCBOR)
        resp.SenderKey.Kty = cose.Kty
        resp.SenderKey.Alg = cose.Alg
        resp.SenderKey.Crv = cose.Crv
        resp.SenderKey.X = b64.Encode(cose.X)
        resp.SenderKey.Y = b64.Encode(cose.Y)
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}

