package webauthn

import (
    "encoding/binary"
    "testing"
)

func mkAD(rp byte, flags byte, sc uint32, extra []byte) []byte {
    b := make([]byte, 37+len(extra))
    for i := 0; i < 32; i++ {
        b[i] = rp
    }
    b[32] = flags
    binary.BigEndian.PutUint32(b[33:37], sc)
    copy(b[37:], extra)
    return b
}

func TestParseAuthData_HappyPath(t *testing.T) {
    in := mkAD(0xaa, FlagUP|FlagUV, 5, nil)
    ad, rem, err := ParseAuthData(in)
    if err != nil { t.Fatalf("parse: %v", err) }
    for _, v := range ad.RpIDHash {
        if v != 0xaa { t.Fatalf("rpidhash mismatch") }
    }
    if ad.Flags != (FlagUP|FlagUV) { t.Fatalf("flags mismatch") }
    if ad.SignCount != 5 { t.Fatalf("signCount mismatch: %d", ad.SignCount) }
    if len(rem) != 0 { t.Fatalf("expected no remainder") }
    if !HasUP(ad.Flags) || !HasUV(ad.Flags) { t.Fatalf("flag helpers failed") }
}

func TestParseAuthData_FlagsCombos(t *testing.T) {
    cases := []struct{
        flags byte
        up, uv bool
    }{
        {0, false, false},
        {FlagUP, true, false},
        {FlagUV, false, true},
        {FlagUP|FlagUV, true, true},
    }
    for _, c := range cases {
        in := mkAD(0x01, c.flags, 0, []byte{1,2,3})
        ad, rem, err := ParseAuthData(in)
        if err != nil { t.Fatalf("parse flags: %v", err) }
        if (HasUP(ad.Flags) != c.up) || (HasUV(ad.Flags) != c.uv) {
            t.Fatalf("helpers mismatch for flags %02x", c.flags)
        }
        if len(rem) != 3 { t.Fatalf("remainder length mismatch") }
    }
}

func TestParseAuthData_InvalidLength(t *testing.T) {
    if _, _, err := ParseAuthData(make([]byte, 36)); err == nil {
        t.Fatalf("expected error for short input")
    }
}

func TestParseAuthData_BigEndianCounter(t *testing.T) {
    // bytes: 01 02 03 04 -> 0x01020304 = 16909060
    in := mkAD(0xbb, 0, 0x01020304, nil)
    ad, _, err := ParseAuthData(in)
    if err != nil { t.Fatalf("parse: %v", err) }
    if ad.SignCount != 0x01020304 { t.Fatalf("be mismatch: %d", ad.SignCount) }
}

