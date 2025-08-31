package config

import (
    "errors"
    "net/url"
    "os"
    "strconv"
    "strings"
)

// Config holds runtime settings for the server.
type Config struct {
    RP_ID           string
    Origin          string
    Port            string
    DBPath          string
    RPAllowlist     []string
    OriginAllowlist []string
}

func getenv(k, d string) string {
    if v := strings.TrimSpace(os.Getenv(k)); v != "" {
        return v
    }
    return d
}

// Load reads configuration from environment variables, applies defaults,
// normalizes values, validates, and derives allowlists.
func Load() (*Config, error) {
    rpID := strings.ToLower(getenv("RP_ID", "localhost"))
    origin := getenv("ORIGIN", "http://localhost:5173")
    port := getenv("PORT", "8080")
    dbPath := getenv("DB_PATH", "server/demo.db")

    // Normalize origin: trim and remove trailing slash
    origin = strings.TrimSpace(origin)
    origin = strings.TrimRight(origin, "/")

    // Parse allowlists
    rpList := parseList(getenv("RP_ID_ALLOWLIST", ""))
    originList := parseList(getenv("ORIGIN_ALLOWLIST", ""))

    // Ensure primary values are included in allowlists
    if !contains(rpList, rpID) {
        rpList = append(rpList, rpID)
    }
    if !contains(originList, origin) {
        originList = append(originList, origin)
    }

    // Validate
    if rpID == "" {
        return nil, errors.New("RP_ID must not be empty")
    }
    u, err := url.Parse(origin)
    if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
        return nil, errors.New("ORIGIN must be a valid http(s) URL with host")
    }
    if _, err := strconv.Atoi(port); err != nil {
        return nil, errors.New("PORT must be numeric")
    }
    // Disallow wildcard/glob patterns in allowlists for safety
    for _, v := range rpList {
        if strings.ContainsAny(v, "*? ") {
            return nil, errors.New("RP_ID_ALLOWLIST entries must be exact, no wildcards/spaces")
        }
    }
    for _, v := range originList {
        if strings.ContainsAny(v, "*? ") {
            return nil, errors.New("ORIGIN_ALLOWLIST entries must be exact, no wildcards/spaces")
        }
    }

    cfg := &Config{
        RP_ID:           rpID,
        Origin:          origin,
        Port:            port,
        DBPath:          dbPath,
        RPAllowlist:     rpList,
        OriginAllowlist: originList,
    }
    return cfg, nil
}

func parseList(s string) []string {
    if s == "" {
        return []string{}
    }
    parts := strings.Split(s, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            out = append(out, p)
        }
    }
    return out
}

func contains(ss []string, v string) bool {
    for _, s := range ss {
        if s == v {
            return true
        }
    }
    return false
}

