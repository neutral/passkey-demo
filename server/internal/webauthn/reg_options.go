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

// RegSession holds server-side state for a pending registration.
type RegSession struct {
    Challenge []byte
    RP_ID     string
    Origin    string
    ExpiresAt time.Time
}

// RegSessionStore is a concurrency-safe in-memory store for registration sessions.
type RegSessionStore struct{ inner *ttl.Store[string, RegSession] }

func NewRegSessionStore(capacity int) *RegSessionStore {
    // TTL 5 minutes with background GC
    return &RegSessionStore{inner: ttl.New[string, RegSession](capacity, 5*time.Minute, true)}
}

func (s *RegSessionStore) Put(id string, v RegSession) error { return s.inner.Put(id, v) }
func (s *RegSessionStore) Get(id string) (RegSession, bool) { return s.inner.Get(id) }
func (s *RegSessionStore) Delete(id string)                 { s.inner.Delete(id) }

// randBytes returns n bytes using crypto/rand.

// RegistrationOptionsResponse defines the JSON returned by the options handler.
type RegistrationOptionsResponse struct {
    RegSessionID string          `json:"reg_session_id"`
    Challenge    string          `json:"challenge"`
    Options      types.RegOptions `json:"options"`
    ExpiresAt    int64           `json:"expires_at"`
}

// BuildRegistrationOptions creates a new registration session, stores it, and returns the response object.
func BuildRegistrationOptions(cfg *cfgpkg.Config, store *RegSessionStore, now func() time.Time) (RegistrationOptionsResponse, error) {
    // 24 bytes session id (>=128 bits)
    sidRaw, err := randutil.BytesE(24)
    if err != nil {
        return RegistrationOptionsResponse{}, err
    }
    sid := b64.Encode(sidRaw)
    // 32-byte challenge
    chRaw, err := randutil.BytesE(32)
    if err != nil {
        return RegistrationOptionsResponse{}, err
    }
    ch := b64.Encode(chRaw)
    exp := now().Add(5 * time.Minute)

    // Store server-side session
    if err := store.Put(sid, RegSession{
        Challenge: chRaw,
        RP_ID:     cfg.RP_ID,
        Origin:    cfg.Origin,
        ExpiresAt: exp,
    }); err != nil {
        return RegistrationOptionsResponse{}, err
    }

    return RegistrationOptionsResponse{
        RegSessionID: sid,
        Challenge:    ch,
        Options: types.RegOptions{
            RP_ID:       cfg.RP_ID,
            Origin:      cfg.Origin,
            UVRequired:  true,
            Attestation: "none",
        },
        ExpiresAt: exp.Unix(),
    }, nil
}

// RegistrationOptionsHandler serves POST /authn/passkey/registration/options.
func RegistrationOptionsHandler(cfg *cfgpkg.Config, store *RegSessionStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        resp, err := BuildRegistrationOptions(cfg, store, time.Now)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        enc := json.NewEncoder(w)
        _ = enc.Encode(resp)
    }
}
