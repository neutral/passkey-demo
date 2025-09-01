# Purpose
Parse `attestationObject` (fmt: none) to extract authenticatorData header, AAGUID, credential ID, and COSE EC2 key for registration finish.

# API
- `ParseAttestationObject(b []byte) (AttestationObject, error)`
- `ParseAttestedCredentialData(rem []byte) (aaguid [16]byte, credID []byte, coseRaw []byte, rest []byte, err error)`
- `ParseCOSEKeyEC2(b []byte) (types.CoseEC2, error)`
- `ExtractRegistrationData(attObjB []byte) (ad AuthData, aaguid [16]byte, credID []byte, cose types.CoseEC2, err error)`

# Errors
- `ErrAttestationCBOR`, `ErrAttestationFormat`, `ErrAttestedDataMissing`, `ErrAuthDataShort`, `ErrCredIDLength`, `ErrCOSEKeyDecode`.

# Notes
- Supports `fmt: "none"` only; ignores `attStmt`. Extensions after COSE key are not parsed.
- Uses fxamacker/cbor decoder with `cbor.RawMessage` to capture exact COSE key bytes.

# Refs
Refs: WebAuthn spec (attestationObject, attestedCredentialData); types.CoseEC2; Step 14 AD parsing; encoding/cbor utilities.

