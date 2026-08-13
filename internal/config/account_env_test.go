package config

import (
	"strings"
	"testing"
)

func TestApplyAccountEnvOverridesFileAccounts(t *testing.T) {
	t.Setenv(accountsJSONFileEnv, "")
	t.Setenv(accountsJSONEnv, `[
		{"name":"primary","email":"one@example.com","password":"pw-one","token":"must-not-survive"},
		{"name":"secondary","email":"two@example.com","password":"pw-two"}
	]`)

	cfg := Config{Accounts: []Account{{Email: "legacy@example.com", Password: "legacy"}}}
	if err := applyAccountEnv(&cfg); err != nil {
		t.Fatalf("applyAccountEnv() error = %v", err)
	}
	if len(cfg.Accounts) != 2 {
		t.Fatalf("accounts = %d, want 2", len(cfg.Accounts))
	}
	if cfg.Accounts[0].Email != "one@example.com" || cfg.Accounts[1].Email != "two@example.com" {
		t.Fatalf("unexpected account order: %#v", cfg.Accounts)
	}
	if cfg.Accounts[0].Token != "" {
		t.Fatal("environment account token must be discarded")
	}
}

func TestApplyAccountEnvRejectsDuplicateAndFailsClosed(t *testing.T) {
	t.Setenv(accountsJSONFileEnv, "")
	t.Setenv(accountsJSONEnv, `[
		{"email":"same@example.com","password":"one"},
		{"email":"same@example.com","password":"two"}
	]`)

	cfg := Config{Accounts: []Account{{Email: "legacy@example.com", Password: "legacy"}}}
	err := applyAccountEnv(&cfg)
	if err == nil || !strings.Contains(err.Error(), "duplicate account") {
		t.Fatalf("error = %v, want duplicate-account error", err)
	}
	if len(cfg.Accounts) != 0 {
		t.Fatalf("accounts = %#v, want fail-closed empty pool", cfg.Accounts)
	}
}

func TestApplyAccountEnvRejectsEmptyPasswordWithoutLeakingIt(t *testing.T) {
	t.Setenv(accountsJSONFileEnv, "")
	t.Setenv(accountsJSONEnv, `[{"email":"secret-user@example.com","password":""}]`)

	cfg := Config{}
	err := applyAccountEnv(&cfg)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if strings.Contains(err.Error(), "secret-user@example.com") {
		t.Fatalf("error leaked full account identifier: %v", err)
	}
	if !strings.Contains(err.Error(), "se***@example.com") {
		t.Fatalf("error does not contain expected redacted identifier: %v", err)
	}
}

func TestApplyAccountEnvLeavesLegacyConfigWhenUnset(t *testing.T) {
	t.Setenv(accountsJSONEnv, "")
	t.Setenv(accountsJSONFileEnv, "")

	cfg := Config{Accounts: []Account{{Email: "legacy@example.com", Password: "legacy"}}}
	if err := applyAccountEnv(&cfg); err != nil {
		t.Fatalf("applyAccountEnv() error = %v", err)
	}
	if len(cfg.Accounts) != 1 || cfg.Accounts[0].Email != "legacy@example.com" {
		t.Fatalf("legacy config unexpectedly changed: %#v", cfg.Accounts)
	}
}
