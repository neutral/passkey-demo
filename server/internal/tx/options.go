package tx

import (
    "context"
    "crypto/rand"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "sync"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    httpctx "github.com/neutral/passkey-demo/internal/http"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
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
type TxSessionStore struct {
    mu       sync.Mutex
    items    map[string]TxSession
    capacity int // 0 = unlimited
}

// NewTxSessionStore constructs a new store with an optional capacity limit.
func NewTxSessionStore(capacity int) *TxSessionStore {
    return &TxSessionStore{items: make(map[string]TxSession), capacity: capacity}
}

// Put inserts or updates a session by id.
func (s *TxSessionStore) Put(id string, v TxSession) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.capacity > 0 && len(s.items) >= s.capacity {
        return errors.New("tx session store at capacity")
    }
    s.items[id] = v
    return nil
}

// Get returns a session by id.
func (s *TxSessionStore) Get(id string) (TxSession, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    v, ok := s.items[id]
    return v, ok
}

// Delete removes a session by id.
func (s *TxSessionStore) Delete(id string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.items, id)
}

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

// randBytes returns n cryptographically secure random bytes.
func randBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil {
        return nil, err
    }
    return b, nil
}

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
    sidRaw, err := randBytes(24)
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

// TxOptionsHandler handles POST /tx/signing/options.
func TxOptionsHandler(cfg *cfgpkg.Config, txStore *TxSessionStore, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        // Resolve session from middleware context (preferred), else fallback to cookie lookup.
        var acctCBOR []byte
        if s, ok := httpctx.FromSession(r.Context()); ok {
            acctCBOR = s.AcctCBOR
        } else {
            c, err := r.Cookie("sid")
            if err != nil || c.Value == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            var expSec int64
            row := db.QueryRow(`SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`, c.Value)
            if err := row.Scan(&acctCBOR, &expSec); err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            if time.Now().Unix() >= expSec {
                http.Error(w, "session expired", http.StatusUnauthorized)
                return
            }
        }
        // Parse inbound JSON
        var in TxOptionsInbound
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            http.Error(w, "bad json", http.StatusBadRequest)
            return
        }
        // Build response via helper
        resp, err := BuildTxOptions(r.Context(), cfg, txStore, db, acctCBOR, in.BundleCBOR, time.Now)
        if err != nil {
            // Map known errors to appropriate statuses
            switch {
            case errors.Is(err, ErrBundleBase64), errors.Is(err, ErrBundleCBOR):
                http.Error(w, "invalid bundle", http.StatusBadRequest)
                return
            case errors.Is(err, ErrSenderKeyMismatch):
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            case errors.Is(err, ErrNonceNotMonotonic), errors.Is(err, ErrNoCredentials):
                http.Error(w, "conflict", http.StatusConflict)
                return
            default:
                http.Error(w, "internal error", http.StatusInternalServerError)
                return
            }
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}
