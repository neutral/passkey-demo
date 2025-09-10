package tx

import (
    "encoding/json"
    "os"
    "path/filepath"
    "runtime"
    "testing"

    vec "github.com/neutral/passkey-demo/internal/vectors"
)

// Test name includes "Golden" so it can be selected with -run Golden.
func TestGoldenVectors_MatchCommittedGolden(t *testing.T) {
    // Recompute vectors programmatically
    have, err := vec.Generate()
    if err != nil {
        t.Fatalf("generate: %v", err)
    }

    // Locate repo root relative to this file: server/internal/tx/ -> server -> .. (repo)
    _, thisFile, _, ok := runtime.Caller(0)
    if !ok {
        t.Fatalf("cannot resolve caller path")
    }
    serverDir := filepath.Dir(filepath.Dir(filepath.Dir(thisFile))) // /.../server
    goldenPath := filepath.Join(serverDir, "..", "specs", "goldens", "tx-bundle-v1.json")

    // Read and unmarshal golden file
    data, err := os.ReadFile(goldenPath)
    if err != nil {
        t.Fatalf("read golden: %v", err)
    }
    var want vec.Vectors
    if err := json.Unmarshal(data, &want); err != nil {
        t.Fatalf("unmarshal golden: %v", err)
    }

    // Compare fields explicitly to provide clearer diffs on failure
    if have.Version != want.Version {
        t.Fatalf("version mismatch: have=%q want=%q", have.Version, want.Version)
    }
    if have.Inputs.Message != want.Inputs.Message || have.Inputs.Nonce != want.Inputs.Nonce {
        t.Fatalf("inputs mismatch: have=%+v want=%+v", have.Inputs, want.Inputs)
    }
    if have.Inputs.SenderKeyCBORB64 != want.Inputs.SenderKeyCBORB64 || have.Inputs.SenderKeyCBORHex != want.Inputs.SenderKeyCBORHex {
        t.Fatalf("sender_key encodings mismatch: have=%+v want=%+v", have.Inputs, want.Inputs)
    }
    if have.Bundle.BundleCBORB64 != want.Bundle.BundleCBORB64 || have.Bundle.BundleCBORHex != want.Bundle.BundleCBORHex {
        t.Fatalf("bundle encodings mismatch: have=%+v want=%+v", have.Bundle, want.Bundle)
    }
    if have.Anchors.ChallengeB64 != want.Anchors.ChallengeB64 || have.Anchors.ChallengeHex != want.Anchors.ChallengeHex || have.Anchors.TxIDHex != want.Anchors.TxIDHex {
        t.Fatalf("anchors mismatch: have=%+v want=%+v", have.Anchors, want.Anchors)
    }
}
