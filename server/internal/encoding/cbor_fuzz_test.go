package encoding

import "testing"

func FuzzDecodeCanonical(f *testing.F) {
    f.Add([]byte{0xa1, 0x61, 0x78, 0x01}) // {"x":1}
    f.Fuzz(func(t *testing.T, data []byte) {
        var v any
        _ = DecodeCanonical(data, &v)
    })
}

