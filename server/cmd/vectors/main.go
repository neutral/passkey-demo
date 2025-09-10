package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"

    vec "github.com/neutral/passkey-demo/internal/vectors"
)

func main() {
    format := flag.String("fmt", "json", "output format: json|text")
    flag.Parse()

    v, err := vec.Generate()
    if err != nil {
        fmt.Fprintf(os.Stderr, "generate: %v\n", err)
        os.Exit(1)
    }

    switch *format {
    case "json":
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        if err := enc.Encode(v); err != nil {
            fmt.Fprintf(os.Stderr, "encode json: %v\n", err)
            os.Exit(1)
        }
    case "text":
        // Human-readable block
        fmt.Printf("version: %s\n", v.Version)
        fmt.Println("inputs:")
        fmt.Printf("  message: %s\n", v.Inputs.Message)
        fmt.Printf("  nonce:   %d\n", v.Inputs.Nonce)
        fmt.Printf("  sender_key_cbor_hex: %s\n", v.Inputs.SenderKeyCBORHex)
        fmt.Printf("  sender_key_cbor_b64: %s\n", v.Inputs.SenderKeyCBORB64)
        fmt.Println("bundle:")
        fmt.Printf("  bundle_cbor_hex: %s\n", v.Bundle.BundleCBORHex)
        fmt.Printf("  bundle_cbor_b64: %s\n", v.Bundle.BundleCBORB64)
        fmt.Println("anchors:")
        fmt.Printf("  challenge_hex: %s\n", v.Anchors.ChallengeHex)
        fmt.Printf("  challenge_b64: %s\n", v.Anchors.ChallengeB64)
        fmt.Printf("  tx_id_hex:     %s\n", v.Anchors.TxIDHex)
    default:
        fmt.Fprintf(os.Stderr, "unknown -fmt: %s\n", *format)
        os.Exit(2)
    }
}

