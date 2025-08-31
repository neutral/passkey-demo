package types

// CoseEC2 represents a COSE EC2 public key as extracted from WebAuthn
// attestation/authenticator data. The integer fields correspond to COSE
// parameters and the byte slices hold the X/Y affine coordinates.
// When encoded as CBOR maps, COSE uses integer keys: 1 (kty), 3 (alg),
// -1 (crv), -2 (x), -3 (y). Struct tags reflect those indexes.
type CoseEC2 struct {
    Kty int    `cbor:"1" json:"kty"`
    Alg int    `cbor:"3" json:"alg"`
    Crv int    `cbor:"-1" json:"crv"`
    X   []byte `cbor:"-2" json:"x"`
    Y   []byte `cbor:"-3" json:"y"`
}

// Bundle is the minimal canonical CBOR structure that the client signs.
// It includes the sender's COSE EC2 public key, a strictly increasing nonce,
// a free-form message string, and an optional valid-until timestamp (unix seconds).
// The CBOR map uses small integer keys (0..3) for stability and compactness.
type Bundle struct {
    SenderKey  CoseEC2 `cbor:"0" json:"sender_key"`
    Nonce      uint64  `cbor:"1" json:"nonce"`
    Message    string  `cbor:"2" json:"message"`
    ValidUntil *uint64 `cbor:"3,omitempty" json:"valid_until,omitempty"`
}

// RegOptions is a minimal shape for registration options returned by the server.
// The frontend will transform these into PublicKeyCredentialCreationOptions.
type RegOptions struct {
    RP_ID       string `json:"rp_id"`
    Origin      string `json:"origin"`
    UVRequired  bool   `json:"uv_required"`
    Attestation string `json:"attestation"`
}

// RegFinish is a minimal shape for the registration completion payload
// sent from the browser to the server. Binary fields are base64url strings.
type RegFinish struct {
    ID       string `json:"id"`
    RawID    string `json:"rawId"`
    Type     string `json:"type"`
    Response struct {
        AttestationObject string `json:"attestationObject"`
        ClientDataJSON    string `json:"clientDataJSON"`
    } `json:"response"`
}

// LoginOptions is a minimal shape for assertion options returned by the server.
// The frontend will transform these into PublicKeyCredentialRequestOptions.
type LoginOptions struct {
    RP_ID      string `json:"rp_id"`
    Origin     string `json:"origin"`
    UVRequired bool   `json:"uv_required"`
}

// LoginFinish is a minimal shape for the login/assertion completion payload
// sent from the browser to the server. Binary fields are base64url strings.
type LoginFinish struct {
    ID       string `json:"id"`
    RawID    string `json:"rawId"`
    Type     string `json:"type"`
    Response struct {
        AuthenticatorData string `json:"authenticatorData"`
        ClientDataJSON    string `json:"clientDataJSON"`
        Signature         string `json:"signature"`
        UserHandle        string `json:"userHandle"`
    } `json:"response"`
}

