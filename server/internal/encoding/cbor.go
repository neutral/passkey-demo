package encoding

import (
    "fmt"

    cbor "github.com/fxamacker/cbor/v2"
)

var (
    encMode cbor.EncMode
    decMode cbor.DecMode
)

func init() {
    // Canonical encoder: deterministic ordering, shortest lengths, etc.
    encOpts := cbor.CanonicalEncOptions()
    em, err := encOpts.EncMode()
    if err != nil {
        panic(fmt.Errorf("cbor enc mode: %w", err))
    }
    encMode = em

    // Decoder with default safe options. (Adjust if stricter behavior is needed.)
    dm, err := (cbor.DecOptions{}).DecMode()
    if err != nil {
        panic(fmt.Errorf("cbor dec mode: %w", err))
    }
    decMode = dm
}

// EncodeCanonical encodes v into canonical CBOR bytes.
func EncodeCanonical(v any) ([]byte, error) {
    return encMode.Marshal(v)
}

// DecodeCanonical decodes canonical CBOR bytes into v.
func DecodeCanonical(data []byte, v any) error {
    return decMode.Unmarshal(data, v)
}
