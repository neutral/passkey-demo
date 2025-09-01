package webauthn

import (
    "errors"
    "net/http"
    "testing"
)

func TestMapVerifyError(t *testing.T) {
    cases := []struct{
        err   error
        wantS int
        wantK string
    }{
        {nil, http.StatusOK, "none"},
        {ErrMalformedDER, http.StatusBadRequest, "ErrMalformedDER"},
        {ErrUnsupportedCurve, http.StatusBadRequest, "ErrUnsupportedCurve"},
        {ErrHighS, http.StatusUnauthorized, "ErrHighS"},
        {ErrBadSignature, http.StatusUnauthorized, "ErrBadSignature"},
        {errors.New("other"), http.StatusInternalServerError, "ErrInternal"},
    }
    for _, c := range cases {
        gotS, gotK := MapVerifyError(c.err)
        if gotS != c.wantS || gotK != c.wantK {
            t.Fatalf("map(%v) -> (%d,%s), want (%d,%s)", c.err, gotS, gotK, c.wantS, c.wantK)
        }
    }
}

func TestHashID(t *testing.T) {
    h := HashID([]byte("abc"))
    if len(h) != 64 { // hex-encoded SHA-256
        t.Fatalf("unexpected hash length: %d", len(h))
    }
}

func TestMapPolicyError(t *testing.T) {
    cases := []struct{
        err   error
        wantS int
        wantK string
    }{
        {nil, http.StatusOK, "none"},
        {ErrRpIdInvalid, http.StatusBadRequest, "ErrRpIdInvalid"},
        {ErrOriginMalformed, http.StatusBadRequest, "ErrOriginMalformed"},
        {ErrOriginScheme, http.StatusBadRequest, "ErrOriginScheme"},
        {ErrRpIdHashMismatch, http.StatusForbidden, "ErrRpIdHashMismatch"},
        {ErrOriginHost, http.StatusForbidden, "ErrOriginHost"},
        {ErrOriginPort, http.StatusForbidden, "ErrOriginPort"},
        {ErrOriginNotAllowed, http.StatusForbidden, "ErrOriginNotAllowed"},
        {errors.New("other"), http.StatusInternalServerError, "ErrInternal"},
    }
    for _, c := range cases {
        gotS, gotK := MapPolicyError(c.err)
        if gotS != c.wantS || gotK != c.wantK {
            t.Fatalf("map(%v) -> (%d,%s), want (%d,%s)", c.err, gotS, gotK, c.wantS, c.wantK)
        }
    }
}
