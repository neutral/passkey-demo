package webauthn

import (
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "log/slog"
    "net/http"
)

// MapVerifyError maps VerifyAssertion errors to HTTP status and a stable error kind string.
//  - ErrMalformedDER, ErrUnsupportedCurve => 400
//  - ErrHighS, ErrBadSignature => 401
//  - other/non-matching => 500
func MapVerifyError(err error) (int, string) {
    switch {
    case err == nil:
        return http.StatusOK, "none"
    case errors.Is(err, ErrMalformedDER):
        return http.StatusBadRequest, "ErrMalformedDER"
    case errors.Is(err, ErrUnsupportedCurve):
        return http.StatusBadRequest, "ErrUnsupportedCurve"
    case errors.Is(err, ErrHighS):
        return http.StatusUnauthorized, "ErrHighS"
    case errors.Is(err, ErrBadSignature):
        return http.StatusUnauthorized, "ErrBadSignature"
    default:
        return http.StatusInternalServerError, "ErrInternal"
    }
}

// VerifyLog contains safe, structured fields to log for a verification attempt.
// Do not include raw AD/CDJ/signature or raw identifiers; prefer hashes.
type VerifyLog struct {
    Outcome           string // "success" | "failure"
    ErrorKind         string // sentinel or "none"
    RP_ID             string
    Origin            string
    UV                bool
    UP                bool
    SignCount         uint32
    CredentialIDHash  string // hex(SHA256(credential_id))
    ChallengeID       string // server-issued id/nonce if available
    TraceID           string // tracing correlation id
    LatencyMS         int64
}

// HashID returns a hex-encoded SHA-256 hash for an identifier.
func HashID(b []byte) string {
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:])
}

// LogAssertion emits a structured log for a verification attempt.
func LogAssertion(l *slog.Logger, v VerifyLog) {
    if l == nil {
        return
    }
    l.Info("webauthn_assert_verify",
        slog.String("outcome", v.Outcome),
        slog.String("error_kind", v.ErrorKind),
        slog.String("rp_id", v.RP_ID),
        slog.String("origin", v.Origin),
        slog.Bool("uv", v.UV),
        slog.Bool("up", v.UP),
        slog.Uint64("sign_count", uint64(v.SignCount)),
        slog.String("credential_id_hash", v.CredentialIDHash),
        slog.String("challenge_id", v.ChallengeID),
        slog.String("trace_id", v.TraceID),
        slog.Int64("latency_ms", v.LatencyMS),
    )
}

// MapPolicyError maps RP ID / Origin policy sentinel errors to HTTP status and kind.
//  - 400: invalid inputs or schemes (ErrRpIdInvalid, ErrOriginMalformed, ErrOriginScheme)
//  - 403: policy mismatches (ErrRpIdHashMismatch, ErrOriginHost, ErrOriginPort, ErrOriginNotAllowed)
//  - 500: other
func MapPolicyError(err error) (int, string) {
    switch {
    case err == nil:
        return http.StatusOK, "none"
    case errors.Is(err, ErrRpIdInvalid):
        return http.StatusBadRequest, "ErrRpIdInvalid"
    case errors.Is(err, ErrOriginMalformed):
        return http.StatusBadRequest, "ErrOriginMalformed"
    case errors.Is(err, ErrOriginScheme):
        return http.StatusBadRequest, "ErrOriginScheme"
    case errors.Is(err, ErrRpIdHashMismatch):
        return http.StatusForbidden, "ErrRpIdHashMismatch"
    case errors.Is(err, ErrOriginHost):
        return http.StatusForbidden, "ErrOriginHost"
    case errors.Is(err, ErrOriginPort):
        return http.StatusForbidden, "ErrOriginPort"
    case errors.Is(err, ErrOriginNotAllowed):
        return http.StatusForbidden, "ErrOriginNotAllowed"
    default:
        return http.StatusInternalServerError, "ErrInternal"
    }
}
