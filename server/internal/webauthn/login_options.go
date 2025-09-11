package webauthn

import (
    "encoding/json"
    "net/http"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
    randutil "github.com/neutral/passkey-demo/internal/util/randutil"
    ttl "github.com/neutral/passkey-demo/internal/util/ttlstore"
)

// LoginSession holds server-side state for a pending login/assertion.
type LoginSession struct {
    Challenge []byte
    RP_ID     string
    Origin    string
    ExpiresAt time.Time
}

// LoginSessionStore is a concurrency-safe in-memory store for login sessions.
type LoginSessionStore struct{ inner *ttl.Store[string, LoginSession] }

func NewLoginSessionStore(capacity int) *LoginSessionStore {
    // TTL 5 minutes with background GC
    return &LoginSessionStore{inner: ttl.New[string, LoginSession](capacity, 5*time.Minute, true)}
}

func (s *LoginSessionStore) Put(id string, v LoginSession) error { return s.inner.Put(id, v) }
func (s *LoginSessionStore) Get(id string) (LoginSession, bool) { return s.inner.Get(id) }
func (s *LoginSessionStore) Delete(id string)                 { s.inner.Delete(id) }

// LoginOptionsResponse defines the JSON returned by the login options handler.
type LoginOptionsResponse struct {
    LoginSessionID string            `json:"login_session_id"`
    Challenge      string            `json:"challenge"`
    Options        types.LoginOptions `json:"options"`
    ExpiresAt      int64             `json:"expires_at"`
}

// BuildLoginOptions creates a new login session, stores it, and returns the response object.
func BuildLoginOptions(cfg *cfgpkg.Config, store *LoginSessionStore, now func() time.Time) (LoginOptionsResponse, error) {
    // 24 bytes session id (>=128 bits)
    sidRaw, err := randutil.BytesE(24)
    if err != nil {
        return LoginOptionsResponse{}, err
    }
    sid := b64.Encode(sidRaw)
    // 32-byte challenge
    chRaw, err := randutil.BytesE(32)
    if err != nil {
        return LoginOptionsResponse{}, err
    }
    ch := b64.Encode(chRaw)
    exp := now().Add(5 * time.Minute)

    // Store server-side session
    if err := store.Put(sid, LoginSession{
        Challenge: chRaw,
        RP_ID:     cfg.RP_ID,
        Origin:    cfg.Origin,
        ExpiresAt: exp,
    }); err != nil {
        return LoginOptionsResponse{}, err
    }

    return LoginOptionsResponse{
        LoginSessionID: sid,
        Challenge:      ch,
        Options: types.LoginOptions{
            RP_ID:           cfg.RP_ID,
            Origin:          cfg.Origin,
            UVRequired:      true,
            AllowCredentials: []string{},
        },
        ExpiresAt: exp.Unix(),
    }, nil
}

// LoginOptionsHandler serves POST /authn/passkey/login/options.
func LoginOptionsHandler(cfg *cfgpkg.Config, store *LoginSessionStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        resp, err := BuildLoginOptions(cfg, store, time.Now)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}
