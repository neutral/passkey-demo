package webauthn

import (
    "crypto/rand"
    "encoding/json"
    "errors"
    "net/http"
    "sync"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
)

// RegSession holds server-side state for a pending registration.
type RegSession struct {
    Challenge []byte
    RP_ID     string
    Origin    string
    ExpiresAt time.Time
}

// RegSessionStore is a concurrency-safe in-memory store for registration sessions.
type RegSessionStore struct {
    mu       sync.Mutex
    items    map[string]RegSession
    capacity int // 0 = unlimited
}

func NewRegSessionStore(capacity int) *RegSessionStore {
    return &RegSessionStore{items: make(map[string]RegSession), capacity: capacity}
}

func (s *RegSessionStore) Put(id string, v RegSession) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.capacity > 0 && len(s.items) >= s.capacity {
        return errors.New("registration session store at capacity")
    }
    s.items[id] = v
    return nil
}

func (s *RegSessionStore) Get(id string) (RegSession, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    v, ok := s.items[id]
    return v, ok
}

func (s *RegSessionStore) Delete(id string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.items, id)
}

// randBytes returns n bytes using crypto/rand.
func randBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil {
        return nil, err
    }
    return b, nil
}

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
    sidRaw, err := randBytes(24)
    if err != nil {
        return RegistrationOptionsResponse{}, err
    }
    sid := b64.Encode(sidRaw)
    // 32-byte challenge
    chRaw, err := randBytes(32)
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

