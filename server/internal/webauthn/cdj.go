package webauthn

import (
    "encoding/json"
    "errors"
    "fmt"

    b64 "github.com/neutral/passkey-demo/internal/encoding"
)

type ClientData struct {
    Type      string `json:"type"`
    Challenge string `json:"challenge"`
    Origin    string `json:"origin"`
}

type ClientDataParsed struct {
    Type      string
    Challenge []byte
    Origin    string
}

func ParseClientDataJSON(b []byte) (ClientDataParsed, error) {
    var raw ClientData
    if err := json.Unmarshal(b, &raw); err != nil {
        return ClientDataParsed{}, fmt.Errorf("parse clientDataJSON: %w", err)
    }
    if raw.Type == "" || raw.Origin == "" || raw.Challenge == "" {
        return ClientDataParsed{}, errors.New("clientDataJSON missing required fields")
    }
    if raw.Type != "webauthn.get" && raw.Type != "webauthn.create" {
        return ClientDataParsed{}, fmt.Errorf("unsupported clientDataJSON type: %s", raw.Type)
    }
    chall, err := b64.Decode(raw.Challenge)
    if err != nil {
        return ClientDataParsed{}, fmt.Errorf("decode challenge: %w", err)
    }
    return ClientDataParsed{Type: raw.Type, Challenge: chall, Origin: raw.Origin}, nil
}

func IsGet(c ClientDataParsed) bool    { return c.Type == "webauthn.get" }
func IsCreate(c ClientDataParsed) bool { return c.Type == "webauthn.create" }

