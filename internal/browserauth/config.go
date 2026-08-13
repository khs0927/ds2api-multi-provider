package browserauth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func providerConfig() (string, string, time.Duration, error) {
	baseURL := strings.TrimSpace(os.Getenv("DS2API_BROWSERLESS_URL"))
	if baseURL == "" {
		baseURL = "https://production-sfo.browserless.io"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil {
		return "", "", 0, errors.New("DS2API_BROWSERLESS_URL must be an http(s) origin")
	}
	providerToken := strings.TrimSpace(os.Getenv("DS2API_BROWSERLESS_TOKEN"))
	if providerToken == "" {
		return "", "", 0, errors.New("DS2API_BROWSERLESS_TOKEN is not configured")
	}
	ttlSeconds := 600
	if raw := strings.TrimSpace(os.Getenv("DS2API_BROWSER_LOGIN_TTL_SECONDS")); raw != "" {
		if n, convErr := strconv.Atoi(raw); convErr == nil && n >= 60 && n <= 600 {
			ttlSeconds = n
		}
	}
	return strings.TrimRight(baseURL, "/"), providerToken, time.Duration(ttlSeconds) * time.Second, nil
}

func randomProfileID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "deepseek-" + hex.EncodeToString(b), nil
}

func safeID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 255 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}
