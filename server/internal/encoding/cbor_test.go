package encoding

import (
    "encoding/hex"
    "reflect"
    "testing"
)

func TestCBOR_DeterministicMapEncoding(t *testing.T) {
    // Two maps with identical content but created in different orders
    m1 := map[int]string{0: "a", 1: "b"}
    m2 := map[int]string{}
    m2[1] = "b"
    m2[0] = "a"

    b1, err := EncodeCanonical(m1)
    if err != nil { t.Fatalf("enc m1: %v", err) }
    b2, err := EncodeCanonical(m2)
    if err != nil { t.Fatalf("enc m2: %v", err) }
    if hex.EncodeToString(b1) != hex.EncodeToString(b2) {
        t.Fatalf("canonical encoding differs: %x vs %x", b1, b2)
    }

    // Golden for {0:"a",1:"b"}
    golden := "a2006161016162"
    if hex.EncodeToString(b1) != golden {
        t.Fatalf("unexpected encoding: got %s want %s", hex.EncodeToString(b1), golden)
    }
}

type sample struct {
    A int
    B string
    C []byte
    D *int `cbor:",omitempty"`
}

func TestCBOR_RoundtripStructAndMap(t *testing.T) {
    v := sample{A: 42, B: "hello", C: []byte{1,2,3}, D: nil}
    b, err := EncodeCanonical(v)
    if err != nil { t.Fatalf("enc: %v", err) }
    var out sample
    if err := DecodeCanonical(b, &out); err != nil { t.Fatalf("dec: %v", err) }
    if !reflect.DeepEqual(v, out) {
        t.Fatalf("roundtrip mismatch: %+v vs %+v", v, out)
    }

    // Map roundtrip (use concrete numeric type to avoid uint64 vs int mismatch)
    m := map[string]int{"x": 1}
    b2, err := EncodeCanonical(m)
    if err != nil { t.Fatalf("enc map: %v", err) }
    var m2 map[string]int
    if err := DecodeCanonical(b2, &m2); err != nil { t.Fatalf("dec map: %v", err) }
    if !reflect.DeepEqual(m, m2) {
        t.Fatalf("map roundtrip mismatch: %+v vs %+v", m, m2)
    }
}

func TestCBOR_DecodeMalformed(t *testing.T) {
    bad := [][]byte{{0xff}} // invalid simple value (break) outside indef context
    for _, b := range bad {
        var v any
        if err := DecodeCanonical(b, &v); err == nil {
            t.Fatalf("expected error for malformed cbor: %x", b)
        }
    }
}
