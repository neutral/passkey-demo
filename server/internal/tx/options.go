package tx

import (
    "context"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    // no sync needed: store is backed by ttlstore
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    httpctx "github.com/neutral/passkey-demo/internal/http"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
    webauthn "github.com/neutral/passkey-demo/internal/webauthn"
    errx "github.com/neutral/passkey-demo/internal/httpx/errors"
    randutil "github.com/neutral/passkey-demo/internal/util/randutil"
    ttl "github.com/neutral/passkey-demo/internal/util/ttlstore"
    repos "github.com/neutral/passkey-demo/internal/repos"
)

// TxSession holds server-side state for a pending transaction signing flow.
type TxSession struct {
    B            []byte
    Challenge    []byte
    AcctCBOR     []byte
    CredentialIDs [][]byte
    ExpiresAt    time.Time
}

// TxSessionStore is a concurrency-safe in-memory store for tx sessions.
type TxSessionStore struct{ inner *ttl.Store[string, TxSession] }

// NewTxSessionStore constructs a new store with an optional capacity limit.
func NewTxSessionStore(capacity int) *TxSessionStore {
    // TTL 5 minutes with background GC
    return &TxSessionStore{inner: ttl.New[string, TxSession](capacity, 5*time.Minute, true)}
}

// Put inserts or updates a session by id.
func (s *TxSessionStore) Put(id string, v TxSession) error { return s.inner.Put(id, v) }

// Get returns a session by id.
func (s *TxSessionStore) Get(id string) (TxSession, bool) { return s.inner.Get(id) }

// Delete removes a session by id.
func (s *TxSessionStore) Delete(id string) { s.inner.Delete(id) }

// Inbound payload for /tx/signing/options.
type TxOptionsInbound struct {
    BundleCBOR string `json:"bundle_cbor_b64"`
}

// TxOptionsResponse defines the JSON returned by the options handler.
type TxOptionsResponse struct {
    TxSessionID string             `json:"tx_session_id"`
    Challenge   string             `json:"challenge"`
    Options     types.LoginOptions `json:"options"`
    TxIDHex     string             `json:"tx_id_hex"`
    ExpiresAt   int64              `json:"expires_at"`
}

// ErrNoCredentials indicates the account has no credentials in the DB.
var ErrNoCredentials = errors.New("no credentials for account")


// BuildTxOptions validates the bundle, derives anchors, collects credential ids, stores a tx session, and returns response JSON fields.
func BuildTxOptions(ctx context.Context, cfg *cfgpkg.Config, store *TxSessionStore, db *sql.DB, acctCBOR []byte, bundleB64 string, now func() time.Time) (TxOptionsResponse, error) {
    // Validate and derive anchors from bundle
    anchored, err := ValidateAndAnchorBundle(ctx, db, acctCBOR, bundleB64)
    if err != nil {
        return TxOptionsResponse{}, err
    }
    // Collect credentials for this account
    rows, err := db.QueryContext(ctx, `SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?`, acctCBOR)
    if err != nil {
        return TxOptionsResponse{}, err
    }
    defer rows.Close()
    var credIDs [][]byte
    for rows.Next() {
        var id []byte
        if err := rows.Scan(&id); err != nil {
            return TxOptionsResponse{}, err
        }
        credIDs = append(credIDs, append([]byte(nil), id...))
    }
    if err := rows.Err(); err != nil {
        return TxOptionsResponse{}, err
    }
    if len(credIDs) == 0 {
        return TxOptionsResponse{}, ErrNoCredentials
    }
    // Session id and expiry
    sidRaw, err := randutil.BytesE(24)
    if err != nil {
        return TxOptionsResponse{}, err
    }
    sid := b64.Encode(sidRaw)
    exp := now().Add(5 * time.Minute)
    // Store tx session
    if err := store.Put(sid, TxSession{
        B:             append([]byte(nil), anchored.B...),
        Challenge:     append([]byte(nil), anchored.Challenge[:]...),
        AcctCBOR:      append([]byte(nil), acctCBOR...),
        CredentialIDs: credIDs,
        ExpiresAt:     exp,
    }); err != nil {
        return TxOptionsResponse{}, err
    }
    // Build allowCredentials (base64url ids)
    allow := make([]string, 0, len(credIDs))
    for _, id := range credIDs {
        allow = append(allow, b64.Encode(id))
    }
    // Response
    return TxOptionsResponse{
        TxSessionID: sid,
        Challenge:   b64.Encode(anchored.Challenge[:]),
        Options: types.LoginOptions{
            RP_ID:           cfg.RP_ID,
            Origin:          cfg.Origin,
            UVRequired:      true,
            AllowCredentials: allow,
        },
        TxIDHex:   hex.EncodeToString(anchored.TxID[:]),
        ExpiresAt: exp.Unix(),
    }, nil
}

// BuildTxOptionsWithRepo is like BuildTxOptions but uses a credentials repo
// to list credential ids for the account.
func BuildTxOptionsWithRepo(ctx context.Context, cfg *cfgpkg.Config, store *TxSessionStore, creds *repos.CredentialsRepo, db *sql.DB, acctCBOR []byte, bundleB64 string, now func() time.Time) (TxOptionsResponse, error) {
    anchored, err := ValidateAndAnchorBundle(ctx, db, acctCBOR, bundleB64)
    if err != nil { return TxOptionsResponse{}, err }
    credIDs, err := creds.ListIDsByAccount(ctx, acctCBOR)
    if err != nil { return TxOptionsResponse{}, err }
    if len(credIDs) == 0 { return TxOptionsResponse{}, ErrNoCredentials }
    sidRaw, err := randutil.BytesE(24)
    if err != nil { return TxOptionsResponse{}, err }
    sid := b64.Encode(sidRaw)
    exp := now().Add(5 * time.Minute)
    if err := store.Put(sid, TxSession{B: append([]byte(nil), anchored.B...), Challenge: append([]byte(nil), anchored.Challenge[:]...), AcctCBOR: append([]byte(nil), acctCBOR...), CredentialIDs: credIDs, ExpiresAt: exp}); err != nil { return TxOptionsResponse{}, err }
    allow := make([]string, 0, len(credIDs))
    for _, id := range credIDs { allow = append(allow, b64.Encode(id)) }
    return TxOptionsResponse{TxSessionID: sid, Challenge: b64.Encode(anchored.Challenge[:]), Options: types.LoginOptions{RP_ID: cfg.RP_ID, Origin: cfg.Origin, UVRequired: true, AllowCredentials: allow}, TxIDHex: hex.EncodeToString(anchored.TxID[:]), ExpiresAt: exp.Unix()}, nil
}

// TxOptionsHandler handles POST /tx/signing/options.
// Requires session middleware to populate account context; no cookie fallback.
func TxOptionsHandler(cfg *cfgpkg.Config, txStore *TxSessionStore, creds *repos.CredentialsRepo, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            errx.WriteReq(w, r, http.StatusMethodNotAllowed, errx.CodeMethodNotAllowed, "method not allowed")
            return
        }
        // Resolve session strictly from middleware context.
        s, ok := httpctx.FromSession(r.Context())
        if !ok {
            errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
            return
        }
        acctCBOR := s.AcctCBOR
        // Parse inbound JSON
        var in TxOptionsInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "bad json")
            return
        }
        // Build response via helper (repo-backed)
        resp, err := BuildTxOptionsWithRepo(r.Context(), cfg, txStore, creds, db, acctCBOR, in.BundleCBOR, time.Now)
        if err != nil {
            // Map known errors to appropriate statuses
            switch {
            case errors.Is(err, ErrBundleBase64), errors.Is(err, ErrBundleCBOR):
                log.Printf("tx_options: invalid bundle acct_hash=%s", webauthn.HashID(acctCBOR))
                errx.WriteReq(w, r, http.StatusBadRequest, errx.CodeBadRequest, "invalid bundle")
                return
            case errors.Is(err, ErrSenderKeyMismatch):
                log.Printf("tx_options: sender_key mismatch acct_hash=%s", webauthn.HashID(acctCBOR))
                errx.WriteReq(w, r, http.StatusUnauthorized, errx.CodeUnauthorized, "unauthorized")
                return
            case errors.Is(err, ErrNonceNotMonotonic), errors.Is(err, ErrNoCredentials):
                if errors.Is(err, ErrNonceNotMonotonic) {
                    log.Printf("tx_options: conflict (nonce not monotonic) acct_hash=%s", webauthn.HashID(acctCBOR))
                } else {
                    log.Printf("tx_options: conflict (no credentials) acct_hash=%s", webauthn.HashID(acctCBOR))
                }
                errx.WriteReq(w, r, http.StatusConflict, errx.CodeConflict, "conflict")
                return
            default:
                log.Printf("tx_options: internal error acct_hash=%s err=%v", webauthn.HashID(acctCBOR), err)
                errx.WriteReq(w, r, http.StatusInternalServerError, errx.CodeInternal, "internal error")
                return
            }
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}
