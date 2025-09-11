package ttlstore

import (
    "testing"
    "time"
)

func TestTTLStore_PutGetExpire(t *testing.T) {
    s := New[string, int](0, 50*time.Millisecond, false)
    if err := s.Put("a", 1); err != nil { t.Fatalf("put: %v", err) }
    if v, ok := s.Get("a"); !ok || v != 1 { t.Fatalf("get: %v %v", v, ok) }
    time.Sleep(60 * time.Millisecond)
    if _, ok := s.Get("a"); ok { t.Fatalf("expected expired") }
}

func TestTTLStore_Capacity(t *testing.T) {
    s := New[int, int](1, time.Second, false)
    if err := s.Put(1, 1); err != nil { t.Fatalf("put1: %v", err) }
    if err := s.Put(2, 2); err == nil { t.Fatalf("expected capacity error") }
}

