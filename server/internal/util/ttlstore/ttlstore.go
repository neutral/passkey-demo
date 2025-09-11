package ttlstore

import (
    "sync"
    "time"
)

// Store is a generic TTL cache with an optional capacity bound and background GC.
// It is concurrency-safe. Items expire after TTL; GC best-effort purges expired entries.
type Store[K comparable, V any] struct {
    mu       sync.Mutex
    items    map[K]entry[V]
    capacity int
    ttl      time.Duration
}

type entry[V any] struct {
    v   V
    exp time.Time
}

// New creates a new TTL store.
// capacity=0 means unbounded. GC runs every ttl/2 (min 1s) when gc=true.
func New[K comparable, V any](capacity int, ttl time.Duration, gc bool) *Store[K, V] {
    s := &Store[K, V]{items: make(map[K]entry[V]), capacity: capacity, ttl: ttl}
    if gc {
        tick := ttl / 2
        if tick < time.Second { tick = time.Second }
        go func() {
            t := time.NewTicker(tick)
            defer t.Stop()
            for range t.C {
                s.gc()
            }
        }()
    }
    return s
}

// Put inserts or replaces an entry with a new expiry.
func (s *Store[K, V]) Put(k K, v V) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.capacity > 0 && len(s.items) >= s.capacity {
        // Evict an arbitrary expired item if possible before refusing.
        s.gcLocked()
        if len(s.items) >= s.capacity {
            return ErrCapacity
        }
    }
    s.items[k] = entry[V]{v: v, exp: time.Now().Add(s.ttl)}
    return nil
}

// Get returns value and presence; expired entries are treated as missing.
func (s *Store[K, V]) Get(k K) (V, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    e, ok := s.items[k]
    if !ok { var zero V; return zero, false }
    if time.Now().After(e.exp) { delete(s.items, k); var zero V; return zero, false }
    return e.v, true
}

// Delete removes the key if present.
func (s *Store[K, V]) Delete(k K) {
    s.mu.Lock(); defer s.mu.Unlock(); delete(s.items, k)
}

// ErrCapacity indicates the store reached its capacity bound.
var ErrCapacity = errCapacity{}

type errCapacity struct{}

func (errCapacity) Error() string { return "ttlstore: at capacity" }

// gc removes expired entries. Caller must NOT hold s.mu.
func (s *Store[K, V]) gc() {
    s.mu.Lock(); defer s.mu.Unlock(); s.gcLocked()
}

// gcLocked assumes s.mu is held.
func (s *Store[K, V]) gcLocked() {
    now := time.Now()
    for k, e := range s.items {
        if now.After(e.exp) {
            delete(s.items, k)
        }
    }
}

