package errors

import (
    "encoding/json"
    "net/http/httptest"
    "testing"
)

func TestWriteEnvelope(t *testing.T) {
    rr := httptest.NewRecorder()
    Write(rr, 401, CodeUnauthorized, "unauthorized")
    if rr.Code != 401 {
        t.Fatalf("status: %d", rr.Code)
    }
    var env Envelope
    if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
        t.Fatalf("json: %v", err)
    }
    if env.Code != CodeUnauthorized || env.Error == "" {
        t.Fatalf("envelope mismatch: %+v", env)
    }
}

