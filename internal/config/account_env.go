package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

const accountsJSONEnv = "DS2API_ACCOUNTS_JSON"
const accountsJSONFileEnv = "DS2API_ACCOUNTS_JSON_FILE"

// applyAccountEnv makes an environment-backed account pool authoritative when
// DS2API_ACCOUNTS_JSON or DS2API_ACCOUNTS_JSON_FILE is set. This keeps account
// credentials out of checked-in config files while preserving legacy config
// behavior when neither variable is configured.
func applyAccountEnv(c *Config) error {
	if c == nil {
		return nil
	}
	raw, configured, err := readAccountsEnv()
	if err != nil {
		return err
	}
	if !configured {
		return nil
	}

	// Once the secret-backed source is configured it is authoritative. Clear
	// any file-backed accounts first so malformed secret input cannot silently
	// fall back to credentials from a checked-in/local config.
	c.Accounts = nil

	var accounts []Account
	if err := json.Unmarshal([]byte(raw), &accounts); err != nil {
		return fmt.Errorf("invalid %s: %w", accountsJSONEnv, err)
	}

	seen := make(map[string]struct{}, len(accounts))
	clean := make([]Account, 0, len(accounts))
	for i, account := range accounts {
		account.Name = strings.TrimSpace(account.Name)
		account.Remark = strings.TrimSpace(account.Remark)
		account.Email = strings.TrimSpace(account.Email)
		account.Mobile = strings.TrimSpace(account.Mobile)
		account.Password = strings.TrimSpace(account.Password)
		account.Token = "" // tokens are always runtime-only
		account.ProxyID = strings.TrimSpace(account.ProxyID)

		id := account.Identifier()
		if id == "" {
			return fmt.Errorf("%s account %d has no email or mobile", accountsJSONEnv, i+1)
		}
		if account.Password == "" {
			return fmt.Errorf("%s account %d (%s) has an empty password", accountsJSONEnv, i+1, redactedIdentifier(id))
		}
		key := strings.ToLower(id)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%s contains duplicate account %s", accountsJSONEnv, redactedIdentifier(id))
		}
		seen[key] = struct{}{}
		clean = append(clean, account)
	}
	if len(clean) == 0 {
		return fmt.Errorf("%s must contain at least one account", accountsJSONEnv)
	}
	c.Accounts = clean
	return nil
}

func readAccountsEnv() (raw string, configured bool, err error) {
	if value := strings.TrimSpace(os.Getenv(accountsJSONEnv)); value != "" {
		return value, true, nil
	}
	path := strings.TrimSpace(os.Getenv(accountsJSONFileEnv))
	if path == "" {
		return "", false, nil
	}
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		return "", true, fmt.Errorf("read %s: %w", accountsJSONFileEnv, readErr)
	}
	value := strings.TrimSpace(string(content))
	if value == "" {
		return "", true, errors.New(accountsJSONFileEnv + " is empty")
	}
	return value, true, nil
}

func redactedIdentifier(id string) string {
	id = strings.TrimSpace(id)
	if at := strings.IndexByte(id, '@'); at > 0 {
		local := id[:at]
		if len(local) > 2 {
			local = local[:2] + "***"
		} else {
			local = "***"
		}
		return local + id[at:]
	}
	if len(id) <= 4 {
		return "***"
	}
	return "***" + id[len(id)-4:]
}
