package randutil

import (
    "crypto/rand"
)

// Bytes returns n cryptographically-secure random bytes.
// It panics only if the system CSPRNG fails, in which case the program
// cannot proceed safely. Callers may also choose to handle the error by
// using BytesE.
func Bytes(n int) []byte {
    b, err := BytesE(n)
    if err != nil {
        panic(err)
    }
    return b
}

// BytesE is the error-returning variant of Bytes.
func BytesE(n int) ([]byte, error) {
    buf := make([]byte, n)
    if _, err := rand.Read(buf); err != nil {
        return nil, err
    }
    return buf, nil
}

